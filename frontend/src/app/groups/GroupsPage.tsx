"use client";

import { useCallback, useEffect, useMemo, useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { ArrowRight, Users } from "lucide-react";
import { motion } from "framer-motion";
import TopNav from "@/components/TopNav";
import LeftNav from "@/components/LeftNav";
import { fadeUp, viewportOnce } from "@/components/Motion";
import { apiFetchJson, getApiBaseUrl } from "@/lib/api";
import { ApiResponse } from "@/lib/types";

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

  const apiBaseUrl = useMemo(() => getApiBaseUrl(), []);

  const loadPageData = useCallback(async () => {
    setIsLoading(true);
    setGroupsError(null);

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
      setUser(meResult.data);

      const { response: groupsResponse, result: groupsResult } = await apiFetchJson<
        ApiResponse<unknown>
      >("/groups", {}, apiBaseUrl);
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
    <div
      className="min-h-screen text-neutral-100"
      style={{
        backgroundImage: "url('/groups-bg.png')",
        backgroundSize: "cover",
        backgroundPosition: "center",
        backgroundAttachment: "fixed",
      }}
    >
      <TopNav
        user={user ?? undefined}
        searchValue={searchQuery}
        onSearchChange={setSearchQuery}
        searchPlaceholder="Search groups..."
        onLogout={() => router.replace("/login")}
        variant="dark"
      />

      <main className="mx-auto grid w-full max-w-6xl gap-6 px-4 py-6 sm:px-6 lg:grid-cols-[220px_minmax(0,1fr)_240px]">
        <aside className="hidden lg:block">
          <LeftNav user={user ?? undefined} activeHref="/groups" variant="dark" />
        </aside>

        <section className="space-y-5">
          <motion.div
            initial="hidden"
            whileInView="show"
            viewport={viewportOnce}
            variants={fadeUp}
            className="rounded-2xl border border-white/10 bg-white/5 p-4 backdrop-blur-sm sm:p-5"
          >
            <div className="flex flex-wrap items-start justify-between gap-4">
              <div>
                <h1 className="text-xl font-semibold tracking-tight text-white">Groups</h1>
                <p className="text-sm text-neutral-400">
                  Browse all available groups and enter the one you want to join.
                </p>
              </div>
              <div className="flex flex-wrap items-center gap-2">
                <Link
                  href="/groups/create"
                  className="inline-flex items-center gap-2 rounded-xl border border-white/20 bg-white/5 px-3 py-2 text-xs font-semibold text-neutral-300 transition hover:bg-white/10 hover:text-white"
                >
                  Create group
                </Link>
                <Link
                  href="/group-invitations"
                  className="inline-flex items-center gap-2 rounded-xl border border-white/20 bg-white/5 px-3 py-2 text-xs font-semibold text-neutral-300 transition hover:bg-white/10 hover:text-white"
                >
                  Invitations
                </Link>
                <span className="rounded-xl border border-white/10 bg-white/5 px-3 py-1 text-xs text-neutral-400">
                  Total groups: {groups.length}
                </span>
                <span className="rounded-xl border border-white/10 bg-white/5 px-3 py-1 text-xs text-neutral-400">
                  Members tracked: {totalMembers}
                </span>
              </div>
            </div>
          </motion.div>

          {isLoading ? (
            <article className="rounded-2xl border border-white/10 bg-white/5 p-6 text-sm text-neutral-400 backdrop-blur-sm">
              Loading groups...
            </article>
          ) : groupsError ? (
            <article className="rounded-2xl border border-rose-500/30 bg-rose-500/10 p-6 text-sm text-rose-400 backdrop-blur-sm">
              {groupsError}
            </article>
          ) : filteredGroups.length === 0 ? (
            <article className="rounded-2xl border border-white/10 bg-white/5 p-6 text-sm text-neutral-400 backdrop-blur-sm">
              No groups found for this search.
            </article>
          ) : (
            <div className="space-y-4">
              {filteredGroups.map((group) => {
                return (
                  <article
                    key={group.id}
                    className="rounded-2xl border border-white/10 bg-white/5 p-5 backdrop-blur-sm"
                  >
                    <div className="flex flex-wrap items-start justify-between gap-3">
                      <div>
                        <h2 className="text-lg font-semibold tracking-tight text-white">
                          {group.name}
                        </h2>
                        <p className="mt-1 text-sm leading-relaxed text-neutral-400">
                          {group.description}
                        </p>
                      </div>
                    </div>

                    <div className="mt-4 flex flex-wrap items-center justify-between gap-3">
                      <span className="inline-flex items-center gap-2 rounded-xl border border-white/10 bg-white/5 px-3 py-1 text-xs text-neutral-400">
                        <Users className="h-3.5 w-3.5" />
                        {group.memberCount} members
                      </span>
                      <Link
                        href={`/groups/${group.id}`}
                        className="group inline-flex items-center gap-2 rounded-xl bg-white px-4 py-2 text-xs font-semibold transition hover:bg-neutral-100"
                        style={{ color: "#000" }}
                      >
                        <span style={{ color: "#000" }}>Enter group</span>
                        <ArrowRight className="h-3.5 w-3.5 transition-transform duration-200 group-hover:translate-x-0.5" style={{ color: "#000" }} />
                      </Link>
                    </div>
                  </article>
                );
              })}
            </div>
          )}
        </section>

        <aside className="hidden space-y-4 md:block">
          <div className="rounded-2xl border border-white/10 bg-white/5 p-4 backdrop-blur-sm">
            <h2 className="text-sm font-semibold text-white">Groups overview</h2>
            <div className="mt-4 grid grid-cols-2 gap-3">
              <div className="rounded-xl border border-white/10 bg-white/5 p-3">
                <p className="text-lg font-semibold text-white">{groups.length}</p>
                <p className="text-xs text-neutral-500">Total groups</p>
              </div>
              <div className="rounded-xl border border-white/10 bg-white/5 p-3">
                <p className="text-lg font-semibold text-white">{totalMembers}</p>
                <p className="text-xs text-neutral-500">Member slots</p>
              </div>
            </div>
          </div>

        </aside>
      </main>
    </div>
  );
}
