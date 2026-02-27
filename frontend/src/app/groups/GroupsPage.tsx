"use client";

import { useCallback, useEffect, useMemo, useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { ArrowRight, Users } from "lucide-react";
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

type GroupApiItem = {
  id?: number | string | null;
  name?: string | null;
  title?: string | null;
  description?: string | null;
  about?: string | null;
  members_count?: number | null;
  member_count?: number | null;
};

type GroupItem = {
  id: number;
  name: string;
  description: string;
  memberCount: number;
};

const trends = [
  { title: "Design systems", posts: "276 group discussions" },
  { title: "Career growth", posts: "241 group discussions" },
  { title: "JavaScript patterns", posts: "198 group discussions" },
];

function extractRawGroups(data: unknown): GroupApiItem[] {
  if (Array.isArray(data)) {
    return data as GroupApiItem[];
  }
  if (data && typeof data === "object") {
    const groups = (data as { groups?: unknown }).groups;
    if (Array.isArray(groups)) {
      return groups as GroupApiItem[];
    }
  }
  return [];
}

function normalizeGroup(item: GroupApiItem): GroupItem | null {
  const id = Number(item.id);
  if (!Number.isFinite(id) || id <= 0) {
    return null;
  }

  const name = (item.name ?? item.title ?? "").trim() || `Group ${id}`;
  const description = (item.description ?? item.about ?? "").trim() || "No description yet.";
  const memberCount = Math.max(0, Number(item.members_count ?? item.member_count ?? 0) || 0);

  return {
    id,
    name,
    description,
    memberCount,
  };
}

export default function GroupsPage() {
  const router = useRouter();
  const [user, setUser] = useState<User | null>(null);
  const [groups, setGroups] = useState<GroupItem[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [groupsError, setGroupsError] = useState<string | null>(null);
  const [searchQuery, setSearchQuery] = useState("");

  const apiBaseUrl = useMemo(
    () =>
      process.env.NEXT_PUBLIC_API_BASE_URL?.trim().replace(/\/+$/, "") ||
      "http://localhost:8080",
    [],
  );

  const loadPageData = useCallback(async () => {
    setIsLoading(true);
    setGroupsError(null);

    try {
      const meResponse = await fetch(`${apiBaseUrl}/auth/me`, {
        credentials: "include",
      });
      const meResult = (await meResponse.json().catch(() => null)) as ApiResponse<User> | null;
      if (!meResponse.ok || !meResult?.success || !meResult.data) {
        router.replace("/login");
        return;
      }
      setUser(meResult.data);

      const groupsResponse = await fetch(`${apiBaseUrl}/groups`, {
        credentials: "include",
      });
      const groupsResult = (await groupsResponse.json().catch(() => null)) as
        | ApiResponse<unknown>
        | null;
      if (!groupsResponse.ok || !groupsResult?.success) {
        if (groupsResponse.status === 404) {
          setGroupsError("Groups endpoint is not available yet on the backend.");
        } else {
          setGroupsError(groupsResult?.error || "Unable to load groups.");
        }
        setGroups([]);
        return;
      }

      const normalized = extractRawGroups(groupsResult.data)
        .map((item) => normalizeGroup(item))
        .filter((item): item is GroupItem => Boolean(item));
      setGroups(normalized);
    } catch {
      setGroupsError("Network error. Please try again.");
      setGroups([]);
    } finally {
      setIsLoading(false);
    }
  }, [apiBaseUrl, router]);

  useEffect(() => {
    void loadPageData();
  }, [loadPageData]);

  const filteredGroups = useMemo(() => {
    const query = searchQuery.trim().toLowerCase();
    if (!query) return groups;
    return groups.filter((group) =>
      `${group.name} ${group.description}`.toLowerCase().includes(query),
    );
  }, [groups, searchQuery]);

  const totalMembers = useMemo(
    () => groups.reduce((total, group) => total + group.memberCount, 0),
    [groups],
  );

  return (
    <div className="min-h-screen bg-neutral-50 text-neutral-900">
      <TopNav
        user={user ?? undefined}
        searchValue={searchQuery}
        onSearchChange={setSearchQuery}
        searchPlaceholder="Search groups..."
        onLogout={() => router.replace("/login")}
      />

      <main className="mx-auto grid w-full max-w-6xl gap-6 px-4 py-6 sm:px-6 lg:grid-cols-[240px_minmax(0,1fr)_280px]">
        <aside className="hidden lg:block">
          <LeftNav user={user ?? undefined} activeHref="/groups" />
        </aside>

        <section className="space-y-5">
          <motion.div
            initial="hidden"
            whileInView="show"
            viewport={viewportOnce}
            variants={fadeUp}
            className="rounded-3xl border border-neutral-200 bg-white p-4 shadow-sm sm:p-5"
          >
            <div className="flex flex-wrap items-start justify-between gap-4">
              <div>
                <h1 className="text-xl font-semibold tracking-tight text-neutral-900">Groups</h1>
                <p className="text-sm text-neutral-600">
                  Browse all available groups and enter the one you want to join.
                </p>
              </div>
              <div className="flex flex-wrap items-center gap-2">
                <Link
                  href="/groups/create"
                  className="inline-flex items-center gap-2 rounded-full border border-neutral-200 bg-white px-3 py-2 text-xs font-semibold text-neutral-700 transition hover:border-neutral-400 hover:text-neutral-900"
                >
                  Create group
                </Link>
                <Link
                  href="/group-invitations"
                  className="inline-flex items-center gap-2 rounded-full border border-neutral-200 bg-white px-3 py-2 text-xs font-semibold text-neutral-700 transition hover:border-neutral-400 hover:text-neutral-900"
                >
                  Invitations
                </Link>
                <span className="rounded-full border border-neutral-200 bg-neutral-50 px-3 py-1 text-xs text-neutral-600">
                  Total groups: {groups.length}
                </span>
                <span className="rounded-full border border-neutral-200 bg-neutral-50 px-3 py-1 text-xs text-neutral-600">
                  Members tracked: {totalMembers}
                </span>
              </div>
            </div>
          </motion.div>

          {isLoading ? (
            <article className="rounded-3xl border border-neutral-200 bg-white p-6 text-sm text-neutral-600 shadow-sm">
              Loading groups...
            </article>
          ) : groupsError ? (
            <article className="rounded-3xl border border-rose-200 bg-rose-50 p-6 text-sm text-rose-700">
              {groupsError}
            </article>
          ) : filteredGroups.length === 0 ? (
            <article className="rounded-3xl border border-neutral-200 bg-white p-6 text-sm text-neutral-600 shadow-sm">
              No groups found for this search.
            </article>
          ) : (
            <div className="space-y-4">
              {filteredGroups.map((group) => {
                return (
                  <article
                    key={group.id}
                    className="rounded-3xl border border-neutral-200 bg-white p-5 shadow-sm"
                  >
                    <div className="flex flex-wrap items-start justify-between gap-3">
                      <div>
                        <h2 className="text-lg font-semibold tracking-tight text-neutral-900">
                          {group.name}
                        </h2>
                        <p className="mt-1 text-sm leading-relaxed text-neutral-600">
                          {group.description}
                        </p>
                      </div>
                    </div>

                    <div className="mt-4 flex flex-wrap items-center justify-between gap-3">
                      <span className="inline-flex items-center gap-2 rounded-full border border-neutral-200 bg-neutral-50 px-3 py-1 text-xs text-neutral-600">
                        <Users className="h-3.5 w-3.5" />
                        {group.memberCount} members
                      </span>
                      <Link
                        href={`/groups/${group.id}`}
                        className="brand-gradient group inline-flex items-center gap-2 rounded-full px-4 py-2 text-xs font-semibold text-white transition hover:-translate-y-0.5 hover:shadow-md"
                      >
                        <span>Enter group</span>
                        <ArrowRight className="h-3.5 w-3.5 transition-transform duration-200 group-hover:translate-x-0.5" />
                      </Link>
                    </div>
                  </article>
                );
              })}
            </div>
          )}
        </section>

        <aside className="hidden space-y-5 md:block">
          <div className="rounded-3xl border border-neutral-200 bg-white p-5 shadow-sm">
            <h2 className="text-sm font-semibold text-neutral-900">Groups overview</h2>
            <div className="mt-4 grid grid-cols-2 gap-3">
              <div className="rounded-2xl border border-neutral-200 bg-neutral-50 p-3">
                <p className="text-lg font-semibold text-neutral-900">{groups.length}</p>
                <p className="text-xs text-neutral-500">Total groups</p>
              </div>
              <div className="rounded-2xl border border-neutral-200 bg-neutral-50 p-3">
                <p className="text-lg font-semibold text-neutral-900">{totalMembers}</p>
                <p className="text-xs text-neutral-500">Member slots</p>
              </div>
            </div>
          </div>

          <div className="rounded-3xl border border-neutral-200 bg-white p-5 shadow-sm">
            <h2 className="text-sm font-semibold text-neutral-900">Trending in groups</h2>
            <div className="mt-4 space-y-3">
              {trends.map((item) => (
                <article key={item.title} className="rounded-2xl border border-neutral-200 bg-neutral-50 p-3">
                  <p className="text-sm font-semibold text-neutral-900">{item.title}</p>
                  <p className="mt-1 text-xs text-neutral-600">{item.posts}</p>
                </article>
              ))}
            </div>
          </div>
        </aside>
      </main>
    </div>
  );
}
