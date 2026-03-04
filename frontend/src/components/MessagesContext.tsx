"use client";

import { createContext, useCallback, useContext, useEffect, useMemo, useRef, useState } from "react";
import { apiFetch, apiFetchJson, getApiBaseUrl } from "@/lib/api";
import { ApiResponse } from "@/lib/types";

type MessageItem = {
  id: number;
  conversation_id: number;
  sender_id: number;
  content?: string | null;
  media_path?: string | null;
  created_at: string;
};

type ConversationItem = {
  id: number;
  type: "direct" | "group" | "private_group" | string;
  other_user_id?: number | null;
  group_id?: number | null;
  unread_count?: number;
  last_message?: MessageItem | null;
};

type UserSummary = {
  id: number;
  first_name: string;
  last_name: string;
  nickname?: string | null;
  avatar_path?: string | null;
};

type GroupSummary = {
  id: number;
  title?: string | null;
  name?: string | null;
};

type MessagesContextValue = {
  conversations: ConversationItem[];
  usersByID: Record<number, UserSummary>;
  groupsByID: Record<number, GroupSummary>;
  unreadCount: number;
  loading: boolean;
  refreshConversations: () => Promise<void>;
  refreshUnreadCounts: () => Promise<void>;
  markConversationRead: (id: number) => Promise<void>;
  markAllRead: () => Promise<void>;
};

const MessagesContext = createContext<MessagesContextValue | undefined>(undefined);

export function useMessages() {
  return useContext(MessagesContext);
}

