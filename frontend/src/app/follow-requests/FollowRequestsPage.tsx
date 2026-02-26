"use client";

import { useCallback, useEffect, useMemo, useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import {
  Compass,
  MessageSquare,
  UserCheck,
  UserMinus,
  UserPlus,
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

type User = {
  id: number;
  email: string;
  first_name: string;
  last_name: string;
  nickname?: string | null;
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

const quickLinks = [
  { label: "Explore", href: "/dashboard", icon: Compass },
  { label: "Groups", href: "/groups", icon: Users },
  { label: "Messages", href: "/messages", icon: MessageSquare },
  { label: "Requests", href: "/follow-requests", icon: UserPlus },
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

export default function FollowRequestsPage() {
  const router = useRouter();
  const [viewer, setViewer] = useState<User | null>(null);
  const [incoming, setIncoming] = useState<FollowRequest[]>([]);
  const [sent, setSent] = useState<FollowRequest[]>([]);
  const [usersByID, setUsersByID] = useState<Record<number, UserListItem>>({});
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

      const [incomingRes, sentRes, usersRes] = await Promise.all([
        fetch(`${apiBaseUrl}/follow-requests`, { credentials: "include" }),
        fetch(`${apiBaseUrl}/follow-requests/sent`, { credentials: "include" }),
        fetch(`${apiBaseUrl}/users?limit=200&offset=0`, { credentials: "include" }),
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

  const resolveUser = (id: number) => usersByID[id];
  const viewerTag =
    viewer?.nickname || (viewer?.email ? viewer.email.split("@")[0] : "member");

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
                <p className="text-xs text-neutral-500">@{viewerTag}</p>
              </div>
            </div>
            <nav className="mt-5 space-y-2">
              {quickLinks.map((item) => {
                const Icon = item.icon;
                const isActive = item.href === "/follow-requests";
                return (
                  <Link
                    key={item.label}
                    href={item.href}
                    className={`flex items-center gap-2 rounded-2xl border px-3 py-2 text-sm transition ${
                      isActive
                        ? "brand-gradient border-transparent text-white"
                        : "border-neutral-200 bg-neutral-50 text-neutral-700 hover:border-neutral-400 hover:text-neutral-900"
                    }`}
                  >
                    <Icon className="h-4 w-4" />
                    <span>{item.label}</span>
                  </Link>
                );
              })}
            </nav>
          </div>
        </aside>

        <section className="space-y-6">
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
                  Follow requests
                </h1>
                <p className="text-sm text-neutral-600">
                  Manage incoming and sent follow requests.
                </p>
              </div>
              <span className="rounded-full border border-neutral-200 bg-neutral-50 px-3 py-1 text-xs text-neutral-600">
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
              className="rounded-3xl border border-neutral-200 bg-white p-5 text-sm text-neutral-600 shadow-sm"
            >
              Loading requests...
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
          ) : (
            <>
              <motion.div
                initial="hidden"
                whileInView="show"
                viewport={viewportOnce}
                variants={fadeUp}
                className="rounded-3xl border border-neutral-200 bg-white p-5 shadow-sm"
              >
                <h2 className="text-sm font-semibold text-neutral-900">Incoming requests</h2>
                {incoming.length === 0 ? (
                  <p className="mt-3 text-sm text-neutral-600">
                    No incoming requests right now.
                  </p>
                ) : (
                  <div className="mt-4 space-y-3">
                    {incoming.map((req) => {
                      const requester = resolveUser(req.requester_id);
                      return (
                        <div
                          key={req.id}
                          className="flex flex-wrap items-center justify-between gap-3 rounded-2xl border border-neutral-200 bg-neutral-50 px-4 py-3"
                        >
                          <div>
                            <p className="text-sm font-semibold text-neutral-800">
                              {requester
                                ? `${requester.first_name} ${requester.last_name}`
                                : `User #${req.requester_id}`}
                            </p>
                            <p className="text-xs text-neutral-500">
                              @{requester?.nickname || `user-${req.requester_id}`} ·{" "}
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
                              className="inline-flex items-center gap-2 rounded-full border border-neutral-300 bg-white px-3 py-2 text-xs font-semibold text-neutral-700 transition hover:border-neutral-400"
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
                className="rounded-3xl border border-neutral-200 bg-white p-5 shadow-sm"
              >
                <h2 className="text-sm font-semibold text-neutral-900">Sent requests</h2>
                {sent.length === 0 ? (
                  <p className="mt-3 text-sm text-neutral-600">No sent requests.</p>
                ) : (
                  <div className="mt-4 space-y-3">
                    {sent.map((req) => {
                      const target = resolveUser(req.target_id);
                      return (
                        <div
                          key={req.id}
                          className="flex flex-wrap items-center justify-between gap-3 rounded-2xl border border-neutral-200 bg-neutral-50 px-4 py-3"
                        >
                          <div>
                            <p className="text-sm font-semibold text-neutral-800">
                              {target
                                ? `${target.first_name} ${target.last_name}`
                                : `User #${req.target_id}`}
                            </p>
                            <p className="text-xs text-neutral-500">
                              @{target?.nickname || `user-${req.target_id}`} ·{" "}
                              {shortDate(req.created_at)}
                            </p>
                          </div>
                          <button
                            type="button"
                            onClick={() => updateRequest(req.id, "canceled")}
                            className="inline-flex items-center gap-2 rounded-full border border-neutral-300 bg-white px-3 py-2 text-xs font-semibold text-neutral-700 transition hover:border-neutral-400"
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
              className="rounded-2xl border border-rose-200 bg-rose-50 px-4 py-3 text-xs text-rose-700"
            >
              {actionError}
            </motion.div>
          ) : null}
        </section>

        <aside className="hidden lg:block">
          <div className="rounded-3xl border border-neutral-200 bg-white p-5 shadow-sm">
            <h3 className="text-sm font-semibold text-neutral-900">Tips</h3>
            <p className="mt-2 text-xs text-neutral-500">
              Accept requests to allow private profiles to connect. Declined requests
              can be re-sent by the other user later.
            </p>
          </div>
        </aside>
      </main>
    </div>
  );
}
