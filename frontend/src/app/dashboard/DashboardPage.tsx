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
import { BrandMark } from "@/components/BrandMark";
import { Footer } from "@/components/Footer";

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

  const pageSize = 5;
  const [user, setUser] = useState<DashboardUser | null>(null);
  const [posts, setPosts] = useState<DashboardPost[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [isPaging, setIsPaging] = useState(false);
  const [feedError, setFeedError] = useState<string | null>(null);
  const [pageIndex, setPageIndex] = useState(0);
  const [hasNextPage, setHasNextPage] = useState(false);
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
      } catch {
        // Non-blocking: keep dashboard usable even if counts fail.
      } finally {
        // no-op
      }
    };

    load();

    return () => {
      cancelled = true;
    };
  }, [apiBaseUrl, router]);

  const pageCacheRef = useRef<Map<string, { posts: DashboardPost[]; hasNext: boolean }>>(
    new Map(),
  );

  useEffect(() => {
    // When feed mode changes, invalidate cached pages.
    pageCacheRef.current.clear();
  }, [groupsOnly]);

  useEffect(() => {
    let cancelled = false;

    const cacheKey = `${groupsOnly ? 1 : 0}:${pageIndex}`;
    const cached = pageCacheRef.current.get(cacheKey);
    if (cached) {
      setFeedError(null);
      setPosts(cached.posts);
      setHasNextPage(cached.hasNext);
      setIsLoading(false);
      setIsPaging(false);
      return () => {
        cancelled = true;
      };
    }

    const loadPage = async () => {
      const isFirstLoad = isLoading && posts.length === 0;
      if (!isFirstLoad) {
        setIsPaging(true);
      }
      setFeedError(null);

      try {
        const params = new URLSearchParams();
        if (groupsOnly) params.set("groups_only", "true");
        params.set("limit", String(pageSize));
        params.set("offset", String(pageIndex * pageSize));
        const feedPath = `/posts?${params.toString()}`;

        const feed = await apiJson(apiBaseUrl, feedPath);
        if (feed.status === 401) {
          router.replace("/login");
          return;
        }
        if (!feed.ok || !feed.json?.success) {
          setFeedError(feed.json?.error || "Unable to load your feed.");
          if (isFirstLoad) {
            setPosts([]);
            setHasNextPage(false);
          }
          return;
        }

        const raw = asArray(feed.json.data) ?? [];
        const nextPosts = raw.map(toDashboardPost).filter(Boolean) as DashboardPost[];
        const nextHasNext = nextPosts.length === pageSize;

        if (cancelled) return;
        pageCacheRef.current.set(cacheKey, { posts: nextPosts, hasNext: nextHasNext });
        setPosts(nextPosts);
        setHasNextPage(nextHasNext);
        setPostReactionMap({});
        setCommentsByPost({});
        setCommentsOpenByPost({});

        // Prefetch the next page to make "Next" feel instant.
        if (nextHasNext) {
          const nextKey = `${groupsOnly ? 1 : 0}:${pageIndex + 1}`;
          if (!pageCacheRef.current.has(nextKey)) {
            const nextParams = new URLSearchParams();
            if (groupsOnly) nextParams.set("groups_only", "true");
            nextParams.set("limit", String(pageSize));
            nextParams.set("offset", String((pageIndex + 1) * pageSize));
            void apiJson(apiBaseUrl, `/posts?${nextParams.toString()}`)
              .then((res) => {
                if (!res.ok || !res.json?.success) return;
                const items = (asArray(res.json.data) ?? [])
                  .map(toDashboardPost)
                  .filter(Boolean) as DashboardPost[];
                pageCacheRef.current.set(nextKey, {
                  posts: items,
                  hasNext: items.length === pageSize,
                });
              })
              .catch(() => undefined);
          }
        }
      } catch {
        if (!cancelled) {
          setFeedError("Network error. Please try again.");
        }
      } finally {
        if (!cancelled) {
          setIsPaging(false);
          setIsLoading(false);
        }
      }
    };

    void loadPage();

    return () => {
      cancelled = true;
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [apiBaseUrl, router, groupsOnly, pageIndex, pageSize]);

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

  // No feed polling here; pagination should be stable while browsing.

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
        if (pageIndex === 0) {
          setPosts((prev) => [created, ...prev].slice(0, pageSize));
        } else {
          setPageIndex(0);
        }
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
    setIsPaging(true);
    setFeedError(null);
    try {
      const params = new URLSearchParams();
      if (groupsOnly) params.set("groups_only", "true");
      params.set("limit", String(pageSize));
      params.set("offset", String(pageIndex * pageSize));
      const feedPath = `/posts?${params.toString()}`;
      const response = await apiJson(apiBaseUrl, feedPath);
      if (!response.ok || !response.json?.success) {
        setFeedError(response.json?.error || "Could not refresh feed.");
        return;
      }
      const raw = asArray(response.json.data) ?? [];
      const nextPosts = raw.map(toDashboardPost).filter(Boolean) as DashboardPost[];
      setPosts(nextPosts);
      setHasNextPage(nextPosts.length === pageSize);
      pageCacheRef.current.set(`${groupsOnly ? 1 : 0}:${pageIndex}`, {
        posts: nextPosts,
        hasNext: nextPosts.length === pageSize,
      });
    } catch {
      setFeedError("Network error. Please try again.");
    } finally {
      setIsPaging(false);
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
    <div className="relative z-[1] min-h-screen bg-[#2b2929] text-white">
      <header className="sticky top-0 z-40 border-b border-white/10 bg-black/30 backdrop-blur-md">
        <div className="mx-auto flex w-full max-w-6xl items-center gap-3 px-4 py-3 sm:px-6">
          <Link href="/" className="inline-flex items-center text-white focus:outline-none focus-visible:ring-2 focus-visible:ring-white/50" aria-label={landingData.productName}>
            <BrandMark label={landingData.productName} size="sm" logoSrc="/vybez-logo-v2.png" />
          </Link>

          <div className="relative ml-2 hidden flex-1 sm:block">
            <Search className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-white/50" />
            <input
              type="search"
              value={searchQuery}
              onChange={(event) => setSearchQuery(event.target.value)}
              placeholder="Search posts, people, topics..."
              className="h-11 w-full rounded-sm border border-white/30 bg-white/5 pl-9 pr-4 text-sm text-white placeholder:text-white/50 outline-none transition focus:border-white/60 focus:ring-2 focus:ring-white/30 focus:ring-offset-2 focus:ring-offset-[#2b2929]"
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
            className="relative inline-flex h-10 w-10 items-center justify-center rounded-full border border-white/20 bg-white/5 text-white/80 transition hover:bg-white/10 hover:text-white"
          >
            <Bell className="h-4 w-4" />
            <span className="absolute -right-1 -top-1 inline-flex h-4 min-w-4 items-center justify-center rounded-full bg-white px-1 text-[10px] font-semibold text-[#2b2929]">
              {notificationCount}
            </span>
          </button>

          <button
            type="button"
            onClick={handleLogout}
            className="inline-flex items-center gap-2 rounded-full border border-white/20 bg-white/5 px-3 py-2 text-xs font-semibold text-white/90 transition hover:bg-white/10"
          >
            <LogOut className="h-3.5 w-3.5" />
            <span className="hidden sm:inline">Log out</span>
          </button>
        </div>
      </header>

      <main className="relative min-h-[80vh]">
        <Image
          src="/dashboard-bg.png"
          alt=""
          fill
          className="object-cover object-center"
          priority
        />
        <div className="absolute inset-0 bg-[#2b2929]/70" aria-hidden />
        <div className="relative z-10 mx-auto grid w-full max-w-6xl gap-6 px-4 py-6 sm:px-6 lg:grid-cols-[minmax(0,1fr)_300px] lg:gap-8">
        <section className="min-w-0 space-y-5">
          <div className="rounded-sm border border-white/10 bg-[#2b2929]/40 p-4 shadow-sm backdrop-blur-sm sm:p-5">
            <div className="flex items-center justify-between gap-3">
              <div>
                <h1 className="text-xl font-semibold tracking-tight text-white">Dashboard</h1>
                <p className="text-sm text-white/70">Create updates and follow your community feed.</p>
              </div>
              <div className="flex items-center gap-2">
                <label className="inline-flex items-center gap-2 text-xs text-white/70">
                  Groups only
                  <input
                    type="checkbox"
                    checked={groupsOnly}
                    onChange={(event) => {
                      setGroupsOnly(event.target.checked);
                      setPageIndex(0);
                    }}
                    className="h-4 w-4 rounded border-white/30 bg-white/5 text-white focus:ring-white/50"
                  />
                </label>
                <button
                  type="button"
                  onClick={refreshFeed}
                  className="rounded-full border border-white/20 bg-white/5 px-3 py-1.5 text-xs font-semibold text-white/90 transition hover:bg-white/10"
                >
                  Refresh feed
                </button>
              </div>
            </div>
          </div>

          <div className="rounded-sm border border-white/10 bg-[#2b2929]/40 p-4 shadow-sm backdrop-blur-sm sm:p-5">
            <textarea
              value={composerText}
              onChange={(event) => setComposerText(event.target.value)}
              rows={4}
              placeholder="Share an update with Vybez..."
              className="w-full resize-none rounded-sm border border-white/30 bg-white/5 px-4 py-3 text-sm text-white placeholder:text-white/50 outline-none transition focus:border-white/60 focus:ring-2 focus:ring-white/30 focus:ring-offset-2 focus:ring-offset-[#2b2929]"
            />
            <div className="mt-3 flex flex-wrap items-center justify-between gap-3">
              <button
                type="button"
                className="inline-flex items-center gap-2 rounded-full border border-white/20 bg-white/5 px-3 py-2 text-xs font-semibold text-white/80 transition hover:bg-white/10"
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
            {composerError ? <p className="mt-3 text-xs text-rose-400">{composerError}</p> : null}
          </div>

          <div className="overflow-hidden rounded-xl border border-white/10 bg-[#1a1919]/90 shadow-lg backdrop-blur-md lg:hidden">
            <div className={`flex items-center gap-2 border-b border-white/10 px-4 py-3 ${chatConnected ? "bg-emerald-500/10" : "bg-white/5"}`}>
              <div className="flex h-8 w-8 shrink-0 items-center justify-center rounded-full bg-white/10">
                <MessageSquare className="h-4 w-4 text-white/80" />
              </div>
              <div className="min-w-0 flex-1">
                <p className="text-sm font-semibold text-white">Messages</p>
                <p className="flex items-center gap-1.5 text-[11px] text-white/60">
                  <span className={`relative flex h-1.5 w-1.5 rounded-full ${chatConnected ? "bg-emerald-400" : "bg-rose-400"}`} />
                  {chatConnected ? "Live" : "Offline"}
                </p>
              </div>
            </div>
            <div className="space-y-2 p-3">
              <div className="flex items-center gap-1.5 rounded-lg bg-white/5 px-2 py-1">
                <span className="text-[10px] font-medium text-white/50">To</span>
                <input
                  type="number"
                  value={chatRecipientID}
                  onChange={(event) => setChatRecipientID(event.target.value)}
                  placeholder="User ID"
                  className="w-20 rounded bg-transparent px-1.5 py-0.5 text-xs text-white placeholder:text-white/40 outline-none"
                />
              </div>
              <div className="flex gap-2">
                <input
                  value={chatDraft}
                  onChange={(event) => setChatDraft(event.target.value)}
                  placeholder="Type a message…"
                  className="min-w-0 flex-1 rounded-lg border border-white/20 bg-white/5 px-3 py-2 text-xs text-white placeholder:text-white/40 outline-none focus:border-white/40"
                />
                <button
                  type="button"
                  onClick={sendChatMessage}
                  className="flex h-9 w-9 shrink-0 items-center justify-center rounded-lg bg-white/15 text-white transition hover:bg-white/25 active:scale-95"
                  aria-label="Send message"
                >
                  <Send className="h-4 w-4" />
                </button>
              </div>
            </div>
          </div>

          <div className="space-y-4">
            {isLoading && posts.length === 0 ? (
              <article className="rounded-xl border border-white/10 bg-[#2b2929]/50 p-6 text-sm text-white/70 shadow-md backdrop-blur-sm ring-1 ring-white/5">
                Loading your feed...
              </article>
            ) : feedError ? (
              <article className="rounded-xl border border-rose-500/30 bg-rose-500/10 p-6 text-sm text-rose-400 backdrop-blur-sm">
                {feedError}
              </article>
            ) : posts.length === 0 ? (
              <article className="rounded-xl border border-white/10 bg-[#2b2929]/50 p-6 text-sm text-white/70 shadow-md backdrop-blur-sm ring-1 ring-white/5">
                No posts yet. Be the first to share an update.
              </article>
            ) : (
              posts.map((post) => (
                <article
                  key={post.id}
                  className="rounded-xl border border-white/10 bg-[#2b2929]/50 p-5 shadow-md backdrop-blur-sm ring-1 ring-white/5"
                >
                  <header className="flex items-start justify-between gap-3">
                    <div className="flex items-center gap-3">
                      <span className="inline-flex h-10 w-10 items-center justify-center rounded-full bg-white/20 text-xs font-semibold text-white">
                        {post.authorInitials}
                      </span>
                      <div>
                        <p className="text-sm font-semibold text-white">{post.authorName}</p>
                        <p className="text-xs text-white/50">{shortDate(post.createdAt)}</p>
                      </div>
                    </div>
                    <div className="flex items-center gap-2">
                      <Link
                        href={`/posts/${post.id}`}
                        className="rounded-full border border-white/20 bg-white/5 px-3 py-1 text-[11px] font-semibold text-white/90 transition hover:bg-white/10"
                      >
                        Open
                      </Link>
                      <span className="rounded-full border border-white/10 bg-white/5 px-2.5 py-1 text-[11px] uppercase tracking-wide text-white/60">
                        {post.privacyLabel}
                      </span>
                    </div>
                  </header>

                  <p className="mt-4 text-sm leading-relaxed text-white/90">{post.content}</p>

                  {post.mediaUrl ? (
                    <div className="mt-4 overflow-hidden rounded-sm border border-white/10">
                      <img src={post.mediaUrl} alt="Post media" className="h-72 w-full object-cover" />
                    </div>
                  ) : null}

                  <footer className="mt-4 flex items-center gap-4 text-xs text-white/70">
                    <button
                      type="button"
                      onClick={() => handlePostReaction(post.id, "like")}
                      className={`inline-flex items-center gap-1 rounded-full px-2 py-1 transition ${
                        postReactionMap[post.id] === "like"
                          ? "bg-brand-primary/20 text-brand-primary"
                          : "bg-white/10 text-white/70 hover:bg-white/15"
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
                          ? "bg-rose-500/20 text-rose-400"
                          : "bg-white/10 text-white/70 hover:bg-white/15"
                      }`}
                    >
                      <ThumbsDown className="h-3.5 w-3.5" />
                      {post.counts.dislikes}
                    </button>
                    <button
                      type="button"
                      onClick={() => toggleComments(post.id)}
                      className="inline-flex items-center gap-1 rounded-full bg-white/10 px-2 py-1 text-white/70 transition hover:bg-white/15"
                    >
                      <MessageCircle className="h-3.5 w-3.5" />
                      {post.counts.comments}
                    </button>
                  </footer>

                  {commentsOpenByPost[post.id] ? (
                    <section className="mt-4 rounded-sm border border-white/10 bg-white/5 p-3 backdrop-blur-sm">
                      <div className="space-y-2">
                        {(commentsByPost[post.id] ?? []).map((comment) => (
                          <article key={comment.id} className="rounded-sm bg-[#2b2929]/60 border border-white/10 p-3">
                            <p className="text-sm text-white/90">{comment.content}</p>
                            <div className="mt-2 flex items-center gap-2 text-xs">
                              <button
                                type="button"
                                onClick={() =>
                                  handleCommentReaction(post.id, comment.id, "like")
                                }
                                className={`inline-flex items-center gap-1 rounded-full px-2 py-1 ${
                                  commentReactionMap[comment.id] === "like"
                                    ? "bg-brand-primary/20 text-brand-primary"
                                    : "bg-white/10 text-white/70"
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
                                    ? "bg-rose-500/20 text-rose-400"
                                    : "bg-white/10 text-white/70"
                                }`}
                              >
                                <ThumbsDown className="h-3 w-3" />
                                {comment.counts.dislikes}
                              </button>
                            </div>
                          </article>
                        ))}

                        {commentsLoadingByPost[post.id] ? (
                          <p className="text-xs text-white/50">Loading comments...</p>
                        ) : null}
                        {commentErrorByPost[post.id] ? (
                          <p className="text-xs text-rose-400">{commentErrorByPost[post.id]}</p>
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
                          className="h-9 flex-1 rounded-sm border border-white/30 bg-white/5 px-3 text-xs text-white placeholder:text-white/50 outline-none focus:border-white/60"
                        />
                        <button
                          type="button"
                          onClick={() => handleCreateComment(post.id)}
                          className="rounded-sm bg-white/20 px-3 text-xs font-semibold text-white border border-white/20 hover:bg-white/30"
                        >
                          Comment
                        </button>
                      </div>
                    </section>
                  ) : null}
                </article>
              ))
            )}
            <div className="flex flex-wrap items-center justify-between gap-2 rounded-sm border border-white/10 bg-[#2b2929]/40 px-4 py-2 text-xs text-white/70 backdrop-blur-sm">
              <span>
                {isPaging
                  ? "Loading page…"
                  : `Showing newest ${pageSize} · Page ${pageIndex + 1}`}
              </span>
              <div className="flex items-center gap-2">
                <button
                  type="button"
                  onClick={() => setPageIndex((prev) => Math.max(0, prev - 1))}
                  disabled={(isLoading && posts.length === 0) || isPaging || pageIndex === 0}
                  className="rounded-full border border-white/20 bg-white/5 px-3 py-1.5 text-[11px] font-semibold text-white/90 transition hover:bg-white/10 disabled:cursor-not-allowed disabled:opacity-60"
                >
                  Prev
                </button>
                <button
                  type="button"
                  onClick={() => setPageIndex((prev) => prev + 1)}
                  disabled={(isLoading && posts.length === 0) || isPaging || !hasNextPage}
                  className="rounded-full border border-white/20 bg-white/5 px-3 py-1.5 text-[11px] font-semibold text-white/90 transition hover:bg-white/10 disabled:cursor-not-allowed disabled:opacity-60"
                >
                  Next
                </button>
              </div>
            </div>
          </div>
        </section>

        <aside className="hidden lg:block lg:space-y-6">
          <div className="overflow-hidden rounded-xl border-l-4 border-white/20 bg-gradient-to-br from-white/10 to-transparent shadow-lg">
            <div className="flex items-center gap-4 p-4">
              <div className="inline-flex h-14 w-14 shrink-0 items-center justify-center rounded-2xl bg-white/15 text-base font-bold text-white shadow-inner">
                {user?.initials ?? "U"}
              </div>
              <div className="min-w-0 flex-1">
                <p className="truncate text-sm font-bold text-white">{displayName}</p>
                <p className="truncate text-xs text-white/50">@{userTag}</p>
              </div>
            </div>
            <nav className="flex flex-col gap-1.5 border-t border-white/10 p-3">
              {quickLinks.map((item) => {
                const Icon = item.icon;
                return (
                  <Link
                    key={item.label}
                    href={item.href}
                    className="inline-flex items-center gap-2 rounded-lg bg-white/10 px-3 py-2 text-xs font-medium text-white/90 transition hover:bg-white/20 hover:text-white"
                  >
                    <Icon className="h-3.5 w-3.5 shrink-0" />
                    <span>{item.label}</span>
                  </Link>
                );
              })}
            </nav>
          </div>

          <div className="overflow-hidden rounded-xl border border-white/10 bg-white/5 shadow-inner">
            <div className="border-b border-white/10 bg-white/10 px-4 py-3">
              <h2 className="text-sm font-semibold text-white">People</h2>
            </div>
            <div className="max-h-56 space-y-1 overflow-y-auto p-2">
              {peopleLoading ? (
                <p className="p-3 text-xs text-white/50">Loading users...</p>
              ) : people.filter((person) => person.id !== user?.id).length === 0 ? (
                <p className="p-3 text-xs text-white/50">No other users found yet.</p>
              ) : (
                people
                  .filter((person) => person.id !== user?.id)
                  .slice(0, 8)
                  .map((person) => (
                    <div key={`right-user-${person.id}`} className="rounded-lg bg-[#2b2929]/60 px-3 py-2">
                      <p className="text-xs font-semibold text-white/90">{person.name}</p>
                      <p className="text-[11px] text-white/50">@{person.handle}</p>
                    </div>
                  ))
              )}
            </div>
          </div>

          <div className="overflow-hidden rounded-xl border border-white/10 bg-[#1a1919]/90 shadow-lg backdrop-blur-md">
            <div className={`flex items-center gap-2 border-b border-white/10 px-4 py-3 ${chatConnected ? "bg-emerald-500/10" : "bg-white/5"}`}>
              <div className="flex h-8 w-8 shrink-0 items-center justify-center rounded-full bg-white/10">
                <MessageSquare className="h-4 w-4 text-white/80" />
              </div>
              <div className="min-w-0 flex-1">
                <p className="text-sm font-semibold text-white">Messages</p>
                <p className="flex items-center gap-1.5 text-[11px] text-white/60">
                  <span className={`relative flex h-1.5 w-1.5 rounded-full ${chatConnected ? "bg-emerald-400 shadow-[0_0_6px_rgba(52,211,153,0.6)]" : "bg-rose-400"}`} />
                  {chatConnected ? "Live" : "Reconnecting…"}
                </p>
              </div>
            </div>

            <div className="flex flex-col">
              <div className="min-h-[120px] max-h-44 overflow-y-auto border-b border-white/10 bg-black/20 p-3">
                {chatMessages.length === 0 ? (
                  <div className="flex h-24 flex-col items-center justify-center gap-1.5 text-center">
                    <MessageSquare className="h-8 w-8 text-white/30" />
                    <p className="text-xs font-medium text-white/50">No messages yet</p>
                    <p className="text-[11px] text-white/40">Pick a user and say hi</p>
                  </div>
                ) : (
                  <div className="space-y-2">
                    {chatMessages.slice(0, 8).map((msg) => (
                      <div key={`${msg.id}-${msg.createdAt}`} className="rounded-lg bg-white/10 px-3 py-2">
                        <p className="text-[10px] font-medium text-white/60">User {msg.senderId}</p>
                        <p className="mt-0.5 text-xs text-white/90">{msg.content || "(media)"}</p>
                        <p className="mt-1 text-[10px] text-white/40">{shortDate(msg.createdAt)}</p>
                      </div>
                    ))}
                  </div>
                )}
              </div>

              <div className="space-y-2 p-3">
                <div className="flex items-center gap-1.5 rounded-lg bg-white/5 px-2 py-1">
                  <span className="text-[10px] font-medium text-white/50">To</span>
                  <input
                    type="number"
                    value={chatRecipientID}
                    onChange={(event) => setChatRecipientID(event.target.value)}
                    placeholder="User ID"
                    className="w-16 rounded bg-transparent px-1.5 py-0.5 text-xs text-white placeholder:text-white/40 outline-none"
                  />
                </div>
                <div className="flex gap-2">
                  <input
                    value={chatDraft}
                    onChange={(event) => setChatDraft(event.target.value)}
                    placeholder="Type a message…"
                    className="min-w-0 flex-1 rounded-lg border border-white/20 bg-white/5 px-3 py-2 text-xs text-white placeholder:text-white/40 outline-none focus:border-white/40"
                  />
                  <button
                    type="button"
                    onClick={sendChatMessage}
                    className="flex h-9 w-9 shrink-0 items-center justify-center rounded-lg bg-white/15 text-white transition hover:bg-white/25 active:scale-95"
                    aria-label="Send message"
                  >
                    <Send className="h-4 w-4" />
                  </button>
                </div>
                {chatError ? <p className="text-[11px] text-rose-400">{chatError}</p> : null}
                {Object.keys(chatUnreadMap).length > 0 ? (
                  <p className="text-[11px] text-white/50">
                    {Object.values(chatUnreadMap).reduce((a, b) => a + b, 0)} unread
                  </p>
                ) : null}
              </div>
            </div>
          </div>

          <div className="overflow-hidden rounded-xl border border-white/10 bg-gradient-to-b from-[#2b2929]/60 to-[#2b2929]/40 shadow-lg">
            <div className="border-b border-white/10 px-4 py-3">
              <h2 className="text-sm font-semibold text-white">Trending topics</h2>
            </div>
            <div className="space-y-2 p-3">
              {trends.map((item, index) => (
                <article key={item.title} className="flex items-start gap-3 rounded-lg border border-white/10 bg-white/5 p-3">
                  <span className="flex h-6 w-6 shrink-0 items-center justify-center rounded-full bg-white/15 text-[11px] font-bold text-white/80">
                    {index + 1}
                  </span>
                  <div className="min-w-0 flex-1">
                    <p className="text-sm font-semibold text-white/90">{item.title}</p>
                    <p className="mt-0.5 text-xs text-white/60">{item.posts}</p>
                  </div>
                </article>
              ))}
            </div>
          </div>
        </aside>
        </div>
      </main>

      <Footer />

      {notificationsOpen ? (
        <div className="fixed right-4 top-16 z-50 w-[320px] rounded-xl border border-white/10 bg-[#2b2929]/95 p-3 shadow-xl backdrop-blur-md">
          <div className="mb-2 flex items-center justify-between gap-2">
            <p className="text-sm font-semibold text-white">Notifications</p>
            <div className="flex items-center gap-1">
              <button
                type="button"
                onClick={markAllNotificationsRead}
                className="text-xs font-semibold text-white/60 hover:text-white"
              >
                Mark all read
              </button>
              <span className="text-white/30">·</span>
              <button
                type="button"
                onClick={() => setNotificationsOpen(false)}
                className="text-xs font-semibold text-white/60 hover:text-white"
              >
                Close
              </button>
            </div>
          </div>
          <div className="max-h-64 space-y-2 overflow-y-auto">
            {notificationsLoading ? (
              <p className="py-4 text-center text-xs text-white/50">Loading...</p>
            ) : notifications.length === 0 ? (
              <p className="py-4 text-center text-xs text-white/50">No notifications yet.</p>
            ) : (
              notifications.map((item) => (
                <button
                  type="button"
                  key={item.id}
                  onClick={() => markNotificationRead(item.id)}
                  className={`w-full rounded-sm border px-3 py-2 text-left text-xs ${
                    item.isRead
                      ? "border-white/10 bg-white/5 text-white/50"
                      : "border-brand-primary/30 bg-brand-primary/15 text-brand-primary"
                  }`}
                >
                  <p className="font-semibold">{item.title}</p>
                  <p className="mt-0.5 text-[11px]">{item.subtitle}</p>
                </button>
              ))
            )}
          </div>
        </div>
      ) : null}
    </div>
  );
}
