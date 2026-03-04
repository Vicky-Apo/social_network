"use client";

import { useCallback, useEffect, useLayoutEffect, useMemo, useRef, useState, type CSSProperties } from "react";
import { createPortal } from "react-dom";
import Image from "next/image";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { Bell, LogOut, MessageSquare, Search, User, Users } from "lucide-react";
import { useAuth } from "./AuthContext";
import { useNotifications } from "./NotificationsContext";
import { useMessages } from "./MessagesContext";
import { landingData } from "@/lib/data";
import Avatar from "@/components/Avatar";
import { toMediaUrl } from "@/lib/media";
import { apiFetch, apiFetchJson, getApiBaseUrl } from "@/lib/api";
import { ApiResponse } from "@/lib/types";
import {
  allowedNotificationTypes,
  getNotificationBody,
  getNotificationHref,
  getNotificationTitle,
} from "@/lib/notifications";

type SearchMode = "users" | "groups";

type UserSearchItem = {
  id: number;
  first_name: string;
  last_name: string;
  nickname?: string | null;
  avatar_path?: string | null;
};

type GroupSearchItem = {
  id: number;
  title?: string | null;
  name?: string | null;
  description?: string | null;
};

export type TopNavUser = {
  id: number;
  email?: string;
  first_name?: string;
  last_name?: string;
  nickname?: string | null;
};

export type TopNavNotification = {
  id: number;
  user_id: number;
  actor_id?: number;
  type: string;
  entity_type: string;
  entity_id: number;
  metadata?: Record<string, unknown>;
  is_read: boolean;
  read_at?: string;
  created_at: string;
};

type Props = {
  user?: TopNavUser | null;
  searchValue?: string;
  onSearchChange?: (value: string) => void;
  searchPlaceholder?: string;
  notifications?: TopNavNotification[];
  notificationCount?: number;
  onNotificationsChange?: (items: TopNavNotification[]) => void;
  onNotificationCountChange?: (count: number) => void;
  onLogout?: () => void;
  variant?: "light" | "dark";
};

