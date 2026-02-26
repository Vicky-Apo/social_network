"use client";
/* eslint-disable @next/next/no-img-element */

import { useEffect, useMemo, useState } from "react";
import Link from "next/link";
import { useParams, useRouter } from "next/navigation";
import {
  ArrowLeft,
  ArrowRight,
  Calendar,
  MessageCircle,
  RefreshCw,
  Send,
  Shield,
  ThumbsDown,
  ThumbsUp,
  Users,
} from "lucide-react";
import { motion } from "framer-motion";
import { fadeUp, viewportOnce } from "@/components/Motion";

type ApiResponse<T> = {
  success?: boolean;
  data?: T;
  error?: string;
};

type GroupDetail = {
  id: number;
  name: string;
  description: string;
  creatorID?: number;
  memberCount: number;
  privacy: "public" | "private" | "unknown";
  createdAt?: string;
  updatedAt?: string;
};

type Post = {
  id: number;
  author_id: number;
  author_first_name: string;
  author_last_name: string;
  content: string;
  media_path?: string | null;
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

function toNumber(value: unknown): number | null {
  const parsed = Number(value);
  return Number.isFinite(parsed) ? parsed : null;
}

function formatDate(value?: string) {
  if (!value) return "N/A";
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return "N/A";
  return date.toLocaleDateString(undefined, { month: "short", day: "numeric", year: "numeric" });
}

function shortDate(value: string) {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return "Just now";
  }
  return date.toLocaleDateString(undefined, { month: "short", day: "numeric" });
}

function toMediaUrl(apiBaseUrl: string, path?: string | null) {
  if (!path) return "";
  if (path.startsWith("http://") || path.startsWith("https://")) return path;
  const normalized = path.startsWith("/") ? path : `/${path}`;
  return `${apiBaseUrl}${normalized}`;
}

function parseGroup(data: unknown): GroupDetail | null {
  if (!data || typeof data !== "object") {
    return null;
  }

  const root = data as Record<string, unknown>;
  const source =
    root.group && typeof root.group === "object"
      ? (root.group as Record<string, unknown>)
      : root;

  const id = toNumber(source.id);
  if (!id || id <= 0) {
    return null;
  }

  const nameRaw = source.title ?? source.name;
  const name = typeof nameRaw === "string" && nameRaw.trim() ? nameRaw.trim() : `Group ${id}`;
  const descriptionRaw = source.description ?? source.about;
  const description =
    typeof descriptionRaw === "string" && descriptionRaw.trim()
      ? descriptionRaw.trim()
      : "No group description yet.";
  const creatorID = toNumber(source.creator_id ?? source.creatorID) ?? undefined;
  const memberCount =
    toNumber(source.members_count ?? source.member_count ?? source.membersCount) ?? 0;
  const privacyText = String(source.privacy ?? "").toLowerCase();
  const privacy: GroupDetail["privacy"] = privacyText.includes("private")
    ? "private"
    : privacyText.includes("public")
      ? "public"
      : "unknown";

  const createdAtRaw = source.created_at ?? source.createdAt;
  const updatedAtRaw = source.updated_at ?? source.updatedAt;

  return {
    id,
    name,
    description,
    creatorID,
    memberCount: Math.max(0, memberCount),
    privacy,
    createdAt: typeof createdAtRaw === "string" ? createdAtRaw : undefined,
    updatedAt: typeof updatedAtRaw === "string" ? updatedAtRaw : undefined,
  };
}