export function MessagesProvider({ children }: { children: React.ReactNode }) {
  const [conversations, setConversations] = useState<ConversationItem[]>([]);
  const [usersByID, setUsersByID] = useState<Record<number, UserSummary>>({});
  const [groupsByID, setGroupsByID] = useState<Record<number, GroupSummary>>({});
  const [unreadCount, setUnreadCount] = useState(0);
  const [loading, setLoading] = useState(false);
  const wsRef = useRef<WebSocket | null>(null);
  const reconnectTimerRef = useRef<number | null>(null);

  const apiBaseUrl = useMemo(() => getApiBaseUrl(), []);
  const wsBaseUrl = useMemo(() => {
    if (apiBaseUrl.startsWith("https://")) return apiBaseUrl.replace("https://", "wss://");
    if (apiBaseUrl.startsWith("http://")) return apiBaseUrl.replace("http://", "ws://");
    return apiBaseUrl;
  }, [apiBaseUrl]);

  const refreshConversations = useCallback(async () => {
    setLoading(true);
    try {
      const { response, result } = await apiFetchJson<ApiResponse<ConversationItem[]>>(
        "/conversations?limit=20&offset=0",
        {},
        apiBaseUrl,
      );
      if (response.ok && result?.success) {
        const items = result.data ?? [];
        setConversations(items);
        const totalUnread = items.reduce((sum, item) => sum + (item.unread_count ?? 0), 0);
        setUnreadCount(totalUnread);
      }
    } finally {
      setLoading(false);
    }
  }, [apiBaseUrl]);

  const refreshUnreadCounts = useCallback(async () => {
    try {
      const { response, result } = await apiFetchJson<
        ApiResponse<Array<{ conversation_id: number; unread_count: number }>>
      >("/conversations/unread-counts", {}, apiBaseUrl);
      if (!response.ok || !result?.success) return;
      const map = new Map(
        (result.data ?? []).map((item) => [item.conversation_id, item.unread_count]),
      );
      setConversations((prev) =>
        prev.map((conv) => ({
          ...conv,
          unread_count: map.get(conv.id) ?? 0,
        })),
      );
      const totalUnread = Array.from(map.values()).reduce((sum, count) => sum + count, 0);
      setUnreadCount(totalUnread);
    } catch {
      // ignore
    }
  }, [apiBaseUrl]);

  const hydrateMissingUsers = useCallback(
    async (ids: number[]) => {
      if (ids.length === 0) return;
      const entries = await Promise.all(
        ids.map(async (id) => {
          try {
            const { response, result } = await apiFetchJson<ApiResponse<{ user?: UserSummary }>>(
              `/profiles/${id}`,
              {},
              apiBaseUrl,
            );
            if (!response.ok || !result?.success || !result.data?.user) return null;
            return result.data.user as UserSummary;
          } catch {
            return null;
          }
        }),
      );
      const mapped: Record<number, UserSummary> = {};
      for (const user of entries) {
        if (user) mapped[user.id] = user;
      }
      if (Object.keys(mapped).length > 0) {
        setUsersByID((prev) => ({ ...prev, ...mapped }));
      }
    },
    [apiBaseUrl],
  );

  const hydrateMissingGroups = useCallback(
    async (ids: number[]) => {
      if (ids.length === 0) return;
      const entries = await Promise.all(
        ids.map(async (id) => {
          try {
            const { response, result } = await apiFetchJson<ApiResponse<GroupSummary>>(
              `/groups/${id}`,
              {},
              apiBaseUrl,
            );
            if (!response.ok || !result?.success || !result.data) return null;
            return result.data as GroupSummary;
          } catch {
            return null;
          }
        }),
      );
      const mapped: Record<number, GroupSummary> = {};
      for (const group of entries) {
        if (group) mapped[group.id] = group;
      }
      if (Object.keys(mapped).length > 0) {
        setGroupsByID((prev) => ({ ...prev, ...mapped }));
      }
    },
    [apiBaseUrl],
  );

  const markConversationRead = useCallback(
    async (id: number) => {
      setConversations((prev) =>
        prev.map((conv) => (conv.id === id ? { ...conv, unread_count: 0 } : conv)),
      );
      setUnreadCount((count) => Math.max(0, count - 1));
      await apiFetch(`/conversations/${id}/read`, { method: "PATCH" }, apiBaseUrl).catch(() => undefined);
      await refreshUnreadCounts();
    },
    [apiBaseUrl, refreshUnreadCounts],
  );

  const markAllRead = useCallback(async () => {
    const unreadIDs = conversations.filter((c) => (c.unread_count ?? 0) > 0).map((c) => c.id);
    await Promise.all(
      unreadIDs.map((id) =>
        apiFetch(`/conversations/${id}/read`, { method: "PATCH" }, apiBaseUrl),
      ),
    );
    setConversations((prev) => prev.map((c) => ({ ...c, unread_count: 0 })));
    setUnreadCount(0);
  }, [apiBaseUrl, conversations]);

  useEffect(() => {
    void refreshConversations();
  }, [refreshConversations]);

  useEffect(() => {
    const missingUsers = new Set<number>();
    const missingGroups = new Set<number>();
    conversations.forEach((conv) => {
      if (conv.other_user_id && !usersByID[conv.other_user_id]) {
        missingUsers.add(conv.other_user_id);
      }
      if (conv.group_id && !groupsByID[conv.group_id]) {
        missingGroups.add(conv.group_id);
      }
    });
    if (missingUsers.size > 0) {
      void hydrateMissingUsers(Array.from(missingUsers));
    }
    if (missingGroups.size > 0) {
      void hydrateMissingGroups(Array.from(missingGroups));
    }
  }, [conversations, groupsByID, hydrateMissingGroups, hydrateMissingUsers, usersByID]);

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
            if (msg.type === "chat_message" || msg.type === "unread_counts") {
              void refreshConversations();
              void refreshUnreadCounts();
            }
          } catch {
            // ignore
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

    connect();
    return () => {
      isMounted = false;
      if (reconnectTimerRef.current) {
        window.clearTimeout(reconnectTimerRef.current);
      }
      wsRef.current?.close();
      wsRef.current = null;
    };
  }, [refreshConversations, refreshUnreadCounts, wsBaseUrl]);

  const value = useMemo(
    () => ({
      conversations,
      usersByID,
      groupsByID,
      unreadCount,
      loading,
      refreshConversations,
      refreshUnreadCounts,
      markConversationRead,
      markAllRead,
    }),
    [
      conversations,
      usersByID,
      groupsByID,
      unreadCount,
      loading,
      refreshConversations,
      refreshUnreadCounts,
      markConversationRead,
      markAllRead,
    ],
  );

  return <MessagesContext.Provider value={value}>{children}</MessagesContext.Provider>;
}
