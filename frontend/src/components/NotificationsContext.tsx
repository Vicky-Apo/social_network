"use client";

import { createContext, useCallback, useContext, useEffect, useMemo, useRef, useState } from "react";
import { usePathname } from "next/navigation";
import { apiFetch, apiFetchJson, getApiBaseUrl } from "@/lib/api";
import { ApiResponse } from "@/lib/types";
import { allowedNotificationTypes } from "@/lib/notifications";

export type NotificationItem = {
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

type NotificationsContextValue = {
  notifications: NotificationItem[];
  count: number;
  loading: boolean;
  refreshNotifications: () => Promise<void>;
  refreshUnreadCount: () => Promise<void>;
  markRead: (id: number) => Promise<void>;
  markAllRead: () => Promise<void>;
  setNotifications: (items: NotificationItem[]) => void;
  setCount: (count: number) => void;
};

const NotificationsContext = createContext<NotificationsContextValue | undefined>(undefined);

export function NotificationsProvider({ children }: { children: React.ReactNode }) {
  const pathname = usePathname();
  const [notifications, setNotifications] = useState<NotificationItem[]>([]);
  const [count, setCount] = useState(0);
  const [loading, setLoading] = useState(false);
  const wsRef = useRef<WebSocket | null>(null);
  const reconnectTimerRef = useRef<number | null>(null);
  const connectRef = useRef<(() => void) | null>(null);

  const apiBaseUrl = useMemo(() => getApiBaseUrl(), []);
  const wsBaseUrl = useMemo(() => {
    if (apiBaseUrl.startsWith("https://")) return apiBaseUrl.replace("https://", "wss://");
    if (apiBaseUrl.startsWith("http://")) return apiBaseUrl.replace("http://", "ws://");
    return apiBaseUrl;
  }, [apiBaseUrl]);

  const refreshNotifications = useCallback(async () => {
    setLoading(true);
    try {
      const { response, result } = await apiFetchJson<ApiResponse<NotificationItem[]>>(
        "/notifications?limit=20",
        {},
        apiBaseUrl,
      );
      if (response.ok && result?.success) {
        const next = (result.data ?? []).filter((item) => allowedNotificationTypes.has(item.type));
        setNotifications(next);
        setCount(next.filter((item) => !item.is_read).length);
      }
    } finally {
      setLoading(false);
    }
  }, [apiBaseUrl]);

  const refreshUnreadCount = useCallback(async () => {
    if (notifications.length > 0) {
      setCount(notifications.filter((item) => !item.is_read).length);
      return;
    }
    try {
      const { response, result } = await apiFetchJson<ApiResponse<{ count: number }>>(
        "/notifications/unread-count",
        {},
        apiBaseUrl,
      );
      if (response.ok && result?.success) {
        setCount(Number(result.data?.count ?? 0));
      }
    } catch {
      // ignore
    }
  }, [apiBaseUrl, notifications]);

  const markRead = useCallback(
    async (id: number) => {
      const prev = notifications;
      setNotifications((items) =>
        items.map((item) => (item.id === id ? { ...item, is_read: true } : item)),
      );
      setCount((value) => Math.max(0, value - 1));
      try {
        const response = await apiFetch(`/notifications/${id}/read`, { method: "PATCH" }, apiBaseUrl);
        if (!response.ok) {
          setNotifications(prev);
        }
      } catch {
        setNotifications(prev);
      }
    },
    [apiBaseUrl, notifications],
  );

  const markAllRead = useCallback(async () => {
    setNotifications((items) => items.map((item) => ({ ...item, is_read: true })));
    setCount(0);
    await apiFetch("/notifications/read-all", { method: "PATCH" }, apiBaseUrl).catch(() => undefined);
  }, [apiBaseUrl]);

  useEffect(() => {
    void refreshUnreadCount();
    void refreshNotifications();
  }, [pathname, refreshNotifications, refreshUnreadCount]);

  useEffect(() => {
    let isMounted = true;

    const connect = () => {
      if (!isMounted) return;
      const ws = new WebSocket(`${wsBaseUrl}/ws`);
      wsRef.current = ws;

      ws.onmessage = (event) => {
        const chunks = String(event.data).split("\n").filter(Boolean);
        chunks.forEach((raw) => {
          try {
            const msg = JSON.parse(raw) as { type?: string; payload?: unknown };
            if (msg.type === "notification" && msg.payload) {
              const payload = msg.payload as NotificationItem;
              if (!allowedNotificationTypes.has(payload.type)) {
                return;
              }
              setNotifications((prev) => {
                if (prev.some((item) => item.id === payload.id)) {
                  return prev;
                }
                return [payload, ...prev].slice(0, 20);
              });
              if (!payload.is_read) {
                setCount((value) => value + 1);
              }
            }
          } catch {
            // ignore malformed payloads
          }
        });
      };

      ws.onclose = () => {
        if (!isMounted) return;
        if (reconnectTimerRef.current) {
          window.clearTimeout(reconnectTimerRef.current);
        }
        reconnectTimerRef.current = window.setTimeout(connect, 2000);
      };
    };

    connectRef.current = connect;
    connect();

    return () => {
      isMounted = false;
      if (reconnectTimerRef.current) {
        window.clearTimeout(reconnectTimerRef.current);
      }
      wsRef.current?.close();
      wsRef.current = null;
    };
  }, [wsBaseUrl]);

  useEffect(() => {
    const handleLogout = () => {
      if (reconnectTimerRef.current) {
        window.clearTimeout(reconnectTimerRef.current);
      }
      wsRef.current?.close();
      wsRef.current = null;
      setNotifications([]);
      setCount(0);
    };

    const handleLogin = () => {
      connectRef.current?.();
      void refreshUnreadCount();
      void refreshNotifications();
    };

    window.addEventListener("app-logout", handleLogout);
    window.addEventListener("app-login", handleLogin);
    return () => {
      window.removeEventListener("app-logout", handleLogout);
      window.removeEventListener("app-login", handleLogin);
    };
  }, [refreshNotifications, refreshUnreadCount]);

  const value = useMemo(
    () => ({
      notifications,
      count,
      loading,
      refreshNotifications,
      refreshUnreadCount,
      markRead,
      markAllRead,
      setNotifications,
      setCount,
    }),
    [
      notifications,
      count,
      loading,
      refreshNotifications,
      refreshUnreadCount,
      markRead,
      markAllRead,
    ],
  );

  return <NotificationsContext.Provider value={value}>{children}</NotificationsContext.Provider>;
}

export function useNotifications() {
  return useContext(NotificationsContext);
}
