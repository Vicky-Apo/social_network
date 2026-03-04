"use client";

import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { createPortal } from "react-dom";
import { useRouter, useSearchParams } from "next/navigation";
import { Plus, Send, Wifi, WifiOff } from "lucide-react";
import { motion } from "framer-motion";
import TopNav from "@/components/TopNav";
import LeftNav from "@/components/LeftNav";
import Avatar from "@/components/Avatar";
import { fadeUp, viewportOnce } from "@/components/Motion";
import { shortDate } from "@/lib/date";
import { toMediaUrl } from "@/lib/media";
import { apiFetch, apiFetchJson, getApiBaseUrl } from "@/lib/api";
import { ApiResponse } from "@/lib/types";

type User = {
  id: number;
  email: string;
  first_name: string;
  last_name: string;
  nickname?: string | null;
  avatar_path?: string | null;
};

type UserListItem = {
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
  description?: string | null;
  is_member?: boolean;
};

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
  created_at: string;
  last_message?: MessageItem | null;
};

type MessageReaction = {
  message_id: number;
  user_id: number;
  emoji: string;
  created_at: string;
};

type WsMessage = {
  type: string;
  payload: unknown;
};

type MessageReactionPayload = {
  message_id: number;
  conversation_id: number;
  user_id: number;
  emoji: string;
  status: "added" | "removed" | string;
};

type TypingPayload = {
  conversation_id: number;
  user_id: number;
  is_typing: boolean;
};

type PresencePayload = {
  user_id: number;
};

type ActiveTab = "private" | "groups";

type PendingTarget =
  | { type: "direct"; userId: number }
  | { type: "group"; groupId: number }
  | null;

const emojiPalette = [
  "😂",
  "❤️",
  "😭",
  "✨",
  "🤣",
  "🔥",
  "🙏",
  "🥰",
  "👍",
  "😍",
  "😊",
  "✅",
  "🥺",
  "💀",
  "👀",
  "🤔",
  "🥳",
  "🎉",
  "🙄",
  "💙",
];
const pageSize = 10;


async function uploadMessageMedia(apiBaseUrl: string, file: File) {
  const formData = new FormData();
  formData.append("file", file);
  formData.append("kind", "message");
  const { response, result } = await apiFetchJson<
    { success?: boolean; data?: { path?: string }; error?: string }
  >("/uploads", { method: "POST", body: formData }, apiBaseUrl);
  if (!response.ok || !result?.success || !result.data?.path) {
    throw new Error(result?.error || "Could not upload media.");
  }
  return result.data.path;
}

function formatChatTitle(
  conversation: ConversationItem,
  usersByID: Record<number, UserListItem>,
  groupsByID: Record<number, GroupSummary>,
) {
  if (conversation.type === "direct" && conversation.other_user_id) {
    const person = usersByID[conversation.other_user_id];
    if (person) {
      return `${person.first_name} ${person.last_name}`;
    }
    return "User";
  }
  if (conversation.group_id) {
    const group = groupsByID[conversation.group_id];
    if (group) {
      return group.title || group.name || "Group";
    }
    return "Group";
  }
  return `Conversation #${conversation.id}`;
}

