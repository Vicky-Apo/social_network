"use client";

import { useEffect, useMemo, useState } from "react";
import Link from "next/link";
import { useParams, useRouter } from "next/navigation";
import { ArrowLeft } from "lucide-react";
import { motion } from "framer-motion";
import TopNav from "@/components/TopNav";
import LeftNav from "@/components/LeftNav";
import Avatar from "@/components/Avatar";
import { fadeUp, viewportOnce } from "@/components/Motion";
import { shortDate } from "@/lib/date";
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

type JoinRequest = {
  id: number;
  group_id: number;
  user_id: number;
  status: string;
  created_at?: string;
  user?: User | null;
};

async function fetchProfileSummary(
  apiBaseUrl: string,
  id: number,
): Promise<User | null> {
  try {
    const { response, result } = await apiFetchJson<ApiResponse<{ user?: User }>>(
      `/profiles/${id}`,
      {},
      apiBaseUrl,
    );
    if (!response.ok || !result?.success || !result.data?.user) {
      return null;
    }
    return result.data.user;
  } catch {
    return null;
  }
}

export default function GroupJoinRequestsPage() {
  const router = useRouter();
  const params = useParams<{ id: string }>();
  const groupID = typeof params?.id === "string" ? Number(params.id) : NaN;

  const [viewer, setViewer] = useState<User | null>(null);
  const [requests, setRequests] = useState<JoinRequest[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [actionError, setActionError] = useState<string | null>(null);

  const apiBaseUrl = useMemo(() => getApiBaseUrl(), []);

  useEffect(() => {
    if (!Number.isFinite(groupID) || groupID <= 0) {
      setError("Invalid group id.");
      setIsLoading(false);
      return;
    }

    let cancelled = false;
    const load = async () => {
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
        if (!cancelled) {
          setViewer(meResult.data);
        }

        const { response: requestsResponse, result: requestsResult } = await apiFetchJson<
          ApiResponse<JoinRequest[]>
        >(`/groups/${groupID}/join-requests`, {}, apiBaseUrl);
        if (!requestsResponse.ok || !requestsResult?.success) {
          setError(requestsResult?.error || "Could not load join requests.");
          setRequests([]);
          return;
        }
        setRequests(requestsResult.data ?? []);
      } catch {
        setError("Network error. Please try again.");
      } finally {
        setIsLoading(false);
      }
    };
    void load();
    return () => {
      cancelled = true;
    };
  }, [apiBaseUrl, groupID, router]);

  const updateRequest = async (id: number, status: "accepted" | "declined") => {
    setActionError(null);
    try {
      const response = await apiFetch(
        `/group-join-requests/${id}`,
        {
          method: "PATCH",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ status }),
        },
        apiBaseUrl,
      );
      if (!response.ok) {
        const result = (await response.json().catch(() => null)) as ApiResponse<unknown> | null;
        setActionError(result?.error || "Could not update join request.");
        return;
      }
      setRequests((prev) => prev.filter((item) => item.id !== id));
    } catch {
      setActionError("Network error. Please try again.");
    }
  };

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
          <LeftNav user={viewer ?? undefined} activeHref="/groups" variant="dark" />
        </aside>

        <section className="space-y-5">
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
                  Join requests
                </h1>
                <p className="text-sm text-neutral-400">
                  Review requests to join this group.
                </p>
              </div>
              <Link
                href={`/groups/${groupID}`}
                className="inline-flex items-center gap-2 rounded-full border border-white/20 bg-white/5 px-3 py-2 text-xs font-semibold text-neutral-300 transition hover:bg-white/10 hover:text-white"
              >
                <ArrowLeft className="h-3.5 w-3.5" />
                Back to group
              </Link>
            </div>
          </motion.div>

          {isLoading ? (
            <motion.div
              initial="hidden"
              whileInView="show"
              viewport={viewportOnce}
              variants={fadeUp}
              className="rounded-3xl border border-white/10 bg-white/5 p-5 text-sm text-neutral-400 backdrop-blur-sm"
            >
              Loading join requests...
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
          ) : requests.length === 0 ? (
            <motion.div
              initial="hidden"
              whileInView="show"
              viewport={viewportOnce}
              variants={fadeUp}
              className="rounded-3xl border border-white/10 bg-white/5 p-5 text-sm text-neutral-400 backdrop-blur-sm"
            >
              No join requests yet.
            </motion.div>
          ) : (
            <motion.div
              initial="hidden"
              whileInView="show"
              viewport={viewportOnce}
              variants={fadeUp}
              className="rounded-3xl border border-white/10 bg-white/5 p-5 backdrop-blur-sm"
            >
              <div className="space-y-3">
                {requests.map((req) => (
                  <div
                    key={req.id}
                    className="flex flex-wrap items-center justify-between gap-3 rounded-2xl border border-white/10 bg-white/5 px-4 py-3"
                  >
                    <div className="flex items-center gap-3">
                      <Avatar
                        src={
                          req.user?.avatar_path
                            ? toMediaUrl(apiBaseUrl, req.user?.avatar_path)
                            : null
                        }
                        name={`${req.user?.first_name ?? ""} ${req.user?.last_name ?? ""}`.trim()}
                        size={36}
                        textClassName="text-[11px]"
                      />
                      <div>
                        <p className="text-sm font-semibold text-white">
                          {req.user
                            ? `${req.user?.first_name} ${req.user?.last_name}`
                            : "User"}
                        </p>
                        <p className="text-xs text-neutral-400">
                          @{req.user?.nickname || "user"} ·
                          Requested {shortDate(req.created_at)}
                        </p>
                      </div>
                    </div>
                    <div className="flex items-center gap-2">
                      <button
                        type="button"
                        onClick={() => updateRequest(req.id, "accepted")}
                        className="inline-flex items-center gap-2 rounded-full bg-emerald-600 px-3 py-2 text-xs font-semibold text-white transition hover:bg-emerald-700"
                      >
                        Accept
                      </button>
                      <button
                        type="button"
                        onClick={() => updateRequest(req.id, "declined")}
                        className="inline-flex items-center gap-2 rounded-full border border-white/20 bg-white/5 px-3 py-2 text-xs font-semibold text-neutral-300 transition hover:bg-white/10 hover:text-white"
                      >
                        Decline
                      </button>
                    </div>
                  </div>
                ))}
              </div>
            </motion.div>
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
      </main>
    </div>
  );
}
