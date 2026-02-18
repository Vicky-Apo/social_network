"use client";

import { useEffect, useMemo, useState } from "react";
import Link from "next/link";
import { useParams, useRouter } from "next/navigation";
import { ArrowLeft, ArrowRight, Calendar, RefreshCw, Shield, Users } from "lucide-react";
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
      return;
    }

    let cancelled = false;
    const load = async () => {
      setIsLoading(true);
      setError(null);

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

        const response = await fetch(`${apiBaseUrl}/groups/${groupIDNumber}`, {
          credentials: "include",
        });
        const result = (await response.json().catch(() => null)) as ApiResponse<unknown> | null;
        if (!response.ok || !result?.success) {
          if (!cancelled) {
            if (response.status === 404) {
              setError("Group endpoint is not available yet or this group does not exist.");
            } else {
              setError(result?.error || "Could not load this group.");
            }
            setGroup(null);
          }
          return;
        }

        const normalized = parseGroup(result.data);
        if (!normalized) {
          if (!cancelled) {
            setError("Received an unexpected group response format.");
            setGroup(null);
          }
          return;
        }

        if (!cancelled) {
          setGroup(normalized);
        }
      } catch {
        if (!cancelled) {
          setError("Network error while loading group details.");
          setGroup(null);
        }
      } finally {
        if (!cancelled) {
          setIsLoading(false);
        }
      }
    };

    void load();
    return () => {
      cancelled = true;
    };
  }, [apiBaseUrl, groupIDNumber, router]);

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
              className="inline-flex items-center gap-2 rounded-full border border-neutral-200 bg-white px-3 py-1.5 text-xs font-semibold text-neutral-700 transition hover:border-neutral-300 hover:text-neutral-900"
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
              className="inline-flex items-center gap-2 rounded-full border border-neutral-200 bg-white px-4 py-2 text-sm font-semibold text-neutral-700 transition hover:border-neutral-300 hover:text-neutral-900"
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
      </main>
    </div>
  );
}
