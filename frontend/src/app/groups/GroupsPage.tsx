"use client";

import { useCallback, useEffect, useMemo, useState } from "react";
import Image from "next/image";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { ArrowRight, Compass, Globe, Lock, LogOut, MessageSquare,
  RefreshCw, Search, Users,} from "lucide-react";
import { motion } from "framer-motion";
import { useAuth } from "../component/AuthContext";
import { landingData } from "@/lib/data";
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

type GroupApiItem = {
  id?: number | string | null;
  name?: string | null;
  title?: string | null;
  description?: string | null;
  about?: string | null;
  privacy?: string | null;
  members_count?: number | null;
  member_count?: number | null;
};

type GroupItem = {
  id: number;
  name: string;
  description: string;
  privacy: "private" | "public";
  memberCount: number;
};

const quickLinks = [
  { label: "Explore", href: "/dashboard", icon: Compass },
  { label: "Groups", href: "/groups", icon: Users },
  { label: "Messages", href: "#", icon: MessageSquare },
];

const trends = [
  { title: "Design systems", posts: "276 group discussions" },
  { title: "Career growth", posts: "241 group discussions" },
  { title: "JavaScript patterns", posts: "198 group discussions" },
];

function initials(first?: string, last?: string) {
  const left = first?.trim().charAt(0) ?? "";
  const right = last?.trim().charAt(0) ?? "";
  return `${left}${right}`.toUpperCase() || "U";
}

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
  const privacyValue = String(item.privacy ?? "").toLowerCase();
  const privacy: GroupItem["privacy"] = privacyValue.includes("private") ? "private" : "public";
  const memberCount = Math.max(0, Number(item.members_count ?? item.member_count ?? 0) || 0);

  return {
    id,
    name,
    description,
    privacy,
    memberCount,
  };
}

export default function GroupsPage() {
  const router = useRouter();
  const { logout } = useAuth();

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

  const handleLogout = async () => {
    try {
      await fetch(`${apiBaseUrl}/auth/logout`, {
        method: "POST",
        credentials: "include",
      });
    } finally {
      logout();
      router.replace("/login");
    }
  };

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

  const displayName = user ? `${user.first_name} ${user.last_name}` : "Loading";
  const userTag =
    user?.nickname || (user?.email ? user.email.split("@")[0] : "community-member");

  return (
    <div className="min-h-screen bg-neutral-50 text-neutral-900">
      <header className="sticky top-0 z-40 border-b border-neutral-200/80 bg-white/85 backdrop-blur-md">
        <div className="mx-auto flex w-full max-w-6xl items-center gap-3 px-4 py-3 sm:px-6">
          <Link href="/" className="inline-flex items-center gap-2">
            <Image
              src="/vybez-logo.png"
              alt={`${landingData.productName} logo`}
              width={32}
              height={32}
              className="h-8 w-8 rounded-full border border-neutral-200 object-cover shadow-sm"
              priority
            />
            <span className="hidden text-sm font-semibold sm:inline">{landingData.productName}</span>
          </Link>

          <div className="relative ml-2 hidden flex-1 sm:block">
            <Search className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-neutral-400" />
            <input
              type="search"
              value={searchQuery}
              onChange={(event) => setSearchQuery(event.target.value)}
              placeholder="Search groups..."
              className="h-11 w-full rounded-2xl border border-neutral-200 bg-neutral-50 pl-9 pr-4 text-sm outline-none transition focus:border-neutral-400"
            />
          </div>

          <button
            type="button"
            onClick={() => void loadPageData()}
            className="inline-flex items-center gap-2 rounded-full border border-neutral-200 bg-white px-3 py-2 text-xs font-semibold text-neutral-700 transition hover:border-neutral-300 hover:text-neutral-900"
          >
            <RefreshCw className="h-3.5 w-3.5" />
            <span className="hidden sm:inline">Refresh</span>
          </button>

          <button
            type="button"
            onClick={handleLogout}
            className="inline-flex items-center gap-2 rounded-full border border-neutral-200 bg-white px-3 py-2 text-xs font-semibold text-neutral-700 transition hover:border-neutral-300 hover:text-neutral-900"
          >
            <LogOut className="h-3.5 w-3.5" />
            <span className="hidden sm:inline">Log out</span>
          </button>
        </div>
      </header>

      <main className="mx-auto grid w-full max-w-6xl gap-6 px-4 py-6 sm:px-6 lg:grid-cols-[240px_minmax(0,1fr)_280px]">
        <aside className="hidden lg:block">
          <div className="rounded-3xl border border-neutral-200 bg-white p-5 shadow-sm">
            <div className="flex items-center gap-3">
              <div className="inline-flex h-11 w-11 items-center justify-center rounded-full bg-neutral-900 text-sm font-semibold text-white">
                {initials(user?.first_name, user?.last_name)}
              </div>
              <div>
                <p className="text-sm font-semibold text-neutral-900">{displayName}</p>
                <p className="text-xs text-neutral-500">@{userTag}</p>
              </div>
            </div>
            <nav className="mt-5 space-y-2">
              {quickLinks.map((item) => {
                const Icon = item.icon;
                const isActive = item.href === "/groups";
                return (
                  <Link
                    key={item.label}
                    href={item.href}
                    className={`flex items-center gap-2 rounded-2xl border px-3 py-2 text-sm transition ${
                      isActive
                        ? "brand-gradient border-transparent text-white"
                        : "border-neutral-200 bg-neutral-50 text-neutral-700 hover:border-neutral-300 hover:text-neutral-900"
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
                const isPrivate = group.privacy === "private";
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
                      <span
                        className={`inline-flex items-center gap-1 rounded-full px-3 py-1 text-xs font-semibold ${
                          isPrivate
                            ? "bg-amber-100 text-amber-800"
                            : "bg-emerald-100 text-emerald-800"
                        }`}
                      >
                        {isPrivate ? <Lock className="h-3.5 w-3.5" /> : <Globe className="h-3.5 w-3.5" />}
                        {isPrivate ? "Private" : "Public"}
                      </span>
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
