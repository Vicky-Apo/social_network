"use client";
/* eslint-disable @next/next/no-img-element */

import { useCallback, useEffect, useMemo, useState } from "react";
import Link from "next/link";
import { useParams, useRouter } from "next/navigation";
import {
  Globe,
  Lock,
  MessageCircle,
  Plus,
  ThumbsDown,
  ThumbsUp,
  UserPlus,
  UserMinus,
} from "lucide-react";
import { motion } from "framer-motion";
import TopNav from "@/components/TopNav";
import LeftNav from "@/components/LeftNav";
import Avatar from "@/components/Avatar";
import Pagination from "@/components/Pagination";
import { fadeUp, viewportOnce } from "@/components/Motion";

type ApiResponse<T> = {
  success?: boolean;
  data?: T;
  error?: string;
};

type MeUser = {
  id: number;
  email: string;
  first_name: string;
  last_name: string;
  nickname?: string | null;
  avatar_path?: string | null;
};

type ProfileUser = {
  id: number;
  email?: string | null;
  first_name: string;
  last_name: string;
  date_of_birth?: string | null;
  avatar_path?: string | null;
  nickname?: string | null;
  about?: string | null;
  is_public: boolean;
  created_at?: string | null;
  updated_at?: string | null;
};

type ProfileDTO = {
  user: ProfileUser;
  followers_count?: number | null;
  following_count?: number | null;
  is_following: boolean;
  is_followed_by: boolean;
  limited?: boolean;
};

