"use client";
/* eslint-disable @next/next/no-img-element */

import { useEffect, useMemo, useRef, useState } from "react";
import Link from "next/link";
import Image from "next/image";
import { useRouter } from "next/navigation";
import {
  Bell,
  Compass,
  MessageCircle,
  LogOut,
  MessageSquare,
  ThumbsDown,
  ThumbsUp,
  Plus,
  Search,
  Send,
  Users,
  Wifi,
  WifiOff,
} from "lucide-react";
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

type Post = {
  id: number;
  author_id: number;
  author_first_name: string;
  author_last_name: string;
  content: string;
  media_path?: string | null;
  privacy: string;
  created_at: string;
  comment_count: number;
  like_count: number;
  dislike_count: number;
};

type Comment = {
  id: number;
  post_id: number;
  author_id: number;
  content: string;
  media_path?: string;
  like_count: number;
  dislike_count: number;
  created_at: string;
};

type Reaction = {
  user_id: number;
  reaction: "like" | "dislike";
};

type ReactionKind = "like" | "dislike";
type ReactionMap = Record<number, ReactionKind | null>;
type WsMessage = {
  type: string;
  payload: unknown;
};

type NotificationItem = {
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

type ChatMessage = {
  id: number;
  conversation_id: number;
  sender_id: number;
  content?: string;
  media_path?: string;
  created_at: string;
};

type UserListItem = {
  id: number;
  first_name: string;
  last_name: string;
  nickname?: string | null;
  avatar_path?: string | null;
};

const quickLinks = [
  { label: "Explore", href: "#", icon: Compass },
  { label: "Groups", href: "/groups", icon: Users },
  { label: "Messages", href: "/messages", icon: MessageSquare },
];

const trends = [
  { title: "Product Feedback", posts: "1.2k posts this week" },
  { title: "Community Showcase", posts: "840 posts this week" },
  { title: "Growth Ideas", posts: "512 posts this week" },
];

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

export default function DashboardPage() {
  const router = useRouter();
  const { logout } = useAuth();

  const [user, setUser] = useState<User | null>(null);
  const [posts, setPosts] = useState<Post[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [feedError, setFeedError] = useState<string | null>(null);
  const [notificationCount, setNotificationCount] = useState(0);
  const [composerText, setComposerText] = useState("");
  const [composerError, setComposerError] = useState<string | null>(null);
  const [isPosting, setIsPosting] = useState(false);
  const [groupsOnly, setGroupsOnly] = useState(false);
  const [postReactionMap, setPostReactionMap] = useState<ReactionMap>({});
  const [commentReactionMap, setCommentReactionMap] = useState<ReactionMap>({});
  const [commentsByPost, setCommentsByPost] = useState<Record<number, Comment[]>>({});
  const [commentsOpenByPost, setCommentsOpenByPost] = useState<Record<number, boolean>>({});
  const [commentsLoadingByPost, setCommentsLoadingByPost] = useState<Record<number, boolean>>({});
  const [commentDraftByPost, setCommentDraftByPost] = useState<Record<number, string>>({});
  const [commentErrorByPost, setCommentErrorByPost] = useState<Record<number, string>>({});
  const [notifications, setNotifications] = useState<NotificationItem[]>([]);
  const [notificationsOpen, setNotificationsOpen] = useState(false);
  const [notificationsLoading, setNotificationsLoading] = useState(false);
  const [chatRecipientID, setChatRecipientID] = useState("");
  const [chatDraft, setChatDraft] = useState("");
  const [chatMessages, setChatMessages] = useState<ChatMessage[]>([]);
  const [chatError, setChatError] = useState<string | null>(null);
  const [chatConnected, setChatConnected] = useState(false);
  const [chatUnreadMap, setChatUnreadMap] = useState<Record<number, number>>({});
  const [searchQuery, setSearchQuery] = useState("");
  const [people, setPeople] = useState<UserListItem[]>([]);
  const [peopleLoading, setPeopleLoading] = useState(false);
  const wsRef = useRef<WebSocket | null>(null);

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

  useEffect(() => {
    let cancelled = false;

    const fetchJson = async <T,>(path: string) => {
      const response = await fetch(`${apiBaseUrl}${path}`, {
        credentials: "include",
      });
      const result = (await response.json().catch(() => null)) as ApiResponse<T> | null;
      return { response, result };
    };

    const load = async () => {
      setIsLoading(true);
      setFeedError(null);

      try {
        const me = await fetchJson<User>("/auth/me");
        if (!me.response.ok || !me.result?.success) {
          if (!cancelled) {
            router.replace("/login");
          }
          return;
        }

        if (!cancelled) {
          setUser(me.result.data ?? null);
        }

        const unread = await fetchJson<{ count: number }>("/notifications/unread-count");
        if (!cancelled && unread.response.ok && unread.result?.success) {
          setNotificationCount(Number(unread.result.data?.count ?? 0));
        }

        const notificationsRes = await fetchJson<NotificationItem[]>("/notifications?limit=10");
        if (!cancelled && notificationsRes.response.ok && notificationsRes.result?.success) {
          setNotifications(notificationsRes.result.data ?? []);
        }

        const feedPath = groupsOnly ? "/posts?groups_only=true" : "/posts";
        const feed = await fetchJson<Post[]>(feedPath);
        if (!feed.response.ok || !feed.result?.success) {
          if (!cancelled) {
            setFeedError(feed.result?.error || "Unable to load your feed.");
            setPosts([]);
          }
          return;
        }

        if (!cancelled) {
          const nextPosts = feed.result.data ?? [];
          setPosts(nextPosts);

          const currentUserID = me.result.data?.id;
          if (currentUserID) {
            void Promise.all(
              nextPosts.map(async (post) => {
                try {
                  const reactionRes = await fetchJson<Reaction[]>(
                    `/posts/${post.id}/reactions`,
                  );
                  if (!reactionRes.response.ok || !reactionRes.result?.success) {
                    return [post.id, null] as const;
                  }
                  const mine = (reactionRes.result.data ?? []).find(
                    (item) => item.user_id === currentUserID,
                  );
                  return [post.id, mine?.reaction ?? null] as const;
                } catch {
                  return [post.id, null] as const;
                }
              }),
            ).then((entries) => {
              if (!cancelled) {
                setPostReactionMap(Object.fromEntries(entries));
              }
            });
          }
        }
      } catch {
        if (!cancelled) {
          setFeedError("Network error. Please try again.");
        }
      } finally {
        if (!cancelled) {
          setIsLoading(false);
        }
      }
    };

    load();

    return () => {
      cancelled = true;
    };
  }, [apiBaseUrl, router, groupsOnly]);

  useEffect(() => {
    if (!user?.id) {
      return;
    }

    let cancelled = false;
    const run = async () => {
      setPeopleLoading(true);
      try {
        const query = searchQuery.trim();
        const path = query ? `/users?q=${encodeURIComponent(query)}` : "/users";
        const response = await fetch(`${apiBaseUrl}${path}`, {
          credentials: "include",
        });
        const result = (await response.json().catch(() => null)) as
          | ApiResponse<UserListItem[]>
          | null;
        if (!cancelled && response.ok && result?.success) {
          setPeople(result.data ?? []);
        }
      } finally {
        if (!cancelled) {
          setPeopleLoading(false);
        }
      }
    };

    const timeoutID = window.setTimeout(run, 220);
    return () => {
      cancelled = true;
      window.clearTimeout(timeoutID);
    };
  }, [apiBaseUrl, searchQuery, user?.id]);

  useEffect(() => {
    if (!user?.id) {
      return;
    }

    const intervalID = window.setInterval(async () => {
      const feedPath = groupsOnly ? "/posts?groups_only=true" : "/posts";
      const response = await fetch(`${apiBaseUrl}${feedPath}`, {
        credentials: "include",
      }).catch(() => null);
      if (!response?.ok) {
        return;
      }
      const result = (await response.json().catch(() => null)) as ApiResponse<Post[]> | null;
      if (result?.success) {
        setPosts(result.data ?? []);
      }
    }, 7000);

    return () => {
      window.clearInterval(intervalID);
    };
  }, [apiBaseUrl, groupsOnly, user?.id]);

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
            const payload = msg.payload as ChatMessage;
            setChatMessages((prev) => [payload, ...prev].slice(0, 50));
          } else if (msg.type === "notification") {
            const payload = msg.payload as NotificationItem;
            setNotifications((prev) => [payload, ...prev].slice(0, 20));
            setNotificationCount((prev) => prev + 1);
          } else if (msg.type === "unread_counts") {
            const payload = msg.payload as Array<{ conversation_id: number; unread_count: number }>;
            setChatUnreadMap((prev) => {
              const next = { ...prev };
              payload.forEach((item) => {
                next[item.conversation_id] = item.unread_count;
              });
              return next;
            });
          } else if (msg.type === "error") {
            const payload = msg.payload as { message?: string };
            setChatError(payload.message || "Chat error.");
          }
        } catch {
          // Ignore malformed websocket payload chunks.
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
  }, [user?.id, wsBaseUrl]);

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

  const handleCreatePost = async () => {
    if (isPosting) return;
    const content = composerText.trim();
    if (!content) {
      setComposerError("Write something before posting.");
      return;
    }

    setIsPosting(true);
    setComposerError(null);

    try {
      const response = await fetch(`${apiBaseUrl}/posts`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        credentials: "include",
        body: JSON.stringify({
          content,
          privacy: "public",
        }),
      });

      const result = (await response.json().catch(() => null)) as ApiResponse<Post> | null;
      if (!response.ok || !result?.success || !result.data) {
        setComposerError(result?.error || "Could not publish your post.");
        return;
      }

      setPosts((prev) => [result.data as Post, ...prev]);
      setComposerText("");
    } catch {
      setComposerError("Network error. Please try again.");
    } finally {
      setIsPosting(false);
    }
  };

  const refreshNotifications = async () => {
    setNotificationsLoading(true);
    try {
      const response = await fetch(`${apiBaseUrl}/notifications?limit=20`, {
        credentials: "include",
      });
      const result = (await response.json().catch(() => null)) as
        | ApiResponse<NotificationItem[]>
        | null;
      if (response.ok && result?.success) {
        setNotifications(result.data ?? []);
      }
    } finally {
      setNotificationsLoading(false);
    }
  };

  const markNotificationRead = async (id: number) => {
    const old = notifications;
    setNotifications((prev) =>
      prev.map((item) => (item.id === id ? { ...item, is_read: true } : item)),
    );
    setNotificationCount((prev) => Math.max(0, prev - 1));

    try {
      const response = await fetch(`${apiBaseUrl}/notifications/${id}/read`, {
        method: "PATCH",
        credentials: "include",
      });
      if (!response.ok) {
        setNotifications(old);
      }
    } catch {
      setNotifications(old);
    }
  };

  const markAllNotificationsRead = async () => {
    setNotifications((prev) => prev.map((item) => ({ ...item, is_read: true })));
    setNotificationCount(0);
    await fetch(`${apiBaseUrl}/notifications/read-all`, {
      method: "PATCH",
      credentials: "include",
    }).catch(() => undefined);
  };

  const sendChatMessage = () => {
    const ws = wsRef.current;
    const recipient = Number(chatRecipientID);
    const content = chatDraft.trim();
    if (!ws || ws.readyState !== WebSocket.OPEN) {
      setChatError("Chat is not connected.");
      return;
    }
    if (!recipient || !content) {
      setChatError("Enter recipient ID and message.");
      return;
    }

    const payload = {
      type: "chat_message",
      payload: {
        recipient_id: recipient,
        content,
      },
    };
    ws.send(JSON.stringify(payload));
    setChatDraft("");
  };

  const refreshFeed = async () => {
    setIsLoading(true);
    setFeedError(null);
    try {
      const feedPath = groupsOnly ? "/posts?groups_only=true" : "/posts";
      const response = await fetch(`${apiBaseUrl}${feedPath}`, {
        credentials: "include",
      });
      const result = (await response.json().catch(() => null)) as ApiResponse<Post[]> | null;
      if (!response.ok || !result?.success) {
        setFeedError(result?.error || "Could not refresh feed.");
        return;
      }
      setPosts(result.data ?? []);
    } catch {
      setFeedError("Network error. Please try again.");
    } finally {
      setIsLoading(false);
    }
  };

  const loadCommentsForPost = async (postID: number) => {
    setCommentsLoadingByPost((prev) => ({ ...prev, [postID]: true }));
    setCommentErrorByPost((prev) => ({ ...prev, [postID]: "" }));

    try {
      const response = await fetch(`${apiBaseUrl}/posts/${postID}/comments`, {
        credentials: "include",
      });
      const result = (await response.json().catch(() => null)) as ApiResponse<Comment[]> | null;

      if (!response.ok || !result?.success) {
        setCommentErrorByPost((prev) => ({
          ...prev,
          [postID]: result?.error || "Could not load comments.",
        }));
        return;
      }

      const comments = result.data ?? [];
      setCommentsByPost((prev) => ({ ...prev, [postID]: comments }));

      if (user?.id && comments.length > 0) {
        const entries = await Promise.all(
          comments.map(async (comment) => {
            const reactionRes = await fetch(
              `${apiBaseUrl}/comments/${comment.id}/reactions`,
              { credentials: "include" },
            );
            const reactionJson = (await reactionRes.json().catch(() => null)) as
              | ApiResponse<Reaction[]>
              | null;
            if (!reactionRes.ok || !reactionJson?.success) {
              return [comment.id, null] as const;
            }
            const mine = (reactionJson.data ?? []).find((item) => item.user_id === user.id);
            return [comment.id, mine?.reaction ?? null] as const;
          }),
        );
        setCommentReactionMap((prev) => ({ ...prev, ...Object.fromEntries(entries) }));
      }
    } catch {
      setCommentErrorByPost((prev) => ({
        ...prev,
        [postID]: "Network error while loading comments.",
      }));
    } finally {
      setCommentsLoadingByPost((prev) => ({ ...prev, [postID]: false }));
    }
  };

  const toggleComments = (postID: number) => {
    const isOpen = commentsOpenByPost[postID] ?? false;
    const nextOpen = !isOpen;
    setCommentsOpenByPost((prev) => ({ ...prev, [postID]: nextOpen }));
    if (nextOpen && !commentsByPost[postID]) {
      void loadCommentsForPost(postID);
    }
  };

  const handleCreateComment = async (postID: number) => {
    const draft = (commentDraftByPost[postID] ?? "").trim();
    if (!draft) {
      setCommentErrorByPost((prev) => ({
        ...prev,
        [postID]: "Write a comment before posting.",
      }));
      return;
    }

    try {
      const response = await fetch(`${apiBaseUrl}/posts/${postID}/comments`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        credentials: "include",
        body: JSON.stringify({ content: draft }),
      });
      const result = (await response.json().catch(() => null)) as ApiResponse<Comment> | null;

      if (!response.ok || !result?.success || !result.data) {
        setCommentErrorByPost((prev) => ({
          ...prev,
          [postID]: result?.error || "Could not post comment.",
        }));
        return;
      }

      setCommentsByPost((prev) => ({
        ...prev,
        [postID]: [result.data as Comment, ...(prev[postID] ?? [])],
      }));
      setCommentDraftByPost((prev) => ({ ...prev, [postID]: "" }));
      setCommentErrorByPost((prev) => ({ ...prev, [postID]: "" }));
      setPosts((prev) =>
        prev.map((post) =>
          post.id === postID ? { ...post, comment_count: post.comment_count + 1 } : post,
        ),
      );
      setCommentsOpenByPost((prev) => ({ ...prev, [postID]: true }));
    } catch {
      setCommentErrorByPost((prev) => ({
        ...prev,
        [postID]: "Network error while posting comment.",
      }));
    }
  };

  const handlePostReaction = async (postID: number, reaction: ReactionKind) => {
    const previous = postReactionMap[postID] ?? null;
    const next = previous === reaction ? null : reaction;

    setPostReactionMap((prev) => ({ ...prev, [postID]: next }));
    setPosts((prev) =>
      prev.map((post) => {
        if (post.id !== postID) return post;
        let like = post.like_count;
        let dislike = post.dislike_count;
        if (previous === "like") like -= 1;
        if (previous === "dislike") dislike -= 1;
        if (next === "like") like += 1;
        if (next === "dislike") dislike += 1;
        return {
          ...post,
          like_count: Math.max(0, like),
          dislike_count: Math.max(0, dislike),
        };
      }),
    );

    try {
      await fetch(`${apiBaseUrl}/posts/${postID}/reactions`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        credentials: "include",
        body: JSON.stringify({ reaction }),
      });
    } catch {
      setPostReactionMap((prev) => ({ ...prev, [postID]: previous }));
      setPosts((prev) =>
        prev.map((post) => {
          if (post.id !== postID) return post;
          let like = post.like_count;
          let dislike = post.dislike_count;
          if (next === "like") like -= 1;
          if (next === "dislike") dislike -= 1;
          if (previous === "like") like += 1;
          if (previous === "dislike") dislike += 1;
          return {
            ...post,
            like_count: Math.max(0, like),
            dislike_count: Math.max(0, dislike),
          };
        }),
      );
    }
  };

  const handleCommentReaction = async (
    postID: number,
    commentID: number,
    reaction: ReactionKind,
  ) => {
    const previous = commentReactionMap[commentID] ?? null;
    const next = previous === reaction ? null : reaction;

    setCommentReactionMap((prev) => ({ ...prev, [commentID]: next }));
    setCommentsByPost((prev) => ({
      ...prev,
      [postID]: (prev[postID] ?? []).map((comment) => {
        if (comment.id !== commentID) return comment;
        let like = comment.like_count;
        let dislike = comment.dislike_count;
        if (previous === "like") like -= 1;
        if (previous === "dislike") dislike -= 1;
        if (next === "like") like += 1;
        if (next === "dislike") dislike += 1;
        return {
          ...comment,
          like_count: Math.max(0, like),
          dislike_count: Math.max(0, dislike),
        };
      }),
    }));

    try {
      await fetch(`${apiBaseUrl}/comments/${commentID}/reactions`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        credentials: "include",
        body: JSON.stringify({
          reaction,
          user_id: user?.id ?? 0,
        }),
      });
    } catch {
      setCommentReactionMap((prev) => ({ ...prev, [commentID]: previous }));
    }
  };

  const displayName = user ? `${user.first_name} ${user.last_name}` : "Loading";
  const userTag =
    user?.nickname || (user?.email ? user.email.split("@")[0] : "community-member");

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

          <div className="relative ml-2 hidden flex-1 sm:block">
            <Search className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-neutral-400" />
            <input
              type="search"
              value={searchQuery}
              onChange={(event) => setSearchQuery(event.target.value)}
              placeholder="Search posts, people, topics..."
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
              {notificationCount}
            </span>
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
                return (
                  <Link
                    key={item.label}
                    href={item.href}
                    className="flex items-center gap-2 rounded-2xl border border-neutral-200 bg-neutral-50 px-3 py-2 text-sm text-neutral-700 transition hover:border-neutral-300 hover:text-neutral-900"
                  >
                    <Icon className="h-4 w-4" />
                    <span>{item.label}</span>
                  </Link>
                );
              })}
            </nav>
            <div className="mt-5">
              <p className="mb-2 text-xs font-semibold uppercase tracking-wide text-neutral-500">
                People
              </p>
              <div className="space-y-2">
                {peopleLoading ? (
                  <p className="text-xs text-neutral-500">Searching users...</p>
                ) : people.filter((person) => person.id !== user?.id).length === 0 ? (
                  <p className="text-xs text-neutral-500">No users found.</p>
                ) : (
                  people
                    .filter((person) => person.id !== user?.id)
                    .slice(0, 5)
                    .map((person) => (
                      <div
                        key={person.id}
                        className="flex items-center justify-between rounded-2xl border border-neutral-200 bg-neutral-50 px-3 py-2"
                      >
                        <div>
                          <p className="text-xs font-semibold text-neutral-800">
                            {person.first_name} {person.last_name}
                          </p>
                          <p className="text-[11px] text-neutral-500">
                            @{person.nickname || `user-${person.id}`}
                          </p>
                        </div>
                        <span className="rounded-full bg-white px-2 py-1 text-[10px] text-neutral-500">
                          ID {person.id}
                        </span>
                      </div>
                    ))
                )}
              </div>
            </div>
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
            <div className="flex items-center justify-between gap-3">
              <div>
                <h1 className="text-xl font-semibold tracking-tight text-neutral-900">Dashboard</h1>
                <p className="text-sm text-neutral-600">Create updates and follow your community feed.</p>
              </div>
              <div className="flex items-center gap-2">
                <label className="inline-flex items-center gap-2 text-xs text-neutral-600">
                  Groups only
                  <input
                    type="checkbox"
                    checked={groupsOnly}
                    onChange={(event) => setGroupsOnly(event.target.checked)}
                    className="h-4 w-4 rounded border-neutral-300 text-neutral-900 focus:ring-neutral-900"
                  />
                </label>
                <button
                  type="button"
                  onClick={refreshFeed}
                  className="rounded-full border border-neutral-200 bg-white px-3 py-1.5 text-xs font-semibold text-neutral-700 transition hover:border-neutral-300 hover:text-neutral-900"
                >
                  Refresh feed
                </button>
              </div>
            </div>
          </motion.div>

          <motion.div
            initial="hidden"
            whileInView="show"
            viewport={viewportOnce}
            variants={fadeUp}
            className="rounded-3xl border border-neutral-200 bg-white p-4 shadow-sm sm:p-5"
          >
            <textarea
              value={composerText}
              onChange={(event) => setComposerText(event.target.value)}
              rows={4}
              placeholder="Share an update with Vybez..."
              className="w-full resize-none rounded-2xl border border-neutral-200 bg-neutral-50 px-4 py-3 text-sm text-neutral-900 placeholder:text-neutral-400 outline-none transition focus:border-neutral-400"
            />
            <div className="mt-3 flex flex-wrap items-center justify-between gap-3">
              <button
                type="button"
                className="inline-flex items-center gap-2 rounded-full border border-neutral-200 bg-white px-3 py-2 text-xs font-semibold text-neutral-700"
              >
                <Plus className="h-3.5 w-3.5" />
                Add media
              </button>
              <button
                type="button"
                onClick={handleCreatePost}
                disabled={isPosting}
                className="brand-gradient inline-flex items-center gap-2 rounded-full px-4 py-2 text-xs font-semibold text-white transition hover:-translate-y-0.5 hover:shadow-md disabled:cursor-not-allowed disabled:opacity-70"
              >
                <Send className="h-3.5 w-3.5" />
                {isPosting ? "Posting..." : "Publish"}
              </button>
            </div>
            {composerError ? <p className="mt-3 text-xs text-rose-600">{composerError}</p> : null}
          </motion.div>

          <motion.div
            initial="hidden"
            whileInView="show"
            viewport={viewportOnce}
            variants={fadeUp}
            className="rounded-3xl border border-neutral-200 bg-white p-4 shadow-sm lg:hidden"
          >
            <div className="flex items-center justify-between">
              <h2 className="text-sm font-semibold text-neutral-900">Live chat</h2>
              <span
                className={`inline-flex items-center gap-1 rounded-full px-2 py-1 text-[10px] font-semibold ${
                  chatConnected
                    ? "bg-emerald-100 text-emerald-800"
                    : "bg-rose-100 text-rose-700"
                }`}
              >
                {chatConnected ? <Wifi className="h-3 w-3" /> : <WifiOff className="h-3 w-3" />}
                {chatConnected ? "Connected" : "Offline"}
              </span>
            </div>
            <div className="mt-3 flex flex-col gap-2">
              <input
                type="number"
                value={chatRecipientID}
                onChange={(event) => setChatRecipientID(event.target.value)}
                placeholder="Recipient user ID"
                className="h-9 rounded-xl border border-neutral-200 bg-neutral-50 px-3 text-xs outline-none focus:border-neutral-400"
              />
              <div className="flex gap-2">
                <input
                  value={chatDraft}
                  onChange={(event) => setChatDraft(event.target.value)}
                  placeholder="Write a direct message..."
                  className="h-9 flex-1 rounded-xl border border-neutral-200 bg-neutral-50 px-3 text-xs outline-none focus:border-neutral-400"
                />
                <button
                  type="button"
                  onClick={sendChatMessage}
                  className="brand-gradient rounded-xl px-3 text-xs font-semibold text-white"
                >
                  Send
                </button>
              </div>
            </div>
          </motion.div>

          <div className="space-y-4">
            <div className="rounded-2xl border border-neutral-200 bg-white px-4 py-2 text-xs text-neutral-600">
              Feed status: {isLoading ? "loading" : `${posts.length} post(s)`}
            </div>
            {isLoading ? (
              <article className="rounded-3xl border border-neutral-200 bg-white p-6 text-sm text-neutral-600 shadow-sm">
                Loading your feed...
              </article>
            ) : feedError ? (
              <article className="rounded-3xl border border-rose-200 bg-rose-50 p-6 text-sm text-rose-700">
                {feedError}
              </article>
            ) : posts.length === 0 ? (
              <article className="rounded-3xl border border-neutral-200 bg-white p-6 text-sm text-neutral-600 shadow-sm">
                No posts yet. Be the first to share an update.
              </article>
            ) : (
              posts.map((post) => (
                <article
                  key={post.id}
                  className="rounded-3xl border border-neutral-200 bg-white p-5 shadow-sm"
                >
                  <header className="flex items-start justify-between gap-3">
                    <div className="flex items-center gap-3">
                      <span className="inline-flex h-10 w-10 items-center justify-center rounded-full bg-neutral-900 text-xs font-semibold text-white">
                        {initials(post.author_first_name, post.author_last_name)}
                      </span>
                      <div>
                        <p className="text-sm font-semibold text-neutral-900">
                          {post.author_first_name} {post.author_last_name}
                        </p>
                        <p className="text-xs text-neutral-500">{shortDate(post.created_at)}</p>
                      </div>
                    </div>
                    <span className="rounded-full border border-neutral-200 bg-neutral-50 px-2.5 py-1 text-[11px] uppercase tracking-wide text-neutral-600">
                      {post.privacy}
                    </span>
                  </header>

                  <p className="mt-4 text-sm leading-relaxed text-neutral-700">{post.content}</p>

                  {post.media_path ? (
                    <div className="mt-4 overflow-hidden rounded-2xl border border-neutral-200">
                      <img src={post.media_path} alt="Post media" className="h-72 w-full object-cover" />
                    </div>
                  ) : null}

                  <footer className="mt-4 flex items-center gap-4 text-xs text-neutral-500">
                    <button
                      type="button"
                      onClick={() => handlePostReaction(post.id, "like")}
                      className={`inline-flex items-center gap-1 rounded-full px-2 py-1 transition ${
                        postReactionMap[post.id] === "like"
                          ? "bg-emerald-100 text-emerald-800"
                          : "bg-neutral-100 text-neutral-600 hover:bg-neutral-200"
                      }`}
                    >
                      <ThumbsUp className="h-3.5 w-3.5" />
                      {post.like_count}
                    </button>
                    <button
                      type="button"
                      onClick={() => handlePostReaction(post.id, "dislike")}
                      className={`inline-flex items-center gap-1 rounded-full px-2 py-1 transition ${
                        postReactionMap[post.id] === "dislike"
                          ? "bg-rose-100 text-rose-800"
                          : "bg-neutral-100 text-neutral-600 hover:bg-neutral-200"
                      }`}
                    >
                      <ThumbsDown className="h-3.5 w-3.5" />
                      {post.dislike_count}
                    </button>
                    <button
                      type="button"
                      onClick={() => toggleComments(post.id)}
                      className="inline-flex items-center gap-1 rounded-full bg-neutral-100 px-2 py-1 text-neutral-600 transition hover:bg-neutral-200"
                    >
                      <MessageCircle className="h-3.5 w-3.5" />
                      {post.comment_count}
                    </button>
                  </footer>

                  {commentsOpenByPost[post.id] ? (
                    <section className="mt-4 rounded-2xl border border-neutral-200 bg-neutral-50 p-3">
                      <div className="space-y-2">
                        {(commentsByPost[post.id] ?? []).map((comment) => (
                          <article key={comment.id} className="rounded-xl bg-white p-3">
                            <p className="text-sm text-neutral-700">{comment.content}</p>
                            <div className="mt-2 flex items-center gap-2 text-xs">
                              <button
                                type="button"
                                onClick={() =>
                                  handleCommentReaction(post.id, comment.id, "like")
                                }
                                className={`inline-flex items-center gap-1 rounded-full px-2 py-1 ${
                                  commentReactionMap[comment.id] === "like"
                                    ? "bg-emerald-100 text-emerald-800"
                                    : "bg-neutral-100 text-neutral-600"
                                }`}
                              >
                                <ThumbsUp className="h-3 w-3" />
                                {comment.like_count}
                              </button>
                              <button
                                type="button"
                                onClick={() =>
                                  handleCommentReaction(post.id, comment.id, "dislike")
                                }
                                className={`inline-flex items-center gap-1 rounded-full px-2 py-1 ${
                                  commentReactionMap[comment.id] === "dislike"
                                    ? "bg-rose-100 text-rose-800"
                                    : "bg-neutral-100 text-neutral-600"
                                }`}
                              >
                                <ThumbsDown className="h-3 w-3" />
                                {comment.dislike_count}
                              </button>
                            </div>
                          </article>
                        ))}

                        {commentsLoadingByPost[post.id] ? (
                          <p className="text-xs text-neutral-500">Loading comments...</p>
                        ) : null}
                        {commentErrorByPost[post.id] ? (
                          <p className="text-xs text-rose-600">{commentErrorByPost[post.id]}</p>
                        ) : null}
                      </div>

                      <div className="mt-3 flex gap-2">
                        <input
                          value={commentDraftByPost[post.id] ?? ""}
                          onChange={(event) =>
                            setCommentDraftByPost((prev) => ({
                              ...prev,
                              [post.id]: event.target.value,
                            }))
                          }
                          placeholder="Write a comment..."
                          className="h-9 flex-1 rounded-xl border border-neutral-200 bg-white px-3 text-xs outline-none focus:border-neutral-400"
                        />
                        <button
                          type="button"
                          onClick={() => handleCreateComment(post.id)}
                          className="rounded-xl bg-neutral-900 px-3 text-xs font-semibold text-white"
                        >
                          Comment
                        </button>
                      </div>
                    </section>
                  ) : null}
                </article>
              ))
            )}
          </div>
        </section>

        <aside className="hidden space-y-5 md:block">
          <div className="rounded-3xl border border-neutral-200 bg-white p-5 shadow-sm">
            <h2 className="text-sm font-semibold text-neutral-900">People</h2>
            <div className="mt-3 space-y-2">
              {peopleLoading ? (
                <p className="text-xs text-neutral-500">Loading users...</p>
              ) : people.filter((person) => person.id !== user?.id).length === 0 ? (
                <p className="text-xs text-neutral-500">No other users found yet.</p>
              ) : (
                people
                  .filter((person) => person.id !== user?.id)
                  .slice(0, 8)
                  .map((person) => (
                    <div
                      key={`right-user-${person.id}`}
                      className="flex items-center justify-between rounded-2xl border border-neutral-200 bg-neutral-50 px-3 py-2"
                    >
                      <div>
                        <p className="text-xs font-semibold text-neutral-800">
                          {person.first_name} {person.last_name}
                        </p>
                        <p className="text-[11px] text-neutral-500">
                          @{person.nickname || `user-${person.id}`}
                        </p>
                      </div>
                      <span className="rounded-full bg-white px-2 py-1 text-[10px] text-neutral-500">
                        ID {person.id}
                      </span>
                    </div>
                  ))
              )}
            </div>
          </div>

          <div className="rounded-3xl border border-neutral-200 bg-white p-5 shadow-sm">
            <div className="flex items-center justify-between">
              <h2 className="text-sm font-semibold text-neutral-900">Notifications</h2>
              <button
                type="button"
                onClick={markAllNotificationsRead}
                className="text-xs font-semibold text-neutral-600 hover:text-neutral-900"
              >
                Mark all read
              </button>
            </div>
            <div className="mt-3 space-y-2">
              {notificationsLoading ? (
                <p className="text-xs text-neutral-500">Loading notifications...</p>
              ) : notifications.length === 0 ? (
                <p className="text-xs text-neutral-500">No notifications yet.</p>
              ) : (
                notifications.slice(0, 5).map((item) => (
                  <button
                    type="button"
                    key={item.id}
                    onClick={() => markNotificationRead(item.id)}
                    className={`w-full rounded-2xl border px-3 py-2 text-left text-xs transition ${
                      item.is_read
                        ? "border-neutral-200 bg-neutral-50 text-neutral-500"
                        : "border-emerald-200 bg-emerald-50 text-emerald-900"
                    }`}
                  >
                    <p className="font-semibold">{item.type.replace(/_/g, " ")}</p>
                    <p className="mt-0.5 text-[11px]">
                      {item.entity_type} #{item.entity_id}
                    </p>
                  </button>
                ))
              )}
            </div>
          </div>

          <div className="rounded-3xl border border-neutral-200 bg-white p-5 shadow-sm">
            <div className="flex items-center justify-between">
              <h2 className="text-sm font-semibold text-neutral-900">Live chat</h2>
              <span
                className={`inline-flex items-center gap-1 rounded-full px-2 py-1 text-[10px] font-semibold ${
                  chatConnected
                    ? "bg-emerald-100 text-emerald-800"
                    : "bg-rose-100 text-rose-700"
                }`}
              >
                {chatConnected ? <Wifi className="h-3 w-3" /> : <WifiOff className="h-3 w-3" />}
                {chatConnected ? "Connected" : "Offline"}
              </span>
            </div>
            <div className="mt-3 space-y-2">
              <input
                type="number"
                value={chatRecipientID}
                onChange={(event) => setChatRecipientID(event.target.value)}
                placeholder="Recipient user ID"
                className="h-9 w-full rounded-xl border border-neutral-200 bg-neutral-50 px-3 text-xs outline-none focus:border-neutral-400"
              />
              <div className="flex gap-2">
                <input
                  value={chatDraft}
                  onChange={(event) => setChatDraft(event.target.value)}
                  placeholder="Write a direct message..."
                  className="h-9 flex-1 rounded-xl border border-neutral-200 bg-neutral-50 px-3 text-xs outline-none focus:border-neutral-400"
                />
                <button
                  type="button"
                  onClick={sendChatMessage}
                  className="brand-gradient rounded-xl px-3 text-xs font-semibold text-white"
                >
                  Send
                </button>
              </div>
              {chatError ? <p className="text-xs text-rose-600">{chatError}</p> : null}
              <div className="rounded-xl border border-neutral-200 bg-neutral-50 p-2">
                <p className="mb-1 text-[10px] font-semibold uppercase tracking-wide text-neutral-500">
                  Recent chat messages
                </p>
                <div className="max-h-36 space-y-1 overflow-y-auto">
                  {chatMessages.length === 0 ? (
                    <p className="text-xs text-neutral-500">No chat messages yet.</p>
                  ) : (
                    chatMessages.slice(0, 8).map((msg) => (
                      <div key={`${msg.id}-${msg.created_at}`} className="rounded-lg bg-white px-2 py-1">
                        <p className="text-[10px] font-semibold text-neutral-700">
                          User {msg.sender_id} · Conv {msg.conversation_id}
                        </p>
                        <p className="text-xs text-neutral-600">{msg.content || "(media)"}</p>
                        <p className="text-[10px] text-neutral-400">{shortDate(msg.created_at)}</p>
                      </div>
                    ))
                  )}
                </div>
              </div>
              {Object.keys(chatUnreadMap).length > 0 ? (
                <p className="text-[11px] text-neutral-500">
                  Unread conversations: {Object.values(chatUnreadMap).reduce((a, b) => a + b, 0)}
                </p>
              ) : null}
            </div>
          </div>

          <div className="rounded-3xl border border-neutral-200 bg-white p-5 shadow-sm">
            <h2 className="text-sm font-semibold text-neutral-900">Trending topics</h2>
            <div className="mt-4 space-y-3">
              {trends.map((item) => (
                <article key={item.title} className="rounded-2xl border border-neutral-200 bg-neutral-50 p-3">
                  <p className="text-sm font-semibold text-neutral-900">{item.title}</p>
                  <p className="mt-1 text-xs text-neutral-600">{item.posts}</p>
                </article>
              ))}
            </div>
          </div>
        </aside>
      </main>

      {notificationsOpen ? (
        <div className="fixed right-4 top-16 z-50 w-[320px] rounded-2xl border border-neutral-200 bg-white p-3 shadow-xl lg:hidden">
          <div className="mb-2 flex items-center justify-between">
            <p className="text-sm font-semibold text-neutral-900">Notifications</p>
            <button
              type="button"
              onClick={() => setNotificationsOpen(false)}
              className="text-xs font-semibold text-neutral-500 hover:text-neutral-900"
            >
              Close
            </button>
          </div>
          <div className="max-h-64 space-y-2 overflow-y-auto">
            {notifications.map((item) => (
              <button
                type="button"
                key={item.id}
                onClick={() => markNotificationRead(item.id)}
                className={`w-full rounded-xl border px-3 py-2 text-left text-xs ${
                  item.is_read
                    ? "border-neutral-200 bg-neutral-50 text-neutral-500"
                    : "border-emerald-200 bg-emerald-50 text-emerald-900"
                }`}
              >
                <p className="font-semibold">{item.type.replace(/_/g, " ")}</p>
                <p className="mt-0.5 text-[11px]">
                  {item.entity_type} #{item.entity_id}
                </p>
              </button>
            ))}
          </div>
        </div>
      ) : null}
    </div>
  );
}