export default function TopNav({
  user,
  searchValue,
  onSearchChange,
  searchPlaceholder = "Search...",
  notifications,
  notificationCount,
  onNotificationsChange,
  onNotificationCountChange,
  onLogout,
  variant = "light",
}: Props) {
  const router = useRouter();
  const { logout } = useAuth();
  const notificationsContext = useNotifications();
  const messagesContext = useMessages();
  const [localSearch, setLocalSearch] = useState("");
  const [notificationsOpen, setNotificationsOpen] = useState(false);
  const [messagesOpen, setMessagesOpen] = useState(false);
  const [localNotifications, setLocalNotifications] = useState<TopNavNotification[]>([]);
  const [localCount, setLocalCount] = useState(0);
  const [loading, setLoading] = useState(false);
  const [searchMode, setSearchMode] = useState<SearchMode>("users");
  const [searchResults, setSearchResults] = useState<
    Array<{ type: "user"; item: UserSearchItem } | { type: "group"; item: GroupSearchItem }>
  >([]);
  const [searchLoading, setSearchLoading] = useState(false);
  const [dropdownStyle, setDropdownStyle] = useState<CSSProperties | null>(null);
  const searchWrapRef = useRef<HTMLDivElement | null>(null);

  const apiBaseUrl = useMemo(() => getApiBaseUrl(), []);

  const resolvedNotifications =
    notifications ?? notificationsContext?.notifications ?? localNotifications;
  const resolvedCount = notificationCount ?? notificationsContext?.count ?? localCount;
  const resolvedLoading = notificationsContext?.loading ?? loading;
  const messageCount = messagesContext?.unreadCount ?? 0;
  const messageLoading = messagesContext?.loading ?? false;
  const messageConversations = messagesContext?.conversations ?? [];
  const usersByID = messagesContext?.usersByID ?? {};
  const groupsByID = messagesContext?.groupsByID ?? {};
  const usingControlledNotifications =
    notifications !== undefined ||
    notificationCount !== undefined ||
    onNotificationsChange !== undefined ||
    onNotificationCountChange !== undefined;

  const setNotificationsSafe = useCallback(
    (items: TopNavNotification[]) => {
      if (onNotificationsChange) {
        onNotificationsChange(items);
      } else {
        setLocalNotifications(items);
      }
    },
    [onNotificationsChange],
  );

  const setCountSafe = useCallback(
    (count: number) => {
      if (onNotificationCountChange) {
        onNotificationCountChange(count);
      } else {
        setLocalCount(count);
      }
    },
    [onNotificationCountChange],
  );

  const refreshNotificationsLocal = useCallback(async () => {
    setLoading(true);
    try {
      const { response, result } = await apiFetchJson<ApiResponse<TopNavNotification[]>>(
        "/notifications?limit=20",
        {},
        apiBaseUrl,
      );
      if (response.ok && result?.success) {
        setNotificationsSafe(result.data ?? []);
      }
    } finally {
      setLoading(false);
    }
  }, [apiBaseUrl, setNotificationsSafe]);

  useEffect(() => {
    if (usingControlledNotifications || notificationsContext) {
      return;
    }
    let cancelled = false;

    const run = async () => {
      try {
        const { response, result } = await apiFetchJson<ApiResponse<{ count: number }>>(
          "/notifications/unread-count",
          {},
          apiBaseUrl,
        );
        if (!cancelled && response.ok && result?.success) {
          setCountSafe(Number(result.data?.count ?? 0));
        }
      } catch {
        // ignore
      }
    };

    void run();
    return () => {
      cancelled = true;
    };
  }, [apiBaseUrl, notificationsContext, setCountSafe, usingControlledNotifications]);

  const markNotificationRead = async (id: number) => {
    if (!usingControlledNotifications && notificationsContext?.markRead) {
      await notificationsContext.markRead(id);
      return;
    }
    const old = resolvedNotifications;
    setNotificationsSafe(
      resolvedNotifications.map((item) =>
        item.id === id ? { ...item, is_read: true } : item,
      ),
    );
    setCountSafe(Math.max(0, resolvedCount - 1));

    try {
      const response = await apiFetch(`/notifications/${id}/read`, { method: "PATCH" }, apiBaseUrl);
      if (!response.ok) {
        setNotificationsSafe(old);
      }
    } catch {
      setNotificationsSafe(old);
    }
  };

  const markAllRead = async () => {
    if (!usingControlledNotifications && notificationsContext?.markAllRead) {
      await notificationsContext.markAllRead();
      return;
    }
    setNotificationsSafe(resolvedNotifications.map((item) => ({ ...item, is_read: true })));
    setCountSafe(0);
    await apiFetch("/notifications/read-all", { method: "PATCH" }, apiBaseUrl).catch(() => undefined);
  };

  const unreadMessageConversations = messageConversations
    .filter((conv) => (conv.unread_count ?? 0) > 0)
    .sort((a, b) => {
      const aTime = a.last_message?.created_at ? Date.parse(a.last_message.created_at) : 0;
      const bTime = b.last_message?.created_at ? Date.parse(b.last_message.created_at) : 0;
      return bTime - aTime;
    })
    .slice(0, 5);

  const handleLogout = async () => {
    try {
      await apiFetch("/auth/logout", { method: "POST" }, apiBaseUrl);
    } finally {
      logout();
      if (onLogout) {
        onLogout();
      } else {
        router.replace("/login");
      }
    }
  };

  const currentSearchValue = searchValue ?? localSearch;

  useLayoutEffect(() => {
    if (!currentSearchValue.trim()) {
      setDropdownStyle(null);
      return;
    }

    const updatePosition = () => {
      const node = searchWrapRef.current;
      if (!node) return;
      const rect = node.getBoundingClientRect();
      setDropdownStyle({
        position: "fixed",
        left: Math.max(12, rect.left),
        top: rect.bottom + 8,
        width: rect.width,
        zIndex: 80,
      });
    };

    updatePosition();
    window.addEventListener("resize", updatePosition);
    window.addEventListener("scroll", updatePosition, true);
    return () => {
      window.removeEventListener("resize", updatePosition);
      window.removeEventListener("scroll", updatePosition, true);
    };
  }, [currentSearchValue]);

  useEffect(() => {
    if (!currentSearchValue.trim()) {
      setSearchResults([]);
      setSearchLoading(false);
      return;
    }

    let cancelled = false;
    const controller = new AbortController();
    const query = currentSearchValue.trim();

    const timeoutID = window.setTimeout(async () => {
      setSearchLoading(true);
      try {
        if (searchMode === "users") {
          const { response, result } = await apiFetchJson<ApiResponse<UserSearchItem[]>>(
            `/users?q=${encodeURIComponent(query)}&limit=6&offset=0`,
            { signal: controller.signal },
            apiBaseUrl,
          );
          if (!cancelled && response.ok && result?.success) {
            setSearchResults(
              (result.data ?? []).map((item) => ({ type: "user" as const, item })),
            );
          } else if (!cancelled) {
            setSearchResults([]);
          }
        } else {
          const { response, result } = await apiFetchJson<ApiResponse<GroupSearchItem[]>>(
            `/groups?q=${encodeURIComponent(query)}&limit=6&offset=0`,
            { signal: controller.signal },
            apiBaseUrl,
          );
          if (!cancelled && response.ok && result?.success) {
            setSearchResults(
              (result.data ?? []).map((item) => ({ type: "group" as const, item })),
            );
          } else if (!cancelled) {
            setSearchResults([]);
          }
        }
      } catch {
        if (!cancelled) {
          setSearchResults([]);
        }
      } finally {
        if (!cancelled) {
          setSearchLoading(false);
        }
      }
    }, 400);

    return () => {
      cancelled = true;
      window.clearTimeout(timeoutID);
      controller.abort();
    };
  }, [apiBaseUrl, currentSearchValue, searchMode]);

  const isDark = variant === "dark";

  return (
    <header
      className={
        isDark
          ? "sticky top-0 z-[70] overflow-visible border-b border-white/10 bg-[#2b2929]/95 backdrop-blur-md"
          : "sticky top-0 z-[70] overflow-visible border-b border-neutral-200/80 bg-white/85 backdrop-blur-md"
      }
    >
      <div className="mx-auto flex w-full max-w-6xl items-center gap-3 px-4 py-3 sm:px-6">
        <Link href="/dashboard" className="inline-flex items-center gap-2">
          <Image
            src="/vybez-logo.png"
            alt={`${landingData.productName} logo`}
            width={110}
            height={40}
            className="h-8 w-auto object-contain"
            priority
          />
        </Link>

        <div ref={searchWrapRef} className="relative ml-2 hidden flex-1 sm:block">
          <Search
            className={`pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 ${
              isDark ? "text-neutral-500" : "text-neutral-400"
            }`}
          />
          <input
            type="search"
            value={currentSearchValue}
            onChange={(event) => {
              const next = event.target.value;
              if (onSearchChange) {
                onSearchChange(next);
              } else {
                setLocalSearch(next);
              }
            }}
            placeholder={searchPlaceholder}
            className={
              isDark
                ? "h-11 w-full rounded-2xl border border-white/20 bg-white/5 pl-9 pr-24 text-sm text-white placeholder:text-neutral-500 outline-none transition focus:border-white/40"
                : "h-11 w-full rounded-2xl border border-neutral-200 bg-neutral-50 pl-9 pr-24 text-sm outline-none transition focus:border-neutral-400"
            }
          />
          <div
            className={`absolute right-2 top-1/2 flex -translate-y-1/2 items-center gap-1 rounded-full p-1 text-[11px] font-semibold ${
              isDark
                ? "border border-white/20 bg-white/5"
                : "border border-neutral-200 bg-white text-neutral-600"
            }`}
          >
            <button
              type="button"
              onClick={() => setSearchMode("users")}
              className={
                isDark
                  ? `inline-flex items-center gap-1 rounded-full px-2 py-1 transition ${
                      searchMode === "users"
                        ? "bg-white text-[#2b2929]"
                        : "text-neutral-400 hover:text-white"
                    }`
                  : `inline-flex items-center gap-1 rounded-full px-2 py-1 transition ${
                      searchMode === "users"
                        ? "bg-neutral-900 text-white"
                        : "text-neutral-600 hover:text-neutral-900"
                    }`
              }
            >
              <User className="h-3 w-3" />
              Users
            </button>
            <button
              type="button"
              onClick={() => setSearchMode("groups")}
              className={
                isDark
                  ? `inline-flex items-center gap-1 rounded-full px-2 py-1 transition ${
                      searchMode === "groups"
                        ? "bg-white text-[#2b2929]"
                        : "text-neutral-400 hover:text-white"
                    }`
                  : `inline-flex items-center gap-1 rounded-full px-2 py-1 transition ${
                      searchMode === "groups"
                        ? "bg-neutral-900 text-white"
                        : "text-neutral-600 hover:text-neutral-900"
                    }`
              }
            >
              <Users className="h-3 w-3" />
              Groups
            </button>
          </div>

          {currentSearchValue.trim() && dropdownStyle
            ? createPortal(
                <div
                  style={dropdownStyle}
                  className={
                    isDark
                      ? "rounded-2xl border border-white/10 bg-[#2b2929] p-3 shadow-2xl"
                      : "rounded-3xl border border-neutral-200 bg-white p-3 shadow-2xl"
                  }
                >
                  {searchLoading ? (
                    <p className={`text-xs ${isDark ? "text-neutral-400" : "text-neutral-500"}`}>
                      Searching...
                    </p>
                  ) : searchResults.length === 0 ? (
                    <p className={`text-xs ${isDark ? "text-neutral-400" : "text-neutral-500"}`}>
                      No results found.
                    </p>
                  ) : (
                    <div className="space-y-2">
                      {searchResults.map((result) => {
                        if (result.type === "user") {
                          const item = result.item;
                          return (
                            <button
                              key={`user-${item.id}`}
                              type="button"
                              onClick={() => router.push(`/profile/${item.id}`)}
                              className={
                                isDark
                                  ? "flex w-full items-center gap-3 rounded-xl border border-white/10 bg-white/5 px-3 py-2 text-left text-xs text-neutral-300 transition hover:bg-white/10 hover:text-white"
                                  : "flex w-full items-center gap-3 rounded-2xl border border-neutral-200 bg-neutral-50 px-3 py-2 text-left text-xs text-neutral-700 transition hover:border-neutral-400 hover:bg-white"
                              }
                            >
                              <Avatar
                                src={item.avatar_path ? toMediaUrl(apiBaseUrl, item.avatar_path) : null}
                                name={`${item.first_name} ${item.last_name}`}
                                size={32}
                                textClassName="text-[10px]"
                              />
                              <div>
                                <p
                                  className={
                                    isDark
                                      ? "text-xs font-semibold text-white"
                                      : "text-xs font-semibold text-neutral-900"
                                  }
                                >
                                  {item.first_name} {item.last_name}
                                </p>
                                <p
                                  className={
                                    isDark ? "text-[11px] text-neutral-400" : "text-[11px] text-neutral-500"
                                  }
                                >
                                  @{item.nickname || `user-${item.id}`}
                                </p>
                              </div>
                            </button>
                          );
                        }
                        const group = result.item;
                        const title = group.title || group.name || `Group ${group.id}`;
                        return (
                          <button
                            key={`group-${group.id}`}
                            type="button"
                            onClick={() => router.push(`/groups/${group.id}`)}
                            className={
                              isDark
                                ? "flex w-full items-center gap-3 rounded-xl border border-white/10 bg-white/5 px-3 py-2 text-left text-xs text-neutral-300 transition hover:bg-white/10 hover:text-white"
                                : "flex w-full items-center gap-3 rounded-2xl border border-neutral-200 bg-neutral-50 px-3 py-2 text-left text-xs text-neutral-700 transition hover:border-neutral-400 hover:bg-white"
                            }
                          >
                            <div className="inline-flex h-8 w-8 items-center justify-center rounded-full bg-emerald-600 text-[10px] font-semibold text-white">
                              {title.slice(0, 2).toUpperCase()}
                            </div>
                            <div>
                              <p
                                className={
                                  isDark
                                    ? "text-xs font-semibold text-white"
                                    : "text-xs font-semibold text-neutral-900"
                                }
                              >
                                {title}
                              </p>
                              {group.description ? (
                                <p
                                  className={
                                    isDark
                                      ? "text-[11px] text-neutral-400"
                                      : "text-[11px] text-neutral-500"
                                  }
                                >
                                  {group.description}
                                </p>
                              ) : null}
                            </div>
                          </button>
                        );
                      })}
                    </div>
                  )}
                </div>,
                document.body,
              )
            : null}
        </div>

        <button
          type="button"
          aria-label="Messages"
          onClick={() => {
            const next = !messagesOpen;
            setMessagesOpen(next);
            if (next && messagesContext) {
              void messagesContext.refreshConversations();
              void messagesContext.refreshUnreadCounts();
            }
          }}
          className={
            isDark
              ? "relative inline-flex h-10 w-10 items-center justify-center rounded-full border border-white/20 bg-white/5 text-neutral-400 transition hover:bg-white/10 hover:text-white"
              : "relative inline-flex h-10 w-10 items-center justify-center rounded-full border border-neutral-200 bg-white text-neutral-600 transition hover:text-neutral-900"
          }
        >
          <MessageSquare className="h-4 w-4" />
          <span
            className={`absolute -right-1 -top-1 inline-flex h-4 min-w-4 items-center justify-center rounded-full px-1 text-[10px] font-semibold ${
              isDark ? "bg-white text-[#2b2929]" : "bg-neutral-900 text-white"
            }`}
          >
            {messageCount}
          </span>
        </button>

        <button
          type="button"
          aria-label="Notifications"
          onClick={() => {
            const next = !notificationsOpen;
            setNotificationsOpen(next);
            if (next) {
              if (!usingControlledNotifications && notificationsContext?.refreshNotifications) {
                void notificationsContext.refreshNotifications({ force: true });
              } else {
                void refreshNotificationsLocal();
              }
            }
          }}
          className={
            isDark
              ? "relative inline-flex h-10 w-10 items-center justify-center rounded-full border border-white/20 bg-white/5 text-neutral-400 transition hover:bg-white/10 hover:text-white"
              : "relative inline-flex h-10 w-10 items-center justify-center rounded-full border border-neutral-200 bg-white text-neutral-600 transition hover:text-neutral-900"
          }
        >
          <Bell className="h-4 w-4" />
          <span
            className={`absolute -right-1 -top-1 inline-flex h-4 min-w-4 items-center justify-center rounded-full px-1 text-[10px] font-semibold ${
              isDark ? "bg-white text-[#2b2929]" : "bg-neutral-900 text-white"
            }`}
          >
            {resolvedCount}
          </span>
        </button>

        {user ? (
          <Link
            href={`/profile/${user.id}`}
            className={
              isDark
                ? "inline-flex items-center gap-2 rounded-full border border-white/20 bg-white/5 px-3 py-2 text-xs font-semibold text-neutral-300 transition hover:bg-white/10 hover:text-white"
                : "inline-flex items-center gap-2 rounded-full border border-neutral-200 bg-white px-3 py-2 text-xs font-semibold text-neutral-700 transition hover:border-neutral-400 hover:text-neutral-900"
            }
          >
            <User className="h-3.5 w-3.5" />
            <span className="hidden sm:inline">My profile</span>
          </Link>
        ) : null}

        <button
          type="button"
          onClick={handleLogout}
          className={
            isDark
              ? "inline-flex items-center gap-2 rounded-full border border-white/20 bg-white/5 px-3 py-2 text-xs font-semibold text-neutral-300 transition hover:bg-white/10 hover:text-white"
              : "inline-flex items-center gap-2 rounded-full border border-neutral-200 bg-white px-3 py-2 text-xs font-semibold text-neutral-700 transition hover:border-neutral-400 hover:text-neutral-900"
          }
        >
          <LogOut className="h-3.5 w-3.5" />
          <span className="hidden sm:inline">Log out</span>
        </button>
      </div>

      {notificationsOpen ? (
        <div
          className={
            isDark
              ? "absolute right-88 top-16 z-50 w-80 rounded-2xl border border-white/10 bg-[#2b2929] p-4 shadow-xl"
              : "absolute right-88 top-16 z-50 w-80 rounded-3xl border border-neutral-200 bg-white p-4 shadow-xl"
          }
        >
          <div className="flex items-center justify-between">
            <p
              className={
                isDark ? "text-sm font-semibold text-white" : "text-sm font-semibold text-neutral-900"
              }
            >
              Notifications
            </p>
            <button
              type="button"
              onClick={markAllRead}
              className={
                isDark
                  ? "text-xs font-semibold text-neutral-400 transition hover:text-white"
                  : "text-xs font-semibold text-neutral-600 transition hover:text-neutral-900"
              }
            >
              Mark all read
            </button>
          </div>
          <div className="mt-3 space-y-2">
            {resolvedLoading ? (
              <p className={`text-xs ${isDark ? "text-neutral-400" : "text-neutral-500"}`}>
                Loading notifications...
              </p>
            ) : resolvedNotifications.length === 0 ? (
              <p className={`text-xs ${isDark ? "text-neutral-400" : "text-neutral-500"}`}>
                No notifications yet.
              </p>
            ) : (
              resolvedNotifications
                .filter((item) => allowedNotificationTypes.has(item.type))
                .slice(0, 5)
                .map((item) => {
                  const href = getNotificationHref(item);
                  return (
                    <button
                      key={item.id}
                      type="button"
                      onClick={() => {
                        void markNotificationRead(item.id);
                        if (href) {
                          router.push(href);
                          setNotificationsOpen(false);
                        }
                      }}
                      className={
                        isDark
                          ? `flex w-full flex-col rounded-xl border px-3 py-2 text-left text-xs transition ${
                              item.is_read
                                ? "border-white/10 bg-white/5 text-neutral-400"
                                : "border-emerald-500/30 bg-emerald-500/20 text-emerald-300"
                            }`
                          : `flex w-full flex-col rounded-2xl border px-3 py-2 text-left text-xs transition ${
                              item.is_read
                                ? "border-neutral-200 bg-neutral-50 text-neutral-500"
                                : "border-emerald-400 bg-emerald-50 text-emerald-900 shadow-[0_0_0_1px_rgba(16,185,129,0.25)]"
                            }`
                      }
                    >
                      <span className="text-[11px] font-semibold uppercase tracking-wide">
                        {getNotificationTitle(item.type)}
                      </span>
                      <span
                        className={`mt-1 text-[11px] ${isDark ? "text-neutral-400" : "text-neutral-500"}`}
                      >
                        {getNotificationBody(item.type, item.metadata)}
                      </span>
                    </button>
                  );
                })
            )}
          </div>
          <button
            type="button"
            onClick={() => {
              setNotificationsOpen(false);
              router.push("/notifications");
            }}
            className={
              isDark
                ? "mt-3 w-full rounded-xl border border-white/20 bg-white/5 px-3 py-2 text-xs font-semibold text-neutral-300 transition hover:bg-white/10 hover:text-white"
                : "mt-3 w-full rounded-2xl border border-neutral-200 bg-white px-3 py-2 text-xs font-semibold text-neutral-600 transition hover:border-neutral-400 hover:text-neutral-900"
            }
          >
            View all notifications
          </button>
        </div>
      ) : null}

      {messagesOpen ? (
        <div
          className={
            isDark
              ? "absolute right-48 top-16 z-50 w-80 rounded-2xl border border-white/10 bg-[#2b2929] p-4 shadow-xl"
              : "absolute right-48 top-16 z-50 w-80 rounded-3xl border border-neutral-200 bg-white p-4 shadow-xl"
          }
        >
          <div className="flex items-center justify-between">
            <p
              className={
                isDark ? "text-sm font-semibold text-white" : "text-sm font-semibold text-neutral-900"
              }
            >
              Messages
            </p>
            <button
              type="button"
              onClick={() => messagesContext?.markAllRead()}
              className={
                isDark
                  ? "text-xs font-semibold text-neutral-400 transition hover:text-white"
                  : "text-xs font-semibold text-neutral-600 transition hover:text-neutral-900"
              }
            >
              Mark all read
            </button>
          </div>
          <div className="mt-3 space-y-2">
            {messageLoading ? (
              <p className={`text-xs ${isDark ? "text-neutral-400" : "text-neutral-500"}`}>
                Loading messages...
              </p>
            ) : unreadMessageConversations.length === 0 ? (
              <p className={`text-xs ${isDark ? "text-neutral-400" : "text-neutral-500"}`}>
                No new messages.
              </p>
            ) : (
              unreadMessageConversations.map((conv) => {
                const otherUser = conv.other_user_id ? usersByID[conv.other_user_id] : null;
                const group = conv.group_id ? groupsByID[conv.group_id] : null;
                const title =
                  conv.type === "direct"
                    ? otherUser
                      ? `${otherUser.first_name} ${otherUser.last_name}`
                      : "Direct message"
                    : group
                      ? group.title || group.name || "Group"
                      : "Group chat";
                const preview =
                  conv.last_message?.content ||
                  (conv.last_message?.media_path ? "(media)" : "New message");
                return (
                  <button
                    key={conv.id}
                    type="button"
                    onClick={async () => {
                      await messagesContext?.markConversationRead(conv.id);
                      setMessagesOpen(false);
                      router.push(`/messages?conversation=${conv.id}`);
                    }}
                    className={
                      isDark
                        ? "flex w-full flex-col rounded-xl border border-emerald-500/30 bg-emerald-500/20 px-3 py-2 text-left text-xs text-emerald-300 transition"
                        : "flex w-full flex-col rounded-2xl border border-emerald-300 bg-emerald-50 px-3 py-2 text-left text-xs text-emerald-900 transition"
                    }
                  >
                    <span className="text-[11px] font-semibold uppercase tracking-wide">
                      {title}
                    </span>
                    <span
                      className={`mt-1 text-[11px] ${isDark ? "text-emerald-400" : "text-emerald-800"}`}
                    >
                      {preview}
                    </span>
                  </button>
                );
              })
            )}
          </div>
          <button
            type="button"
            onClick={() => {
              setMessagesOpen(false);
              router.push("/messages");
            }}
            className={
              isDark
                ? "mt-3 w-full rounded-xl border border-white/20 bg-white/5 px-3 py-2 text-xs font-semibold text-neutral-300 transition hover:bg-white/10 hover:text-white"
                : "mt-3 w-full rounded-2xl border border-neutral-200 bg-white px-3 py-2 text-xs font-semibold text-neutral-600 transition hover:border-neutral-400 hover:text-neutral-900"
            }
          >
            View all messages
          </button>
        </div>
      ) : null}
    </header>
  );
}
