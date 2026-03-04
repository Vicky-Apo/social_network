"use client";

import { useCallback, useEffect, useMemo, useState } from "react";
import Link from "next/link";
import { useParams, useRouter } from "next/navigation";
import { motion } from "framer-motion";
import TopNav from "@/components/TopNav";
import LeftNav from "@/components/LeftNav";
import Avatar from "@/components/Avatar";
import { fadeUp, viewportOnce } from "@/components/Motion";
import { toMediaUrl } from "@/lib/media";
import { apiFetch, apiFetchJson, getApiBaseUrl } from "@/lib/api";
import { ApiResponse } from "@/lib/types";

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

export default function FollowingPage() {
  const router = useRouter();
  const params = useParams<{ id: string }>();
  const profileID = typeof params?.id === "string" ? Number(params.id) : NaN;

  const [viewer, setViewer] = useState<User | null>(null);
  const [following, setFollowing] = useState<UserListItem[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [actionError, setActionError] = useState<string | null>(null);

  const apiBaseUrl = useMemo(() => getApiBaseUrl(), []);

  const loadFollowing = useCallback(async () => {
    if (!Number.isFinite(profileID) || profileID <= 0) {
      setError("Invalid profile id.");
      setIsLoading(false);
      return;
    }

    setIsLoading(true);
    setError(null);

    try {
      const { response: meResponse, result: meResult } = await apiFetchJson<ApiResponse<User>>(
        "/auth/me",
        {},
        apiBaseUrl,
      );
      if (!meResponse.ok || !meResult?.success || !meResult.data) {
        router.replace("/login");
        return;
      }
      setViewer(meResult.data);

      const { response, result } = await apiFetchJson<ApiResponse<UserListItem[]>>(
        `/profiles/${profileID}/following`,
        {},
        apiBaseUrl,
      );
      if (!response.ok || !result?.success) {
        setError(result?.error || "Could not load following list.");
        return;
      }
      setFollowing(result.data ?? []);
    } catch {
      setError("Network error. Please try again.");
    } finally {
      setIsLoading(false);
    }
  }, [apiBaseUrl, profileID, router]);

  useEffect(() => {
    void loadFollowing();
  }, [loadFollowing]);

  const unfollow = async (id: number) => {
    setActionError(null);
    try {
      const response = await apiFetch(`/users/${id}/followers`, { method: "DELETE" }, apiBaseUrl);
      if (!response.ok) {
        const result = (await response.json().catch(() => null)) as ApiResponse<unknown> | null;
        setActionError(result?.error || "Could not unfollow user.");
        return;
      }
      setFollowing((prev) => prev.filter((item) => item.id !== id));
    } catch {
      setActionError("Network error. Please try again.");
    }
  };

  const isOwner = viewer?.id === profileID;

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

      <main className="mx-auto grid w-full max-w-6xl gap-6 px-4 py-6 sm:px-6 lg:grid-cols-[240px_minmax(0,1fr)_280px]">
        <aside className="hidden lg:block">
          <LeftNav user={viewer ?? undefined} activeHref="/dashboard" variant="dark" />
        </aside>

        <section className="space-y-5">
          <motion.div
            initial="hidden"
            whileInView="show"
            viewport={viewportOnce}
            variants={fadeUp}
            className="rounded-3xl border border-neutral-200 bg-white p-5 shadow-sm"
          >
            <div className="flex flex-wrap items-center justify-between gap-3">
              <div>
                <h1 className="text-xl font-semibold tracking-tight text-neutral-900">Following</h1>
                <p className="text-sm text-neutral-600">
                  {isOwner ? "People you follow." : "Following list."}
                </p>
              </div>
              <button
                type="button"
                onClick={() => router.back()}
                className="inline-flex items-center gap-2 rounded-full border border-neutral-200 bg-white px-3 py-2 text-xs font-semibold text-neutral-700 transition hover:border-neutral-400 hover:text-neutral-900"
              >
                Back
              </button>
            </div>
          </motion.div>

          {isLoading ? (
            <motion.div
              initial="hidden"
              whileInView="show"
              viewport={viewportOnce}
              variants={fadeUp}
              className="rounded-3xl border border-neutral-200 bg-white p-5 text-sm text-neutral-600 shadow-sm"
            >
              Loading following...
            </motion.div>
          ) : error ? (
            <motion.div
              initial="hidden"
              whileInView="show"
              viewport={viewportOnce}
              variants={fadeUp}
              className="rounded-3xl border border-rose-200 bg-rose-50 p-5 text-sm text-rose-700"
            >
              {error}
            </motion.div>
          ) : following.length === 0 ? (
            <motion.div
              initial="hidden"
              whileInView="show"
              viewport={viewportOnce}
              variants={fadeUp}
              className="rounded-3xl border border-neutral-200 bg-white p-5 text-sm text-neutral-600 shadow-sm"
            >
              No following yet.
            </motion.div>
          ) : (
            <motion.div
              initial="hidden"
              whileInView="show"
              viewport={viewportOnce}
              variants={fadeUp}
              className="rounded-3xl border border-neutral-200 bg-white p-5 shadow-sm"
            >
              <div className="space-y-3">
                {following.map((person) => (
                  <div
                    key={person.id}
                    className="flex flex-wrap items-center justify-between gap-3 rounded-2xl border border-neutral-200 bg-neutral-50 px-4 py-3 transition hover:border-neutral-400 hover:bg-white"
                  >
                    <Link href={`/profile/${person.id}`} className="flex items-center gap-3">
                      <Avatar
                        src={
                          person.avatar_path ? toMediaUrl(apiBaseUrl, person.avatar_path) : null
                        }
                        name={`${person.first_name} ${person.last_name}`}
                        size={40}
                        textClassName="text-xs"
                      />
                      <div>
                        <p className="text-sm font-semibold text-neutral-800">
                          {person.first_name} {person.last_name}
                        </p>
                        <p className="text-xs text-neutral-500">
                          @{person.nickname || "user"}
                        </p>
                      </div>
                    </Link>
                    {isOwner ? (
                      <button
                        type="button"
                        onClick={() => unfollow(person.id)}
                        className="inline-flex items-center gap-2 rounded-full border border-neutral-300 bg-white px-3 py-2 text-xs font-semibold text-neutral-700 transition hover:border-neutral-400"
                      >
                        Unfollow
                      </button>
                    ) : null}
                  </div>
                ))}
              </div>
              {actionError ? (
                <p className="mt-3 text-xs text-rose-600">{actionError}</p>
              ) : null}
            </motion.div>
          )}
        </section>

        <aside className="hidden lg:block">
          <div className="rounded-3xl border border-neutral-200 bg-white p-5 shadow-sm">
            <h3 className="text-sm font-semibold text-neutral-900">Tips</h3>
            <p className="mt-2 text-xs text-neutral-500">
              Unfollowing will hide their private content from your feed.
            </p>
          </div>
        </aside>
      </main>
    </div>
  );
}
