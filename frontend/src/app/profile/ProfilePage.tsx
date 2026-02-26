"use client";
/* eslint-disable @next/next/no-img-element */

import { useCallback, useEffect, useMemo, useState } from "react";
import Link from "next/link";
import { useParams, useRouter } from "next/navigation";
import {
  Compass,
  Globe,
  Lock,
  MessageSquare,
  UserPlus,
  UserMinus,
  Users,
} from "lucide-react";
import { motion } from "framer-motion";
import TopNav from "../component/TopNav";
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

const quickLinks = [
  { label: "Explore", href: "/dashboard", icon: Compass },
  { label: "Groups", href: "/groups", icon: Users },
  { label: "Messages", href: "/messages", icon: MessageSquare },
  { label: "Requests", href: "/follow-requests", icon: UserPlus },
];

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
      setPosts(profileResult.data.posts ?? []);

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
  const userTag = profile?.user.nickname || `user-${profile?.user.id ?? ""}`;
  const visibilityLabel = profile?.user.is_public ? "Public profile" : "Private profile";
  const privacyIcon = profile?.user.is_public ? Globe : Lock;
  const PrivacyIcon = privacyIcon;

  return (
    <div className="min-h-screen bg-neutral-50 text-neutral-900">
      <TopNav user={viewer ?? undefined} onLogout={() => router.replace("/login")} />

      <main className="mx-auto grid w-full max-w-6xl gap-6 px-4 py-6 sm:px-6 lg:grid-cols-[240px_minmax(0,1fr)_280px]">
        <aside className="hidden lg:block">
          <div className="rounded-3xl border border-neutral-200 bg-white p-5 shadow-sm">
            <div className="flex items-center gap-3">
              <div className="inline-flex h-11 w-11 items-center justify-center rounded-full bg-neutral-900 text-sm font-semibold text-white">
                {initials(viewer?.first_name, viewer?.last_name)}
              </div>
              <div>
                <p className="text-sm font-semibold text-neutral-900">
                  {viewer ? `${viewer.first_name} ${viewer.last_name}` : "Loading"}
                </p>
                <p className="text-xs text-neutral-500">
                  @{viewer?.nickname || (viewer?.email ? viewer.email.split("@")[0] : "member")}
                </p>
              </div>
            </div>
            <nav className="mt-5 space-y-2">
              {quickLinks.map((item) => {
                const Icon = item.icon;
                return (
                  <Link
                    key={item.label}
                    href={item.href}
                    className="flex items-center gap-2 rounded-2xl border border-neutral-200 bg-neutral-50 px-3 py-2 text-sm text-neutral-700 transition hover:border-neutral-400 hover:text-neutral-900"
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
            className="rounded-3xl border border-neutral-200 bg-white p-5 shadow-sm"
          >
            {isLoading ? (
              <p className="text-sm text-neutral-600">Loading profile...</p>
            ) : error ? (
              <p className="text-sm text-rose-600">{error}</p>
            ) : profile ? (
              <div className="flex flex-col gap-5 sm:flex-row sm:items-center sm:justify-between">
                <div className="flex items-center gap-4">
                  {profile.user.avatar_path ? (
                    <img
                      src={toMediaUrl(apiBaseUrl, profile.user.avatar_path)}
                      alt={displayName}
                      className="h-16 w-16 rounded-full border border-neutral-200 object-cover"
                    />
                  ) : (
                    <div className="inline-flex h-16 w-16 items-center justify-center rounded-full bg-neutral-900 text-lg font-semibold text-white">
                      {initials(profile.user.first_name, profile.user.last_name)}
                    </div>
                  )}
                  <div>
                    <div className="flex flex-wrap items-center gap-2">
                      <h1 className="text-2xl font-semibold tracking-tight text-neutral-900">
                        {displayName}
                      </h1>
                      <span className="inline-flex items-center gap-1 rounded-full border border-neutral-200 bg-neutral-50 px-2 py-1 text-[11px] uppercase tracking-wide text-neutral-600">
                        <PrivacyIcon className="h-3 w-3" />
                        {visibilityLabel}
                      </span>
                      {profile.is_followed_by ? (
                        <span className="rounded-full bg-emerald-50 px-2 py-1 text-[11px] font-semibold text-emerald-700">
                          Follows you
                        </span>
                      ) : null}
                    </div>
                    <p className="text-sm text-neutral-500">@{userTag}</p>
                  </div>
                </div>

                {!isOwner ? (
                  <div className="flex flex-wrap items-center gap-2">
                    {followState === "following" ? (
                      <button
                        type="button"
                        onClick={handleUnfollow}
                        className="inline-flex items-center gap-2 rounded-full border border-neutral-200 bg-white px-4 py-2 text-xs font-semibold text-neutral-700 transition hover:border-neutral-400 hover:text-neutral-900"
                      >
                        <UserMinus className="h-3.5 w-3.5" />
                        Unfollow
                      </button>
                    ) : followState === "requested" ? (
                      <button
                        type="button"
                        onClick={handleCancelRequest}
                        className="inline-flex items-center gap-2 rounded-full border border-neutral-200 bg-white px-4 py-2 text-xs font-semibold text-neutral-700 transition hover:border-neutral-400 hover:text-neutral-900"
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
                      <span className="text-xs text-neutral-500">Updating...</span>
                    ) : null}
                    {followError ? <p className="text-xs text-rose-600">{followError}</p> : null}
                  </div>
                ) : (
                  <div className="flex flex-wrap items-center gap-2">
                    <Link
                      href="/profile/edit"
                      className="inline-flex items-center gap-2 rounded-full border border-neutral-200 bg-white px-4 py-2 text-xs font-semibold text-neutral-700 transition hover:border-neutral-400 hover:text-neutral-900"
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
              className="grid gap-4 rounded-3xl border border-neutral-200 bg-white p-5 shadow-sm sm:grid-cols-3"
            >
              <div className="rounded-2xl border border-neutral-200 bg-neutral-50 px-4 py-3 transition hover:border-neutral-400 hover:bg-white">
                <Link href={`/profile/${profile.user.id}/followers`} className="block">
                  <p className="text-xs uppercase tracking-wide text-neutral-500">Followers</p>
                  <p className="mt-1 text-xl font-semibold text-neutral-900">
                    {profile.followers_count ?? 0}
                  </p>
                </Link>
              </div>
              <div className="rounded-2xl border border-neutral-200 bg-neutral-50 px-4 py-3 transition hover:border-neutral-400 hover:bg-white">
                <Link href={`/profile/${profile.user.id}/following`} className="block">
                  <p className="text-xs uppercase tracking-wide text-neutral-500">Following</p>
                  <p className="mt-1 text-xl font-semibold text-neutral-900">
                    {profile.following_count ?? 0}
                  </p>
                </Link>
              </div>
              <div className="rounded-2xl border border-neutral-200 bg-neutral-50 px-4 py-3">
                <p className="text-xs uppercase tracking-wide text-neutral-500">Joined</p>
                <p className="mt-1 text-sm font-semibold text-neutral-900">
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
              className="rounded-3xl border border-neutral-200 bg-white p-5 shadow-sm"
            >
              <h2 className="text-sm font-semibold text-neutral-900">About</h2>
              <div className="mt-3 grid gap-3 text-sm text-neutral-700 sm:grid-cols-2">
                <div>
                  <p className="text-xs uppercase tracking-wide text-neutral-500">Email</p>
                  <p className="mt-1">{profile.user.email ?? "Hidden"}</p>
                </div>
                <div>
                  <p className="text-xs uppercase tracking-wide text-neutral-500">Date of birth</p>
                  <p className="mt-1">{profile.user.date_of_birth ?? "Hidden"}</p>
                </div>
                <div className="sm:col-span-2">
                  <p className="text-xs uppercase tracking-wide text-neutral-500">Bio</p>
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
            className="rounded-3xl border border-neutral-200 bg-white p-5 shadow-sm"
          >
            <div className="flex items-center justify-between">
              <h2 className="text-lg font-semibold text-neutral-900">Posts</h2>
              <span className="text-xs text-neutral-500">{posts.length} post(s)</span>
            </div>

            {!profile ? null : profile.limited ? (
              <p className="mt-4 text-sm text-neutral-600">
                This profile is private. Follow to see posts and activity.
              </p>
            ) : posts.length === 0 ? (
              <p className="mt-4 text-sm text-neutral-600">No posts yet.</p>
            ) : (
              <div className="mt-4 space-y-4">
                {posts.map((post) => (
                  <article
                    key={post.id}
                    className="rounded-2xl border border-neutral-200 bg-neutral-50 p-4"
                  >
                    <header className="flex items-start justify-between gap-3">
                      <div>
                        <p className="text-sm font-semibold text-neutral-900">
                          {post.author_first_name} {post.author_last_name}
                        </p>
                        <p className="text-xs text-neutral-500">{shortDate(post.created_at)}</p>
                      </div>
                      <span className="rounded-full border border-neutral-200 bg-white px-2.5 py-1 text-[11px] uppercase tracking-wide text-neutral-600">
                        {post.privacy}
                      </span>
                    </header>

                    <p className="mt-3 text-sm text-neutral-700">{post.content}</p>

                    {post.media_path ? (
                      <div className="mt-3 overflow-hidden rounded-2xl border border-neutral-200 bg-white">
                        <img
                          src={toMediaUrl(apiBaseUrl, post.media_path)}
                          alt="Post media"
                          className="max-h-[520px] w-full object-contain bg-white"
                        />
                      </div>
                    ) : null}

                    <footer className="mt-3 flex items-center gap-4 text-xs text-neutral-500">
                      <span>{post.comment_count} comments</span>
                      <span>{post.like_count} likes</span>
                      <span>{post.dislike_count} dislikes</span>
                    </footer>
                  </article>
                ))}
              </div>
            )}
          </motion.div>
        </section>

        <aside className="hidden lg:block">
          <div className="rounded-3xl border border-neutral-200 bg-white p-5 shadow-sm">
            <h3 className="text-sm font-semibold text-neutral-900">Profile status</h3>
            <p className="mt-2 text-xs text-neutral-500">
              {profile?.user.is_public
                ? "Public profiles are visible to all members."
                : "Private profiles require approval to view posts."}
            </p>
            <div className="mt-4 space-y-2 text-xs text-neutral-600">
              <div className="flex items-center justify-between rounded-2xl border border-neutral-200 bg-neutral-50 px-3 py-2">
                <span>Visibility</span>
                <span className="font-semibold text-neutral-800">
                  {profile?.user.is_public ? "Public" : "Private"}
                </span>
              </div>
              <div className="flex items-center justify-between rounded-2xl border border-neutral-200 bg-neutral-50 px-3 py-2">
                <span>Followers</span>
                <span className="font-semibold text-neutral-800">
                  {profile?.followers_count ?? 0}
                </span>
              </div>
              <div className="flex items-center justify-between rounded-2xl border border-neutral-200 bg-neutral-50 px-3 py-2">
                <span>Following</span>
                <span className="font-semibold text-neutral-800">
                  {profile?.following_count ?? 0}
                </span>
              </div>
            </div>
          </div>
        </aside>
      </main>
    </div>
  );
}
