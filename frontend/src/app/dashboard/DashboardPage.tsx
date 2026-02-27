"use client";
/* eslint-disable @next/next/no-img-element */

import { useEffect, useMemo, useRef, useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { MessageCircle, User, ThumbsDown, ThumbsUp, Plus, Send, Wifi, WifiOff } from "lucide-react";
import { motion } from "framer-motion";
import { useAuth } from "@/components/AuthContext";
import { fadeUp, viewportOnce } from "@/components/Motion";
import TopNav from "@/components/TopNav";
import LeftNav from "@/components/LeftNav";
import { useNotifications } from "@/components/NotificationsContext";

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

type Post = {
  id: number;
  author_id: number;
  group_id?: number | null;
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

type PostPrivacy = "public" | "followers" | "private";

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

function normalizePrivacy(value?: string | null): PostPrivacy {
  if (value === "followers" || value === "private") return value;
  return "public";
}

function toMediaUrl(apiBaseUrl: string, path?: string | null) {
  if (!path) return "";
  if (path.startsWith("http://") || path.startsWith("https://")) return path;
  const normalized = path.startsWith("/") ? path : `/${path}`;
  return `${apiBaseUrl}${normalized}`;
}

function notificationActorName(item: NotificationItem) {
  const meta = item.metadata ?? {};
  const requester = meta["requester_name"];
  if (typeof requester === "string" && requester.trim()) return requester;
  return "Someone";
}

function notificationGroupName(item: NotificationItem) {
  const meta = item.metadata ?? {};
  const groupName = meta["group_name"];
  if (typeof groupName === "string" && groupName.trim()) return groupName;
  return "your group";
}

function notificationTitle(item: NotificationItem) {
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

function notificationBody(item: NotificationItem) {
  switch (item.type) {
    case "follow_request":
      return `${notificationActorName(item)} sent you a follow request.`;
    case "group_invitation":
      return `${notificationActorName(item)} invited you to ${notificationGroupName(item)}.`;
    case "group_join_request":
      return `${notificationActorName(item)} requested to join ${notificationGroupName(item)}.`;
    case "event_created":
      return `New event in ${notificationGroupName(item)}.`;
    default:
      return "Notification update.";
  }
}

export default function DashboardPage() {
  const router = useRouter();
  const { logout } = useAuth();
  const notificationsContext = useNotifications();

  const [user, setUser] = useState<User | null>(null);
  const [posts, setPosts] = useState<Post[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [feedError, setFeedError] = useState<string | null>(null);
  const [composerText, setComposerText] = useState("");
  const [composerFile, setComposerFile] = useState<File | null>(null);
  const [composerFileName, setComposerFileName] = useState("");
  const [composerPrivacy, setComposerPrivacy] = useState<PostPrivacy>("public");
  const [composerAllowedIDs, setComposerAllowedIDs] = useState<number[]>([]);
  const [composerError, setComposerError] = useState<string | null>(null);
  const [isPosting, setIsPosting] = useState(false);
  const [groupsOnly, setGroupsOnly] = useState(false);
  const [postReactionMap, setPostReactionMap] = useState<ReactionMap>({});
  const [commentReactionMap, setCommentReactionMap] = useState<ReactionMap>({});
  const [commentsByPost, setCommentsByPost] = useState<Record<number, Comment[]>>({});
  const [commentsOpenByPost, setCommentsOpenByPost] = useState<Record<number, boolean>>({});
  const [commentsLoadingByPost, setCommentsLoadingByPost] = useState<Record<number, boolean>>({});
  const [commentDraftByPost, setCommentDraftByPost] = useState<Record<number, string>>({});
  const [commentFileByPost, setCommentFileByPost] = useState<Record<number, File | null>>({});
  const [commentFileNameByPost, setCommentFileNameByPost] = useState<Record<number, string>>({});
  const [commentErrorByPost, setCommentErrorByPost] = useState<Record<number, string>>({});
  const [chatRecipientID, setChatRecipientID] = useState("");
  const [chatDraft, setChatDraft] = useState("");
  const [chatMessages, setChatMessages] = useState<ChatMessage[]>([]);
  const [chatError, setChatError] = useState<string | null>(null);
  const [chatConnected, setChatConnected] = useState(false);
  const [chatUnreadMap, setChatUnreadMap] = useState<Record<number, number>>({});
  const [followers, setFollowers] = useState<UserListItem[]>([]);
  const [followersLoading, setFollowersLoading] = useState(false);
  const [editingPostID, setEditingPostID] = useState<number | null>(null);
  const [editPostText, setEditPostText] = useState("");
  const [editPostFile, setEditPostFile] = useState<File | null>(null);
  const [editPostFileName, setEditPostFileName] = useState("");
  const [editPostClearMedia, setEditPostClearMedia] = useState(false);
  const [editPostPrivacy, setEditPostPrivacy] = useState<PostPrivacy>("public");
  const [editPostAllowedIDs, setEditPostAllowedIDs] = useState<number[]>([]);
  const [editPostError, setEditPostError] = useState<string | null>(null);
  const [editingCommentID, setEditingCommentID] = useState<number | null>(null);
  const [editCommentText, setEditCommentText] = useState("");
  const [editCommentFile, setEditCommentFile] = useState<File | null>(null);
  const [editCommentFileName, setEditCommentFileName] = useState("");
  const [editCommentClearMedia, setEditCommentClearMedia] = useState(false);
  const [editCommentError, setEditCommentError] = useState<string | null>(null);
  const [isLoadingMore, setIsLoadingMore] = useState(false);
  const [hasMorePosts, setHasMorePosts] = useState(true);
  const wsRef = useRef<WebSocket | null>(null);
  const feedLimit = 20;

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

  const notifications = notificationsContext?.notifications ?? [];
  const notificationsLoading = notificationsContext?.loading ?? false;
  const markNotificationRead =
    notificationsContext?.markRead ?? (async () => Promise.resolve());
  const markAllNotificationsRead =
    notificationsContext?.markAllRead ?? (async () => Promise.resolve());


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

        setFollowersLoading(true);
        const followerList = await fetchJson<UserListItem[]>(
          `/profiles/${me.result.data?.id ?? 0}/followers`,
        );
        if (!cancelled && followerList.response.ok && followerList.result?.success) {
          setFollowers(followerList.result.data ?? []);
        }
        if (!cancelled) {
          setFollowersLoading(false);
        }

        const feedPath = groupsOnly
          ? `/posts?groups_only=true&limit=${feedLimit}&offset=0`
          : `/posts?limit=${feedLimit}&offset=0`;
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
          setHasMorePosts(nextPosts.length >= feedLimit);

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
          setFollowersLoading(false);
        }
      }
    };

    load();

    return () => {
      cancelled = true;
    };
  }, [apiBaseUrl, router, groupsOnly]);

  const loadMorePosts = async () => {
    if (isLoadingMore || !hasMorePosts) return;
    setIsLoadingMore(true);
    setFeedError(null);
    try {
      const offset = posts.length;
      const feedPath = groupsOnly
        ? `/posts?groups_only=true&limit=${feedLimit}&offset=${offset}`
        : `/posts?limit=${feedLimit}&offset=${offset}`;
      const response = await fetch(`${apiBaseUrl}${feedPath}`, {
        credentials: "include",
      });
      const result = (await response.json().catch(() => null)) as ApiResponse<Post[]> | null;
      if (!response.ok || !result?.success) {
        setFeedError(result?.error || "Could not load more posts.");
        return;
      }
      const nextPosts = result.data ?? [];
      setPosts((prev) => [...prev, ...nextPosts]);
      setHasMorePosts(nextPosts.length >= feedLimit);
    } finally {
      setIsLoadingMore(false);
    }
  };

  useEffect(() => {
    if (!user?.id) {
      return;
    }

    const intervalID = window.setInterval(async () => {
      const feedPath = groupsOnly
        ? `/posts?groups_only=true&limit=${feedLimit}&offset=0`
        : `/posts?limit=${feedLimit}&offset=0`;
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
            if (notificationsContext) {
              void notificationsContext.refreshNotifications();
              void notificationsContext.refreshUnreadCount();
            }
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
    if (!content && !composerFile) {
      setComposerError("Write something or attach media before posting.");
      return;
    }
    if (composerPrivacy === "private" && composerAllowedIDs.length === 0) {
      setComposerError("Select at least one follower for a private post.");
      return;
    }

    setIsPosting(true);
    setComposerError(null);

    try {
      let mediaPath: string | undefined;
      if (composerFile) {
        const formData = new FormData();
        formData.append("file", composerFile);
        formData.append("kind", "post");
        const uploadRes = await fetch(`${apiBaseUrl}/uploads`, {
          method: "POST",
          credentials: "include",
          body: formData,
        });
        const uploadJson = (await uploadRes.json().catch(() => null)) as
          | ApiResponse<{ path?: string }>
          | null;
        if (!uploadRes.ok || !uploadJson?.success || !uploadJson.data?.path) {
          setComposerError(uploadJson?.error || "Could not upload media.");
          return;
        }
        mediaPath = uploadJson.data.path;
      }

      const response = await fetch(`${apiBaseUrl}/posts`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        credentials: "include",
        body: JSON.stringify({
          content: content || undefined,
          media_path: mediaPath,
          privacy: composerPrivacy,
          allowed_user_ids: composerPrivacy === "private" ? composerAllowedIDs : undefined,
        }),
      });

      const result = (await response.json().catch(() => null)) as ApiResponse<Post> | null;
      if (!response.ok || !result?.success || !result.data) {
        setComposerError(result?.error || "Could not publish your post.");
        return;
      }

      setPosts((prev) => [result.data as Post, ...prev]);
      setComposerText("");
      setComposerFile(null);
      setComposerFileName("");
      setComposerAllowedIDs([]);
      setComposerPrivacy("public");
    } catch {
      setComposerError("Network error. Please try again.");
    } finally {
      setIsPosting(false);
    }
  };

  const uploadMedia = async (file: File, kind: "post" | "comment") => {
    const formData = new FormData();
    formData.append("file", file);
    formData.append("kind", kind);
    const uploadRes = await fetch(`${apiBaseUrl}/uploads`, {
      method: "POST",
      credentials: "include",
      body: formData,
    });
    const uploadJson = (await uploadRes.json().catch(() => null)) as
      | ApiResponse<{ path?: string }>
      | null;
    if (!uploadRes.ok || !uploadJson?.success || !uploadJson.data?.path) {
      throw new Error(uploadJson?.error || "Upload failed");
    }
    return uploadJson.data.path;
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
      setChatError("Pick a recipient and write a message.");
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
      const feedPath = groupsOnly
        ? `/posts?groups_only=true&limit=${feedLimit}&offset=0`
        : `/posts?limit=${feedLimit}&offset=0`;
      const response = await fetch(`${apiBaseUrl}${feedPath}`, {
        credentials: "include",
      });
      const result = (await response.json().catch(() => null)) as ApiResponse<Post[]> | null;
      if (!response.ok || !result?.success) {
        setFeedError(result?.error || "Could not refresh feed.");
        return;
      }
      setPosts(result.data ?? []);
      setHasMorePosts((result.data ?? []).length >= feedLimit);
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
    const attachment = commentFileByPost[postID] ?? null;
    if (!draft && !attachment) {
      setCommentErrorByPost((prev) => ({
        ...prev,
        [postID]: "Write a comment or attach media before posting.",
      }));
      return;
    }

    try {
      let mediaPath: string | undefined;
      if (attachment) {
        const formData = new FormData();
        formData.append("file", attachment);
        formData.append("kind", "comment");
        const uploadRes = await fetch(`${apiBaseUrl}/uploads`, {
          method: "POST",
          credentials: "include",
          body: formData,
        });
        const uploadJson = (await uploadRes.json().catch(() => null)) as
          | ApiResponse<{ path?: string }>
          | null;
        if (!uploadRes.ok || !uploadJson?.success || !uploadJson.data?.path) {
          setCommentErrorByPost((prev) => ({
            ...prev,
            [postID]: uploadJson?.error || "Could not upload comment media.",
          }));
          return;
        }
        mediaPath = uploadJson.data.path;
      }

      const response = await fetch(`${apiBaseUrl}/posts/${postID}/comments`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        credentials: "include",
        body: JSON.stringify({ content: draft || undefined, media_path: mediaPath }),
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
      setCommentFileByPost((prev) => ({ ...prev, [postID]: null }));
      setCommentFileNameByPost((prev) => ({ ...prev, [postID]: "" }));
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

  const startEditPost = (post: Post) => {
    setEditingPostID(post.id);
    setEditPostText(post.content || "");
    setEditPostFile(null);
    setEditPostFileName("");
    setEditPostClearMedia(false);
    setEditPostPrivacy(normalizePrivacy(post.privacy));
    setEditPostAllowedIDs([]);
    setEditPostError(null);
  };

  const cancelEditPost = () => {
    setEditingPostID(null);
    setEditPostText("");
    setEditPostFile(null);
    setEditPostFileName("");
    setEditPostClearMedia(false);
    setEditPostPrivacy("public");
    setEditPostAllowedIDs([]);
    setEditPostError(null);
  };

  const saveEditPost = async (post: Post) => {
    const isGroupPost = post.group_id != null;
    const content = editPostText.trim();
    if (!content && !editPostFile && !post.media_path && !editPostClearMedia) {
      setEditPostError("Content or media is required.");
      return;
    }
    if (!isGroupPost && editPostPrivacy === "private" && editPostAllowedIDs.length === 0) {
      setEditPostError("Select at least one follower for a private post.");
      return;
    }
    setEditPostError(null);
    try {
      let mediaPath: string | undefined;
      if (editPostClearMedia && !editPostFile) {
        mediaPath = "";
      }
      if (editPostFile) {
        mediaPath = await uploadMedia(editPostFile, "post");
      }
      const response = await fetch(`${apiBaseUrl}/posts/${post.id}`, {
        method: "PATCH",
        headers: { "Content-Type": "application/json" },
        credentials: "include",
        body: JSON.stringify({
          content: content || undefined,
          media_path: mediaPath ?? undefined,
          privacy: isGroupPost ? undefined : editPostPrivacy,
          allowed_user_ids:
            !isGroupPost && editPostPrivacy === "private" ? editPostAllowedIDs : undefined,
        }),
      });
      const result = (await response.json().catch(() => null)) as ApiResponse<Post> | null;
      if (!response.ok || !result?.success || !result.data) {
        setEditPostError(result?.error || "Could not update post.");
        return;
      }
      setPosts((prev) => prev.map((item) => (item.id === post.id ? result.data as Post : item)));
      cancelEditPost();
    } catch (err) {
      setEditPostError(err instanceof Error ? err.message : "Network error.");
    }
  };

  const deletePost = async (postID: number) => {
    try {
      const response = await fetch(`${apiBaseUrl}/posts/${postID}`, {
        method: "DELETE",
        credentials: "include",
      });
      if (!response.ok) {
        return;
      }
      setPosts((prev) => prev.filter((post) => post.id !== postID));
    } catch {
      // ignore
    }
  };

  const startEditComment = (comment: Comment) => {
    setEditingCommentID(comment.id);
    setEditCommentText(comment.content || "");
    setEditCommentFile(null);
    setEditCommentFileName("");
    setEditCommentClearMedia(false);
    setEditCommentError(null);
  };

  const cancelEditComment = () => {
    setEditingCommentID(null);
    setEditCommentText("");
    setEditCommentFile(null);
    setEditCommentFileName("");
    setEditCommentClearMedia(false);
    setEditCommentError(null);
  };

  const saveEditComment = async (postID: number, comment: Comment) => {
    const content = editCommentText.trim();
    if (!content && !editCommentFile && !comment.media_path && !editCommentClearMedia) {
      setEditCommentError("Content or media is required.");
      return;
    }
    setEditCommentError(null);
    try {
      let mediaPath: string | undefined;
      if (editCommentClearMedia && !editCommentFile) {
        mediaPath = "";
      }
      if (editCommentFile) {
        mediaPath = await uploadMedia(editCommentFile, "comment");
      }
      const response = await fetch(`${apiBaseUrl}/comments/${comment.id}`, {
        method: "PATCH",
        headers: { "Content-Type": "application/json" },
        credentials: "include",
        body: JSON.stringify({
          content: content || undefined,
          media_path: mediaPath ?? undefined,
        }),
      });
      const result = (await response.json().catch(() => null)) as ApiResponse<Comment> | null;
      if (!response.ok || !result?.success || !result.data) {
        setEditCommentError(result?.error || "Could not update comment.");
        return;
      }
      setCommentsByPost((prev) => ({
        ...prev,
        [postID]: (prev[postID] ?? []).map((item) =>
          item.id === comment.id ? (result.data as Comment) : item,
        ),
      }));
      cancelEditComment();
    } catch (err) {
      setEditCommentError(err instanceof Error ? err.message : "Network error.");
    }
  };

  const deleteComment = async (postID: number, commentID: number) => {
    try {
      const response = await fetch(`${apiBaseUrl}/comments/${commentID}`, {
        method: "DELETE",
        credentials: "include",
      });
      if (!response.ok) {
        return;
      }
      setCommentsByPost((prev) => ({
        ...prev,
        [postID]: (prev[postID] ?? []).filter((item) => item.id !== commentID),
      }));
      setPosts((prev) =>
        prev.map((post) =>
          post.id === postID ? { ...post, comment_count: Math.max(0, post.comment_count - 1) } : post,
        ),
      );
    } catch {
      // ignore
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
      <TopNav
        user={user ?? undefined}
        onLogout={handleLogout}
      />

      <main className="mx-auto grid w-full max-w-6xl gap-6 px-4 py-6 sm:px-6 lg:grid-cols-[240px_minmax(0,1fr)_280px]">
        <aside className="hidden lg:block">
          <LeftNav user={user ?? undefined} activeHref="/dashboard" />
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
                  className="rounded-full border border-neutral-200 bg-white px-3 py-1.5 text-xs font-semibold text-neutral-700 transition hover:border-neutral-400 hover:text-neutral-900"
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
            <div className="mt-3 flex flex-wrap items-center gap-3">
              <label className="text-xs font-semibold text-neutral-600">
                Privacy
                <select
                  value={composerPrivacy}
                  onChange={(event) => {
                    const next = event.target.value as PostPrivacy;
                    setComposerPrivacy(next);
                    if (next !== "private") {
                      setComposerAllowedIDs([]);
                    }
                  }}
                  className="mt-2 h-10 w-full rounded-2xl border border-neutral-200 bg-white px-3 text-xs text-neutral-700 outline-none focus:border-neutral-400 sm:w-48"
                >
                  <option value="public">Public</option>
                  <option value="followers">Followers</option>
                  <option value="private">Private (select followers)</option>
                </select>
              </label>
              {composerPrivacy === "private" ? (
                <div className="flex flex-1 flex-wrap gap-2 rounded-2xl border border-neutral-200 bg-white px-3 py-2 text-xs text-neutral-600">
                  {followersLoading ? (
                    <span className="text-xs text-neutral-500">Loading followers...</span>
                  ) : followers.length === 0 ? (
                    <span className="text-xs text-neutral-500">No followers available.</span>
                  ) : (
                    followers.map((follower) => {
                      const checked = composerAllowedIDs.includes(follower.id);
                      return (
                        <label key={follower.id} className="inline-flex items-center gap-2">
                          <input
                            type="checkbox"
                            checked={checked}
                            onChange={(event) => {
                              const nextChecked = event.target.checked;
                              setComposerAllowedIDs((prev) =>
                                nextChecked
                                  ? [...prev, follower.id]
                                  : prev.filter((id) => id !== follower.id),
                              );
                            }}
                            className="h-4 w-4 rounded border-neutral-300 text-neutral-900 focus:ring-neutral-900"
                          />
                          <span>@{follower.nickname || "user"}</span>
                        </label>
                      );
                    })
                  )}
                </div>
              ) : null}
            </div>
            <div className="mt-3 flex flex-wrap items-center justify-between gap-3">
              <label className="inline-flex items-center gap-2 rounded-full border border-neutral-200 bg-white px-3 py-2 text-xs font-semibold text-neutral-700 transition hover:border-neutral-400 hover:text-neutral-900">
                <input
                  type="file"
                  accept="image/png,image/jpeg,image/gif"
                  className="hidden"
                  onChange={(event) => {
                    const file = event.target.files?.[0] ?? null;
                    setComposerFile(file);
                    setComposerFileName(file?.name ?? "");
                  }}
                />
                <Plus className="h-3.5 w-3.5" />
                {composerFileName ? "Change media" : "Add media"}
              </label>
              {composerFileName ? (
                <span className="text-xs text-neutral-500">{composerFileName}</span>
              ) : null}
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
                placeholder="Recipient"
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

                  {editingPostID === post.id ? (
                    <div className="mt-4 space-y-3">
                      <textarea
                        value={editPostText}
                        onChange={(event) => setEditPostText(event.target.value)}
                        rows={3}
                        className="w-full rounded-2xl border border-neutral-200 bg-neutral-50 px-4 py-3 text-sm outline-none focus:border-neutral-400"
                      />
                      {post.group_id == null ? (
                        <div className="flex flex-wrap items-start gap-3">
                          <label className="text-xs font-semibold text-neutral-600">
                            Privacy
                            <select
                              value={editPostPrivacy}
                              onChange={(event) => {
                                const next = event.target.value as PostPrivacy;
                                setEditPostPrivacy(next);
                                if (next !== "private") {
                                  setEditPostAllowedIDs([]);
                                }
                              }}
                              className="mt-2 h-10 w-full rounded-2xl border border-neutral-200 bg-white px-3 text-xs text-neutral-700 outline-none focus:border-neutral-400 sm:w-48"
                            >
                              <option value="public">Public</option>
                              <option value="followers">Followers</option>
                              <option value="private">Private (select followers)</option>
                            </select>
                          </label>
                          {editPostPrivacy === "private" ? (
                            <div className="flex flex-1 flex-wrap gap-2 rounded-2xl border border-neutral-200 bg-white px-3 py-2 text-xs text-neutral-600">
                              {followersLoading ? (
                                <span className="text-xs text-neutral-500">
                                  Loading followers...
                                </span>
                              ) : followers.length === 0 ? (
                                <span className="text-xs text-neutral-500">
                                  No followers available.
                                </span>
                              ) : (
                                followers.map((follower) => {
                                  const checked = editPostAllowedIDs.includes(follower.id);
                                  return (
                                    <label
                                      key={follower.id}
                                      className="inline-flex items-center gap-2"
                                    >
                                      <input
                                        type="checkbox"
                                        checked={checked}
                                        onChange={(event) => {
                                          const nextChecked = event.target.checked;
                                          setEditPostAllowedIDs((prev) =>
                                            nextChecked
                                              ? [...prev, follower.id]
                                              : prev.filter((id) => id !== follower.id),
                                          );
                                        }}
                                        className="h-4 w-4 rounded border-neutral-300 text-neutral-900 focus:ring-neutral-900"
                                      />
                                      <span>@{follower.nickname || "user"}</span>
                                    </label>
                                  );
                                })
                              )}
                            </div>
                          ) : null}
                        </div>
                      ) : (
                        <p className="text-xs text-neutral-500">
                          Group post privacy cannot be changed.
                        </p>
                      )}
                      <div className="flex flex-wrap items-center gap-2">
                        <label className="inline-flex items-center gap-2 rounded-full border border-neutral-200 bg-white px-3 py-2 text-xs font-semibold text-neutral-700 transition hover:border-neutral-400 hover:text-neutral-900">
                          <input
                              type="file"
                              accept="image/png,image/jpeg,image/gif"
                              className="hidden"
                              onChange={(event) => {
                                const file = event.target.files?.[0] ?? null;
                                setEditPostFile(file);
                                setEditPostFileName(file?.name ?? "");
                                setEditPostClearMedia(false);
                              }}
                            />
                            Change media
                          </label>
                          {editPostFileName ? (
                            <span className="text-xs text-neutral-500">{editPostFileName}</span>
                          ) : null}
                          {post.media_path ? (
                            <button
                              type="button"
                              onClick={() => {
                                setEditPostClearMedia(true);
                                setEditPostFile(null);
                                setEditPostFileName("");
                              }}
                              className={`rounded-full border px-3 py-2 text-xs font-semibold transition ${
                                editPostClearMedia
                                  ? "border-rose-200 bg-rose-50 text-rose-700"
                                  : "border-neutral-200 bg-white text-neutral-700 hover:border-neutral-400"
                              }`}
                            >
                              {editPostClearMedia ? "Media removed" : "Remove media"}
                            </button>
                          ) : null}
                          <button
                            type="button"
                            onClick={() => saveEditPost(post)}
                            className="rounded-full bg-neutral-900 px-3 py-2 text-xs font-semibold text-white"
                        >
                          Save
                        </button>
                        <button
                          type="button"
                          onClick={cancelEditPost}
                          className="rounded-full border border-neutral-200 bg-white px-3 py-2 text-xs font-semibold text-neutral-700"
                        >
                          Cancel
                        </button>
                      </div>
                      {editPostError ? <p className="text-xs text-rose-600">{editPostError}</p> : null}
                    </div>
                  ) : (
                    <p className="mt-4 text-sm leading-relaxed text-neutral-700">{post.content}</p>
                  )}

                  {post.media_path ? (
                    <div className="mt-4 overflow-hidden rounded-2xl border border-neutral-200">
                      <img
                        src={toMediaUrl(apiBaseUrl, post.media_path)}
                        alt="Post media"
                        className="max-h-[520px] w-full object-contain bg-white"
                      />
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
                    {user?.id === post.author_id ? (
                      <>
                        <button
                          type="button"
                          onClick={() => startEditPost(post)}
                          className="inline-flex items-center gap-1 rounded-full bg-neutral-100 px-2 py-1 text-neutral-600 transition hover:bg-neutral-200"
                        >
                          Edit
                        </button>
                        <button
                          type="button"
                          onClick={() => deletePost(post.id)}
                          className="inline-flex items-center gap-1 rounded-full bg-neutral-100 px-2 py-1 text-neutral-600 transition hover:bg-neutral-200"
                        >
                          Delete
                        </button>
                      </>
                    ) : null}
                  </footer>

                  {commentsOpenByPost[post.id] ? (
                    <section className="mt-4 rounded-2xl border border-neutral-200 bg-neutral-50 p-3">
                      <div className="space-y-2">
                        {(commentsByPost[post.id] ?? []).map((comment) => (
                          <article key={comment.id} className="rounded-xl bg-white p-3">
                            {editingCommentID === comment.id ? (
                              <div className="space-y-2">
                                <textarea
                                  value={editCommentText}
                                  onChange={(event) => setEditCommentText(event.target.value)}
                                  rows={2}
                                  className="w-full rounded-xl border border-neutral-200 bg-neutral-50 px-3 py-2 text-xs outline-none focus:border-neutral-400"
                                />
                                <div className="flex flex-wrap items-center gap-2">
                                  <label className="inline-flex items-center gap-2 rounded-full border border-neutral-200 bg-white px-3 py-1 text-[11px] font-semibold text-neutral-700 transition hover:border-neutral-400 hover:text-neutral-900">
                                    <input
                                      type="file"
                                      accept="image/png,image/jpeg,image/gif"
                                      className="hidden"
                                      onChange={(event) => {
                                        const file = event.target.files?.[0] ?? null;
                                        setEditCommentFile(file);
                                        setEditCommentFileName(file?.name ?? "");
                                        setEditCommentClearMedia(false);
                                      }}
                                    />
                                    Change media
                                  </label>
                                  {editCommentFileName ? (
                                    <span className="text-[11px] text-neutral-500">
                                      {editCommentFileName}
                                    </span>
                                  ) : null}
                                  {comment.media_path ? (
                                    <button
                                      type="button"
                                      onClick={() => {
                                        setEditCommentClearMedia(true);
                                        setEditCommentFile(null);
                                        setEditCommentFileName("");
                                      }}
                                      className={`rounded-full border px-3 py-1 text-[11px] font-semibold transition ${
                                        editCommentClearMedia
                                          ? "border-rose-200 bg-rose-50 text-rose-700"
                                          : "border-neutral-200 bg-white text-neutral-700 hover:border-neutral-400"
                                      }`}
                                    >
                                      {editCommentClearMedia ? "Media removed" : "Remove media"}
                                    </button>
                                  ) : null}
                                  <button
                                    type="button"
                                    onClick={() => saveEditComment(post.id, comment)}
                                    className="rounded-full bg-neutral-900 px-3 py-1 text-[11px] font-semibold text-white"
                                  >
                                    Save
                                  </button>
                                  <button
                                    type="button"
                                    onClick={cancelEditComment}
                                    className="rounded-full border border-neutral-200 bg-white px-3 py-1 text-[11px] font-semibold text-neutral-700"
                                  >
                                    Cancel
                                  </button>
                                </div>
                                {editCommentError ? (
                                  <p className="text-[11px] text-rose-600">{editCommentError}</p>
                                ) : null}
                              </div>
                            ) : (
                              <>
                                <p className="text-sm text-neutral-700">{comment.content}</p>
                                {comment.media_path ? (
                                  <div className="mt-2 overflow-hidden rounded-xl border border-neutral-200 bg-white">
                                    <img
                                      src={toMediaUrl(apiBaseUrl, comment.media_path)}
                                      alt="Comment media"
                                      className="max-h-64 w-full object-contain bg-white"
                                    />
                                  </div>
                                ) : null}
                              </>
                            )}
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
                              {user?.id === comment.author_id ? (
                                <>
                                  <button
                                    type="button"
                                    onClick={() => startEditComment(comment)}
                                    className="inline-flex items-center gap-1 rounded-full bg-neutral-100 px-2 py-1 text-neutral-600"
                                  >
                                    Edit
                                  </button>
                                  <button
                                    type="button"
                                    onClick={() => deleteComment(post.id, comment.id)}
                                    className="inline-flex items-center gap-1 rounded-full bg-neutral-100 px-2 py-1 text-neutral-600"
                                  >
                                    Delete
                                  </button>
                                </>
                              ) : null}
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
                        <label className="inline-flex h-9 items-center gap-2 rounded-xl border border-neutral-200 bg-white px-3 text-xs font-semibold text-neutral-700 transition hover:border-neutral-400 hover:text-neutral-900">
                          <input
                            type="file"
                            accept="image/png,image/jpeg,image/gif"
                            className="hidden"
                            onChange={(event) => {
                              const file = event.target.files?.[0] ?? null;
                              setCommentFileByPost((prev) => ({ ...prev, [post.id]: file }));
                              setCommentFileNameByPost((prev) => ({
                                ...prev,
                                [post.id]: file?.name ?? "",
                              }));
                            }}
                          />
                          <Plus className="h-3.5 w-3.5" />
                        </label>
                        <button
                          type="button"
                          onClick={() => handleCreateComment(post.id)}
                          className="rounded-xl bg-neutral-900 px-3 text-xs font-semibold text-white"
                        >
                          Comment
                        </button>
                      </div>
                      {commentFileNameByPost[post.id] ? (
                        <p className="mt-2 text-[11px] text-neutral-500">
                          Attached: {commentFileNameByPost[post.id]}
                        </p>
                      ) : null}
                    </section>
                  ) : null}
                </article>
              ))
            )}
          </div>
          {hasMorePosts && !isLoading && !feedError ? (
            <div className="mt-5">
              <button
                type="button"
                onClick={loadMorePosts}
                disabled={isLoadingMore}
                className="inline-flex items-center gap-2 rounded-full border border-neutral-200 bg-white px-4 py-2 text-xs font-semibold text-neutral-700 transition hover:border-neutral-400 hover:text-neutral-900 disabled:cursor-not-allowed disabled:opacity-70"
              >
                {isLoadingMore ? "Loading..." : "Load more posts"}
              </button>
            </div>
          ) : null}
        </section>

        <aside className="hidden space-y-5 md:block">
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
                        : "border-emerald-400 bg-emerald-50 text-emerald-900 shadow-[0_0_0_1px_rgba(16,185,129,0.25)]"
                    }`}
                  >
                    <p className="font-semibold">{notificationTitle(item)}</p>
                    <p className="mt-0.5 text-[11px]">{notificationBody(item)}</p>
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
                placeholder="Recipient"
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

    </div>
  );
}