type Post = {
  id: number;
  author_id: number;
  group_id?: number | null;
  author_first_name: string;
  author_last_name: string;
  author_nickname?: string | null;
  author_avatar_path?: string | null;
  content: string;
  media_path?: string | null;
  privacy: string;
  comment_count: number;
  like_count: number;
  dislike_count: number;
  created_at: string;
  updated_at?: string;
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

type FullProfileResponse = {
  profile: ProfileDTO;
  posts: Post[];
  activity?: {
    recent_posts?: Post[];
  };
};

type FollowRequest = {
  id: number;
  requester_id: number;
  target_id: number;
  status: string;
  created_at: string;
};

type FollowState = "none" | "requested" | "following" | "loading";

function initials(first?: string | null, last?: string | null) {
  const left = first?.trim().charAt(0) ?? "";
  const right = last?.trim().charAt(0) ?? "";
  return `${left}${right}`.toUpperCase() || "U";
}

function formatDate(value?: string | null) {
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

export default function ProfilePage() {
  const router = useRouter();
  const params = useParams<{ id: string }>();

  const profileID = typeof params?.id === "string" ? Number(params.id) : NaN;

  const [viewer, setViewer] = useState<MeUser | null>(null);
  const [profile, setProfile] = useState<ProfileDTO | null>(null);
  const [posts, setPosts] = useState<Post[]>([]);
  const [postReactionMap, setPostReactionMap] = useState<ReactionMap>({});
  const [commentReactionMap, setCommentReactionMap] = useState<ReactionMap>({});
  const [commentsByPost, setCommentsByPost] = useState<Record<number, Comment[]>>({});
  const [commentsOpenByPost, setCommentsOpenByPost] = useState<Record<number, boolean>>({});
  const [commentsLoadingByPost, setCommentsLoadingByPost] = useState<Record<number, boolean>>({});
  const [commentDraftByPost, setCommentDraftByPost] = useState<Record<number, string>>({});
  const [commentFileByPost, setCommentFileByPost] = useState<Record<number, File | null>>({});
  const [commentFileNameByPost, setCommentFileNameByPost] = useState<Record<number, string>>({});
  const [commentErrorByPost, setCommentErrorByPost] = useState<Record<number, string>>({});
  const [commentPageByPost, setCommentPageByPost] = useState<Record<number, number>>({});
  const [commentTotalByPost, setCommentTotalByPost] = useState<Record<number, number>>({});
  const [totalPosts, setTotalPosts] = useState(0);
  const [currentPage, setCurrentPage] = useState(1);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [followState, setFollowState] = useState<FollowState>("none");
  const [followError, setFollowError] = useState<string | null>(null);
  const [pendingRequestID, setPendingRequestID] = useState<number | null>(null);

  const apiBaseUrl = useMemo(
    () =>
      process.env.NEXT_PUBLIC_API_BASE_URL?.trim().replace(/\/+$/, "") ||
      "http://localhost:8080",
    [],
  );
  const feedLimit = 10;
  const totalPages = Math.max(1, Math.ceil(totalPosts / feedLimit));
  const commentLimit = 10;

  const uploadMedia = async (file: File, kind: "comment") => {
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

  const loadProfile = useCallback(async () => {
    if (!Number.isFinite(profileID) || profileID <= 0) {
      setError("Invalid profile id.");
      setIsLoading(false);
      return;
    }

    setIsLoading(true);
    setError(null);

    try {
      const meResponse = await fetch(`${apiBaseUrl}/auth/me`, {
        credentials: "include",
      });
      const meResult = (await meResponse.json().catch(() => null)) as ApiResponse<MeUser> | null;
      if (!meResponse.ok || !meResult?.success || !meResult.data) {
        router.replace("/login");
        return;
      }
      setViewer(meResult.data);

      const [profileResponse, sentResponse] = await Promise.all([
        fetch(`${apiBaseUrl}/profiles/${profileID}/full`, {
          credentials: "include",
        }),
        fetch(`${apiBaseUrl}/follow-requests/sent`, {
          credentials: "include",
        }),
      ]);

      const profileResult = (await profileResponse.json().catch(() => null)) as
        | ApiResponse<FullProfileResponse>
        | null;
      const sentResult = (await sentResponse.json().catch(() => null)) as
        | ApiResponse<FollowRequest[]>
        | null;

      if (!profileResponse.ok || !profileResult?.success || !profileResult.data) {
        setError(profileResult?.error || "Could not load profile.");
        setProfile(null);
        setPosts([]);
        return;
      }

      const nextProfile = profileResult.data.profile;
      setProfile(nextProfile);
      setPosts([]);

      const sentRequests = sentResponse.ok && sentResult?.success ? sentResult.data ?? [] : [];
      const pendingRequest = sentRequests.find((req) => req.target_id === profileID);
      setPendingRequestID(pendingRequest?.id ?? null);

      if (nextProfile.is_following) {
        setFollowState("following");
      } else if (pendingRequest) {
        setFollowState("requested");
      } else {
        setFollowState("none");
      }
    } catch {
      setError("Network error. Please try again.");
      setProfile(null);
      setPosts([]);
    } finally {
      setIsLoading(false);
    }
  }, [apiBaseUrl, profileID, router]);

  useEffect(() => {
    void loadProfile();
  }, [loadProfile]);

  useEffect(() => {
    setCurrentPage(1);
  }, [profileID]);

  const fetchPostsPage = async (page: number) => {
    const offset = (page - 1) * feedLimit;
    const response = await fetch(
      `${apiBaseUrl}/posts?author_id=${profileID}&limit=${feedLimit}&offset=${offset}`,
      { credentials: "include" },
    );
    const result = (await response.json().catch(() => null)) as ApiResponse<Post[]> | null;
    if (!response.ok || !result?.success) {
      throw new Error(result?.error || "Could not load posts.");
    }
    const nextPosts = result.data ?? [];
    const totalHeader = response.headers.get("X-Total-Count");
    const parsedTotal = totalHeader ? Number(totalHeader) : Number.NaN;
    setTotalPosts(Number.isFinite(parsedTotal) ? parsedTotal : nextPosts.length);
    setPosts(nextPosts);
  };

  useEffect(() => {
    if (!profile || profile.limited) {
      setPosts([]);
      return;
    }
    void fetchPostsPage(currentPage);
  }, [apiBaseUrl, profile, profileID, currentPage]);

  useEffect(() => {
    if (!viewer?.id || posts.length === 0) {
      return;
    }

    let cancelled = false;
    Promise.all(
      posts.map(async (post) => {
        try {
          const res = await fetch(`${apiBaseUrl}/posts/${post.id}/reactions`, {
            credentials: "include",
          });
          const json = (await res.json().catch(() => null)) as
            | ApiResponse<Reaction[]>
            | null;
          if (!res.ok || !json?.success) {
            return [post.id, null] as const;
          }
          const mine = (json.data ?? []).find((item) => item.user_id === viewer.id);
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

    return () => {
      cancelled = true;
    };
  }, [apiBaseUrl, posts, viewer?.id]);

  const handleFollow = async () => {
    if (!profile || followState === "loading") return;
    setFollowState("loading");
    setFollowError(null);

    try {
      const response = await fetch(`${apiBaseUrl}/follow-requests`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        credentials: "include",
        body: JSON.stringify({ target_id: profile.user.id }),
      });
      const result = (await response.json().catch(() => null)) as
        | ApiResponse<{ status?: string }>
        | null;
      if (!response.ok || !result?.success) {
        setFollowState("none");
        setFollowError(result?.error || "Could not send follow request.");
        return;
      }

      if (result?.data?.status === "followed") {
        setFollowState("following");
        setProfile((prev) => (prev ? { ...prev, is_following: true } : prev));
      } else {
        setFollowState("requested");
        setPendingRequestID(
          typeof (result?.data as { request?: { id?: number } } | undefined)?.request?.id === "number"
            ? (result?.data as { request?: { id?: number } }).request?.id ?? null
            : null,
        );
      }
    } catch {
      setFollowState("none");
      setFollowError("Network error. Please try again.");
    }
  };

  const handleUnfollow = async () => {
    if (!profile || followState === "loading") return;
    setFollowState("loading");
    setFollowError(null);

    try {
      const response = await fetch(`${apiBaseUrl}/users/${profile.user.id}/followers`, {
        method: "DELETE",
        credentials: "include",
      });
      if (!response.ok) {
        const result = (await response.json().catch(() => null)) as ApiResponse<unknown> | null;
        setFollowError(result?.error || "Could not unfollow this user.");
        setFollowState("following");
        return;
      }
      setFollowState("none");
      setProfile((prev) => (prev ? { ...prev, is_following: false } : prev));
    } catch {
      setFollowState("following");
      setFollowError("Network error. Please try again.");
    }
  };

  const loadCommentsForPost = async (postID: number, page: number) => {
    setCommentsLoadingByPost((prev) => ({ ...prev, [postID]: true }));
    setCommentErrorByPost((prev) => ({ ...prev, [postID]: "" }));

    try {
      const offset = (page - 1) * commentLimit;
      const response = await fetch(
        `${apiBaseUrl}/posts/${postID}/comments?limit=${commentLimit}&offset=${offset}`,
        { credentials: "include" },
      );
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
      setCommentPageByPost((prev) => ({ ...prev, [postID]: page }));
      const totalHeader = response.headers.get("X-Total-Count");
      const parsedTotal = totalHeader ? Number(totalHeader) : Number.NaN;
      setCommentTotalByPost((prev) => ({
        ...prev,
        [postID]: Number.isFinite(parsedTotal) ? parsedTotal : comments.length,
      }));

      if (viewer?.id && comments.length > 0) {
        const entries = await Promise.all(
          comments.map(async (comment) => {
            try {
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
              const mine = (reactionJson.data ?? []).find((item) => item.user_id === viewer.id);
              return [comment.id, mine?.reaction ?? null] as const;
            } catch {
              return [comment.id, null] as const;
            }
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
    if (nextOpen) {
      const page = commentPageByPost[postID] ?? 1;
      void loadCommentsForPost(postID, page);
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

      const nextPage = commentPageByPost[postID] ?? 1;
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
      setCommentTotalByPost((prev) => ({
        ...prev,
        [postID]: (prev[postID] ?? 0) + 1,
      }));
      await loadCommentsForPost(postID, nextPage);
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
        if (previous === "like") like = Math.max(0, like - 1);
        if (previous === "dislike") dislike = Math.max(0, dislike - 1);
        if (next === "like") like += 1;
        if (next === "dislike") dislike += 1;
        return { ...post, like_count: like, dislike_count: dislike };
      }),
    );

    try {
      await fetch(`${apiBaseUrl}/posts/${postID}/reactions`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        credentials: "include",
        body: JSON.stringify({ reaction: next }),
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
        if (previous === "like") like = Math.max(0, like - 1);
        if (previous === "dislike") dislike = Math.max(0, dislike - 1);
        if (next === "like") like += 1;
        if (next === "dislike") dislike += 1;
        return { ...comment, like_count: like, dislike_count: dislike };
      }),
    }));

    try {
      await fetch(`${apiBaseUrl}/comments/${commentID}/reactions`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        credentials: "include",
        body: JSON.stringify({ reaction: next }),
      });
    } catch {
      setCommentReactionMap((prev) => ({ ...prev, [commentID]: previous }));
    }
  };

  const handleCancelRequest = async () => {
    if (!pendingRequestID || followState === "loading") return;
    setFollowState("loading");
    setFollowError(null);

    try {
      const response = await fetch(`${apiBaseUrl}/follow-requests/${pendingRequestID}`, {
        method: "PATCH",
        headers: { "Content-Type": "application/json" },
        credentials: "include",
        body: JSON.stringify({ status: "canceled" }),
      });
      if (!response.ok) {
        const result = (await response.json().catch(() => null)) as ApiResponse<unknown> | null;
        setFollowError(result?.error || "Could not cancel request.");
        setFollowState("requested");
        return;
      }
      setFollowState("none");
      setPendingRequestID(null);
    } catch {
      setFollowState("requested");
      setFollowError("Network error. Please try again.");
    }
  };

  const isOwner = profile?.user?.id === viewer?.id;
  const displayName = profile
    ? `${profile.user.first_name} ${profile.user.last_name}`
    : "Profile";
  const userTag =
    profile?.user.nickname?.trim() ||
    (profile?.user.email ? profile.user.email.split("@")[0] : null) ||
    "user";
  const visibilityLabel = profile?.user.is_public ? "Public profile" : "Private profile";
  const privacyIcon = profile?.user.is_public ? Globe : Lock;
  const PrivacyIcon = privacyIcon;

  return (
    <div
      className="min-h-screen text-neutral-100"
      style={{
        backgroundImage: "url('/groups-bg.png')",
        backgroundSize: "cover",
        backgroundPosition: "center",
        backgroundAttachment: "fixed",
      }}
    >
      <TopNav user={viewer ?? undefined} onLogout={() => router.replace("/login")} variant="dark" />

      <main className="mx-auto grid w-full max-w-6xl gap-6 px-4 py-6 sm:px-6 lg:grid-cols-[240px_minmax(0,1fr)]">
        <aside className="hidden lg:block">
          <LeftNav user={viewer ?? undefined} activeHref="/dashboard" variant="dark" />
        </aside>

        <section className="space-y-5">
          <motion.div
            initial="hidden"
            whileInView="show"
            viewport={viewportOnce}
            variants={fadeUp}
            className="rounded-3xl border border-white/10 bg-white/5 p-6 backdrop-blur-sm"
          >
            {isLoading ? (
              <p className="text-sm text-neutral-400">Loading profile...</p>
            ) : error ? (
              <p className="text-sm text-rose-400">{error}</p>
            ) : profile ? (
              <div className="flex flex-col gap-5 sm:flex-row sm:items-center sm:justify-between">
                <div className="flex items-center gap-4">
                  <Avatar
                    src={
                      profile.user.avatar_path
                        ? toMediaUrl(apiBaseUrl, profile.user.avatar_path)
                        : null
                    }
                    name={displayName}
                    size={64}
                    textClassName="text-lg"
                  />
                  <div>
                    <div className="flex flex-wrap items-center gap-2">
                      <h1 className="text-2xl font-semibold tracking-tight text-white">
                        {displayName}
                      </h1>
                      <span className="inline-flex items-center gap-1 rounded-full border border-white/20 bg-white/5 px-2 py-1 text-[11px] uppercase tracking-wide text-neutral-400">
                        <PrivacyIcon className="h-3 w-3" />
                        {visibilityLabel}
                      </span>
                      {profile.is_followed_by ? (
                        <span className="rounded-full bg-emerald-500/20 px-2 py-1 text-[11px] font-semibold text-emerald-400">
                          Follows you
                        </span>
                      ) : null}
                    </div>
                    <p className="text-sm text-neutral-400">@{userTag}</p>
                  </div>
                </div>

                {!isOwner ? (
                  <div className="flex flex-wrap items-center gap-2">
                    {followState === "following" ? (
                      <button
                        type="button"
                        onClick={handleUnfollow}
                        className="inline-flex items-center gap-2 rounded-full border border-white/20 bg-white/5 px-4 py-2 text-xs font-semibold text-neutral-300 transition hover:bg-white/10 hover:text-white"
                      >
                        <UserMinus className="h-3.5 w-3.5" />
                        Unfollow
                      </button>
                    ) : followState === "requested" ? (
                      <button
                        type="button"
                        onClick={handleCancelRequest}
                        className="inline-flex items-center gap-2 rounded-full border border-white/20 bg-white/5 px-4 py-2 text-xs font-semibold text-neutral-300 transition hover:bg-white/10 hover:text-white"
                      >
                        <UserMinus className="h-3.5 w-3.5" />
                        Cancel request
                      </button>
                    ) : (
                      <button
                        type="button"
                        onClick={handleFollow}
                        className="brand-gradient inline-flex items-center gap-2 rounded-full px-4 py-2 text-xs font-semibold text-white transition hover:-translate-y-0.5 hover:shadow-md"
                      >
                        <UserPlus className="h-3.5 w-3.5" />
                        Follow
                      </button>
                    )}
                    {followState === "loading" ? (
                      <span className="text-xs text-neutral-400">Updating...</span>
                    ) : null}
                    {followError ? <p className="text-xs text-rose-400">{followError}</p> : null}
                  </div>
                ) : (
                  <div className="flex flex-wrap items-center gap-2">
                    <Link
                      href="/profile/edit"
                      className="inline-flex items-center gap-2 rounded-full border border-white/20 bg-white/5 px-4 py-2 text-xs font-semibold text-neutral-300 transition hover:bg-white/10 hover:text-white"
                    >
                      Edit profile
                    </Link>
                  </div>
                )}
              </div>
            ) : null}
          </motion.div>

          {profile ? (
            <motion.div
              initial="hidden"
              whileInView="show"
              viewport={viewportOnce}
              variants={fadeUp}
              className="grid gap-4 rounded-3xl border border-white/10 bg-white/5 p-6 backdrop-blur-sm sm:grid-cols-3"
            >
              <div className="rounded-2xl border border-white/10 bg-white/5 px-4 py-3 transition hover:border-white/20 hover:bg-white/10">
                <Link href={`/profile/${profile.user.id}/followers`} className="block">
                  <p className="text-xs uppercase tracking-wide text-neutral-400">Followers</p>
                  <p className="mt-1 text-xl font-semibold text-white">
                    {profile.followers_count ?? 0}
                  </p>
                </Link>
              </div>
              <div className="rounded-2xl border border-white/10 bg-white/5 px-4 py-3 transition hover:border-white/20 hover:bg-white/10">
                <Link href={`/profile/${profile.user.id}/following`} className="block">
                  <p className="text-xs uppercase tracking-wide text-neutral-400">Following</p>
                  <p className="mt-1 text-xl font-semibold text-white">
                    {profile.following_count ?? 0}
                  </p>
                </Link>
              </div>
              <div className="rounded-2xl border border-white/10 bg-white/5 px-4 py-3">
                <p className="text-xs uppercase tracking-wide text-neutral-400">Joined</p>
                <p className="mt-1 text-sm font-semibold text-white">
                  {formatDate(profile.user.created_at)}
                </p>
              </div>
            </motion.div>
          ) : null}

          {profile ? (
            <motion.div
              initial="hidden"
              whileInView="show"
              viewport={viewportOnce}
              variants={fadeUp}
              className="rounded-3xl border border-white/10 bg-white/5 p-6 backdrop-blur-sm"
            >
              <h2 className="text-sm font-semibold text-white">About</h2>
              <div className="mt-3 grid gap-3 text-sm text-neutral-300 sm:grid-cols-2">
                <div>
                  <p className="text-xs uppercase tracking-wide text-neutral-400">Email</p>
                  <p className="mt-1">{profile.user.email ?? "Hidden"}</p>
                </div>
                <div>
                  <p className="text-xs uppercase tracking-wide text-neutral-400">Date of birth</p>
                  <p className="mt-1">{profile.user.date_of_birth ?? "Hidden"}</p>
                </div>
                <div className="sm:col-span-2">
                  <p className="text-xs uppercase tracking-wide text-neutral-400">Bio</p>
                  <p className="mt-1">
                    {profile.user.about ? profile.user.about : "No bio yet."}
                  </p>
                </div>
              </div>
            </motion.div>
          ) : null}

          <motion.div
            initial="hidden"
            whileInView="show"
            viewport={viewportOnce}
            variants={fadeUp}
            className="rounded-3xl border border-white/10 bg-white/5 p-6 backdrop-blur-sm"
          >
            <div className="flex items-center justify-between">
              <h2 className="text-lg font-semibold text-white">Posts</h2>
              <span className="text-xs text-neutral-400">{posts.length} post(s)</span>
            </div>

            {!profile ? null : profile.limited ? (
              <p className="mt-4 text-sm text-neutral-400">
                This profile is private. Follow to see posts and activity.
              </p>
            ) : posts.length === 0 ? (
              <p className="mt-4 text-sm text-neutral-400">No posts yet.</p>
            ) : (
              <div className="mt-4 space-y-4">
                {posts.map((post) => (
                  <article
                    key={post.id}
                    className="rounded-2xl border border-white/10 bg-white/5 p-4"
                  >
                    <header className="flex items-start justify-between gap-3">
                      <div>
                        <p className="text-sm font-semibold text-white">
                          {post.author_first_name} {post.author_last_name}
                        </p>
                        <p className="text-xs text-neutral-400">{shortDate(post.created_at)}</p>
                      </div>
                      <span className="rounded-full border border-white/20 bg-white/5 px-2.5 py-1 text-[11px] uppercase tracking-wide text-neutral-400">
                        {post.privacy}
                      </span>
                    </header>

                    <p className="mt-3 text-sm text-neutral-300">{post.content}</p>

                    {post.media_path ? (
                      <div className="mt-3 overflow-hidden rounded-2xl border border-white/10 bg-white/5">
                        <img
                          src={toMediaUrl(apiBaseUrl, post.media_path)}
                          alt="Post media"
                          className="max-h-[520px] w-full object-contain"
                        />
                      </div>
                    ) : null}

                    <footer className="mt-3 flex items-center gap-4 text-xs text-neutral-400">
                      <button
                        type="button"
                        onClick={() => handlePostReaction(post.id, "like")}
                        className={`inline-flex items-center gap-1 rounded-full px-2 py-1 transition ${
                          postReactionMap[post.id] === "like"
                            ? "bg-emerald-500/20 text-emerald-400"
                            : "bg-white/5 text-neutral-400 hover:bg-white/10"
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
                            ? "bg-rose-500/20 text-rose-400"
                            : "bg-white/5 text-neutral-400 hover:bg-white/10"
                        }`}
                      >
                        <ThumbsDown className="h-3.5 w-3.5" />
                        {post.dislike_count}
                      </button>
                      <button
                        type="button"
                        onClick={() => toggleComments(post.id)}
                        className="inline-flex items-center gap-1 rounded-full bg-white/5 px-2 py-1 text-neutral-400 transition hover:bg-white/10"
                      >
                        <MessageCircle className="h-3.5 w-3.5" />
                        {post.comment_count}
                      </button>
                    </footer>

                    {commentsOpenByPost[post.id] ? (
                      <section className="mt-4 rounded-2xl border border-white/10 bg-white/5 p-3">
                        <div className="space-y-2">
                          {(commentsByPost[post.id] ?? []).map((comment) => (
                            <article key={comment.id} className="rounded-xl bg-white/5 p-3">
                              <p className="text-sm text-neutral-300">{comment.content}</p>
                              {comment.media_path ? (
                                <div className="mt-2 overflow-hidden rounded-xl border border-white/10">
                                  <img
                                    src={toMediaUrl(apiBaseUrl, comment.media_path)}
                                    alt="Comment media"
                                    className="max-h-64 w-full object-contain"
                                  />
                                </div>
                              ) : null}
                              <div className="mt-2 flex items-center gap-2 text-xs">
                                <button
                                  type="button"
                                  onClick={() =>
                                    handleCommentReaction(post.id, comment.id, "like")
                                  }
                                  className={`inline-flex items-center gap-1 rounded-full px-2 py-1 ${
                                    commentReactionMap[comment.id] === "like"
                                      ? "bg-emerald-500/20 text-emerald-400"
                                      : "bg-white/5 text-neutral-400"
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
                                      ? "bg-rose-500/20 text-rose-400"
                                      : "bg-white/5 text-neutral-400"
                                  }`}
                                >
                                  <ThumbsDown className="h-3 w-3" />
                                  {comment.dislike_count}
                                </button>
                              </div>
                            </article>
                          ))}

                          {commentsLoadingByPost[post.id] ? (
                            <p className="text-xs text-neutral-400">Loading comments...</p>
                          ) : null}
                          {commentErrorByPost[post.id] ? (
                            <p className="text-xs text-rose-400">{commentErrorByPost[post.id]}</p>
                          ) : null}
                        </div>

                        <Pagination
                          currentPage={commentPageByPost[post.id] ?? 1}
                          totalPages={Math.max(
                            1,
                            Math.ceil(
                              (commentTotalByPost[post.id] ??
                                (commentsByPost[post.id]?.length ?? 0)) / commentLimit,
                            ),
                          )}
                          onPageChange={(page) => loadCommentsForPost(post.id, page)}
                          className="mt-3"
                        />

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
                            className="h-9 flex-1 rounded-xl border border-neutral-200 bg-white px-3 text-xs text-black outline-none focus:border-neutral-400 placeholder:text-neutral-500"
                          />
                          <label className="inline-flex h-9 items-center gap-2 rounded-xl border border-white/20 bg-white/5 px-3 text-xs font-semibold text-neutral-300 transition hover:bg-white/10 hover:text-white">
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
                            className="rounded-xl bg-white/10 px-3 text-xs font-semibold text-white transition hover:bg-white/20"
                          >
                            Comment
                          </button>
                        </div>
                        {commentFileNameByPost[post.id] ? (
                          <p className="mt-2 text-[11px] text-neutral-400">
                            Attached: {commentFileNameByPost[post.id]}
                          </p>
                        ) : null}
                      </section>
                    ) : null}
                  </article>
                ))}
              </div>
            )}
            {!profile?.limited ? (
              <Pagination
                currentPage={currentPage}
                totalPages={totalPages}
                onPageChange={setCurrentPage}
                className="mt-5"
              />
            ) : null}
          </motion.div>
        </section>
      </main>
    </div>
  );
}
