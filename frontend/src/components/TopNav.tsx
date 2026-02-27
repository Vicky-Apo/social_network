"use client";

import { useCallback, useEffect, useLayoutEffect, useMemo, useRef, useState, type CSSProperties } from "react";
import { createPortal } from "react-dom";
import Image from "next/image";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { Bell, LogOut, Search, User, Users } from "lucide-react";
import { useAuth } from "./AuthContext";
import { landingData } from "@/lib/data";

type ApiResponse<T> = {
  success?: boolean;
  data?: T;
  error?: string;
};

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
};

function formatNotificationTitle(item: TopNavNotification) {
  switch (item.type) {
    case "follow_request":
      return "Follow request";
    case "group_invitation":
      return "Group invitation";
    case "group_join_request":
      return "Join request";
    case "event_created":
      return "New group event";
    default:
      return "Notification";
  }
}

const allowedNotificationTypes = new Set([
  "follow_request",
  "group_invitation",
  "group_join_request",
  "event_created",
]);

function getActorName(item: TopNavNotification) {
  const meta = item.metadata ?? {};
  const requester = meta["requester_name"];
  if (typeof requester === "string" && requester.trim()) return requester;
  const actorID = item.actor_id;
  if (actorID) return `User #${actorID}`;
  return "Someone";
}

function getGroupName(item: TopNavNotification) {
  const meta = item.metadata ?? {};
  const groupName = meta["group_name"];
  if (typeof groupName === "string" && groupName.trim()) return groupName;
  const groupID = meta["group_id"];
  if (typeof groupID === "number") return `Group #${groupID}`;
  return "your group";
}

function getNotificationBody(item: TopNavNotification) {
  switch (item.type) {
    case "follow_request":
      return `${getActorName(item)} sent you a follow request.`;
    case "group_invitation":
      return `${getActorName(item)} invited you to ${getGroupName(item)}.`;
    case "group_join_request":
      return `${getActorName(item)} requested to join ${getGroupName(item)}.`;
    case "event_created":
      return `New event in ${getGroupName(item)}.`;
    default:
      return "Notification update.";
  }
}

function getNotificationHref(item: TopNavNotification) {
  const meta = item.metadata ?? {};
  switch (item.type) {
    case "follow_request":
      return "/follow-requests";
    case "group_invitation":
      return "/group-invitations";
    case "group_join_request": {
      const groupID = meta["group_id"];
      if (typeof groupID === "number") {
        return `/groups/${groupID}/join-requests`;
      }
      return "/groups";
    }
    case "event_created":
      return `/events/${item.entity_id}`;
    default:
      return null;
  }
}

