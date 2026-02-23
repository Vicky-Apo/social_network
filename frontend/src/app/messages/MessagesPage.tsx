"use client";

import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import Image from "next/image";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { Compass, LogOut, MessageSquare, RefreshCw, Send, Users, Wifi, WifiOff } from "lucide-react";
import { motion } from "framer-motion";
import { useAuth } from "../component/AuthContext";
import { landingData } from "@/lib/data";
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
};

type UserListItem = {
  id: number;
  first_name: string;
  last_name: string;
  nickname?: string | null;
  avatar_path?: string | null;
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

const quickLinks = [
  { label: "Explore", href: "/dashboard", icon: Compass },
  { label: "Groups", href: "/groups", icon: Users },
  { label: "Messages", href: "/messages", icon: MessageSquare },
];

const quickReactions = ["👍", "❤️", "😂", "🔥"];
const pageSize = 30;

function initials(first?: string, last?: string) {
  const left = first?.trim().charAt(0) ?? "";
  const right = last?.trim().charAt(0) ?? "";
  return `${left}${right}`.toUpperCase() || "U";
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
) {
  if (conversation.type === "direct" && conversation.other_user_id) {
    const person = usersByID[conversation.other_user_id];
    if (person) {
      return `${person.first_name} ${person.last_name}`;
    }
    return `User #${conversation.other_user_id}`;
  }
  if (conversation.group_id) {
    return `Group #${conversation.group_id}`;
  }
  return `Conversation #${conversation.id}`;
}

export default function MessagesPage() {
  const router = useRouter();
  const { logout } = useAuth();
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
  const [recipientIDInput, setRecipientIDInput] = useState("");
  const [people, setPeople] = useState<UserListItem[]>([]);
  const [usersByID, setUsersByID] = useState<Record<number, UserListItem>>({});
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
  }, [apiBaseUrl, fetchConversations, router]);

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

  const handleLogout = async () => {
    try {
      await fetch(`${apiBaseUrl}/auth/logout`, {
        method: "POST",
        credentials: "include",
      });
    } finally {
      logout();
      router.replace("/login");
    }
  };

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
    const recipientID = Number(recipientIDInput);
    const content = quickDirectDraft.trim();
    if (!ws || ws.readyState !== WebSocket.OPEN) {
      setChatError("Chat is not connected.");
      return;
    }
    if (!recipientID || !content) {
      setChatError("Enter recipient ID and message.");
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
    setQuickDirectDraft("");
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
  const displayName = user ? `${user.first_name} ${user.last_name}` : "Loading";
  const userTag =
    user?.nickname || (user?.email ? user.email.split("@")[0] : "community-member");
  const messagesOldestFirst = useMemo(
    () => [...messagesNewestFirst].reverse(),
    [messagesNewestFirst],
  );
  const typingUserIDs = activeConversationID ? typingByConversation[activeConversationID] ?? [] : [];
  const typingLabel = typingUserIDs
    .map((id) => {
      const person = usersByID[id];
      return person ? `${person.first_name}` : `User #${id}`;
    })
    .join(", ");

  return (
    <div className="min-h-screen bg-neutral-50 text-neutral-900">
      <header className="sticky top-0 z-40 border-b border-neutral-200/80 bg-white/85 backdrop-blur-md">
        <div className="mx-auto flex w-full max-w-6xl items-center gap-3 px-4 py-3 sm:px-6">
          <Link href="/" className="inline-flex items-center gap-2">
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

          <button
            type="button"
            onClick={() => void fetchConversations()}
            className="ml-auto inline-flex items-center gap-2 rounded-full border border-neutral-200 bg-white px-3 py-2 text-xs font-semibold text-neutral-700 transition hover:border-neutral-300 hover:text-neutral-900"
          >
            <RefreshCw className="h-3.5 w-3.5" />
            <span className="hidden sm:inline">Refresh</span>
          </button>

          <button
            type="button"
            onClick={handleLogout}
            className="inline-flex items-center gap-2 rounded-full border border-neutral-200 bg-white px-3 py-2 text-xs font-semibold text-neutral-700 transition hover:border-neutral-300 hover:text-neutral-900"
          >
            <LogOut className="h-3.5 w-3.5" />
            <span className="hidden sm:inline">Log out</span>
          </button>
        </div>
      </header>

      <main className="mx-auto grid w-full max-w-6xl gap-6 px-4 py-6 sm:px-6 lg:grid-cols-[240px_minmax(0,1fr)_280px]">
        <aside className="hidden lg:block">
          <div className="rounded-3xl border border-neutral-200 bg-white p-5 shadow-sm">
            <div className="flex items-center gap-3">
              <div className="inline-flex h-11 w-11 items-center justify-center rounded-full bg-neutral-900 text-sm font-semibold text-white">
                {initials(user?.first_name, user?.last_name)}
              </div>
              <div>
                <p className="text-sm font-semibold text-neutral-900">{displayName}</p>
                <p className="text-xs text-neutral-500">@{userTag}</p>
              </div>
            </div>
            <nav className="mt-5 space-y-2">
              {quickLinks.map((item) => {
                const Icon = item.icon;
                const isActive = item.href === "/messages";
                return (
                  <Link
                    key={item.label}
                    href={item.href}
                    className={`flex items-center gap-2 rounded-2xl border px-3 py-2 text-sm transition ${
                      isActive
                        ? "brand-gradient border-transparent text-white"
                        : "border-neutral-200 bg-neutral-50 text-neutral-700 hover:border-neutral-300 hover:text-neutral-900"
                    }`}
                  >
                    <Icon className="h-4 w-4" />
                    <span>{item.label}</span>
                  </Link>
                );
              })}
            </nav>
          </div>
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
                      return (
                        <button
                          key={conversation.id}
                          type="button"
                          onClick={() => setActiveConversationID(conversation.id)}
                          className={`w-full rounded-2xl border px-3 py-2 text-left transition ${
                            active
                              ? "border-neutral-900 bg-neutral-900 text-white"
                              : "border-neutral-200 bg-neutral-50 text-neutral-700 hover:border-neutral-300"
                          }`}
                        >
                          <p className="text-xs font-semibold">
                            {formatChatTitle(conversation, usersByID)}
                          </p>
                          <p className={`mt-1 text-[11px] ${active ? "text-neutral-200" : "text-neutral-500"}`}>
                            {conversation.last_message?.content || "(no message yet)"}
                          </p>
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
                      {activeConversation ? formatChatTitle(activeConversation, usersByID) : "Select a conversation"}
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
                        className="rounded-full border border-neutral-200 bg-neutral-50 px-3 py-1 text-[11px] font-semibold text-neutral-600 transition hover:border-neutral-300 hover:text-neutral-900 disabled:opacity-60"
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
                          <p>{message.content || "(media only)"}</p>
                          <p className={`mt-1 text-[10px] ${mine ? "text-neutral-300" : "text-neutral-500"}`}>
                            User {message.sender_id} • {shortDate(message.created_at)}
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
                type="number"
                value={recipientIDInput}
                onChange={(event) => setRecipientIDInput(event.target.value)}
                placeholder="Recipient user ID"
                className="h-10 w-full rounded-2xl border border-neutral-200 bg-neutral-50 px-3 text-sm outline-none focus:border-neutral-400"
              />
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
                  onClick={() => setRecipientIDInput(String(person.id))}
                  className="flex w-full items-center justify-between rounded-2xl border border-neutral-200 bg-neutral-50 px-3 py-2 text-left transition hover:border-neutral-300"
                >
                  <div>
                    <p className="text-xs font-semibold text-neutral-800">
                      {person.first_name} {person.last_name}
                    </p>
                    <p className="text-[11px] text-neutral-500">@{person.nickname || `user-${person.id}`}</p>
                  </div>
                  <span className="rounded-full bg-white px-2 py-1 text-[10px] text-neutral-500">
                    ID {person.id}
                  </span>
                </button>
              ))}
            </div>
          </div>
        </aside>
      </main>
    </div>
  );
}
