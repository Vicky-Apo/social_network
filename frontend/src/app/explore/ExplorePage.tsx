"use client";

import { useEffect, useMemo, useState } from "react";
import Link from "next/link";
import Image from "next/image";
import { useRouter } from "next/navigation";
import { Bell, Search, ArrowLeft } from "lucide-react";
import { landingData } from "@/lib/data";
import { apiJson, asArray, asNumber, asString, isRecord } from "@/lib/api";
import { BrandMark } from "@/components/BrandMark";
import { Footer } from "@/components/Footer";

type Post = {
  id: number;
  authorName: string;
  content: string;
  media_path?: string | null;
  privacyLabel: string;
  createdAt: string;
  counts: { likes: number; comments: number; dislikes: number };
};

function initials(first?: string, last?: string) {
  const left = first?.trim().charAt(0) ?? "";
  const right = last?.trim().charAt(0) ?? "";
  return `${left}${right}`.toUpperCase() || "U";
}

function shortDate(value: string) {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return "Just now";
  return date.toLocaleDateString(undefined, { month: "short", day: "numeric" });
}

function toPost(value: unknown): Post | null {
  if (!isRecord(value)) return null;

  const id = asNumber(value.id);
  const first = asString(value.author_first_name) ?? "";
  const last = asString(value.author_last_name) ?? "";
  const content = asString(value.content) ?? "";
  const createdAt = asString(value.created_at) ?? "";
  const privacyLabel = asString(value.privacy) ?? "public";
  const media_path = asString(value.media_path);
  const likes = asNumber(value.like_count) ?? 0;
  const comments = asNumber(value.comment_count) ?? 0;
  const dislikes = asNumber(value.dislike_count) ?? 0;

  if (!id) return null;

  return {
    id,
    authorName: `${first} ${last}`.trim() || "Member",
    content,
    media_path: media_path ?? null,
    privacyLabel,
    createdAt,
    counts: { likes, comments, dislikes },
  };
}

