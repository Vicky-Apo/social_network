"use client";

import { useCallback, useEffect, useMemo, useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { UserCheck, UserMinus, UserPlus } from "lucide-react";
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

type FollowRequest = {
  id: number;
  requester_id: number;
  target_id: number;
  status: string;
  created_at: string;
};

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

function toMediaUrl(apiBaseUrl: string, path?: string | null) {
  if (!path) return "";
  if (path.startsWith("http://") || path.startsWith("https://")) return path;
  const normalized = path.startsWith("/") ? path : `/${path}`;
  return `${apiBaseUrl}${normalized}`;
}

export default function FollowRequestsPage() {
  const router = useRouter();
  const [viewer, setViewer] = useState<User | null>(null);
  const [incoming, setIncoming] = useState<FollowRequest[]>([]);
  const [sent, setSent] = useState<FollowRequest[]>([]);
  const [usersByID, setUsersByID] = useState<Record<number, UserListItem>>({});
  const [followers, setFollowers] = useState<UserListItem[]>([]);
  const [following, setFollowing] = useState<UserListItem[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [actionError, setActionError] = useState<string | null>(null);

  const apiBaseUrl = useMemo(
    () =>
      process.env.NEXT_PUBLIC_API_BASE_URL?.trim().replace(/\/+$/, "") ||
      "http://localhost:8080",
    [],
  );

  const loadData = useCallback(async () => {
    setIsLoading(true);
    setError(null);
    setActionError(null);

    try {
      const meResponse = await fetch(`${apiBaseUrl}/auth/me`, {
        credentials: "include",
      });
      const meResult = (await meResponse.json().catch(() => null)) as ApiResponse<User> | null;
      if (!meResponse.ok || !meResult?.success || !meResult.data) {
        router.replace("/login");
        return;
      }
      setViewer(meResult.data);

      const [incomingRes, sentRes, usersRes, followersRes, followingRes] = await Promise.all([
        fetch(`${apiBaseUrl}/follow-requests`, { credentials: "include" }),
        fetch(`${apiBaseUrl}/follow-requests/sent`, { credentials: "include" }),
        fetch(`${apiBaseUrl}/users?limit=200&offset=0`, { credentials: "include" }),
        fetch(`${apiBaseUrl}/profiles/${meResult.data.id}/followers`, {
          credentials: "include",
        }),
        fetch(`${apiBaseUrl}/profiles/${meResult.data.id}/following`, {
          credentials: "include",
        }),
      ]);

      const incomingJson = (await incomingRes.json().catch(() => null)) as
        | ApiResponse<FollowRequest[]>
        | null;
      const sentJson = (await sentRes.json().catch(() => null)) as
        | ApiResponse<FollowRequest[]>
        | null;
      const usersJson = (await usersRes.json().catch(() => null)) as
        | ApiResponse<UserListItem[]>
        | null;
      const followersJson = (await followersRes.json().catch(() => null)) as
        | ApiResponse<UserListItem[]>
        | null;
      const followingJson = (await followingRes.json().catch(() => null)) as
        | ApiResponse<UserListItem[]>
        | null;

      if (!incomingRes.ok || !incomingJson?.success) {
        setError(incomingJson?.error || "Could not load incoming requests.");
        return;
      }
      if (!sentRes.ok || !sentJson?.success) {
        setError(sentJson?.error || "Could not load sent requests.");
        return;
      }

      setIncoming(incomingJson.data ?? []);
      setSent(sentJson.data ?? []);

      if (usersRes.ok && usersJson?.success) {
        const next = (usersJson.data ?? []).reduce<Record<number, UserListItem>>((acc, item) => {
          acc[item.id] = item;
          return acc;
        }, {});
        setUsersByID(next);
      }
      if (followersRes.ok && followersJson?.success) {
        setFollowers(followersJson.data ?? []);
      }
      if (followingRes.ok && followingJson?.success) {
        setFollowing(followingJson.data ?? []);
      }
    } catch {
      setError("Network error. Please try again.");
    } finally {
      setIsLoading(false);
    }
  }, [apiBaseUrl, router]);

  useEffect(() => {
    void loadData();
  }, [loadData]);

  const updateRequest = async (id: number, status: "accepted" | "declined" | "canceled") => {
    setActionError(null);
    try {
      const response = await fetch(`${apiBaseUrl}/follow-requests/${id}`, {
        method: "PATCH",
        headers: { "Content-Type": "application/json" },
        credentials: "include",
        body: JSON.stringify({ status }),
      });
      if (!response.ok) {
        const result = (await response.json().catch(() => null)) as ApiResponse<unknown> | null;
        setActionError(result?.error || "Could not update request.");
        return;
      }
      setIncoming((prev) => prev.filter((item) => item.id !== id));
      setSent((prev) => prev.filter((item) => item.id !== id));
    } catch {
      setActionError("Network error. Please try again.");
    }
  };

  const removeFollower = async (followerID: number) => {
    setActionError(null);
    try {
      const response = await fetch(`${apiBaseUrl}/followers/${followerID}`, {
        method: "DELETE",
        credentials: "include",
      });
      if (!response.ok) {
        const result = (await response.json().catch(() => null)) as ApiResponse<unknown> | null;
        setActionError(result?.error || "Could not remove follower.");
        return;
      }
      setFollowers((prev) => prev.filter((item) => item.id !== followerID));
    } catch {
      setActionError("Network error. Please try again.");
    }
  };

  const unfollowUser = async (targetID: number) => {
    setActionError(null);
    try {
      const response = await fetch(`${apiBaseUrl}/users/${targetID}/followers`, {
        method: "DELETE",
        credentials: "include",
      });
      if (!response.ok) {
        const result = (await response.json().catch(() => null)) as ApiResponse<unknown> | null;
        setActionError(result?.error || "Could not unfollow user.");
        return;
      }
      setFollowing((prev) => prev.filter((item) => item.id !== targetID));
    } catch {
      setActionError("Network error. Please try again.");
    }
  };

  const resolveUser = (id: number) => usersByID[id];
  return (
    <div
      className="min-h-screen text-neutral-100"
      style={{
        backgroundImage: "url('/requests-bg.png')",
        backgroundSize: "cover",
        backgroundPosition: "center",
        backgroundAttachment: "fixed",
      }}
    >
      <TopNav user={viewer ?? undefined} onLogout={() => router.replace("/login")} variant="dark" />

      <main className="mx-auto grid w-full max-w-6xl gap-6 px-4 py-6 sm:px-6 lg:grid-cols-[240px_minmax(0,1fr)_280px]">
        <aside className="hidden lg:block">
          <LeftNav user={viewer ?? undefined} activeHref="/follow-requests" variant="dark" />
        </aside>

        <section className="space-y-6">
          <motion.div
            initial="hidden"
            whileInView="show"
            viewport={viewportOnce}
            variants={fadeUp}
            className="rounded-3xl border border-white/10 bg-white/5 p-5 backdrop-blur-sm"
          >
            <div className="flex flex-wrap items-center justify-between gap-3">
              <div>
                <h1 className="text-xl font-semibold tracking-tight text-white">
                  Follow requests
                </h1>
                <p className="text-sm text-white">
                  Manage incoming and sent follow requests.
                </p>
              </div>
              <span className="rounded-full border border-white/20 bg-white/5 px-3 py-1 text-xs text-white">
                {incoming.length + sent.length} total
              </span>
            </div>
          </motion.div>

          {isLoading ? (
            <motion.div
              initial="hidden"
              whileInView="show"
              viewport={viewportOnce}
              variants={fadeUp}
              className="rounded-3xl border border-white/10 bg-white/5 p-5 text-sm text-white backdrop-blur-sm"
            >
              Loading requests...
            </motion.div>
          ) : error ? (
            <motion.div
              initial="hidden"
              whileInView="show"
              viewport={viewportOnce}
              variants={fadeUp}
              className="rounded-3xl border border-rose-500/30 bg-rose-500/10 p-5 text-sm text-rose-400"
            >
              {error}
            </motion.div>
          ) : (
            <>
              <motion.div
                initial="hidden"
                whileInView="show"
                viewport={viewportOnce}
                variants={fadeUp}
                className="rounded-3xl border border-white/10 bg-white/5 p-5 backdrop-blur-sm"
              >
                <h2 className="text-sm font-semibold text-white">Incoming requests</h2>
                {incoming.length === 0 ? (
                  <p className="mt-3 text-sm text-white">
                    No incoming requests right now.
                  </p>
                ) : (
                  <div className="mt-4 space-y-3">
                    {incoming.map((req) => {
                      const requester = resolveUser(req.requester_id);
                      return (
                        <div
                          key={req.id}
                          className="flex flex-wrap items-center justify-between gap-3 rounded-2xl border border-white/10 bg-white/5 px-4 py-3"
                        >
                          <div>
                            <p className="text-sm font-semibold text-white">
                              {requester
                                ? `${requester.first_name} ${requester.last_name}`
                                : "User"}
                            </p>
                            <p className="text-xs text-neutral-400">
                              @{requester?.nickname || "user"} ·{" "}
                              {shortDate(req.created_at)}
                            </p>
                          </div>
                          <div className="flex items-center gap-2">
                            <button
                              type="button"
                              onClick={() => updateRequest(req.id, "accepted")}
                              className="inline-flex items-center gap-2 rounded-full bg-emerald-600 px-3 py-2 text-xs font-semibold text-white transition hover:bg-emerald-700"
                            >
                              <UserCheck className="h-3.5 w-3.5" />
                              Accept
                            </button>
                            <button
                              type="button"
                              onClick={() => updateRequest(req.id, "declined")}
                              className="inline-flex items-center gap-2 rounded-full border border-white/20 bg-white/5 px-3 py-2 text-xs font-semibold text-neutral-300 transition hover:bg-white/10 hover:text-white"
                            >
                              <UserMinus className="h-3.5 w-3.5" />
                              Decline
                            </button>
                          </div>
                        </div>
                      );
                    })}
                  </div>
                )}
              </motion.div>

              <motion.div
                initial="hidden"
                whileInView="show"
                viewport={viewportOnce}
                variants={fadeUp}
                className="rounded-3xl border border-white/10 bg-white/5 p-5 backdrop-blur-sm"
              >
                <h2 className="text-sm font-semibold text-white">Sent requests</h2>
                {sent.length === 0 ? (
                  <p className="mt-3 text-sm text-white">No sent requests.</p>
                ) : (
                  <div className="mt-4 space-y-3">
                    {sent.map((req) => {
                      const target = resolveUser(req.target_id);
                      return (
                        <div
                          key={req.id}
                          className="flex flex-wrap items-center justify-between gap-3 rounded-2xl border border-white/10 bg-white/5 px-4 py-3"
                        >
                          <div>
                            <p className="text-sm font-semibold text-white">
                              {target
                                ? `${target.first_name} ${target.last_name}`
                                : "User"}
                            </p>
                            <p className="text-xs text-neutral-400">
                              @{target?.nickname || "user"} ·{" "}
                              {shortDate(req.created_at)}
                            </p>
                          </div>
                          <button
                            type="button"
                            onClick={() => updateRequest(req.id, "canceled")}
                            className="inline-flex items-center gap-2 rounded-full border border-white/20 bg-white/5 px-3 py-2 text-xs font-semibold text-neutral-300 transition hover:bg-white/10 hover:text-white"
                          >
                            <UserMinus className="h-3.5 w-3.5" />
                            Cancel
                          </button>
                        </div>
                      );
                    })}
                  </div>
                )}
              </motion.div>

            </>
          )}

          {actionError ? (
            <motion.div
              initial="hidden"
              whileInView="show"
              viewport={viewportOnce}
              variants={fadeUp}
              className="rounded-2xl border border-rose-500/30 bg-rose-500/10 px-4 py-3 text-xs text-rose-400"
            >
              {actionError}
            </motion.div>
          ) : null}
        </section>

        <aside className="hidden space-y-4 lg:block">
          <motion.div
            initial="hidden"
            whileInView="show"
            viewport={viewportOnce}
            variants={fadeUp}
            className="rounded-3xl border border-white/10 bg-white/5 p-5 backdrop-blur-sm"
          >
            <h2 className="text-sm font-semibold text-white">Followers</h2>
            {followers.length === 0 ? (
              <p className="mt-3 text-sm text-white">No followers yet.</p>
            ) : (
              <div className="mt-4 space-y-3">
                {followers.map((follower) => (
                  <div
                    key={follower.id}
                    className="flex flex-wrap items-center justify-between gap-3 rounded-2xl border border-white/10 bg-white/5 px-4 py-3"
                  >
                    <div>
                      <p className="text-sm font-semibold text-white">
                        {follower.first_name} {follower.last_name}
                      </p>
                      <p className="text-xs text-neutral-400">
                        @{follower.nickname || "user"}
                      </p>
                    </div>
                    <button
                      type="button"
                      onClick={() => removeFollower(follower.id)}
                      className="inline-flex items-center gap-2 rounded-full border border-white/20 bg-white/5 px-3 py-2 text-xs font-semibold text-neutral-300 transition hover:bg-white/10 hover:text-white"
                    >
                      <UserMinus className="h-3.5 w-3.5" />
                      Remove
                    </button>
                  </div>
                ))}
              </div>
            )}
          </motion.div>

          <motion.div
            initial="hidden"
            whileInView="show"
            viewport={viewportOnce}
            variants={fadeUp}
            className="rounded-3xl border border-white/10 bg-white/5 p-5 backdrop-blur-sm"
          >
            <h2 className="text-sm font-semibold text-white">Following</h2>
            {following.length === 0 ? (
              <p className="mt-3 text-sm text-white">Not following anyone.</p>
            ) : (
              <div className="mt-4 space-y-3">
                {following.map((target) => (
                  <div
                    key={target.id}
                    className="flex flex-wrap items-center justify-between gap-3 rounded-2xl border border-white/10 bg-white/5 px-4 py-3"
                  >
                    <div>
                      <p className="text-sm font-semibold text-white">
                        {target.first_name} {target.last_name}
                      </p>
                      <p className="text-xs text-neutral-400">
                        @{target.nickname || "user"}
                      </p>
                    </div>
                    <button
                      type="button"
                      onClick={() => unfollowUser(target.id)}
                      className="inline-flex items-center gap-2 rounded-full border border-white/20 bg-white/5 px-3 py-2 text-xs font-semibold text-neutral-300 transition hover:bg-white/10 hover:text-white"
                    >
                      <UserMinus className="h-3.5 w-3.5" />
                      Unfollow
                    </button>
                  </div>
                ))}
              </div>
            )}
          </motion.div>
        </aside>
      </main>
    </div>
  );
}
