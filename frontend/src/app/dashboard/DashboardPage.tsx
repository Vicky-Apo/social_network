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
import { useAuth } from "../component/AuthContext";
import { landingData } from "@/lib/data";
import { apiJson, asArray, asNumber, asString, isRecord } from "@/lib/api";

type ReactionKind = "like" | "dislike";
type ReactionMap = Record<number, ReactionKind | null>;
type WsMessage = {
  type: string;
  payload: unknown;
};

type DashboardUser = {
  id: number;
  displayName: string;
  handle: string;
  initials: string;
};

type DashboardPost = {
  id: number;
  authorName: string;
  authorInitials: string;
  content: string;
  mediaUrl?: string | null;
  privacyLabel: string;
  createdAt: string;
  counts: { comments: number; likes: number; dislikes: number };
};

type DashboardComment = {
  id: number;
  content: string;
  createdAt: string;
  counts: { likes: number; dislikes: number };
};

type DashboardNotification = {
  id: number;
  title: string;
  subtitle: string;
  isRead: boolean;
  createdAt: string;
};

type DashboardChatMessage = {
  id: number;
  conversationId: number;
  senderId: number;
  content: string;
  createdAt: string;
};

type DashboardPerson = {
  id: number;
  name: string;
  handle: string;
  avatarUrl?: string | null;
};

