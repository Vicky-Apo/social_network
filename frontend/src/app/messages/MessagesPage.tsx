"use client";

import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { MessageSquare, Send, Wifi, WifiOff } from "lucide-react";
import { motion } from "framer-motion";
import TopNav from "@/components/TopNav";
import LeftNav from "@/components/LeftNav";
import { fadeUp, viewportOnce } from "@/components/Motion";

type ApiResponse<T> = {
  success?: boolean;
  data?: T;
  error?: string;
};

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

type TypingPayload = {
  conversation_id: number;
  user_id: number;
  is_typing: boolean;
};

const quickReactions = ["👍", "❤️", "😂", "🔥"];
const pageSize = 30;

function initials(first?: string, last?: string) {
  const left = first?.trim().charAt(0) ?? "";
  const right = last?.trim().charAt(0) ?? "";
  return `${left}${right}`.toUpperCase() || "U";
}

function toMediaUrl(apiBaseUrl: string, path?: string | null) {
  if (!path) return "";
  if (path.startsWith("http://") || path.startsWith("https://")) return path;
  const normalized = path.startsWith("/") ? path : `/${path}`;
  return `${apiBaseUrl}${normalized}`;
}

function shortDate(value: string) {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return "Just now";
  }
  return date.toLocaleDateString(undefined, { month: "short", day: "numeric" });
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
  const wsRef = useRef<WebSocket | null>(null);
  const typingTimerRef = useRef<number | null>(null);
  const typingSentRef = useRef(false);
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
  const [quickDirectDraft, setQuickDirectDraft] = useState("");
  const [chatError, setChatError] = useState<string | null>(null);
  const [chatConnected, setChatConnected] = useState(false);
  const [directQuery, setDirectQuery] = useState("");
  const [selectedDirectUser, setSelectedDirectUser] = useState<UserListItem | null>(null);
  const [people, setPeople] = useState<UserListItem[]>([]);
  const [usersByID, setUsersByID] = useState<Record<number, UserListItem>>({});
  const [groupsByID, setGroupsByID] = useState<Record<number, GroupSummary>>({});
  const [isHydratingNames, setIsHydratingNames] = useState(false);
  const [reactionsByMessage, setReactionsByMessage] = useState<Record<number, MessageReaction[]>>({});
  const [typingByConversation, setTypingByConversation] = useState<Record<number, number[]>>({});

  const apiBaseUrl = useMemo(
    () =>
      process.env.NEXT_PUBLIC_API_BASE_URL?.trim().replace(/\/+$/, "") ||
      "http://localhost:8080",
    [],
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
    const response = await fetch(`${apiBaseUrl}/conversations?limit=50`, {
      credentials: "include",
    });
    const result = (await response.json().catch(() => null)) as
      | ApiResponse<ConversationItem[]>
      | null;
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
  }, [apiBaseUrl]);

  const fetchReactionsForMessage = useCallback(
    async (messageID: number) => {
      const response = await fetch(`${apiBaseUrl}/messages/${messageID}/reactions`, {
        credentials: "include",
      });
      const result = (await response.json().catch(() => null)) as
        | ApiResponse<MessageReaction[]>
        | null;
      if (!response.ok || !result?.success) {
        return;
      }
      setReactionsByMessage((prev) => ({
        ...prev,
        [messageID]: result.data ?? [],
      }));
    },
    [apiBaseUrl],
  );

  const fetchReactionsForBatch = useCallback(
    async (batch: MessageItem[]) => {
      await Promise.all(batch.map((message) => fetchReactionsForMessage(message.id)));
    },
    [fetchReactionsForMessage],
  );

  const fetchGroups = useCallback(async () => {
    try {
      const response = await fetch(`${apiBaseUrl}/groups?limit=200&offset=0`, {
        credentials: "include",
      });
      const result = (await response.json().catch(() => null)) as
        | ApiResponse<GroupSummary[]>
        | null;
      if (!response.ok || !result?.success) return;
      const items = result.data ?? [];
      const mapped: Record<number, GroupSummary> = {};
      for (const item of items) {
        if (typeof item.id === "number") {
          mapped[item.id] = item;
        }
      }
      setGroupsByID(mapped);
    } catch {
      // ignore
    }
  }, [apiBaseUrl]);

  const hydrateMissingUsers = useCallback(
    async (ids: number[]) => {
      if (ids.length === 0) return;
      setIsHydratingNames(true);
      try {
        const entries = await Promise.all(
          ids.map(async (id) => {
            try {
              const response = await fetch(`${apiBaseUrl}/profiles/${id}`, {
                credentials: "include",
              });
              const result = (await response.json().catch(() => null)) as
                | ApiResponse<{ user?: UserListItem }>
                | null;
              if (!response.ok || !result?.success || !result.data?.user) {
                return null;
              }
              return result.data.user as UserListItem;
            } catch {
              return null;
            }
          }),
        );
        const mapped: Record<number, UserListItem> = {};
        for (const user of entries) {
          if (user) mapped[user.id] = user;
        }
        if (Object.keys(mapped).length > 0) {
          setUsersByID((prev) => ({ ...prev, ...mapped }));
        }
      } finally {
        setIsHydratingNames(false);
      }
    },
    [apiBaseUrl],
  );

  const hydrateMissingGroups = useCallback(
    async (ids: number[]) => {
      if (ids.length === 0) return;
      try {
        const entries = await Promise.all(
          ids.map(async (id) => {
            try {
              const response = await fetch(`${apiBaseUrl}/groups/${id}`, {
                credentials: "include",
              });
              const result = (await response.json().catch(() => null)) as
                | ApiResponse<{ id?: number; title?: string; name?: string }>
                | null;
              if (!response.ok || !result?.success || !result.data) {
                return null;
              }
              const data = result.data as { id?: number; title?: string; name?: string };
              const groupID = typeof data.id === "number" ? data.id : id;
              return { id: groupID, title: data.title ?? data.name ?? `Group ${groupID}` } as GroupSummary;
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
      } catch {
        // ignore
      }
    },
    [apiBaseUrl],
  );

  const loadMessagesPage = useCallback(
    async (conversationID: number, offset: number, mode: "replace" | "append") => {
      if (mode === "replace") {
        setIsMessagesLoading(true);
        setMessagesError(null);
      } else {
        setIsLoadingOlder(true);
      }

      try {
        const response = await fetch(
          `${apiBaseUrl}/conversations/${conversationID}/messages?limit=${pageSize}&offset=${offset}`,
          { credentials: "include" },
        );
        const result = (await response.json().catch(() => null)) as
          | ApiResponse<MessageItem[]>
          | null;
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
          await fetch(`${apiBaseUrl}/conversations/${conversationID}/read`, {
            method: "PATCH",
            credentials: "include",
          }).catch(() => undefined);
        } else {
          setMessagesNewestFirst((prev) => [...prev, ...batch]);
        }
        setHasMoreMessages(batch.length >= pageSize);
        void fetchReactionsForBatch(batch);
      } finally {
        if (mode === "replace") {
          setIsMessagesLoading(false);
        } else {
          setIsLoadingOlder(false);
        }
      }
    },
    [apiBaseUrl, fetchReactionsForBatch],
  );

  useEffect(() => {
    let cancelled = false;

    const load = async () => {
      setIsLoading(true);
      setError(null);
      try {
        const meResponse = await fetch(`${apiBaseUrl}/auth/me`, {
          credentials: "include",
        });
        const meResult = (await meResponse.json().catch(() => null)) as ApiResponse<User> | null;
        if (!meResponse.ok || !meResult?.success || !meResult.data) {
          if (!cancelled) {
            router.replace("/login");
          }
          return;
        }
        if (!cancelled) {
          setUser(meResult.data);
        }

        const usersResponse = await fetch(`${apiBaseUrl}/users?limit=100`, {
          credentials: "include",
        });
        const usersResult = (await usersResponse.json().catch(() => null)) as
          | ApiResponse<UserListItem[]>
          | null;
        if (!cancelled && usersResponse.ok && usersResult?.success) {
          const list = usersResult.data ?? [];
          setPeople(list);
          setUsersByID(
            Object.fromEntries(list.map((person) => [person.id, person])) as Record<number, UserListItem>,
          );
        }

        await Promise.all([fetchConversations(), fetchGroups()]);
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
  }, [apiBaseUrl, fetchConversations, router]);

  useEffect(() => {
    const missingUserIDs = new Set<number>();
    conversations.forEach((conv) => {
      if (conv.other_user_id && !usersByID[conv.other_user_id]) {
        missingUserIDs.add(conv.other_user_id);
      }
      if (conv.group_id && !groupsByID[conv.group_id]) {
        // handled below
      }
    });
    messagesNewestFirst.forEach((msg) => {
      if (msg.sender_id && !usersByID[msg.sender_id]) {
        missingUserIDs.add(msg.sender_id);
      }
    });
    Object.values(typingByConversation).forEach((ids) => {
      ids.forEach((id) => {
        if (!usersByID[id]) missingUserIDs.add(id);
      });
    });

    const missingGroupIDs = new Set<number>();
    conversations.forEach((conv) => {
      if (conv.group_id && !groupsByID[conv.group_id]) {
        missingGroupIDs.add(conv.group_id);
      }
    });

    if (missingUserIDs.size > 0) {
      void hydrateMissingUsers(Array.from(missingUserIDs));
    }
    if (missingGroupIDs.size > 0) {
      void hydrateMissingGroups(Array.from(missingGroupIDs));
    }
  }, [conversations, groupsByID, hydrateMissingGroups, hydrateMissingUsers, messagesNewestFirst, typingByConversation, usersByID]);

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
              void fetchReactionsForMessage(payload.id);
              void fetch(`${apiBaseUrl}/conversations/${payload.conversation_id}/read`, {
                method: "PATCH",
                credentials: "include",
              }).catch(() => undefined);
            }
            void fetchConversations();
          } else if (msg.type === "typing") {
            const payload = msg.payload as TypingPayload;
            if (!payload?.conversation_id || !payload?.user_id || payload.user_id === user.id) {
              return;
            }
            setTypingByConversation((prev) => {
              const current = prev[payload.conversation_id] ?? [];
              if (payload.is_typing) {
                if (current.includes(payload.user_id)) return prev;
                return {
                  ...prev,
                  [payload.conversation_id]: [...current, payload.user_id],
                };
              }
              return {
                ...prev,
                [payload.conversation_id]: current.filter((id) => id !== payload.user_id),
              };
            });
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

  const sendMessageToActiveConversation = () => {
    const ws = wsRef.current;
    const content = chatDraft.trim();
    if (!ws || ws.readyState !== WebSocket.OPEN) {
      setChatError("Chat is not connected.");
      return;
    }
    if (!content) {
      setChatError("Write a message before sending.");
      return;
    }
    const activeConversation = conversations.find((item) => item.id === activeConversationID);
    if (!activeConversation) {
      setChatError("Select a conversation first.");
      return;
    }

    if (activeConversation.type === "direct") {
      const recipientID = Number(activeConversation.other_user_id ?? 0);
      if (!recipientID) {
        setChatError("Recipient is missing for this conversation.");
        return;
      }
      ws.send(
        JSON.stringify({
          type: "chat_message",
          payload: {
            recipient_id: recipientID,
            content,
          },
        }),
      );
    } else {
      const groupID = Number(activeConversation.group_id ?? 0);
      if (!groupID) {
        setChatError("Group is missing for this conversation.");
        return;
      }
      ws.send(
        JSON.stringify({
          type: "chat_message",
          payload: {
            group_id: groupID,
            content,
          },
        }),
      );
    }

    stopTyping();
    setChatDraft("");
    setChatError(null);
  };

  const sendDirectMessageByRecipientID = () => {
    const ws = wsRef.current;
    const content = quickDirectDraft.trim();
    if (!ws || ws.readyState !== WebSocket.OPEN) {
      setChatError("Chat is not connected.");
      return;
    }
    if (!selectedDirectUser || !content) {
      setChatError("Pick a recipient and write a message.");
      return;
    }

    ws.send(
      JSON.stringify({
        type: "chat_message",
        payload: {
          recipient_id: selectedDirectUser.id,
          content,
        },
      }),
    );
    setQuickDirectDraft("");
    setSelectedDirectUser(null);
    setDirectQuery("");
    setChatError(null);
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
    const response = await fetch(`${apiBaseUrl}/messages/${messageID}/reactions`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      credentials: "include",
      body: JSON.stringify({ emoji }),
    });
    const result = (await response.json().catch(() => null)) as ApiResponse<unknown> | null;
    if (!response.ok || !result?.success) {
      setChatError(result?.error || "Could not update reaction.");
      return;
    }
    await fetchReactionsForMessage(messageID);
  };

  const activeConversation = conversations.find((item) => item.id === activeConversationID) ?? null;
  const activeConversationTitle = activeConversation
    ? formatChatTitle(activeConversation, usersByID, groupsByID)
    : "Conversation";
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
    <div className="min-h-screen bg-neutral-50 text-neutral-900">
      <TopNav user={user ?? undefined} onLogout={() => router.replace("/login")} />

      <main className="mx-auto grid w-full max-w-6xl gap-6 px-4 py-6 sm:px-6 lg:grid-cols-[240px_minmax(0,1fr)_280px]">
        <aside className="hidden lg:block">
          <LeftNav user={user ?? undefined} activeHref="/messages" />
        </aside>

        <section className="space-y-5">
          <motion.div
            initial="hidden"
            whileInView="show"
            viewport={viewportOnce}
            variants={fadeUp}
            className="rounded-3xl border border-neutral-200 bg-white p-4 shadow-sm sm:p-5"
          >
            <div className="flex flex-wrap items-start justify-between gap-4">
              <div>
                <h1 className="text-xl font-semibold tracking-tight text-neutral-900">Messages</h1>
                <p className="text-sm text-neutral-600">
                  Realtime direct and group chats in one place.
                </p>
              </div>
              <span
                className={`inline-flex items-center gap-1 rounded-full px-3 py-1 text-xs font-semibold ${
                  chatConnected
                    ? "bg-emerald-100 text-emerald-800"
                    : "bg-rose-100 text-rose-700"
                }`}
              >
                {chatConnected ? <Wifi className="h-3.5 w-3.5" /> : <WifiOff className="h-3.5 w-3.5" />}
                {chatConnected ? "Connected" : "Offline"}
              </span>
            </div>
          </motion.div>

          {isLoading ? (
            <article className="rounded-3xl border border-neutral-200 bg-white p-6 text-sm text-neutral-600 shadow-sm">
              Loading conversations...
            </article>
          ) : error ? (
            <article className="rounded-3xl border border-rose-200 bg-rose-50 p-6 text-sm text-rose-700">
              {error}
            </article>
          ) : (
            <div className="grid gap-4 lg:grid-cols-[260px_minmax(0,1fr)]">
              <aside className="rounded-3xl border border-neutral-200 bg-white p-3 shadow-sm">
                <h2 className="px-2 text-sm font-semibold text-neutral-900">Conversations</h2>
                <div className="mt-3 space-y-2">
                  {conversations.length === 0 ? (
                    <p className="px-2 text-xs text-neutral-500">No conversations yet.</p>
                  ) : (
                    conversations.map((conversation) => {
                      const active = conversation.id === activeConversationID;
                      const title = formatChatTitle(conversation, usersByID, groupsByID);
                      const isDirect = conversation.type === "direct" && conversation.other_user_id;
                      const user = isDirect ? usersByID[conversation.other_user_id ?? 0] : null;
                      return (
                        <button
                          key={conversation.id}
                          type="button"
                          onClick={() => setActiveConversationID(conversation.id)}
                          className={`w-full rounded-2xl border px-3 py-2 text-left transition ${
                            active
                              ? "border-neutral-900 bg-neutral-900 text-white"
                              : "border-neutral-200 bg-neutral-50 text-neutral-700 hover:border-neutral-400"
                          }`}
                        >
                          <div className="flex items-center gap-2">
                            {isDirect && user?.avatar_path ? (
                              <div className="h-8 w-8 overflow-hidden rounded-full border border-neutral-200 bg-white">
                                <img
                                  src={toMediaUrl(apiBaseUrl, user.avatar_path)}
                                  alt={title}
                                  className="h-full w-full object-contain"
                                />
                              </div>
                            ) : (
                              <div className="inline-flex h-8 w-8 items-center justify-center rounded-full bg-neutral-900 text-[10px] font-semibold text-white">
                                {initials(
                                  isDirect ? user?.first_name : title,
                                  isDirect ? user?.last_name : undefined,
                                )}
                              </div>
                            )}
                            <div className="min-w-0">
                              <p className="text-xs font-semibold truncate">{title}</p>
                              <p className={`mt-1 text-[11px] ${active ? "text-neutral-200" : "text-neutral-500"} truncate`}>
                                {conversation.last_message?.content || "(no message yet)"}
                              </p>
                            </div>
                          </div>
                        </button>
                      );
                    })
                  )}
                </div>
              </aside>

              <article className="rounded-3xl border border-neutral-200 bg-white p-4 shadow-sm">
                <div className="flex flex-wrap items-center justify-between gap-3 border-b border-neutral-200 pb-3">
                  <div>
                    <h2 className="text-sm font-semibold text-neutral-900">
                      {activeConversation ? activeConversationTitle : "Select a conversation"}
                    </h2>
                    <p className="text-xs text-neutral-500">
                      {activeConversation ? `Type: ${activeConversation.type}` : "Pick one on the left or start direct chat."}
                    </p>
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
                  className="mt-4 max-h-[420px] space-y-3 overflow-y-auto pr-1"
                >
                  {hasMoreMessages && !isMessagesLoading ? (
                    <div className="flex justify-center">
                      <button
                        type="button"
                        onClick={() => void handleLoadOlder()}
                        disabled={isLoadingOlder}
                        className="rounded-full border border-neutral-200 bg-neutral-50 px-3 py-1 text-[11px] font-semibold text-neutral-600 transition hover:border-neutral-400 hover:text-neutral-900 disabled:opacity-60"
                      >
                        {isLoadingOlder ? "Loading..." : "Load older messages"}
                      </button>
                    </div>
                  ) : null}

                  {isMessagesLoading ? (
                    <p className="text-sm text-neutral-500">Loading messages...</p>
                  ) : messagesError ? (
                    <p className="rounded-2xl border border-rose-200 bg-rose-50 px-3 py-2 text-xs text-rose-700">
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
                          className={`max-w-[82%] rounded-2xl px-3 py-2 text-sm ${
                            mine
                              ? "ml-auto bg-neutral-900 text-white"
                              : "bg-neutral-100 text-neutral-800"
                          }`}
                        >
                          <div className="flex items-center gap-2">
                            {sender?.avatar_path ? (
                              <div className="h-7 w-7 overflow-hidden rounded-full border border-neutral-200 bg-white">
                                <img
                                  src={toMediaUrl(apiBaseUrl, sender.avatar_path)}
                                  alt={senderName}
                                  className="h-full w-full object-contain"
                                />
                              </div>
                            ) : (
                              <div className="inline-flex h-7 w-7 items-center justify-center rounded-full bg-neutral-900 text-[10px] font-semibold text-white">
                                {initials(sender?.first_name, sender?.last_name)}
                              </div>
                            )}
                            <p className={`text-[11px] font-semibold ${mine ? "text-neutral-200" : "text-neutral-600"}`}>
                              {senderName}
                            </p>
                          </div>
                          <p>{message.content || "(media only)"}</p>
                          <p className={`mt-1 text-[10px] ${mine ? "text-neutral-300" : "text-neutral-500"}`}>
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
                                    ? "border-emerald-300 bg-emerald-100 text-emerald-900"
                                    : "border-neutral-300 bg-white text-neutral-700"
                                }`}
                              >
                                {emoji} {count}
                              </button>
                            ))}
                            {quickReactions.map((emoji) => (
                              <button
                                key={`${message.id}-quick-${emoji}`}
                                type="button"
                                onClick={() => void toggleMessageReaction(message.id, emoji)}
                                className="rounded-full border border-neutral-300 bg-white px-2 py-0.5 text-[11px] text-neutral-700"
                              >
                                {emoji}
                              </button>
                            ))}
                          </div>
                        </div>
                      );
                    })
                  )}
                </div>

                <div className="mt-4 rounded-2xl border border-neutral-200 bg-neutral-50 p-3">
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
                    className="w-full resize-none rounded-2xl border border-neutral-200 bg-white px-3 py-2 text-sm outline-none transition focus:border-neutral-400"
                  />
                  <div className="mt-3 flex flex-wrap items-center gap-2">
                    <button
                      type="button"
                      onClick={sendMessageToActiveConversation}
                      className="brand-gradient inline-flex items-center gap-2 rounded-full px-4 py-2 text-xs font-semibold text-white transition hover:-translate-y-0.5 hover:shadow-md"
                    >
                      <Send className="h-3.5 w-3.5" />
                      Send to current chat
                    </button>
                  </div>
                  {chatError ? <p className="mt-2 text-xs text-rose-600">{chatError}</p> : null}
                </div>
              </article>
            </div>
          )}
        </section>

        <aside className="hidden space-y-5 md:block">
          <div className="rounded-3xl border border-neutral-200 bg-white p-5 shadow-sm">
            <h2 className="text-sm font-semibold text-neutral-900">Start direct chat</h2>
            <div className="mt-3 space-y-2">
              <input
                value={directQuery}
                onChange={(event) => {
                  setDirectQuery(event.target.value);
                  if (selectedDirectUser) {
                    setSelectedDirectUser(null);
                  }
                }}
                placeholder="Search a user..."
                className="h-10 w-full rounded-2xl border border-neutral-200 bg-neutral-50 px-3 text-sm outline-none focus:border-neutral-400"
              />
              {directQuery.trim() ? (
                <div className="rounded-2xl border border-neutral-200 bg-white p-2 shadow-sm">
                  {people
                    .filter((item) => item.id !== user?.id)
                    .filter((item) =>
                      `${item.first_name} ${item.last_name} ${item.nickname ?? ""}`
                        .toLowerCase()
                        .includes(directQuery.trim().toLowerCase()),
                    )
                    .slice(0, 6)
                    .map((person) => (
                      <button
                        type="button"
                        key={person.id}
                        onClick={() => {
                          setSelectedDirectUser(person);
                          setDirectQuery("");
                        }}
                        className="flex w-full items-center justify-between rounded-2xl border border-neutral-200 bg-neutral-50 px-3 py-2 text-left transition hover:border-neutral-400"
                      >
                        <div>
                          <p className="text-xs font-semibold text-neutral-800">
                            {person.first_name} {person.last_name}
                          </p>
                          <p className="text-[11px] text-neutral-500">
                            @{person.nickname || "user"}
                          </p>
                        </div>
                      </button>
                    ))}
                </div>
              ) : null}
              {selectedDirectUser ? (
                <div className="flex items-center justify-between rounded-2xl border border-neutral-200 bg-neutral-50 px-3 py-2 text-xs text-neutral-700">
                  <span className="font-semibold">
                    {selectedDirectUser.first_name} {selectedDirectUser.last_name}
                  </span>
                  <button
                    type="button"
                    onClick={() => setSelectedDirectUser(null)}
                    className="rounded-full border border-neutral-200 bg-white px-2 py-1 text-[10px] font-semibold text-neutral-600"
                  >
                    Clear
                  </button>
                </div>
              ) : null}
              <textarea
                value={quickDirectDraft}
                onChange={(event) => setQuickDirectDraft(event.target.value)}
                rows={2}
                placeholder="Quick direct message"
                className="w-full resize-none rounded-2xl border border-neutral-200 bg-neutral-50 px-3 py-2 text-sm outline-none focus:border-neutral-400"
              />
              <button
                type="button"
                onClick={sendDirectMessageByRecipientID}
                className="brand-gradient inline-flex items-center gap-2 rounded-full px-4 py-2 text-xs font-semibold text-white transition hover:-translate-y-0.5 hover:shadow-md"
              >
                <Send className="h-3.5 w-3.5" />
                Send direct
              </button>
            </div>
          </div>

          <div className="rounded-3xl border border-neutral-200 bg-white p-5 shadow-sm">
            <h2 className="text-sm font-semibold text-neutral-900">People</h2>
            <div className="mt-3 space-y-2">
              {people.filter((item) => item.id !== user?.id).slice(0, 8).map((person) => (
                <button
                  type="button"
                  key={person.id}
                  onClick={() => setSelectedDirectUser(person)}
                  className="flex w-full items-center justify-between rounded-2xl border border-neutral-200 bg-neutral-50 px-3 py-2 text-left transition hover:border-neutral-400"
                >
                  <div className="flex items-center gap-3">
                    {person.avatar_path ? (
                      <div className="h-8 w-8 overflow-hidden rounded-full border border-neutral-200 bg-white">
                        <img
                          src={toMediaUrl(apiBaseUrl, person.avatar_path)}
                          alt={`${person.first_name} ${person.last_name}`}
                          className="h-full w-full object-contain"
                        />
                      </div>
                    ) : (
                      <div className="inline-flex h-8 w-8 items-center justify-center rounded-full bg-neutral-900 text-[10px] font-semibold text-white">
                        {initials(person.first_name, person.last_name)}
                      </div>
                    )}
                    <div>
                    <p className="text-xs font-semibold text-neutral-800">
                      {person.first_name} {person.last_name}
                    </p>
                    <p className="text-[11px] text-neutral-500">@{person.nickname || "user"}</p>
                    </div>
                  </div>
                </button>
              ))}
            </div>
          </div>
        </aside>
      </main>
    </div>
  );
}
