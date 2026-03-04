"use client";

import { useEffect, useMemo, useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { motion } from "framer-motion";
import TopNav from "@/components/TopNav";
import LeftNav from "@/components/LeftNav";
import { fadeUp, viewportOnce } from "@/components/Motion";
import { shortDate } from "@/lib/date";
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

type GroupInvitation = {
  id: number;
  group_id: number;
  group_title?: string | null;
  inviter_id: number;
  invitee_id: number;
  status: string;
  created_at?: string;
};

export default function GroupInvitationsPage() {
  const router = useRouter();
  const [viewer, setViewer] = useState<User | null>(null);
  const [invites, setInvites] = useState<GroupInvitation[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [actionError, setActionError] = useState<string | null>(null);

  const apiBaseUrl = useMemo(() => getApiBaseUrl(), []);

  useEffect(() => {
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

        const { response: invitesResponse, result: invitesResult } = await apiFetchJson<
          ApiResponse<GroupInvitation[]>
        >("/group-invitations", {}, apiBaseUrl);
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

  const updateInvite = async (id: number, status: "accepted" | "declined") => {
    setActionError(null);
    try {
      const response = await apiFetch(
        `/group-invitations/${id}`,
        {
          method: "PATCH",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ status }),
        },
        apiBaseUrl,
      );
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

      <main className="mx-auto grid w-full max-w-6xl gap-6 px-4 py-6 sm:px-6 lg:grid-cols-[220px_minmax(0,1fr)]">
        <aside className="hidden lg:block">
          <LeftNav user={viewer ?? undefined} activeHref="/groups" variant="dark" />
        </aside>

        <section className="space-y-5">
          <motion.div
            initial="hidden"
            whileInView="show"
            viewport={viewportOnce}
            variants={fadeUp}
            className="rounded-2xl border border-white/10 bg-white/5 p-5 backdrop-blur-sm"
          >
            <div className="flex flex-wrap items-center justify-between gap-3">
              <div>
                <h1 className="text-xl font-semibold tracking-tight text-white">
                  Group invitations
                </h1>
                <p className="text-sm text-neutral-400">
                  Accept or decline invitations you have received.
                </p>
              </div>
              <Link
                href="/groups"
                className="inline-flex items-center gap-2 rounded-xl border border-white/20 bg-white/5 px-3 py-2 text-xs font-semibold text-neutral-300 transition hover:bg-white/10 hover:text-white"
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
              className="rounded-2xl border border-white/10 bg-white/5 p-5 text-sm text-neutral-400 backdrop-blur-sm"
            >
              Loading invitations...
            </motion.div>
          ) : error ? (
            <motion.div
              initial="hidden"
              whileInView="show"
              viewport={viewportOnce}
              variants={fadeUp}
              className="rounded-2xl border border-rose-500/30 bg-rose-500/10 p-5 text-sm text-rose-400 backdrop-blur-sm"
            >
              {error}
            </motion.div>
          ) : invites.length === 0 ? (
            <motion.div
              initial="hidden"
              whileInView="show"
              viewport={viewportOnce}
              variants={fadeUp}
              className="rounded-2xl border border-white/10 bg-white/5 p-5 text-sm text-neutral-400 backdrop-blur-sm"
            >
              No invitations right now.
            </motion.div>
          ) : (
            <motion.div
              initial="hidden"
              whileInView="show"
              viewport={viewportOnce}
              variants={fadeUp}
              className="rounded-2xl border border-white/10 bg-white/5 p-5 backdrop-blur-sm"
            >
              <div className="space-y-3">
                {invites.map((invite) => (
                  <div
                    key={invite.id}
                    className="flex flex-wrap items-center justify-between gap-3 rounded-xl border border-white/10 bg-white/5 px-4 py-3"
                  >
                    <div>
                      <p className="text-sm font-semibold text-white">
                        {invite.group_title || `Group ${invite.group_id}`}
                      </p>
                      <p className="text-xs text-neutral-500">
                        Invited by user #{invite.inviter_id} · {shortDate(invite.created_at)}
                      </p>
                    </div>
                    <div className="flex items-center gap-2">
                      <button
                        type="button"
                        onClick={() => updateInvite(invite.id, "accepted")}
                        className="inline-flex items-center gap-2 rounded-xl bg-white px-3 py-2 text-xs font-semibold text-[#2b2929] transition hover:bg-neutral-100"
                      >
                        Accept
                      </button>
                      <button
                        type="button"
                        onClick={() => updateInvite(invite.id, "declined")}
                        className="inline-flex items-center gap-2 rounded-xl border border-white/20 bg-white/5 px-3 py-2 text-xs font-semibold text-neutral-300 transition hover:bg-white/10 hover:text-white"
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
              className="rounded-xl border border-rose-500/30 bg-rose-500/10 px-4 py-3 text-xs text-rose-400"
            >
              {actionError}
            </motion.div>
          ) : null}
        </section>
      </main>
    </div>
  );
}