function toMediaUrl(apiBaseUrl: string, path?: string | null) {
  if (!path) return "";
  if (path.startsWith("http://") || path.startsWith("https://")) return path;
  const normalized = path.startsWith("/") ? path : `/${path}`;
  return `${apiBaseUrl}${normalized}`;
}

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
}: Props) {
  const router = useRouter();
  const { logout } = useAuth();
  const [localSearch, setLocalSearch] = useState("");
  const [notificationsOpen, setNotificationsOpen] = useState(false);
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

  const apiBaseUrl = useMemo(
    () =>
      process.env.NEXT_PUBLIC_API_BASE_URL?.trim().replace(/\/+$/, "") ||
      "http://localhost:8080",
    [],
  );

  const resolvedNotifications = notifications ?? localNotifications;
  const resolvedCount = notificationCount ?? localCount;

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

  const refreshNotifications = useCallback(async () => {
    setLoading(true);
    try {
      const response = await fetch(`${apiBaseUrl}/notifications?limit=20`, {
        credentials: "include",
      });
      const result = (await response.json().catch(() => null)) as
        | ApiResponse<TopNavNotification[]>
        | null;
      if (response.ok && result?.success) {
        setNotificationsSafe(result.data ?? []);
      }
    } finally {
      setLoading(false);
    }
  }, [apiBaseUrl, setNotificationsSafe]);

  useEffect(() => {
    if (notifications && notificationCount !== undefined) {
      return;
    }
    let cancelled = false;

    const run = async () => {
      try {
        const response = await fetch(`${apiBaseUrl}/notifications/unread-count`, {
          credentials: "include",
        });
        const result = (await response.json().catch(() => null)) as
          | ApiResponse<{ count: number }>
          | null;
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
  }, [apiBaseUrl, notificationCount, notifications, setCountSafe]);

  const markNotificationRead = async (id: number) => {
    const old = resolvedNotifications;
    setNotificationsSafe(
      resolvedNotifications.map((item) =>
        item.id === id ? { ...item, is_read: true } : item,
      ),
    );
    setCountSafe(Math.max(0, resolvedCount - 1));

    try {
      const response = await fetch(`${apiBaseUrl}/notifications/${id}/read`, {
        method: "PATCH",
        credentials: "include",
      });
      if (!response.ok) {
        setNotificationsSafe(old);
      }
    } catch {
      setNotificationsSafe(old);
    }
  };

  const markAllRead = async () => {
    setNotificationsSafe(resolvedNotifications.map((item) => ({ ...item, is_read: true })));
    setCountSafe(0);
    await fetch(`${apiBaseUrl}/notifications/read-all`, {
      method: "PATCH",
      credentials: "include",
    }).catch(() => undefined);
  };

  const handleLogout = async () => {
    try {
      await fetch(`${apiBaseUrl}/auth/logout`, {
        method: "POST",
        credentials: "include",
      });
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
          const res = await fetch(
            `${apiBaseUrl}/users?q=${encodeURIComponent(query)}&limit=6&offset=0`,
            { credentials: "include", signal: controller.signal },
          );
          const json = (await res.json().catch(() => null)) as
            | ApiResponse<UserSearchItem[]>
            | null;
          if (!cancelled && res.ok && json?.success) {
            setSearchResults(
              (json.data ?? []).map((item) => ({ type: "user" as const, item })),
            );
          } else if (!cancelled) {
            setSearchResults([]);
          }
        } else {
          const res = await fetch(
            `${apiBaseUrl}/groups?q=${encodeURIComponent(query)}&limit=6&offset=0`,
            { credentials: "include", signal: controller.signal },
          );
          const json = (await res.json().catch(() => null)) as
            | ApiResponse<GroupSearchItem[]>
            | null;
          if (!cancelled && res.ok && json?.success) {
            setSearchResults(
              (json.data ?? []).map((item) => ({ type: "group" as const, item })),
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

  return (
    <header className="sticky top-0 z-[70] overflow-visible border-b border-neutral-200/80 bg-white/85 backdrop-blur-md">
      <div className="mx-auto flex w-full max-w-6xl items-center gap-3 px-4 py-3 sm:px-6">
        <Link href="/dashboard" className="inline-flex items-center gap-2">
          <Image
            src="/vybez-logo.png"
            alt={`${landingData.productName} logo`}
            width={32}
            height={32}
            className="h-8 w-8 rounded-full border border-neutral-200 object-cover shadow-sm"
            priority
          />
          <span className="hidden text-sm font-semibold sm:inline">{landingData.productName}</span>
        </Link>

        <div ref={searchWrapRef} className="relative ml-2 hidden flex-1 sm:block">
          <Search className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-neutral-400" />
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
            className="h-11 w-full rounded-2xl border border-neutral-200 bg-neutral-50 pl-9 pr-24 text-sm outline-none transition focus:border-neutral-400"
          />
          <div className="absolute right-2 top-1/2 flex -translate-y-1/2 items-center gap-1 rounded-full border border-neutral-200 bg-white p-1 text-[11px] font-semibold text-neutral-600">
            <button
              type="button"
              onClick={() => setSearchMode("users")}
              className={`inline-flex items-center gap-1 rounded-full px-2 py-1 transition ${
                searchMode === "users"
                  ? "bg-neutral-900 text-white"
                  : "text-neutral-600 hover:text-neutral-900"
              }`}
            >
              <User className="h-3 w-3" />
              Users
            </button>
            <button
              type="button"
              onClick={() => setSearchMode("groups")}
              className={`inline-flex items-center gap-1 rounded-full px-2 py-1 transition ${
                searchMode === "groups"
                  ? "bg-neutral-900 text-white"
                  : "text-neutral-600 hover:text-neutral-900"
              }`}
            >
              <Users className="h-3 w-3" />
              Groups
            </button>
          </div>

          {currentSearchValue.trim() && dropdownStyle
            ? createPortal(
                <div
                  style={dropdownStyle}
                  className="rounded-3xl border border-neutral-200 bg-white p-3 shadow-2xl"
                >
                  {searchLoading ? (
                    <p className="text-xs text-neutral-500">Searching...</p>
                  ) : searchResults.length === 0 ? (
                    <p className="text-xs text-neutral-500">No results found.</p>
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
                              className="flex w-full items-center gap-3 rounded-2xl border border-neutral-200 bg-neutral-50 px-3 py-2 text-left text-xs text-neutral-700 transition hover:border-neutral-400 hover:bg-white"
                            >
                              {item.avatar_path ? (
                                <div className="h-8 w-8 overflow-hidden rounded-full border border-neutral-200 bg-white">
                                  <img
                                    src={toMediaUrl(apiBaseUrl, item.avatar_path)}
                                    alt={`${item.first_name} ${item.last_name}`}
                                    className="h-full w-full object-contain"
                                  />
                                </div>
                              ) : (
                                <div className="inline-flex h-8 w-8 items-center justify-center rounded-full bg-neutral-900 text-[10px] font-semibold text-white">
                                  {item.first_name?.charAt(0)}
                                  {item.last_name?.charAt(0)}
                                </div>
                              )}
                              <div>
                                <p className="text-xs font-semibold text-neutral-900">
                                  {item.first_name} {item.last_name}
                                </p>
                                <p className="text-[11px] text-neutral-500">
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
                            className="flex w-full items-center gap-3 rounded-2xl border border-neutral-200 bg-neutral-50 px-3 py-2 text-left text-xs text-neutral-700 transition hover:border-neutral-400 hover:bg-white"
                          >
                            <div className="inline-flex h-8 w-8 items-center justify-center rounded-full bg-emerald-600 text-[10px] font-semibold text-white">
                              {title.slice(0, 2).toUpperCase()}
                            </div>
                            <div>
                              <p className="text-xs font-semibold text-neutral-900">{title}</p>
                              {group.description ? (
                                <p className="text-[11px] text-neutral-500">
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
          aria-label="Notifications"
          onClick={() => {
            const next = !notificationsOpen;
            setNotificationsOpen(next);
            if (next) {
              void refreshNotifications();
            }
          }}
          className="relative inline-flex h-10 w-10 items-center justify-center rounded-full border border-neutral-200 bg-white text-neutral-600 transition hover:text-neutral-900"
        >
          <Bell className="h-4 w-4" />
          <span className="absolute -right-1 -top-1 inline-flex h-4 min-w-4 items-center justify-center rounded-full bg-neutral-900 px-1 text-[10px] font-semibold text-white">
            {resolvedCount}
          </span>
        </button>

        {user ? (
          <Link
            href={`/profile/${user.id}`}
            className="inline-flex items-center gap-2 rounded-full border border-neutral-200 bg-white px-3 py-2 text-xs font-semibold text-neutral-700 transition hover:border-neutral-400 hover:text-neutral-900"
          >
            <User className="h-3.5 w-3.5" />
            <span className="hidden sm:inline">My profile</span>
          </Link>
        ) : null}

        <button
          type="button"
          onClick={handleLogout}
          className="inline-flex items-center gap-2 rounded-full border border-neutral-200 bg-white px-3 py-2 text-xs font-semibold text-neutral-700 transition hover:border-neutral-400 hover:text-neutral-900"
        >
          <LogOut className="h-3.5 w-3.5" />
          <span className="hidden sm:inline">Log out</span>
        </button>
      </div>

      {notificationsOpen ? (
        <div className="absolute right-88 top-16 z-50 w-80 rounded-3xl border border-neutral-200 bg-white p-4 shadow-xl">
          <div className="flex items-center justify-between">
            <p className="text-sm font-semibold text-neutral-900">Notifications</p>
            <button
              type="button"
              onClick={markAllRead}
              className="text-xs font-semibold text-neutral-600 transition hover:text-neutral-900"
            >
              Mark all read
            </button>
          </div>
          <div className="mt-3 space-y-2">
            {loading ? (
              <p className="text-xs text-neutral-500">Loading notifications...</p>
            ) : resolvedNotifications.length === 0 ? (
              <p className="text-xs text-neutral-500">No notifications yet.</p>
            ) : (
              resolvedNotifications
                .filter((item) => allowedNotificationTypes.has(item.type))
                .slice(0, 8)
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
                      className={`flex w-full flex-col rounded-2xl border px-3 py-2 text-left text-xs transition ${
                        item.is_read
                          ? "border-neutral-200 bg-neutral-50 text-neutral-500"
                          : "border-neutral-900/10 bg-white text-neutral-800"
                      }`}
                    >
                      <span className="text-[11px] font-semibold uppercase tracking-wide">
                        {formatNotificationTitle(item)}
                      </span>
                      <span className="mt-1 text-[11px] text-neutral-500">
                        {getNotificationBody(item)}
                      </span>
                    </button>
                  );
                })
            )}
          </div>
        </div>
      ) : null}
    </header>
  );
}