export default function ExplorePage() {
  const router = useRouter();
  const [posts, setPosts] = useState<Post[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [isRefreshing, setIsRefreshing] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [notificationCount, setNotificationCount] = useState(0);
  const [query, setQuery] = useState("");
  const [filter, setFilter] = useState<"all" | "media">("all");

  const apiBaseUrl = useMemo(
    () =>
      process.env.NEXT_PUBLIC_API_BASE_URL?.trim().replace(/\/+$/, "") ||
      "http://localhost:8080",
    [],
  );

  const load = async (mode: "initial" | "refresh" = "initial") => {
    if (mode === "initial") {
      setIsLoading(true);
    } else {
      setIsRefreshing(true);
    }
    setError(null);

    try {
      const [postsRes, unreadRes] = await Promise.all([
        apiJson(apiBaseUrl, "/posts"),
        apiJson(apiBaseUrl, "/notifications/unread-count").catch(() => null),
      ]);

      if (postsRes.status === 401) {
        router.replace("/login");
        return;
      }

      if (!postsRes.ok || !postsRes.json?.success) {
        setError(postsRes.json?.error || "Unable to load explore feed.");
        setPosts([]);
        return;
      }

      const rawPosts = asArray(postsRes.json.data) ?? [];
      const nextPosts = rawPosts.map(toPost).filter(Boolean) as Post[];
      setPosts(nextPosts);

      if (unreadRes?.ok && unreadRes.json?.success && isRecord(unreadRes.json.data)) {
        const count = asNumber(unreadRes.json.data.count) ?? 0;
        setNotificationCount(count);
      }
    } catch {
      setError("Network error. Please try again.");
    } finally {
      setIsLoading(false);
      setIsRefreshing(false);
    }
  };

  useEffect(() => {
    void load("initial");
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [apiBaseUrl]);

  const filtered = useMemo(() => {
    let list = posts;
    if (query.trim()) {
      const q = query.trim().toLowerCase();
      list = list.filter(
        (post) =>
          post.authorName.toLowerCase().includes(q) || post.content.toLowerCase().includes(q),
      );
    }
    if (filter === "media") {
      list = list.filter((post) => Boolean(post.media_path));
    }
    return list;
  }, [posts, query, filter]);

  return (
    <div className="relative z-[1] min-h-screen bg-[#2b2929] text-white">
      <header className="sticky top-0 z-40 border-b border-white/10 bg-black/30 backdrop-blur-md">
        <div className="mx-auto flex w-full max-w-6xl items-center gap-3 px-4 py-3 sm:px-6">
          <Link
            href="/dashboard"
            className="inline-flex items-center gap-2 rounded-full border border-white/20 bg-white/5 px-3 py-2 text-xs font-semibold text-white/90 transition hover:bg-white/10"
          >
            <ArrowLeft className="h-3.5 w-3.5" />
            <span className="hidden sm:inline">Dashboard</span>
          </Link>

          <Link href="/" className="inline-flex items-center text-white focus:outline-none focus-visible:ring-2 focus-visible:ring-white/50" aria-label={landingData.productName}>
            <BrandMark label={landingData.productName} size="sm" logoSrc="/vybez-logo-v2.png" />
          </Link>

          <div className="relative ml-2 flex-1">
            <Search className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-white/50" />
            <input
              type="search"
              value={query}
              onChange={(event) => setQuery(event.target.value)}
              placeholder="Search posts..."
              className="h-11 w-full rounded-xl border border-white/30 bg-white/5 pl-9 pr-4 text-sm text-white placeholder:text-white/50 outline-none transition focus:border-white/60 focus:ring-2 focus:ring-white/30 focus:ring-offset-2 focus:ring-offset-[#2b2929]"
            />
          </div>

          <button
            type="button"
            aria-label="Notifications"
            className="relative inline-flex h-10 w-10 items-center justify-center rounded-full border border-white/20 bg-white/5 text-white/80 transition hover:bg-white/10 hover:text-white"
          >
            <Bell className="h-4 w-4" />
            <span className="absolute -right-1 -top-1 inline-flex h-4 min-w-4 items-center justify-center rounded-full bg-white px-1 text-[10px] font-semibold text-[#2b2929]">
              {notificationCount}
            </span>
          </button>
        </div>
      </header>

      <main className="relative min-h-[80vh]">
        <Image
          src="/explore-bg.png"
          alt=""
          fill
          className="object-cover object-center"
          priority
        />
        <div className="absolute inset-0 bg-[#2b2929]/55" aria-hidden />
        <div className="relative z-10 mx-auto w-full max-w-6xl px-4 py-6 sm:px-6">
          <section className="mb-8">
            <div className="flex flex-wrap items-center justify-between gap-4">
              <div>
                <h1 className="text-2xl font-bold tracking-tight text-white sm:text-3xl">Explore</h1>
                <p className="mt-1 text-sm text-white/70">Find posts and people from the community.</p>
              </div>
              <button
                type="button"
                onClick={() => void load("refresh")}
                disabled={isRefreshing}
                className="rounded-full border border-white/20 bg-white/10 px-4 py-2 text-xs font-semibold text-white transition hover:bg-white/20 disabled:opacity-60"
              >
                {isRefreshing ? "Refreshing…" : "Refresh"}
              </button>
            </div>

            <div className="mt-4 flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
              <div className="relative flex-1">
                <Search className="pointer-events-none absolute left-4 top-1/2 h-5 w-5 -translate-y-1/2 text-white/50" />
                <input
                  type="search"
                  value={query}
                  onChange={(event) => setQuery(event.target.value)}
                  placeholder="Search posts, authors, keywords…"
                  className="h-12 w-full rounded-2xl border border-white/20 bg-white/10 pl-12 pr-4 text-sm text-white placeholder:text-white/50 outline-none transition focus:border-white/40 focus:bg-white/15 focus:ring-2 focus:ring-white/20"
                />
              </div>
              <div className="flex gap-2">
                {(["all", "media"] as const).map((key) => (
                  <button
                    key={key}
                    type="button"
                    onClick={() => setFilter(key)}
                    className={`rounded-full px-4 py-2 text-xs font-semibold transition ${
                      filter === key
                        ? "bg-white text-[#2b2929]"
                        : "border border-white/20 bg-white/5 text-white/90 hover:bg-white/10"
                    }`}
                  >
                    {key === "all" ? "All posts" : "With media"}
                  </button>
                ))}
              </div>
            </div>

            {!isLoading && !error && filtered.length > 0 && (
              <p className="mt-3 text-xs text-white/50">
                {filtered.length} {filtered.length === 1 ? "post" : "posts"}
                {query.trim() ? " matching your search" : ""}
              </p>
            )}
          </section>

          <div>
            {isLoading ? (
              <div className="flex flex-col items-center justify-center rounded-2xl border border-white/10 bg-white/5 py-16 backdrop-blur-sm">
                <div className="h-8 w-8 animate-pulse rounded-full bg-white/20" />
                <p className="mt-4 text-sm text-white/60">Loading discover feed…</p>
              </div>
            ) : error ? (
              <article className="rounded-2xl border border-rose-500/30 bg-rose-500/10 p-8 text-center backdrop-blur-sm">
                <p className="text-sm font-medium text-rose-300">{error}</p>
                <button
                  type="button"
                  onClick={() => void load("refresh")}
                  className="mt-4 rounded-full bg-white/10 px-4 py-2 text-xs font-semibold text-white hover:bg-white/20"
                >
                  Try again
                </button>
              </article>
            ) : filtered.length === 0 ? (
              <div className="rounded-2xl border border-white/10 bg-white/5 py-16 text-center backdrop-blur-sm">
                <p className="text-sm font-medium text-white/80">
                  {query.trim() || filter === "media"
                    ? "No posts match. Try a different search or filter."
                    : "No posts yet. Be the first to share something."}
                </p>
                {(query.trim() || filter === "media") && (
                  <button
                    type="button"
                    onClick={() => { setQuery(""); setFilter("all"); }}
                    className="mt-4 rounded-full border border-white/20 bg-white/10 px-4 py-2 text-xs font-semibold text-white hover:bg-white/20"
                  >
                    Clear filters
                  </button>
                )}
              </div>
            ) : (
              <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
                {filtered.map((post, index) => {
                  const hasMedia = Boolean(post.media_path);
                  const isFeatured = index === 0 && hasMedia;
                  return (
                    <article
                      key={post.id}
                      className={`group overflow-hidden rounded-2xl border border-white/10 bg-[#2b2929]/50 shadow-lg backdrop-blur-sm transition hover:-translate-y-0.5 hover:border-white/20 hover:shadow-xl ${
                        isFeatured ? "sm:col-span-2 sm:row-span-1" : ""
                      }`}
                    >
                      <Link href={`/posts/${post.id}`} className="flex h-full flex-col">
                        {hasMedia && (
                          <div className={`relative overflow-hidden bg-white/5 ${isFeatured ? "aspect-video" : "aspect-[4/3]"}`}>
                            {/* eslint-disable-next-line @next/next/no-img-element */}
                            <img
                              src={post.media_path!}
                              alt=""
                              className="h-full w-full object-cover transition group-hover:scale-105"
                            />
                            <div className="absolute inset-0 bg-gradient-to-t from-black/60 via-transparent to-transparent opacity-0 transition group-hover:opacity-100" />
                          </div>
                        )}
                        <div className="flex flex-1 flex-col p-4">
                          <header className="flex items-start justify-between gap-2">
                            <div className="flex items-center gap-2">
                              <span className="inline-flex h-8 w-8 shrink-0 items-center justify-center rounded-full bg-white/20 text-xs font-semibold text-white">
                                {initials(post.authorName.split(" ")[0], post.authorName.split(" ")[1])}
                              </span>
                              <div className="min-w-0">
                                <p className="truncate text-sm font-semibold text-white">{post.authorName}</p>
                                <p className="text-[11px] text-white/50">{shortDate(post.createdAt)}</p>
                              </div>
                            </div>
                            <span className="shrink-0 rounded-full bg-white/10 px-2 py-0.5 text-[10px] font-medium uppercase text-white/70">
                              {post.privacyLabel}
                            </span>
                          </header>
                          <p className={`mt-2 text-sm leading-relaxed text-white/90 ${hasMedia ? "line-clamp-2" : "line-clamp-4"}`}>
                            {post.content}
                          </p>
                          <footer className="mt-auto flex items-center gap-4 pt-3 text-xs text-white/50">
                            <span>{post.counts.likes} likes</span>
                            <span>{post.counts.comments} comments</span>
                          </footer>
                        </div>
                      </Link>
                    </article>
                  );
                })}
              </div>
            )}
          </div>
        </div>
      </main>

      <Footer />
    </div>
  );
}

