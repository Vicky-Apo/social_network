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

type JoinRequest = {
  id: number;
  group_id: number;
  user_id: number;
  status: string;
  created_at?: string;
};

function shortDate(value?: string) {
  if (!value) return "Just now";
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return "Just now";
  return date.toLocaleDateString(undefined, { month: "short", day: "numeric" });
}

function toMediaUrl(apiBaseUrl: string, path?: string | null) {
  if (!path) return "";
  if (path.startsWith("http://") || path.startsWith("https://")) return path;
  const normalized = path.startsWith("/") ? path : `/${path}`;
  return `${apiBaseUrl}${normalized}`;
}

async function fetchProfileSummary(
  apiBaseUrl: string,
  id: number,
): Promise<User | null> {
  try {
    const response = await fetch(`${apiBaseUrl}/profiles/${id}`, { credentials: "include" });
    const result = (await response.json().catch(() => null)) as
      | ApiResponse<{ user?: User }>
      | null;
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
  const [requesterProfiles, setRequesterProfiles] = useState<Record<number, User>>({});
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [actionError, setActionError] = useState<string | null>(null);

  const apiBaseUrl = useMemo(
    () =>
      process.env.NEXT_PUBLIC_API_BASE_URL?.trim().replace(/\/+$/, "") ||
      "http://localhost:8080",
    [],
  );

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
        const meResponse = await fetch(`${apiBaseUrl}/auth/me`, { credentials: "include" });
        const meResult = (await meResponse.json().catch(() => null)) as ApiResponse<User> | null;
        if (!meResponse.ok || !meResult?.success || !meResult.data) {
          router.replace("/login");
          return;
        }
        if (!cancelled) {
          setViewer(meResult.data);
        }

        const requestsResponse = await fetch(
          `${apiBaseUrl}/groups/${groupID}/join-requests`,
          { credentials: "include" },
        );
        const requestsResult = (await requestsResponse.json().catch(() => null)) as
          | ApiResponse<JoinRequest[]>
          | null;
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

  useEffect(() => {
    if (requests.length === 0) return;
    const missing = Array.from(new Set(requests.map((req) => req.user_id))).filter(
      (id) => !requesterProfiles[id],
    );
    if (missing.length === 0) return;
    let cancelled = false;

    Promise.all(
      missing.map((id) => fetchProfileSummary(apiBaseUrl, id).then((user) => [id, user] as const)),
    )
      .then((entries) => {
        if (cancelled) return;
        const updates: Record<number, User> = {};
        for (const [id, user] of entries) {
          if (user) {
            updates[id] = user;
          }
        }
        if (Object.keys(updates).length > 0) {
          setRequesterProfiles((prev) => ({ ...prev, ...updates }));
        }
      })
      .catch(() => undefined);

    return () => {
      cancelled = true;
    };
  }, [apiBaseUrl, requests, requesterProfiles]);

  const updateRequest = async (id: number, status: "accepted" | "declined") => {
    setActionError(null);
    try {
      const response = await fetch(`${apiBaseUrl}/group-join-requests/${id}`, {
        method: "PATCH",
        headers: { "Content-Type": "application/json" },
        credentials: "include",
        body: JSON.stringify({ status }),
      });
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
    <div className="min-h-screen bg-neutral-50 text-neutral-900">
      <TopNav user={viewer ?? undefined} onLogout={() => router.replace("/login")} />

      <main className="mx-auto grid w-full max-w-6xl gap-6 px-4 py-6 sm:px-6 lg:grid-cols-[240px_minmax(0,1fr)]">
        <aside className="hidden lg:block">
          <LeftNav user={viewer ?? undefined} activeHref="/groups" />
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
                <h1 className="text-xl font-semibold tracking-tight text-neutral-900">
                  Join requests
                </h1>
                <p className="text-sm text-neutral-600">
                  Review requests to join this group.
                </p>
              </div>
              <Link
                href={`/groups/${groupID}`}
                className="inline-flex items-center gap-2 rounded-full border border-neutral-200 bg-white px-3 py-2 text-xs font-semibold text-neutral-700 transition hover:border-neutral-400 hover:text-neutral-900"
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
              className="rounded-3xl border border-neutral-200 bg-white p-5 text-sm text-neutral-600 shadow-sm"
            >
              Loading join requests...
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
          ) : requests.length === 0 ? (
            <motion.div
              initial="hidden"
              whileInView="show"
              viewport={viewportOnce}
              variants={fadeUp}
              className="rounded-3xl border border-neutral-200 bg-white p-5 text-sm text-neutral-600 shadow-sm"
            >
              No join requests yet.
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
                {requests.map((req) => (
                  <div
                    key={req.id}
                    className="flex flex-wrap items-center justify-between gap-3 rounded-2xl border border-neutral-200 bg-neutral-50 px-4 py-3"
                  >
                    <div className="flex items-center gap-3">
                      <Avatar
                        src={
                          requesterProfiles[req.user_id]?.avatar_path
                            ? toMediaUrl(apiBaseUrl, requesterProfiles[req.user_id]?.avatar_path)
                            : null
                        }
                        name={`${requesterProfiles[req.user_id]?.first_name ?? ""} ${requesterProfiles[req.user_id]?.last_name ?? ""}`.trim()}
                        size={36}
                        textClassName="text-[11px]"
                      />
                      <div>
                        <p className="text-sm font-semibold text-neutral-800">
                          {requesterProfiles[req.user_id]
                            ? `${requesterProfiles[req.user_id]?.first_name} ${requesterProfiles[req.user_id]?.last_name}`
                            : "User"}
                        </p>
                        <p className="text-xs text-neutral-500">
                          @{requesterProfiles[req.user_id]?.nickname || "user"} ·
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
                        className="inline-flex items-center gap-2 rounded-full border border-neutral-300 bg-white px-3 py-2 text-xs font-semibold text-neutral-700 transition hover:border-neutral-400"
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
              className="rounded-2xl border border-rose-200 bg-rose-50 px-4 py-3 text-xs text-rose-700"
            >
              {actionError}
            </motion.div>
          ) : null}
        </section>
      </main>
    </div>
  );
}
