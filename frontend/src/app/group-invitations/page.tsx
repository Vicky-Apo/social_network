"use client";

import { useEffect, useMemo, useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
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

type GroupInvitation = {
  id: number;
  group_id: number;
  inviter_id: number;
  invitee_id: number;
  status: string;
  created_at?: string;
};

type GroupSummary = {
  id: number;
  title?: string | null;
  name?: string | null;
};

function shortDate(value?: string) {
  if (!value) return "Just now";
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return "Just now";
  return date.toLocaleDateString(undefined, { month: "short", day: "numeric" });
}

export default function GroupInvitationsPage() {
  const router = useRouter();
  const [viewer, setViewer] = useState<User | null>(null);
  const [invites, setInvites] = useState<GroupInvitation[]>([]);
  const [groupsByID, setGroupsByID] = useState<Record<number, GroupSummary>>({});
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

        const invitesResponse = await fetch(`${apiBaseUrl}/group-invitations`, {
          credentials: "include",
        });
        const invitesResult = (await invitesResponse.json().catch(() => null)) as
          | ApiResponse<GroupInvitation[]>
          | null;
        if (!invitesResponse.ok || !invitesResult?.success) {
          setError(invitesResult?.error || "Could not load invitations.");
          setInvites([]);
          return;
        }
        setInvites(invitesResult.data ?? []);
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
  }, [apiBaseUrl, router]);

  useEffect(() => {
    const missing = Array.from(new Set(invites.map((inv) => inv.group_id))).filter(
      (id) => !groupsByID[id],
    );
    if (missing.length === 0) return;
    let cancelled = false;
    Promise.all(
      missing.map(async (id) => {
        try {
          const response = await fetch(`${apiBaseUrl}/groups/${id}`, {
            credentials: "include",
          });
          const result = (await response.json().catch(() => null)) as
            | ApiResponse<GroupSummary>
            | null;
          if (!response.ok || !result?.success || !result.data) return null;
          return result.data;
        } catch {
          return null;
        }
      }),
    ).then((items) => {
      if (cancelled) return;
      const mapped: Record<number, GroupSummary> = {};
      items.forEach((item) => {
        if (item && typeof item.id === "number") mapped[item.id] = item;
      });
      if (Object.keys(mapped).length > 0) {
        setGroupsByID((prev) => ({ ...prev, ...mapped }));
      }
    });
    return () => {
      cancelled = true;
    };
  }, [apiBaseUrl, groupsByID, invites]);

  const updateInvite = async (id: number, status: "accepted" | "declined") => {
    setActionError(null);
    try {
      const response = await fetch(`${apiBaseUrl}/group-invitations/${id}`, {
        method: "PATCH",
        headers: { "Content-Type": "application/json" },
        credentials: "include",
        body: JSON.stringify({ status }),
      });
      if (!response.ok) {
        const result = (await response.json().catch(() => null)) as ApiResponse<unknown> | null;
        setActionError(result?.error || "Could not update invitation.");
        return;
      }
      setInvites((prev) => prev.filter((item) => item.id !== id));
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
                  Group invitations
                </h1>
                <p className="text-sm text-neutral-600">
                  Accept or decline invitations you have received.
                </p>
              </div>
              <Link
                href="/groups"
                className="inline-flex items-center gap-2 rounded-full border border-neutral-200 bg-white px-3 py-2 text-xs font-semibold text-neutral-700 transition hover:border-neutral-400 hover:text-neutral-900"
              >
                Back to groups
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
              Loading invitations...
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
          ) : invites.length === 0 ? (
            <motion.div
              initial="hidden"
              whileInView="show"
              viewport={viewportOnce}
              variants={fadeUp}
              className="rounded-3xl border border-neutral-200 bg-white p-5 text-sm text-neutral-600 shadow-sm"
            >
              No invitations right now.
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
                {invites.map((invite) => (
                  <div
                    key={invite.id}
                    className="flex flex-wrap items-center justify-between gap-3 rounded-2xl border border-neutral-200 bg-neutral-50 px-4 py-3"
                  >
                    <div>
                      <p className="text-sm font-semibold text-neutral-800">
                        {groupsByID[invite.group_id]?.title ||
                          groupsByID[invite.group_id]?.name ||
                          "Group"}
                      </p>
                      <p className="text-xs text-neutral-500">
                        Invited by user #{invite.inviter_id} · {shortDate(invite.created_at)}
                      </p>
                    </div>
                    <div className="flex items-center gap-2">
                      <button
                        type="button"
                        onClick={() => updateInvite(invite.id, "accepted")}
                        className="inline-flex items-center gap-2 rounded-full bg-emerald-600 px-3 py-2 text-xs font-semibold text-white transition hover:bg-emerald-700"
                      >
                        Accept
                      </button>
                      <button
                        type="button"
                        onClick={() => updateInvite(invite.id, "declined")}
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
