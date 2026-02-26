"use client";

import { useCallback, useEffect, useMemo, useState } from "react";
import Image from "next/image";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { Bell, LogOut, Search, User } from "lucide-react";
import { useAuth } from "./AuthContext";
import { landingData } from "@/lib/data";

type ApiResponse<T> = {
  success?: boolean;
  data?: T;
  error?: string;
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
      return "New follow request";
    case "group_invitation":
      return "Group invitation";
    case "group_join_request":
      return "Join request";
    case "event_created":
      return "New group event";
    case "post_reaction":
      return "New post reaction";
    case "comment_reaction":
      return "New comment reaction";
    case "comment_on_post":
      return "New comment";
    default:
      return "Notification";
  }
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

  return (
    <header className="sticky top-0 z-40 border-b border-neutral-200/80 bg-white/85 backdrop-blur-md">
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

        <div className="relative ml-2 hidden flex-1 sm:block">
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
            className="h-11 w-full rounded-2xl border border-neutral-200 bg-neutral-50 pl-9 pr-4 text-sm outline-none transition focus:border-neutral-400"
          />
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
        <div className="absolute right-4 top-16 z-50 w-80 rounded-3xl border border-neutral-200 bg-white p-4 shadow-xl">
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
              resolvedNotifications.slice(0, 8).map((item) => (
                <button
                  key={item.id}
                  type="button"
                  onClick={() => markNotificationRead(item.id)}
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
                    #{item.entity_id} · {item.type}
                  </span>
                </button>
              ))
            )}
          </div>
        </div>
      ) : null}
    </header>
  );
}