export default function MessagesPage() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const wsRef = useRef<WebSocket | null>(null);
  const typingTimerRef = useRef<number | null>(null);
  const typingSentRef = useRef(false);
  const typingClearTimersRef = useRef<Record<string, number>>({});
  const messageListRef = useRef<HTMLDivElement | null>(null);
  const preserveScrollRef = useRef<{ active: boolean; prevHeight: number }>({
    active: false,
    prevHeight: 0,
  });
  const autoScrollRef = useRef(true);

  const [user, setUser] = useState<User | null>(null);
  const [conversations, setConversations] = useState<ConversationItem[]>([]);
  const [messagesNewestFirst, setMessagesNewestFirst] = useState<MessageItem[]>([]);
  const [activeConversationID, setActiveConversationID] = useState<number | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [isMessagesLoading, setIsMessagesLoading] = useState(false);
  const [isLoadingOlder, setIsLoadingOlder] = useState(false);
  const [hasMoreMessages, setHasMoreMessages] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [messagesError, setMessagesError] = useState<string | null>(null);
  const [chatDraft, setChatDraft] = useState("");
  const [chatFile, setChatFile] = useState<File | null>(null);
  const [chatFileName, setChatFileName] = useState("");
  const [isSending, setIsSending] = useState(false);
  const [chatError, setChatError] = useState<string | null>(null);
  const [chatConnected, setChatConnected] = useState(false);
  const [activeTab, setActiveTab] = useState<ActiveTab>("private");
  const [directQuery, setDirectQuery] = useState("");
  const [contacts, setContacts] = useState<UserListItem[]>([]);
  const [memberGroups, setMemberGroups] = useState<GroupSummary[]>([]);
  const [pendingTarget, setPendingTarget] = useState<PendingTarget>(null);
  const [usersByID, setUsersByID] = useState<Record<number, UserListItem>>({});
  const [groupsByID, setGroupsByID] = useState<Record<number, GroupSummary>>({});
  const [reactionsByMessage, setReactionsByMessage] = useState<Record<number, MessageReaction[]>>({});
  const [typingByConversation, setTypingByConversation] = useState<Record<number, number[]>>({});
  const [onlineUsers, setOnlineUsers] = useState<Record<number, boolean>>({});
  const [openReactionPicker, setOpenReactionPicker] = useState<{
    messageId: number;
    top: number;
    left: number;
  } | null>(null);

  const apiBaseUrl = useMemo(() => getApiBaseUrl(), []);
  const fetchJson = useCallback(
    async <T,>(path: string, options: RequestInit = {}) =>
      apiFetchJson<ApiResponse<T>>(path, options, apiBaseUrl),
    [apiBaseUrl],
  );
  const wsBaseUrl = useMemo(() => {
    if (apiBaseUrl.startsWith("https://")) return apiBaseUrl.replace("https://", "wss://");
    if (apiBaseUrl.startsWith("http://")) return apiBaseUrl.replace("http://", "ws://");
    return apiBaseUrl;
  }, [apiBaseUrl]);

  const sendTypingIndicator = useCallback((isTyping: boolean) => {
    const ws = wsRef.current;
    if (!ws || ws.readyState !== WebSocket.OPEN || !activeConversationID) {
      return;
    }
    ws.send(
      JSON.stringify({
        type: "typing",
        payload: {
          conversation_id: activeConversationID,
          is_typing: isTyping,
        },
      }),
    );
  }, [activeConversationID]);

  const fetchConversations = useCallback(async () => {
    const { response, result } = await fetchJson<ConversationItem[]>(
      "/conversations?limit=50",
    );
    if (!response.ok || !result?.success) {
      throw new Error(result?.error || "Could not load conversations.");
    }
    const items = result.data ?? [];
    setConversations(items);
    setActiveConversationID((prev) => {
      if (prev && items.some((item) => item.id === prev)) {
        return prev;
      }
      return items[0]?.id ?? null;
    });
  }, [fetchJson]);

  const fetchReactionsForMessage = useCallback(
    async (messageID: number) => {
      const { response, result } = await fetchJson<MessageReaction[]>(
        `/messages/${messageID}/reactions`,
      );
      if (!response.ok || !result?.success) {
        return;
      }
      setReactionsByMessage((prev) => ({
        ...prev,
        [messageID]: result.data ?? [],
      }));
    },
    [fetchJson],
  );

  const fetchGroups = useCallback(async () => {
    try {
      const { response, result } = await fetchJson<GroupSummary[]>(
        "/groups?limit=200&offset=0",
      );
      if (!response.ok || !result?.success) return;
      const items = result.data ?? [];
      const mapped: Record<number, GroupSummary> = {};
      for (const item of items) {
        if (typeof item.id === "number") {
          mapped[item.id] = item;
        }
      }
      setGroupsByID(mapped);
      setMemberGroups(items.filter((group) => group.is_member));
    } catch {
      // ignore
    }
  }, [fetchJson]);

  const loadMessagesPage = useCallback(
    async (conversationID: number, offset: number, mode: "replace" | "append") => {
      if (mode === "replace") {
        setIsMessagesLoading(true);
        setMessagesError(null);
      } else {
        setIsLoadingOlder(true);
      }

      try {
        const { response, result } = await fetchJson<MessageItem[]>(
          `/conversations/${conversationID}/messages?limit=${pageSize}&offset=${offset}`,
        );
        if (!response.ok || !result?.success) {
          setMessagesError(result?.error || "Could not load messages.");
          if (mode === "replace") {
            setMessagesNewestFirst([]);
          }
          return;
        }

        const batch = result.data ?? [];
        if (mode === "replace") {
          setMessagesNewestFirst(batch);
          setReactionsByMessage({});
          autoScrollRef.current = true;
          await apiFetch(
            `/conversations/${conversationID}/read`,
            { method: "PATCH" },
            apiBaseUrl,
          ).catch(() => undefined);
        } else {
          setMessagesNewestFirst((prev) => [...prev, ...batch]);
        }
        if (batch.length > 0) {
          void Promise.all(batch.map((item) => fetchReactionsForMessage(item.id)));
        }
        setHasMoreMessages(batch.length >= pageSize);
      } finally {
        if (mode === "replace") {
          setIsMessagesLoading(false);
        } else {
          setIsLoadingOlder(false);
        }
      }
    },
    [apiBaseUrl],
  );

  useEffect(() => {
    let cancelled = false;

    const load = async () => {
      setIsLoading(true);
      setError(null);
      try {
        const { response: meResponse, result: meResult } = await fetchJson<User>("/auth/me");
        if (!meResponse.ok || !meResult?.success || !meResult.data) {
          if (!cancelled) {
            router.replace("/login");
          }
          return;
        }
        if (!cancelled) {
          setUser(meResult.data);
        }

        const [followers, following] = await Promise.all([
          fetchJson<UserListItem[]>(`/profiles/${meResult.data.id}/followers`),
          fetchJson<UserListItem[]>(`/profiles/${meResult.data.id}/following`),
        ]);
        if (!cancelled) {
          const combined = new Map<number, UserListItem>();
          if (followers.response.ok && followers.result?.success) {
            (followers.result.data ?? []).forEach((person) => combined.set(person.id, person));
          }
          if (following.response.ok && following.result?.success) {
            (following.result.data ?? []).forEach((person) => combined.set(person.id, person));
          }
          const list = Array.from(combined.values());
          setContacts(list);
          setUsersByID(
            Object.fromEntries(list.map((person) => [person.id, person])) as Record<number, UserListItem>,
          );
        }

        await fetchConversations();
      } catch {
        if (!cancelled) {
          setError("Network error while loading chats.");
        }
      } finally {
        if (!cancelled) {
          setIsLoading(false);
        }
      }
    };

    void load();
    return () => {
      cancelled = true;
    };
  }, [fetchConversations, fetchJson, router]);

  useEffect(() => {
    if (activeTab !== "groups") return;
    void fetchGroups();
  }, [activeTab, fetchGroups]);

  useEffect(() => {
    if (!activeConversationID) {
      setMessagesNewestFirst([]);
      setReactionsByMessage({});
      setHasMoreMessages(false);
      return;
    }
    void loadMessagesPage(activeConversationID, 0, "replace");
  }, [activeConversationID, loadMessagesPage]);

  useEffect(() => {
    const param = searchParams?.get("conversation");
    if (!param) return;
    const id = Number(param);
    if (!Number.isNaN(id) && id > 0) {
      setActiveConversationID(id);
      setPendingTarget(null);
    }
  }, [searchParams]);

  useEffect(() => {
    if (!user?.id) {
      return;
    }

    const ws = new WebSocket(`${wsBaseUrl}/ws`);
    wsRef.current = ws;

    ws.onopen = () => {
      setChatConnected(true);
      setChatError(null);
    };

    ws.onmessage = (event) => {
      const chunks = String(event.data).split("\n").filter(Boolean);
      chunks.forEach((raw) => {
        try {
          const msg = JSON.parse(raw) as WsMessage;
          if (msg.type === "chat_message") {
            const payload = msg.payload as MessageItem;
            if (payload.conversation_id === activeConversationID) {
              setMessagesNewestFirst((prev) => {
                if (prev.some((item) => item.id === payload.id)) {
                  return prev;
                }
                return [payload, ...prev];
              });
              // New messages start with no reactions; skip extra fetch.
              void apiFetch(
                `/conversations/${payload.conversation_id}/read`,
                { method: "PATCH" },
                apiBaseUrl,
              ).catch(() => undefined);
            }
            void fetchConversations();
          } else if (msg.type === "message_reaction") {
            const payload = msg.payload as MessageReactionPayload;
            setReactionsByMessage((prev) => {
              const current = prev[payload.message_id] ?? [];
              const exists = current.some(
                (reaction) =>
                  reaction.user_id === payload.user_id &&
                  reaction.emoji === payload.emoji,
              );
              let next = current;
              if (payload.status === "added" && !exists) {
                next = [
                  ...current,
                  {
                    message_id: payload.message_id,
                    user_id: payload.user_id,
                    emoji: payload.emoji,
                    created_at: new Date().toISOString(),
                  },
                ];
              } else if (payload.status === "removed" && exists) {
                next = current.filter(
                  (reaction) =>
                    !(
                      reaction.user_id === payload.user_id &&
                      reaction.emoji === payload.emoji
                    ),
                );
              }
              if (next === current) return prev;
              return { ...prev, [payload.message_id]: next };
            });
          } else if (msg.type === "typing") {
            const payload = msg.payload as TypingPayload;
            if (!payload?.conversation_id || !payload?.user_id || payload.user_id === user.id) {
              return;
            }
            const key = `${payload.conversation_id}:${payload.user_id}`;
            const existingTimer = typingClearTimersRef.current[key];
            if (existingTimer) {
              window.clearTimeout(existingTimer);
            }
            if (payload.is_typing) {
              setTypingByConversation((prev) => {
                const current = prev[payload.conversation_id] ?? [];
                if (current.includes(payload.user_id)) return prev;
                return {
                  ...prev,
                  [payload.conversation_id]: [...current, payload.user_id],
                };
              });
            }
            const timeoutID = window.setTimeout(() => {
              setTypingByConversation((prev) => {
                const current = prev[payload.conversation_id] ?? [];
                if (!current.includes(payload.user_id)) return prev;
                const next = current.filter((id) => id !== payload.user_id);
                return { ...prev, [payload.conversation_id]: next };
              });
              delete typingClearTimersRef.current[key];
            }, 3000);
            typingClearTimersRef.current[key] = timeoutID;
          } else if (msg.type === "user_online") {
            const payload = msg.payload as PresencePayload;
            if (payload?.user_id) {
              setOnlineUsers((prev) => ({ ...prev, [payload.user_id]: true }));
            }
          } else if (msg.type === "user_offline") {
            const payload = msg.payload as PresencePayload;
            if (payload?.user_id) {
              setOnlineUsers((prev) => ({ ...prev, [payload.user_id]: false }));
            }
          } else if (msg.type === "error") {
            const payload = msg.payload as { message?: string };
            setChatError(payload.message || "Chat error.");
          }
        } catch {
          // Ignore malformed websocket chunks.
        }
      });
    };

    ws.onclose = () => {
      setChatConnected(false);
    };

    ws.onerror = () => {
      setChatError("Could not connect to realtime chat.");
    };

    return () => {
      ws.close();
      wsRef.current = null;
    };
  }, [activeConversationID, apiBaseUrl, fetchConversations, fetchReactionsForMessage, user?.id, wsBaseUrl]);

  useEffect(() => {
    const scroller = messageListRef.current;
    if (!scroller) return;

    if (preserveScrollRef.current.active) {
      const previousHeight = preserveScrollRef.current.prevHeight;
      const nextHeight = scroller.scrollHeight;
      scroller.scrollTop += nextHeight - previousHeight;
      preserveScrollRef.current.active = false;
      return;
    }

    if (autoScrollRef.current) {
      scroller.scrollTop = scroller.scrollHeight;
    }
  }, [messagesNewestFirst]);

  const stopTyping = useCallback(() => {
    if (typingTimerRef.current) {
      window.clearTimeout(typingTimerRef.current);
      typingTimerRef.current = null;
    }
    if (typingSentRef.current) {
      sendTypingIndicator(false);
      typingSentRef.current = false;
    }
  }, [sendTypingIndicator]);

  const sendMessageToActiveConversation = async () => {
    const ws = wsRef.current;
    const content = chatDraft.trim();
    if (!ws || ws.readyState !== WebSocket.OPEN) {
      setChatError("Chat is not connected.");
      return;
    }
    const activeConversation = conversations.find((item) => item.id === activeConversationID);
    if (!content && !chatFile) {
      setChatError("Write a message or attach media before sending.");
      return;
    }

    if (isSending) {
      return;
    }
    setIsSending(true);
    setChatError(null);

    let mediaPath: string | undefined;
    try {
      if (chatFile) {
        mediaPath = await uploadMessageMedia(apiBaseUrl, chatFile);
      }
    } catch (err) {
      setChatError(err instanceof Error ? err.message : "Could not upload media.");
      setIsSending(false);
      return;
    }

    const targetDirectID =
      activeConversation?.type === "direct"
        ? Number(activeConversation.other_user_id ?? 0)
        : pendingTarget?.type === "direct"
          ? pendingTarget.userId
          : 0;
    const targetGroupID =
      activeConversation?.type !== "direct"
        ? Number(activeConversation?.group_id ?? 0)
        : pendingTarget?.type === "group"
          ? pendingTarget.groupId
          : 0;

    if (targetDirectID) {
      const recipientID = targetDirectID;
      if (!recipientID) {
        setChatError("Recipient is missing for this conversation.");
        setIsSending(false);
        return;
      }
      ws.send(
        JSON.stringify({
          type: "chat_message",
          payload: {
            recipient_id: recipientID,
            content: content || undefined,
            media_path: mediaPath,
          },
        }),
      );
    } else if (targetGroupID) {
      const groupID = targetGroupID;
      if (!groupID) {
        setChatError("Group is missing for this conversation.");
        setIsSending(false);
        return;
      }
      ws.send(
        JSON.stringify({
          type: "chat_message",
          payload: {
            group_id: groupID,
            content: content || undefined,
            media_path: mediaPath,
          },
        }),
      );
    } else {
      setChatError("Select a conversation first.");
      setIsSending(false);
      return;
    }

    stopTyping();
    setChatDraft("");
    setChatFile(null);
    setChatFileName("");
    setPendingTarget(null);
    setChatError(null);
    setIsSending(false);
    void fetchConversations();
  };

  const handleLoadOlder = async () => {
    if (!activeConversationID || isLoadingOlder || !hasMoreMessages) {
      return;
    }
    const scroller = messageListRef.current;
    if (scroller) {
      preserveScrollRef.current = {
        active: true,
        prevHeight: scroller.scrollHeight,
      };
    }
    await loadMessagesPage(activeConversationID, messagesNewestFirst.length, "append");
  };

  const toggleMessageReaction = async (messageID: number, emoji: string) => {
    const { response, result } = await fetchJson<unknown>(`/messages/${messageID}/reactions`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ emoji }),
    });
    if (!response.ok || !result?.success) {
      setChatError(result?.error || "Could not update reaction.");
      return;
    }
    await fetchReactionsForMessage(messageID);
  };

  const directConversations = useMemo(
    () => conversations.filter((item) => item.type === "direct"),
    [conversations],
  );
  const groupConversations = useMemo(
    () => conversations.filter((item) => item.type !== "direct" && item.group_id),
    [conversations],
  );
  const filteredContacts = useMemo(() => {
    const query = directQuery.trim().toLowerCase();
    const base = contacts.filter((item) => item.id !== user?.id);
    if (!query) return base;
    return base.filter((item) =>
      `${item.first_name} ${item.last_name} ${item.nickname ?? ""}`.toLowerCase().includes(query),
    );
  }, [contacts, directQuery, user?.id]);
  const activeConversation = conversations.find((item) => item.id === activeConversationID) ?? null;
  const pendingTitle =
    pendingTarget?.type === "direct"
      ? (() => {
          const person = usersByID[pendingTarget.userId] || contacts.find((c) => c.id === pendingTarget.userId);
          return person ? `${person.first_name} ${person.last_name}` : "Direct message";
        })()
      : pendingTarget?.type === "group"
        ? (() => {
            const group = groupsByID[pendingTarget.groupId] || memberGroups.find((g) => g.id === pendingTarget.groupId);
            return group ? group.title || group.name || "Group" : "Group chat";
          })()
        : null;
  const activeConversationTitle = activeConversation
    ? formatChatTitle(activeConversation, usersByID, groupsByID)
    : pendingTitle ?? "Select a conversation";
  const activeConversationType = activeConversation
    ? activeConversation.type
    : pendingTarget?.type === "direct"
      ? "direct"
      : pendingTarget?.type === "group"
        ? "group"
        : "";
  const messagesOldestFirst = useMemo(
    () => [...messagesNewestFirst].reverse(),
    [messagesNewestFirst],
  );
  const typingUserIDs = activeConversationID ? typingByConversation[activeConversationID] ?? [] : [];
  const typingLabel = typingUserIDs
    .map((id) => {
      const person = usersByID[id];
      return person ? `${person.first_name} ${person.last_name}`.trim() : "User";
    })
    .join(", ");

  return (
    <div
      className="min-h-screen text-neutral-100"
      style={{
        backgroundImage: "url('/messages-bg.png')",
        backgroundSize: "cover",
        backgroundPosition: "center",
        backgroundAttachment: "fixed",
      }}
    >
      <TopNav user={user ?? undefined} onLogout={() => router.replace("/login")} variant="dark" />

      <main className="mx-auto grid w-full max-w-6xl gap-6 px-4 py-6 sm:px-6 lg:grid-cols-[220px_minmax(0,1fr)]">
        <aside className="hidden lg:block">
          <LeftNav user={user ?? undefined} activeHref="/messages" variant="dark" />
        </aside>

        <section className="space-y-5">
          <motion.div
            initial="hidden"
            whileInView="show"
            viewport={viewportOnce}
            variants={fadeUp}
            className="rounded-2xl border border-white/10 bg-white/5 p-4 backdrop-blur-sm sm:p-5"
          >
            <div className="flex flex-wrap items-start justify-between gap-4">
              <div>
                <h1 className="text-xl font-semibold tracking-tight text-white">Messages</h1>
              </div>
              <span
                className={`inline-flex items-center gap-1 rounded-full px-3 py-1 text-xs font-medium ${
                  chatConnected
                    ? "bg-white/10 text-white"
                    : "bg-rose-500/20 text-rose-400"
                }`}
              >
                {chatConnected ? <Wifi className="h-3.5 w-3.5" /> : <WifiOff className="h-3.5 w-3.5" />}
                {chatConnected ? "Connected" : "Offline"}
              </span>
            </div>
          </motion.div>

          {isLoading ? (
            <article className="rounded-2xl border border-white/10 bg-white/5 p-6 text-sm text-neutral-400 backdrop-blur-sm">
              Loading conversations...
            </article>
          ) : error ? (
            <article className="rounded-2xl border border-rose-500/30 bg-rose-500/10 p-6 text-sm text-rose-400 backdrop-blur-sm">
              {error}
            </article>
          ) : (
            <div className="grid gap-4 lg:grid-cols-[280px_minmax(0,1fr)]">
              <aside className="rounded-2xl border border-white/10 bg-white/5 p-3 backdrop-blur-sm">
                <div className="flex items-center gap-2 px-2">
                  <button
                    type="button"
                    onClick={() => setActiveTab("private")}
                    className={`rounded-xl px-3 py-1 text-xs font-semibold transition ${
                      activeTab === "private"
                        ? "bg-white text-[#2b2929]"
                        : "border border-white/20 bg-white/5 text-neutral-400 hover:bg-white/10 hover:text-white"
                    }`}
                  >
                    Private
                  </button>
                  <button
                    type="button"
                    onClick={() => setActiveTab("groups")}
                    className={`rounded-xl px-3 py-1 text-xs font-semibold transition ${
                      activeTab === "groups"
                        ? "bg-white text-[#2b2929]"
                        : "border border-white/20 bg-white/5 text-neutral-400 hover:bg-white/10 hover:text-white"
                    }`}
                  >
                    Groups
                  </button>
                </div>

                {activeTab === "private" ? (
                  <div className="mt-4 space-y-4">
                    <div>
                      <h2 className="px-2 text-sm font-semibold text-white">Direct chats</h2>
                      <div className="mt-3 space-y-2">
                        {directConversations.length === 0 ? (
                          <p className="px-2 text-xs text-neutral-400">No direct chats yet.</p>
                        ) : (
                          directConversations.map((conversation) => {
                            const active = conversation.id === activeConversationID;
                            const title = formatChatTitle(conversation, usersByID, groupsByID);
                            const isDirect = conversation.other_user_id;
                            const userItem = isDirect ? usersByID[conversation.other_user_id ?? 0] : null;
                            const isOnline =
                              userItem?.id !== undefined ? Boolean(onlineUsers[userItem.id]) : false;
                            const lastMessage = conversation.last_message;
                            const typingUsers = typingByConversation[conversation.id] ?? [];
                            const preview =
                              lastMessage?.content ||
                              (lastMessage?.media_path ? "(media)" : "(no message yet)");
                            const previewText = typingUsers.length > 0 ? "Typing..." : preview;
                            return (
                              <button
                                key={conversation.id}
                                type="button"
                                onClick={() => {
                                  setActiveConversationID(conversation.id);
                                  setPendingTarget(null);
                                }}
                                className={`w-full rounded-2xl border px-3 py-2 text-left transition ${
                                  active
                                    ? "border-white/20 bg-white/10 text-white"
                                    : "border-white/10 bg-white/5 text-neutral-300 hover:bg-white/10 hover:text-white"
                                }`}
                              >
                                <div className="flex items-center gap-2">
                                  <Avatar
                                    src={
                                      userItem?.avatar_path
                                        ? toMediaUrl(apiBaseUrl, userItem.avatar_path)
                                        : null
                                    }
                                    name={title}
                                    size={32}
                                    textClassName="text-[10px]"
                                  />
                                  <div className="min-w-0">
                                    <p className="text-xs font-semibold truncate">{title}</p>
                                    <p
                                      className={`mt-1 text-[11px] ${
                                        active ? "text-neutral-300" : "text-neutral-500"
                                      } truncate`}
                                    >
                                      {preview}
                                    </p>
                                  </div>
                                  <span
                                    className={`ml-auto h-2 w-2 rounded-full ${
                                      isOnline ? "bg-emerald-400" : "bg-neutral-500"
                                    }`}
                                  />
                                </div>
                              </button>
                            );
                          })
                        )}
                      </div>
                    </div>

                    <div>
                      <h3 className="px-2 text-sm font-semibold text-white">Contacts</h3>
                      <input
                        value={directQuery}
                        onChange={(event) => setDirectQuery(event.target.value)}
                        placeholder="Search contacts..."
                        className="mt-2 h-9 w-full rounded-xl border border-white/20 bg-white/5 px-3 text-xs text-white placeholder:text-neutral-500 outline-none focus:border-white/40"
                      />
                      <div className="mt-3 space-y-2">
                        {filteredContacts.length === 0 ? (
                          <p className="px-2 text-xs text-neutral-500">No contacts found.</p>
                        ) : (
                          filteredContacts.map((person) => {
                            const existing = directConversations.find(
                              (conv) => conv.other_user_id === person.id,
                            );
                            const isOnline = Boolean(onlineUsers[person.id]);
                            return (
                              <button
                                type="button"
                                key={person.id}
                                onClick={() => {
                                  setActiveConversationID(existing?.id ?? null);
                                  setPendingTarget(
                                    existing ? null : { type: "direct", userId: person.id },
                                  );
                                }}
                                className="flex w-full items-center justify-between rounded-xl border border-white/10 bg-white/5 px-3 py-2 text-left text-neutral-300 transition hover:bg-white/10 hover:text-white"
                              >
                                <div className="flex items-center gap-3">
                                  <Avatar
                                    src={
                                      person.avatar_path
                                        ? toMediaUrl(apiBaseUrl, person.avatar_path)
                                        : null
                                    }
                                    name={`${person.first_name} ${person.last_name}`}
                                    size={32}
                                    textClassName="text-[10px]"
                                  />
                                  <div>
                                    <p className="text-xs font-semibold text-white">
                                      {person.first_name} {person.last_name}
                                    </p>
                                    <p className="text-[11px] text-neutral-500">
                                      @{person.nickname || "user"}
                                    </p>
                                  </div>
                                </div>
                                <span
                                  className={`h-2 w-2 rounded-full ${
                                    isOnline ? "bg-emerald-400" : "bg-neutral-500"
                                  }`}
                                />
                              </button>
                            );
                          })
                        )}
                      </div>
                    </div>
                  </div>
                ) : (
                  <div className="mt-4 space-y-4">
                    <div>
                      <h2 className="px-2 text-sm font-semibold text-white">Group chats</h2>
                      <div className="mt-3 space-y-2">
                        {groupConversations.length === 0 ? (
                          <p className="px-2 text-xs text-neutral-500">No group chats yet.</p>
                        ) : (
                          groupConversations.map((conversation) => {
                            const active = conversation.id === activeConversationID;
                            const title = formatChatTitle(conversation, usersByID, groupsByID);
                            const lastMessage = conversation.last_message;
                            const typingUsers = typingByConversation[conversation.id] ?? [];
                            const preview =
                              lastMessage?.content ||
                              (lastMessage?.media_path ? "(media)" : "(no message yet)");
                            const previewText = typingUsers.length > 0 ? "Typing..." : preview;
                            return (
                              <button
                                key={conversation.id}
                                type="button"
                                onClick={() => {
                                  setActiveConversationID(conversation.id);
                                  setPendingTarget(null);
                                }}
                                className={`w-full rounded-2xl border px-3 py-2 text-left transition ${
                                  active
                                    ? "border-white/20 bg-white/10 text-white"
                                    : "border-white/10 bg-white/5 text-neutral-300 hover:bg-white/10 hover:text-white"
                                }`}
                              >
                                <div className="flex items-center gap-2">
                                  <Avatar
                                    name={title}
                                    size={32}
                                    className="border-white/20 bg-white/20"
                                    textClassName="text-[10px] text-white"
                                  />
                                  <div className="min-w-0">
                                    <p className="text-xs font-semibold truncate">{title}</p>
                                    <p
                                      className={`mt-1 text-[11px] ${
                                        typingUsers.length > 0
                                          ? "text-white"
                                          : active
                                            ? "text-neutral-300"
                                            : "text-neutral-500"
                                      } truncate`}
                                    >
                                      {previewText}
                                    </p>
                                  </div>
                                </div>
                              </button>
                            );
                          })
                        )}
                      </div>
                    </div>

                    <div>
                      <h3 className="px-2 text-sm font-semibold text-white">Your groups</h3>
                      <div className="mt-3 space-y-2">
                        {memberGroups.length === 0 ? (
                          <p className="px-2 text-xs text-neutral-500">You are not in any groups.</p>
                        ) : (
                          memberGroups.map((group) => {
                            const existing = groupConversations.find(
                              (conv) => conv.group_id === group.id,
                            );
                            const typingUsers = existing?.id
                              ? typingByConversation[existing.id] ?? []
                              : [];
                            const title = group.title || group.name || "Group";
                            return (
                              <button
                                key={group.id}
                                type="button"
                                onClick={() => {
                                  setActiveConversationID(existing?.id ?? null);
                                  setPendingTarget(
                                    existing ? null : { type: "group", groupId: group.id },
                                  );
                                }}
                                className="flex w-full items-center gap-2 rounded-xl border border-white/10 bg-white/5 px-3 py-2 text-left text-neutral-300 transition hover:bg-white/10 hover:text-white"
                              >
                                <div className="inline-flex h-8 w-8 items-center justify-center rounded-full bg-emerald-600 text-[10px] font-semibold text-white">
                                  {title.slice(0, 2).toUpperCase()}
                                </div>
                                <div className="min-w-0">
                                  <p className="text-xs font-semibold text-white truncate">{title}</p>
                                  {typingUsers.length > 0 ? (
                                    <p className="text-[11px] text-white truncate">Typing...</p>
                                  ) : group.description ? (
                                    <p className="text-[11px] text-neutral-500 truncate">
                                      {group.description}
                                    </p>
                                  ) : null}
                                </div>
                              </button>
                            );
                          })
                        )}
                      </div>
                    </div>
                  </div>
                )}
              </aside>

              <article className="rounded-2xl border border-white/10 bg-white/5 p-4 backdrop-blur-sm min-h-[720px]">
                <div className="flex flex-wrap items-center justify-between gap-3 border-b border-white/10 pb-3">
                  <div className="flex items-center gap-3">
                    {activeConversation?.type === "direct" && activeConversation?.other_user_id ? (
                      <Avatar
                        src={
                          usersByID[activeConversation.other_user_id]?.avatar_path
                            ? toMediaUrl(apiBaseUrl, usersByID[activeConversation.other_user_id].avatar_path!)
                            : null
                        }
                        name={
                          usersByID[activeConversation.other_user_id]
                            ? `${usersByID[activeConversation.other_user_id].first_name} ${usersByID[activeConversation.other_user_id].last_name}`
                            : "User"
                        }
                        size={40}
                        textClassName="text-xs"
                      />
                    ) : pendingTarget?.type === "direct" ? (
                      <Avatar
                        src={
                          (usersByID[pendingTarget.userId] || contacts.find((c) => c.id === pendingTarget.userId))?.avatar_path
                            ? toMediaUrl(
                                apiBaseUrl,
                                (usersByID[pendingTarget.userId] || contacts.find((c) => c.id === pendingTarget.userId))!.avatar_path!,
                              )
                            : null
                        }
                        name={
                          (() => {
                            const p = usersByID[pendingTarget.userId] || contacts.find((c) => c.id === pendingTarget.userId);
                            return p ? `${p.first_name} ${p.last_name}` : "User";
                          })()
                        }
                        size={40}
                        textClassName="text-xs"
                      />
                    ) : null}
                    <div>
                      <h2 className="text-sm font-semibold text-white">{activeConversationTitle}</h2>
                      <p className="text-xs text-white">
                        {activeConversationType ? `Type: ${activeConversationType}` : "Pick a chat on the left to start."}
                      </p>
                    </div>
                  </div>
                </div>

                {typingUserIDs.length > 0 ? (
                  <p className="mt-3 text-xs text-neutral-500">{typingLabel} typing...</p>
                ) : null}

                <div
                  ref={messageListRef}
                  onScroll={(event) => {
                    const node = event.currentTarget;
                    const distanceToBottom = node.scrollHeight - node.scrollTop - node.clientHeight;
                    autoScrollRef.current = distanceToBottom < 120;
                  }}
                  className="mt-4 max-h-[840px] space-y-3 overflow-y-auto pr-1"
                >
                  {hasMoreMessages && !isMessagesLoading ? (
                    <div className="flex justify-center">
                      <button
                        type="button"
                        onClick={() => void handleLoadOlder()}
                        disabled={isLoadingOlder}
                        className="rounded-xl border border-white/20 bg-white/5 px-3 py-1 text-[11px] font-semibold text-neutral-300 transition hover:bg-white/10 hover:text-white disabled:opacity-60"
                      >
                        {isLoadingOlder ? "Loading..." : "Load older messages"}
                      </button>
                    </div>
                  ) : null}

                  {isMessagesLoading ? (
                    <p className="text-sm text-neutral-500">Loading messages...</p>
                  ) : messagesError ? (
                    <p className="rounded-xl border border-rose-500/30 bg-rose-500/10 px-3 py-2 text-xs text-rose-400">
                      {messagesError}
                    </p>
                  ) : messagesOldestFirst.length === 0 ? (
                    <p className="text-sm text-neutral-500">No messages yet.</p>
                  ) : (
                    messagesOldestFirst.map((message) => {
                      const mine = message.sender_id === user?.id;
                      const sender = usersByID[message.sender_id];
                      const senderName = sender
                        ? `${sender.first_name} ${sender.last_name}`
                        : "User";
                      const reactions = reactionsByMessage[message.id] ?? [];
                      const reactionCounts = reactions.reduce<Record<string, number>>((acc, item) => {
                        acc[item.emoji] = (acc[item.emoji] ?? 0) + 1;
                        return acc;
                      }, {});
                      const myEmojis = new Set(
                        reactions.filter((item) => item.user_id === user?.id).map((item) => item.emoji),
                      );

                      return (
                        <div
                          key={`${message.id}-${message.created_at}`}
                          className={`relative max-w-[82%] overflow-visible rounded-xl px-3 py-2 text-sm ${
                            mine
                              ? "ml-auto bg-neutral-700 text-white"
                              : "bg-neutral-800 text-white"
                          }`}
                        >
                          <div className={`flex items-center gap-2 ${mine ? "w-full justify-end" : ""}`}>
                            <p className={`text-[11px] font-semibold ${mine ? "text-neutral-600" : "text-neutral-400"}`}>
                              {senderName}
                            </p>
                          </div>
                          {message.content ? <p className="text-white">{message.content}</p> : null}
                          {message.media_path ? (
                            <div className="mt-2 overflow-hidden rounded-xl border border-white/10 bg-slate-900/60">
                              <img
                                src={toMediaUrl(apiBaseUrl, message.media_path)}
                                alt="Message media"
                                className="max-h-64 w-full object-contain"
                              />
                            </div>
                          ) : null}
                          {!message.content && !message.media_path ? <p>(empty)</p> : null}
                          <p className={`mt-1 text-[10px] ${mine ? "text-neutral-500" : "text-neutral-500"}`}>
                            {shortDate(message.created_at)}
                          </p>

                          <div className="mt-2 flex flex-wrap gap-1">
                            {Object.entries(reactionCounts).map(([emoji, count]) => (
                              <button
                                key={`${message.id}-${emoji}`}
                                type="button"
                                onClick={() => void toggleMessageReaction(message.id, emoji)}
                                className={`rounded-full border px-2 py-0.5 text-[11px] ${
                                  myEmojis.has(emoji)
                                    ? "border-emerald-500/30 bg-emerald-500/20 text-emerald-400"
                                    : "border-slate-600 bg-slate-700/80 text-neutral-100"
                                }`}
                              >
                                {emoji} {count}
                              </button>
                            ))}
                            <div className="relative">
                              <button
                                type="button"
                                onClick={(event) => {
                                  const rect = event.currentTarget.getBoundingClientRect();
                                  setOpenReactionPicker((prev) => {
                                    if (prev?.messageId === message.id) {
                                      return null;
                                    }
                                    return {
                                      messageId: message.id,
                                      top: Math.max(12, rect.top - 12),
                                      left: Math.max(12, rect.left),
                                    };
                                  });
                                }}
                                className="inline-flex items-center justify-center rounded-full border border-slate-600 bg-slate-700/80 px-2 py-0.5 text-[11px] text-neutral-100"
                              >
                                <Plus className="h-3 w-3" />
                              </button>
                            </div>
                          </div>
                        </div>
                      );
                    })
                  )}
                </div>

                <div className="mt-4 rounded-xl border border-white/10 bg-white/5 p-3">
                  <textarea
                    value={chatDraft}
                    onChange={(event) => {
                      setChatDraft(event.target.value);
                      if (!typingSentRef.current) {
                        sendTypingIndicator(true);
                        typingSentRef.current = true;
                      }
                      if (typingTimerRef.current) {
                        window.clearTimeout(typingTimerRef.current);
                      }
                      typingTimerRef.current = window.setTimeout(() => {
                        sendTypingIndicator(false);
                        typingSentRef.current = false;
                        typingTimerRef.current = null;
                      }, 1200);
                    }}
                    onBlur={stopTyping}
                    rows={3}
                    placeholder="Write a message (emoji supported)"
                    className="w-full resize-none rounded-xl border border-white/20 bg-white/5 px-3 py-2 text-sm text-white placeholder:text-neutral-500 outline-none transition focus:border-white/40"
                  />
                  <div className="mt-3 flex flex-wrap items-center gap-2">
                    <label className="inline-flex h-9 items-center gap-2 rounded-xl border border-white/20 bg-white/5 px-3 text-xs font-semibold text-neutral-300 transition hover:bg-white/10 hover:text-white">
                      <input
                        type="file"
                        accept="image/png,image/jpeg,image/gif"
                        className="hidden"
                        onChange={(event) => {
                          const file = event.target.files?.[0] ?? null;
                          setChatFile(file);
                          setChatFileName(file?.name ?? "");
                        }}
                      />
                      <Plus className="h-3.5 w-3.5" />
                      Add media
                    </label>
                    <button
                      type="button"
                      onClick={sendMessageToActiveConversation}
                      disabled={isSending}
                      className="inline-flex items-center gap-2 rounded-xl bg-white px-4 py-2 text-xs font-semibold text-[#2b2929] transition hover:bg-neutral-100"
                    >
                      <Send className="h-3.5 w-3.5" />
                      {isSending ? "Sending..." : "Send"}
                    </button>
                  </div>
                  {chatFileName ? (
                    <div className="mt-2 flex items-center justify-between rounded-xl border border-white/10 bg-white/5 px-3 py-2 text-xs text-neutral-400">
                      <span>Attached: {chatFileName}</span>
                      <button
                        type="button"
                        onClick={() => {
                          setChatFile(null);
                          setChatFileName("");
                        }}
                        className="rounded-xl border border-white/20 bg-white/5 px-2 py-1 text-[10px] font-semibold text-neutral-400"
                      >
                        Remove
                      </button>
                    </div>
                  ) : null}
                  {chatError ? <p className="mt-2 text-xs text-rose-400">{chatError}</p> : null}
                </div>
              </article>
            </div>
          )}
        </section>

      </main>

      {openReactionPicker
        ? createPortal(
            <div
              className="fixed z-[120]"
              style={{
                left: openReactionPicker.left,
                top: openReactionPicker.top,
                transform: "translateY(-100%)",
              }}
            >
              <div className="w-56 rounded-xl border border-white/10 bg-white/5 p-2 shadow-2xl backdrop-blur-sm">
                <div className="grid grid-cols-5 gap-1">
                  {emojiPalette.map((emoji) => (
                    <button
                      key={`picker-${openReactionPicker.messageId}-${emoji}`}
                      type="button"
                      onClick={() => {
                        void toggleMessageReaction(openReactionPicker.messageId, emoji);
                        setOpenReactionPicker(null);
                      }}
                      className="flex h-9 w-9 items-center justify-center rounded-xl border border-transparent text-base transition hover:border-white/20 hover:bg-white/10"
                    >
                      {emoji}
                    </button>
                  ))}
                </div>
              </div>
            </div>,
            document.body,
          )
        : null}
    </div>
  );
}