export default function GroupDetailsPage() {
  const router = useRouter();
  const params = useParams<{ id: string }>();
  const groupID = typeof params?.id === "string" ? params.id : "";
  const groupIDNumber = Number(groupID);
  const [group, setGroup] = useState<GroupDetail | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [userID, setUserID] = useState<number | null>(null);
  const [posts, setPosts] = useState<Post[]>([]);
  const [postsLoading, setPostsLoading] = useState(true);
  const [postsError, setPostsError] = useState<string | null>(null);
  const [isLoadingMore, setIsLoadingMore] = useState(false);
  const [hasMorePosts, setHasMorePosts] = useState(true);
  const [pageSize] = useState(8);
  const [composerText, setComposerText] = useState("");
  const [mediaUrl, setMediaUrl] = useState("");
  const [composerFile, setComposerFile] = useState<File | null>(null);
  const [composerFileName, setComposerFileName] = useState("");
  const [composerError, setComposerError] = useState<string | null>(null);
  const [isPosting, setIsPosting] = useState(false);
  const [postReactionMap, setPostReactionMap] = useState<ReactionMap>({});
  const [commentReactionMap, setCommentReactionMap] = useState<ReactionMap>({});
  const [commentsByPost, setCommentsByPost] = useState<Record<number, Comment[]>>({});
  const [commentsOpenByPost, setCommentsOpenByPost] = useState<Record<number, boolean>>({});
  const [commentsLoadingByPost, setCommentsLoadingByPost] = useState<Record<number, boolean>>({});
  const [commentDraftByPost, setCommentDraftByPost] = useState<Record<number, string>>({});
  const [commentFileByPost, setCommentFileByPost] = useState<Record<number, File | null>>({});
  const [commentFileNameByPost, setCommentFileNameByPost] = useState<Record<number, string>>({});
  const [commentErrorByPost, setCommentErrorByPost] = useState<Record<number, string>>({});
  const [editingPostID, setEditingPostID] = useState<number | null>(null);
  const [editPostText, setEditPostText] = useState("");
  const [editPostFile, setEditPostFile] = useState<File | null>(null);
  const [editPostFileName, setEditPostFileName] = useState("");
  const [editPostClearMedia, setEditPostClearMedia] = useState(false);
  const [editPostError, setEditPostError] = useState<string | null>(null);
  const [editingCommentID, setEditingCommentID] = useState<number | null>(null);
  const [editCommentText, setEditCommentText] = useState("");
  const [editCommentFile, setEditCommentFile] = useState<File | null>(null);
  const [editCommentFileName, setEditCommentFileName] = useState("");
  const [editCommentClearMedia, setEditCommentClearMedia] = useState(false);
  const [editCommentError, setEditCommentError] = useState<string | null>(null);

  const apiBaseUrl = useMemo(
    () =>
      process.env.NEXT_PUBLIC_API_BASE_URL?.trim().replace(/\/+$/, "") ||
      "http://localhost:8080",
    [],
  );

  useEffect(() => {
    if (!Number.isFinite(groupIDNumber) || groupIDNumber <= 0) {
      setError("Invalid group id.");
      setIsLoading(false);
      setPostsLoading(false);
      return;
    }

    let cancelled = false;
    const load = async () => {
      setIsLoading(true);
      setError(null);
      setPostsLoading(true);
      setPostsError(null);
      setHasMorePosts(true);

      try {
        const meResponse = await fetch(`${apiBaseUrl}/auth/me`, {
          credentials: "include",
        });
        const meResult = (await meResponse.json().catch(() => null)) as ApiResponse<unknown> | null;
        if (!meResponse.ok || !meResult?.success) {
          if (!cancelled) {
            router.replace("/login");
          }
          return;
        }
        const meUser = meResult.data as { id?: number } | null;
        if (!cancelled) {
          setUserID(typeof meUser?.id === "number" ? meUser.id : null);
        }

        const [groupResponse, postsResponse] = await Promise.all([
          fetch(`${apiBaseUrl}/groups/${groupIDNumber}`, {
            credentials: "include",
          }),
          fetch(`${apiBaseUrl}/groups/${groupIDNumber}/posts?limit=${pageSize}&offset=0`, {
            credentials: "include",
          }),
        ]);

        const groupResult = (await groupResponse.json().catch(() => null)) as
          | ApiResponse<unknown>
          | null;
        if (!groupResponse.ok || !groupResult?.success) {
          if (!cancelled) {
            if (groupResponse.status === 404) {
              setError("Group endpoint is not available yet or this group does not exist.");
            } else {
              setError(groupResult?.error || "Could not load this group.");
            }
            setGroup(null);
          }
        } else {
          const normalized = parseGroup(groupResult.data);
          if (!normalized) {
            if (!cancelled) {
              setError("Received an unexpected group response format.");
              setGroup(null);
            }
          } else if (!cancelled) {
            setGroup(normalized);
          }
        }

        const postsResult = (await postsResponse.json().catch(() => null)) as
          | ApiResponse<Post[]>
          | null;
        if (!postsResponse.ok || !postsResult?.success) {
          if (!cancelled) {
            if (postsResponse.status === 404) {
              setPostsError("Group posts endpoint is not available yet.");
            } else {
              setPostsError(postsResult?.error || "Could not load group posts.");
            }
            setPosts([]);
          }
        } else if (!cancelled) {
          const nextPosts = postsResult.data ?? [];
          setPosts(nextPosts);
          setHasMorePosts(nextPosts.length >= pageSize);

          const currentUserID = typeof meUser?.id === "number" ? meUser.id : null;
          if (currentUserID && nextPosts.length > 0) {
            void Promise.all(
              nextPosts.map(async (post) => {
                try {
                  const reactionRes = await fetch(`${apiBaseUrl}/posts/${post.id}/reactions`, {
                    credentials: "include",
                  });
                  const reactionJson = (await reactionRes.json().catch(() => null)) as
                    | ApiResponse<Reaction[]>
                    | null;
                  if (!reactionRes.ok || !reactionJson?.success) {
                    return [post.id, null] as const;
                  }
                  const mine = (reactionJson.data ?? []).find(
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
          setError("Network error while loading group details.");
          setGroup(null);
          setPostsError("Network error while loading group posts.");
          setPosts([]);
          setHasMorePosts(false);
        }
      } finally {
        if (!cancelled) {
          setIsLoading(false);
          setPostsLoading(false);
        }
      }
    };

    void load();
    return () => {
      cancelled = true;
    };
  }, [apiBaseUrl, groupIDNumber, pageSize, router]);

  const loadMorePosts = async () => {
    if (isLoadingMore || !hasMorePosts) return;
    setIsLoadingMore(true);
    setPostsError(null);
    try {
      const offset = posts.length;
      const response = await fetch(
        `${apiBaseUrl}/groups/${groupIDNumber}/posts?limit=${pageSize}&offset=${offset}`,
        { credentials: "include" },
      );
      const result = (await response.json().catch(() => null)) as ApiResponse<Post[]> | null;
      if (!response.ok || !result?.success) {
        setPostsError(result?.error || "Could not load more posts.");
        return;
      }
      const nextPosts = result.data ?? [];
      setPosts((prev) => [...prev, ...nextPosts]);
      setHasMorePosts(nextPosts.length >= pageSize);

      if (userID && nextPosts.length > 0) {
        const entries = await Promise.all(
          nextPosts.map(async (post) => {
            try {
              const reactionRes = await fetch(`${apiBaseUrl}/posts/${post.id}/reactions`, {
                credentials: "include",
              });
              const reactionJson = (await reactionRes.json().catch(() => null)) as
                | ApiResponse<Reaction[]>
                | null;
              if (!reactionRes.ok || !reactionJson?.success) {
                return [post.id, null] as const;
              }
              const mine = (reactionJson.data ?? []).find((item) => item.user_id === userID);
              return [post.id, mine?.reaction ?? null] as const;
            } catch {
              return [post.id, null] as const;
            }
          }),
        );
        setPostReactionMap((prev) => ({ ...prev, ...Object.fromEntries(entries) }));
      }
    } finally {
      setIsLoadingMore(false);
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
      throw new Error(uploadJson?.error || "Could not upload media.");
    }
    return uploadJson.data.path;
  };

  const handleCreatePost = async () => {
    if (isPosting) return;
    const content = composerText.trim();
    const media = mediaUrl.trim();
    if (!content && !media && !composerFile) {
      setComposerError("Add a message or media before posting.");
      return;
    }

    setIsPosting(true);
    setComposerError(null);

    try {
      let mediaPath: string | undefined;
      if (composerFile) {
        mediaPath = await uploadMedia(composerFile, "post");
      }

      const response = await fetch(`${apiBaseUrl}/groups/${groupIDNumber}/posts`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        credentials: "include",
        body: JSON.stringify({
          content: content || undefined,
          media_path: mediaPath || media || undefined,
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
      setMediaUrl("");
      setComposerFile(null);
      setComposerFileName("");
    } catch {
      setComposerError("Network error. Please try again.");
    } finally {
      setIsPosting(false);
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

      if (userID && comments.length > 0) {
        const entries = await Promise.all(
          comments.map(async (comment) => {
            const reactionRes = await fetch(`${apiBaseUrl}/comments/${comment.id}/reactions`, {
              credentials: "include",
            });
            const reactionJson = (await reactionRes.json().catch(() => null)) as
              | ApiResponse<Reaction[]>
              | null;
            if (!reactionRes.ok || !reactionJson?.success) {
              return [comment.id, null] as const;
            }
            const mine = (reactionJson.data ?? []).find((item) => item.user_id === userID);
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
        try {
          mediaPath = await uploadMedia(attachment, "comment");
        } catch (err) {
          setCommentErrorByPost((prev) => ({
            ...prev,
            [postID]: err instanceof Error ? err.message : "Could not upload comment media.",
          }));
          return;
        }
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
    setEditPostError(null);
  };

  const cancelEditPost = () => {
    setEditingPostID(null);
    setEditPostText("");
    setEditPostFile(null);
    setEditPostFileName("");
    setEditPostClearMedia(false);
    setEditPostError(null);
  };

  const saveEditPost = async (post: Post) => {
    const content = editPostText.trim();
    if (!content && !editPostFile && !post.media_path && !editPostClearMedia) {
      setEditPostError("Content or media is required.");
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
        body: JSON.stringify({
          reaction,
        }),
      });
    } catch {
      setPostReactionMap((prev) => ({ ...prev, [postID]: previous }));
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
        }),
      });
    } catch {
      setCommentReactionMap((prev) => ({ ...prev, [commentID]: previous }));
    }
  };

  return (
    <div className="min-h-screen bg-neutral-50 px-4 py-10 text-neutral-900 sm:px-6">
      <main className="mx-auto w-full max-w-3xl">
        <motion.section
          initial="hidden"
          whileInView="show"
          viewport={viewportOnce}
          variants={fadeUp}
          className="rounded-3xl border border-neutral-200 bg-white p-6 shadow-sm"
        >
          <div className="flex flex-wrap items-center justify-between gap-2">
            <span className="inline-flex items-center gap-2 rounded-full border border-neutral-200 bg-neutral-50 px-3 py-1 text-xs font-semibold text-neutral-600">
              <Users className="h-3.5 w-3.5" />
              Group #{groupID}
            </span>
            <button
              type="button"
              onClick={() => window.location.reload()}
              className="inline-flex items-center gap-2 rounded-full border border-neutral-200 bg-white px-3 py-1.5 text-xs font-semibold text-neutral-700 transition hover:border-neutral-400 hover:text-neutral-900"
            >
              <RefreshCw className="h-3.5 w-3.5" />
              Refresh
            </button>
          </div>

          {isLoading ? (
            <p className="mt-4 text-sm text-neutral-600">Loading group details...</p>
          ) : error ? (
            <p className="mt-4 rounded-2xl border border-rose-200 bg-rose-50 px-4 py-3 text-sm text-rose-700">
              {error}
            </p>
          ) : group ? (
            <>
              <h1 className="mt-3 text-2xl font-semibold tracking-tight text-neutral-900">{group.name}</h1>
              <p className="mt-2 text-sm text-neutral-600">{group.description}</p>

              <div className="mt-5 grid grid-cols-1 gap-3 sm:grid-cols-3">
                <div className="rounded-2xl border border-neutral-200 bg-neutral-50 p-3">
                  <p className="text-[11px] uppercase tracking-wide text-neutral-500">Privacy</p>
                  <p className="mt-1 inline-flex items-center gap-1 text-sm font-semibold text-neutral-800">
                    <Shield className="h-3.5 w-3.5" />
                    {group.privacy}
                  </p>
                </div>
                <div className="rounded-2xl border border-neutral-200 bg-neutral-50 p-3">
                  <p className="text-[11px] uppercase tracking-wide text-neutral-500">Members</p>
                  <p className="mt-1 text-sm font-semibold text-neutral-800">{group.memberCount}</p>
                </div>
                <div className="rounded-2xl border border-neutral-200 bg-neutral-50 p-3">
                  <p className="text-[11px] uppercase tracking-wide text-neutral-500">Creator ID</p>
                  <p className="mt-1 text-sm font-semibold text-neutral-800">{group.creatorID ?? "N/A"}</p>
                </div>
              </div>

              <div className="mt-4 grid grid-cols-1 gap-3 sm:grid-cols-2">
                <div className="rounded-2xl border border-neutral-200 bg-neutral-50 p-3">
                  <p className="inline-flex items-center gap-1 text-[11px] uppercase tracking-wide text-neutral-500">
                    <Calendar className="h-3.5 w-3.5" />
                    Created
                  </p>
                  <p className="mt-1 text-sm font-semibold text-neutral-800">{formatDate(group.createdAt)}</p>
                </div>
                <div className="rounded-2xl border border-neutral-200 bg-neutral-50 p-3">
                  <p className="inline-flex items-center gap-1 text-[11px] uppercase tracking-wide text-neutral-500">
                    <Calendar className="h-3.5 w-3.5" />
                    Updated
                  </p>
                  <p className="mt-1 text-sm font-semibold text-neutral-800">{formatDate(group.updatedAt)}</p>
                </div>
              </div>
            </>
          ) : (
            <p className="mt-4 text-sm text-neutral-600">Group details are not available.</p>
          )}

          <div className="mt-6 flex flex-wrap gap-3">
            <Link
              href="/groups"
              className="inline-flex items-center gap-2 rounded-full border border-neutral-200 bg-white px-4 py-2 text-sm font-semibold text-neutral-700 transition hover:border-neutral-400 hover:text-neutral-900"
            >
              <ArrowLeft className="h-4 w-4" />
              Back to groups
            </Link>
            <Link
              href="/dashboard"
              className="brand-gradient inline-flex items-center gap-2 rounded-full px-4 py-2 text-sm font-semibold text-white transition hover:-translate-y-0.5 hover:shadow-md"
            >
              Open dashboard
              <ArrowRight className="h-4 w-4" />
            </Link>
          </div>
        </motion.section>

        <motion.section
          initial="hidden"
          whileInView="show"
          viewport={viewportOnce}
          variants={fadeUp}
          className="mt-6 rounded-3xl border border-neutral-200 bg-white p-6 shadow-sm"
        >
          <div className="flex flex-wrap items-center justify-between gap-3">
            <div>
              <h2 className="text-lg font-semibold text-neutral-900">Group posts</h2>
              <p className="text-sm text-neutral-600">
                Latest posts shared with this group.
              </p>
            </div>
            <span className="rounded-full border border-neutral-200 bg-neutral-50 px-3 py-1 text-xs text-neutral-600">
              {posts.length} post(s)
            </span>
          </div>

          <div className="mt-5 rounded-3xl border border-neutral-200 bg-neutral-50 p-4">
            <textarea
              value={composerText}
              onChange={(event) => setComposerText(event.target.value)}
              rows={4}
              placeholder="Share an update with this group..."
              className="w-full resize-none rounded-2xl border border-neutral-200 bg-white px-4 py-3 text-sm text-neutral-900 placeholder:text-neutral-400 outline-none transition focus:border-neutral-400"
            />
            <div className="mt-3 flex flex-wrap items-center gap-3">
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
                Add media
              </label>
              {composerFileName ? (
                <span className="text-xs text-neutral-500">{composerFileName}</span>
              ) : null}
              <input
                value={mediaUrl}
                onChange={(event) => setMediaUrl(event.target.value)}
                placeholder="Or paste media URL"
                className="h-10 flex-1 rounded-2xl border border-neutral-200 bg-white px-4 text-sm text-neutral-900 placeholder:text-neutral-400 outline-none transition focus:border-neutral-400"
              />
            </div>
            <div className="mt-3 flex flex-wrap items-center justify-between gap-3">
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
            {composerError ? (
              <p className="mt-3 text-xs text-rose-600">{composerError}</p>
            ) : null}
          </div>

          {postsLoading ? (
            <p className="mt-4 text-sm text-neutral-600">Loading group posts...</p>
          ) : postsError ? (
            <p className="mt-4 rounded-2xl border border-rose-200 bg-rose-50 px-4 py-3 text-sm text-rose-700">
              {postsError}
            </p>
          ) : posts.length === 0 ? (
            <p className="mt-4 text-sm text-neutral-600">
              No posts yet. Be the first to share something.
            </p>
          ) : (
            <div className="mt-4 space-y-4">
              {posts.map((post) => (
                <article
                  key={post.id}
                  className="rounded-3xl border border-neutral-200 bg-neutral-50 p-5"
                >
                  <header className="flex items-start justify-between gap-3">
                    <div>
                      <p className="text-sm font-semibold text-neutral-900">
                        {post.author_first_name} {post.author_last_name}
                      </p>
                      <p className="text-xs text-neutral-500">{shortDate(post.created_at)}</p>
                    </div>
                    <button
                      type="button"
                      onClick={() => toggleComments(post.id)}
                      className="inline-flex items-center gap-1 rounded-full border border-neutral-200 bg-white px-2.5 py-1 text-[11px] text-neutral-500"
                    >
                      <MessageCircle className="h-3.5 w-3.5" />
                      {post.comment_count} comments
                    </button>
                  </header>

                  {editingPostID === post.id ? (
                    <div className="mt-3 space-y-3">
                      <textarea
                        value={editPostText}
                        onChange={(event) => setEditPostText(event.target.value)}
                        rows={3}
                        className="w-full rounded-2xl border border-neutral-200 bg-white px-4 py-3 text-sm outline-none focus:border-neutral-400"
                      />
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
                      {editPostError ? (
                        <p className="text-xs text-rose-600">{editPostError}</p>
                      ) : null}
                    </div>
                  ) : (
                    <p className="mt-3 text-sm leading-relaxed text-neutral-700">{post.content}</p>
                  )}

                  {post.media_path ? (
                    <div className="mt-4 overflow-hidden rounded-2xl border border-neutral-200 bg-white">
                      <img
                        src={toMediaUrl(apiBaseUrl, post.media_path)}
                        alt="Post media"
                        className="max-h-[520px] w-full object-contain bg-white"
                      />
                    </div>
                  ) : null}

                  <footer className="mt-4 flex items-center gap-3 text-xs text-neutral-500">
                    <button
                      type="button"
                      onClick={() => handlePostReaction(post.id, "like")}
                      className={`inline-flex items-center gap-1 rounded-full px-2 py-1 transition ${
                        postReactionMap[post.id] === "like"
                          ? "bg-emerald-100 text-emerald-800"
                          : "bg-white text-neutral-600 hover:bg-neutral-100"
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
                          : "bg-white text-neutral-600 hover:bg-neutral-100"
                      }`}
                    >
                      <ThumbsDown className="h-3.5 w-3.5" />
                      {post.dislike_count}
                    </button>
                    {userID === post.author_id ? (
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
                    <section className="mt-4 rounded-2xl border border-neutral-200 bg-white p-3">
                      <div className="space-y-2">
                        {(commentsByPost[post.id] ?? []).map((comment) => (
                          <article key={comment.id} className="rounded-xl bg-neutral-50 p-3">
                            {editingCommentID === comment.id ? (
                              <div className="space-y-2">
                                <textarea
                                  value={editCommentText}
                                  onChange={(event) => setEditCommentText(event.target.value)}
                                  rows={2}
                                  className="w-full rounded-xl border border-neutral-200 bg-white px-3 py-2 text-xs outline-none focus:border-neutral-400"
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
                                    : "bg-white text-neutral-600"
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
                                    : "bg-white text-neutral-600"
                                }`}
                              >
                                <ThumbsDown className="h-3 w-3" />
                                {comment.dislike_count}
                              </button>
                              {userID === comment.author_id ? (
                                <>
                                  <button
                                    type="button"
                                    onClick={() => startEditComment(comment)}
                                    className="inline-flex items-center gap-1 rounded-full bg-neutral-100 px-2 py-1 text-neutral-600 transition hover:bg-neutral-200"
                                  >
                                    Edit
                                  </button>
                                  <button
                                    type="button"
                                    onClick={() => deleteComment(post.id, comment.id)}
                                    className="inline-flex items-center gap-1 rounded-full bg-neutral-100 px-2 py-1 text-neutral-600 transition hover:bg-neutral-200"
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
                          className="h-9 flex-1 rounded-xl border border-neutral-200 bg-neutral-50 px-3 text-xs outline-none focus:border-neutral-400"
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
                          Add media
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
              ))}
            </div>
          )}

          {hasMorePosts && !postsLoading && !postsError ? (
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
        </motion.section>
      </main>
    </div>
  );
}