const quickLinks = [
  { label: "Explore", href: "/explore", icon: Compass },
  { label: "Groups", href: "#", icon: Users },
  { label: "Messages", href: "#", icon: MessageSquare },
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

function toDashboardUser(value: unknown): DashboardUser | null {
  if (!isRecord(value)) return null;
  const id = asNumber(value.id);
  if (!id) return null;

  const first = asString(value.first_name) ?? "";
  const last = asString(value.last_name) ?? "";
  const email = asString(value.email) ?? "";
  const nickname = asString(value.nickname) ?? "";

  const displayName = `${first} ${last}`.trim() || "Member";
  const handle =
    nickname.trim() ||
    (email.includes("@") ? email.split("@")[0] : "") ||
    `user-${id}`;

  return { id, displayName, handle, initials: initials(first, last) };
}

function toDashboardPost(value: unknown): DashboardPost | null {
  if (!isRecord(value)) return null;
  const id = asNumber(value.id);
  if (!id) return null;

  const first = asString(value.author_first_name) ?? "";
  const last = asString(value.author_last_name) ?? "";
  const authorName = `${first} ${last}`.trim() || "Member";

  return {
    id,
    authorName,
    authorInitials: initials(first, last),
    content: asString(value.content) ?? "",
    mediaUrl: asString(value.media_path),
    privacyLabel: asString(value.privacy) ?? "public",
    createdAt: asString(value.created_at) ?? "",
    counts: {
      comments: asNumber(value.comment_count) ?? 0,
      likes: asNumber(value.like_count) ?? 0,
      dislikes: asNumber(value.dislike_count) ?? 0,
    },
  };
}

function toDashboardComment(value: unknown): DashboardComment | null {
  if (!isRecord(value)) return null;
  const id = asNumber(value.id);
  if (!id) return null;
  return {
    id,
    content: asString(value.content) ?? "",
    createdAt: asString(value.created_at) ?? "",
    counts: {
      likes: asNumber(value.like_count) ?? 0,
      dislikes: asNumber(value.dislike_count) ?? 0,
    },
  };
}

function toDashboardNotification(value: unknown): DashboardNotification | null {
  if (!isRecord(value)) return null;
  const id = asNumber(value.id);
  if (!id) return null;

  const rawType = (asString(value.type) ?? "notification").replaceAll("_", " ");
  const entityType = (asString(value.entity_type) ?? "").replaceAll("_", " ");
  const createdAt = asString(value.created_at) ?? "";
  const isRead = value.is_read === true;

  const title = rawType.charAt(0).toUpperCase() + rawType.slice(1);
  const subtitle = entityType
    ? `${entityType} · ${shortDate(createdAt)}`
    : shortDate(createdAt);

  return { id, title, subtitle, isRead, createdAt };
}

function toDashboardChatMessage(value: unknown): DashboardChatMessage | null {
  if (!isRecord(value)) return null;
  const id = asNumber(value.id);
  const conversationId = asNumber(value.conversation_id);
  const senderId = asNumber(value.sender_id);
  if (!id || !conversationId || !senderId) return null;

  return {
    id,
    conversationId,
    senderId,
    content: asString(value.content) ?? "",
    createdAt: asString(value.created_at) ?? "",
  };
}

function toDashboardPerson(value: unknown): DashboardPerson | null {
  if (!isRecord(value)) return null;
  const id = asNumber(value.id);
  if (!id) return null;

  const first = asString(value.first_name) ?? "";
  const last = asString(value.last_name) ?? "";
  const nickname = asString(value.nickname) ?? "";
  const name = `${first} ${last}`.trim() || "Member";
  const handle = nickname.trim() || `user-${id}`;

  return { id, name, handle, avatarUrl: asString(value.avatar_path) };
}

export default function DashboardPage() {
  const router = useRouter();
  const { logout } = useAuth();

  const [user, setUser] = useState<DashboardUser | null>(null);
  const [posts, setPosts] = useState<DashboardPost[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [feedError, setFeedError] = useState<string | null>(null);
  const [notificationCount, setNotificationCount] = useState(0);
  const [composerText, setComposerText] = useState("");
  const [composerError, setComposerError] = useState<string | null>(null);
  const [isPosting, setIsPosting] = useState(false);
  const [groupsOnly, setGroupsOnly] = useState(false);
  const [postReactionMap, setPostReactionMap] = useState<ReactionMap>({});
  const [commentReactionMap, setCommentReactionMap] = useState<ReactionMap>({});
  const [commentsByPost, setCommentsByPost] = useState<Record<number, DashboardComment[]>>({});
  const [commentsOpenByPost, setCommentsOpenByPost] = useState<Record<number, boolean>>({});
  const [commentsLoadingByPost, setCommentsLoadingByPost] = useState<Record<number, boolean>>({});
  const [commentDraftByPost, setCommentDraftByPost] = useState<Record<number, string>>({});
  const [commentErrorByPost, setCommentErrorByPost] = useState<Record<number, string>>({});
  const [notifications, setNotifications] = useState<DashboardNotification[]>([]);
  const [notificationsOpen, setNotificationsOpen] = useState(false);
  const [notificationsLoading, setNotificationsLoading] = useState(false);
  const [chatRecipientID, setChatRecipientID] = useState("");
  const [chatDraft, setChatDraft] = useState("");
  const [chatMessages, setChatMessages] = useState<DashboardChatMessage[]>([]);
  const [chatError, setChatError] = useState<string | null>(null);
  const [chatConnected, setChatConnected] = useState(false);
  const [chatUnreadMap, setChatUnreadMap] = useState<Record<number, number>>({});
  const [searchQuery, setSearchQuery] = useState("");
  const [people, setPeople] = useState<DashboardPerson[]>([]);
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

    const load = async () => {
      setIsLoading(true);
      setFeedError(null);

      try {
        const me = await apiJson(apiBaseUrl, "/auth/me");
        if (!me.ok || !me.json?.success || !me.json.data) {
          if (!cancelled) {
            router.replace("/login");
          }
          return;
        }

        if (!cancelled) {
          setUser(toDashboardUser(me.json.data));
        }

        const unread = await apiJson(apiBaseUrl, "/notifications/unread-count").catch(() => null);
        if (!cancelled && unread?.ok && unread.json?.success && isRecord(unread.json.data)) {
          setNotificationCount(asNumber(unread.json.data.count) ?? 0);
        }

        const notificationsRes = await apiJson(apiBaseUrl, "/notifications?limit=10").catch(
          () => null,
        );
        if (!cancelled && notificationsRes?.ok && notificationsRes.json?.success) {
          const raw = asArray(notificationsRes.json.data) ?? [];
          setNotifications(raw.map(toDashboardNotification).filter(Boolean) as DashboardNotification[]);
        }

        const feedPath = groupsOnly ? "/posts?groups_only=true" : "/posts";
        const feed = await apiJson(apiBaseUrl, feedPath);
        if (!feed.ok || !feed.json?.success) {
          if (!cancelled) {
            setFeedError(feed.json?.error || "Unable to load your feed.");
            setPosts([]);
          }
          return;
        }

        if (!cancelled) {
          const raw = asArray(feed.json.data) ?? [];
          setPosts(raw.map(toDashboardPost).filter(Boolean) as DashboardPost[]);
          setPostReactionMap({});
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
        const response = await apiJson(apiBaseUrl, path);
        if (!cancelled && response.ok && response.json?.success) {
          const raw = asArray(response.json.data) ?? [];
          setPeople(raw.map(toDashboardPerson).filter(Boolean) as DashboardPerson[]);
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

  // Intentionally do not surface backend internal objects (follow requests, profiles, etc.)
  // in the UI. The dashboard should only present user-facing content.

  useEffect(() => {
    if (!user?.id) {
      return;
    }

    const intervalID = window.setInterval(async () => {
      const feedPath = groupsOnly ? "/posts?groups_only=true" : "/posts";
      const response = await apiJson(apiBaseUrl, feedPath).catch(() => null);
      if (!response?.ok || !response.json?.success) return;
      const raw = asArray(response.json.data) ?? [];
      setPosts(raw.map(toDashboardPost).filter(Boolean) as DashboardPost[]);
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
          const parsed = JSON.parse(raw) as unknown;
          if (!isRecord(parsed)) return;

          const type = asString(parsed.type) ?? "";
          const payload = parsed.payload;

          if (type === "chat_message") {
            const next = toDashboardChatMessage(payload);
            if (next) {
              setChatMessages((prev) => [next, ...prev].slice(0, 50));
            }
          } else if (type === "notification") {
            const next = toDashboardNotification(payload);
            if (next) {
              setNotifications((prev) => [next, ...prev].slice(0, 20));
              setNotificationCount((prev) => prev + 1);
            }
          } else if (type === "unread_counts") {
            const list = asArray(payload) ?? [];
            setChatUnreadMap((prev) => {
              const next = { ...prev };
              list.forEach((item) => {
                if (!isRecord(item)) return;
                const id = asNumber(item.conversation_id);
                const count = asNumber(item.unread_count);
                if (!id || count === null) return;
                next[id] = count;
              });
              return next;
            });
          } else if (type === "error") {
            const message =
              isRecord(payload) ? asString(payload.message) ?? "Chat error." : "Chat error.";
            setChatError(message);
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
      await apiJson(apiBaseUrl, "/auth/logout", { method: "POST" }).catch(() => undefined);
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
      const response = await apiJson(apiBaseUrl, "/posts", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          content,
          privacy: "public",
        }),
      });

      if (!response.ok || !response.json?.success || !response.json.data) {
        setComposerError(response.json?.error || "Could not publish your post.");
        return;
      }

      const created = toDashboardPost(response.json.data);
      if (created) {
        setPosts((prev) => [created, ...prev]);
      }
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
      const response = await apiJson(apiBaseUrl, "/notifications?limit=20");
      if (response.ok && response.json?.success) {
        const raw = asArray(response.json.data) ?? [];
        setNotifications(raw.map(toDashboardNotification).filter(Boolean) as DashboardNotification[]);
      }
    } finally {
      setNotificationsLoading(false);
    }
  };

  const markNotificationRead = async (id: number) => {
    const old = notifications;
    setNotifications((prev) =>
      prev.map((item) => (item.id === id ? { ...item, isRead: true } : item)),
    );
    setNotificationCount((prev) => Math.max(0, prev - 1));

    try {
      const response = await apiJson(apiBaseUrl, `/notifications/${id}/read`, {
        method: "PATCH",
      });
      if (!response.ok) {
        setNotifications(old);
      }
    } catch {
      setNotifications(old);
    }
  };

  const markAllNotificationsRead = async () => {
    setNotifications((prev) => prev.map((item) => ({ ...item, isRead: true })));
    setNotificationCount(0);
    await apiJson(apiBaseUrl, "/notifications/read-all", { method: "PATCH" }).catch(() => undefined);
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
      const response = await apiJson(apiBaseUrl, feedPath);
      if (!response.ok || !response.json?.success) {
        setFeedError(response.json?.error || "Could not refresh feed.");
        return;
      }
      const raw = asArray(response.json.data) ?? [];
      setPosts(raw.map(toDashboardPost).filter(Boolean) as DashboardPost[]);
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
      const response = await apiJson(apiBaseUrl, `/posts/${postID}/comments`);
      if (!response.ok || !response.json?.success) {
        setCommentErrorByPost((prev) => ({
          ...prev,
          [postID]: response.json?.error || "Could not load comments.",
        }));
        return;
      }

      const raw = asArray(response.json.data) ?? [];
      const comments = raw.map(toDashboardComment).filter(Boolean) as DashboardComment[];
      setCommentsByPost((prev) => ({ ...prev, [postID]: comments }));
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
      const response = await apiJson(apiBaseUrl, `/posts/${postID}/comments`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ content: draft }),
      });

      if (!response.ok || !response.json?.success || !response.json.data) {
        setCommentErrorByPost((prev) => ({
          ...prev,
          [postID]: response.json?.error || "Could not post comment.",
        }));
        return;
      }

      const created = toDashboardComment(response.json.data);
      if (!created) return;

      setCommentsByPost((prev) => ({
        ...prev,
        [postID]: [created, ...(prev[postID] ?? [])],
      }));
      setCommentDraftByPost((prev) => ({ ...prev, [postID]: "" }));
      setCommentErrorByPost((prev) => ({ ...prev, [postID]: "" }));
      setPosts((prev) =>
        prev.map((post) =>
          post.id === postID
            ? { ...post, counts: { ...post.counts, comments: post.counts.comments + 1 } }
            : post,
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
        let like = post.counts.likes;
        let dislike = post.counts.dislikes;
        if (previous === "like") like -= 1;
        if (previous === "dislike") dislike -= 1;
        if (next === "like") like += 1;
        if (next === "dislike") dislike += 1;
        return {
          ...post,
          counts: {
            ...post.counts,
            likes: Math.max(0, like),
            dislikes: Math.max(0, dislike),
          },
        };
      }),
    );

    try {
      await apiJson(apiBaseUrl, `/posts/${postID}/reactions`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ reaction }),
      });
    } catch {
      setPostReactionMap((prev) => ({ ...prev, [postID]: previous }));
      setPosts((prev) =>
        prev.map((post) => {
          if (post.id !== postID) return post;
          let like = post.counts.likes;
          let dislike = post.counts.dislikes;
          if (next === "like") like -= 1;
          if (next === "dislike") dislike -= 1;
          if (previous === "like") like += 1;
          if (previous === "dislike") dislike += 1;
          return {
            ...post,
            counts: {
              ...post.counts,
              likes: Math.max(0, like),
              dislikes: Math.max(0, dislike),
            },
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
        let like = comment.counts.likes;
        let dislike = comment.counts.dislikes;
        if (previous === "like") like -= 1;
        if (previous === "dislike") dislike -= 1;
        if (next === "like") like += 1;
        if (next === "dislike") dislike += 1;
        return {
          ...comment,
          counts: {
            ...comment.counts,
            likes: Math.max(0, like),
            dislikes: Math.max(0, dislike),
          },
        };
      }),
    }));

    try {
      await apiJson(apiBaseUrl, `/comments/${commentID}/reactions`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          reaction,
          user_id: user?.id ?? 0,
        }),
      });
    } catch {
      setCommentReactionMap((prev) => ({ ...prev, [commentID]: previous }));
    }
  };

  const displayName = user ? user.displayName : "Loading";
  const userTag = user ? user.handle : "community-member";

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
                {user?.initials ?? "U"}
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
                        className="rounded-2xl border border-neutral-200 bg-neutral-50 px-3 py-2"
                      >
                        <p className="text-xs font-semibold text-neutral-800">{person.name}</p>
                        <p className="text-[11px] text-neutral-500">@{person.handle}</p>
                      </div>
                    ))
                )}
              </div>
            </div>
          </div>
        </aside>

        <section className="space-y-5">
          <div className="rounded-3xl border border-neutral-200 bg-white p-4 shadow-sm sm:p-5">
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
          </div>

          <div className="rounded-3xl border border-neutral-200 bg-white p-4 shadow-sm sm:p-5">
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
          </div>

          <div className="rounded-3xl border border-neutral-200 bg-white p-4 shadow-sm lg:hidden">
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
          </div>

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
                        {post.authorInitials}
                      </span>
                      <div>
                        <p className="text-sm font-semibold text-neutral-900">{post.authorName}</p>
                        <p className="text-xs text-neutral-500">{shortDate(post.createdAt)}</p>
                      </div>
                    </div>
                    <div className="flex items-center gap-2">
                      <Link
                        href={`/posts/${post.id}`}
                        className="rounded-full border border-neutral-200 bg-white px-3 py-1 text-[11px] font-semibold text-neutral-700 transition hover:border-neutral-300 hover:text-neutral-900"
                      >
                        Open
                      </Link>
                      <span className="rounded-full border border-neutral-200 bg-neutral-50 px-2.5 py-1 text-[11px] uppercase tracking-wide text-neutral-600">
                        {post.privacyLabel}
                      </span>
                    </div>
                  </header>

                  <p className="mt-4 text-sm leading-relaxed text-neutral-700">{post.content}</p>

                  {post.mediaUrl ? (
                    <div className="mt-4 overflow-hidden rounded-2xl border border-neutral-200">
                      <img src={post.mediaUrl} alt="Post media" className="h-72 w-full object-cover" />
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
                      {post.counts.likes}
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
                      {post.counts.dislikes}
                    </button>
                    <button
                      type="button"
                      onClick={() => toggleComments(post.id)}
                      className="inline-flex items-center gap-1 rounded-full bg-neutral-100 px-2 py-1 text-neutral-600 transition hover:bg-neutral-200"
                    >
                      <MessageCircle className="h-3.5 w-3.5" />
                      {post.counts.comments}
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
                                {comment.counts.likes}
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
                                {comment.counts.dislikes}
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
                    <div key={`right-user-${person.id}`} className="rounded-2xl border border-neutral-200 bg-neutral-50 px-3 py-2">
                      <p className="text-xs font-semibold text-neutral-800">{person.name}</p>
                      <p className="text-[11px] text-neutral-500">@{person.handle}</p>
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
                      item.isRead
                        ? "border-neutral-200 bg-neutral-50 text-neutral-500"
                        : "border-emerald-200 bg-emerald-50 text-emerald-900"
                    }`}
                  >
                    <p className="font-semibold">{item.title}</p>
                    <p className="mt-0.5 text-[11px]">{item.subtitle}</p>
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
                      <div key={`${msg.id}-${msg.createdAt}`} className="rounded-lg bg-white px-2 py-1">
                        <p className="text-[10px] font-semibold text-neutral-700">
                          User {msg.senderId} · Conv {msg.conversationId}
                        </p>
                        <p className="text-xs text-neutral-600">{msg.content || "(media)"}</p>
                        <p className="text-[10px] text-neutral-400">{shortDate(msg.createdAt)}</p>
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
                  item.isRead
                    ? "border-neutral-200 bg-neutral-50 text-neutral-500"
                    : "border-emerald-200 bg-emerald-50 text-emerald-900"
                }`}
              >
                <p className="font-semibold">{item.title}</p>
                <p className="mt-0.5 text-[11px]">{item.subtitle}</p>
              </button>
            ))}
          </div>
        </div>
      ) : null}
    </div>
  );
}
