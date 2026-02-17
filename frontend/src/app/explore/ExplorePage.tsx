"use client";

import { useEffect, useMemo, useState } from "react";
import Image from "next/image";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { Bell, Search, ArrowLeft } from "lucide-react";
import { landingData } from "@/lib/data";
import { apiJson, asArray, asNumber, asString, isRecord } from "@/lib/api";

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

  const filtered = query.trim()
    ? posts.filter((post) => {
        const q = query.trim().toLowerCase();
        return post.authorName.toLowerCase().includes(q) || post.content.toLowerCase().includes(q);
      })
    : posts;

  return (
    <div className="min-h-screen bg-neutral-50 text-neutral-900">
      <header className="sticky top-0 z-40 border-b border-neutral-200/80 bg-white/85 backdrop-blur-md">
        <div className="mx-auto flex w-full max-w-6xl items-center gap-3 px-4 py-3 sm:px-6">
          <Link href="/dashboard" className="inline-flex items-center gap-2 text-sm font-semibold text-neutral-700">
            <ArrowLeft className="h-4 w-4" />
            <span className="hidden sm:inline">Dashboard</span>
          </Link>

          <div className="flex items-center gap-2">
            <Image
              src="/vybez-logo.png"
              alt={`${landingData.productName} logo`}
              width={32}
              height={32}
              className="h-8 w-8 rounded-full border border-neutral-200 object-cover shadow-sm"
              priority
            />
            <span className="hidden text-sm font-semibold sm:inline">{landingData.productName}</span>
          </div>

          <div className="relative ml-2 flex-1">
            <Search className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-neutral-400" />
            <input
              type="search"
              value={query}
              onChange={(event) => setQuery(event.target.value)}
              placeholder="Search explore feed..."
              className="h-11 w-full rounded-2xl border border-neutral-200 bg-neutral-50 pl-9 pr-4 text-sm outline-none transition focus:border-neutral-400"
            />
          </div>

          <button
            type="button"
            aria-label="Notifications"
            className="relative inline-flex h-10 w-10 items-center justify-center rounded-full border border-neutral-200 bg-white text-neutral-600 transition hover:text-neutral-900"
          >
            <Bell className="h-4 w-4" />
            <span className="absolute -right-1 -top-1 inline-flex h-4 min-w-4 items-center justify-center rounded-full bg-neutral-900 px-1 text-[10px] font-semibold text-white">
              {notificationCount}
            </span>
          </button>
        </div>
      </header>

      <main className="mx-auto w-full max-w-6xl px-4 py-6 sm:px-6">
        <div className="rounded-3xl border border-neutral-200 bg-white p-4 shadow-sm sm:p-5">
          <div className="flex flex-wrap items-center justify-between gap-3">
            <div>
              <h1 className="text-xl font-semibold tracking-tight text-neutral-900">Explore</h1>
              <p className="text-sm text-neutral-600">Browse the latest posts from the community.</p>
            </div>
            <button
              type="button"
              onClick={() => void load("refresh")}
              className="rounded-full border border-neutral-200 bg-white px-3 py-1.5 text-xs font-semibold text-neutral-700 transition hover:border-neutral-300 hover:text-neutral-900"
            >
              {isRefreshing ? "Refreshing..." : "Refresh"}
            </button>
          </div>
        </div>

        <div className="mt-5">
          {isLoading ? (
            <article className="rounded-3xl border border-neutral-200 bg-white p-6 text-sm text-neutral-600 shadow-sm">
              Loading explore feed...
            </article>
          ) : error ? (
            <article className="rounded-3xl border border-rose-200 bg-rose-50 p-6 text-sm text-rose-700">
              {error}
            </article>
          ) : filtered.length === 0 ? (
            <article className="rounded-3xl border border-neutral-200 bg-white p-6 text-sm text-neutral-600 shadow-sm">
              No posts to show yet.
            </article>
          ) : (
            <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
              {filtered.map((post) => (
                <article
                  key={post.id}
                  className="group overflow-hidden rounded-3xl border border-neutral-200 bg-white shadow-sm transition hover:-translate-y-0.5 hover:shadow-md"
                >
                  <Link href={`/posts/${post.id}`} className="block">
                    <div className="p-4">
                      <header className="flex items-start justify-between gap-3">
                        <div className="flex items-center gap-3">
                          <span className="inline-flex h-10 w-10 items-center justify-center rounded-full bg-neutral-900 text-xs font-semibold text-white">
                            {initials(post.authorName.split(" ")[0], post.authorName.split(" ")[1])}
                          </span>
                          <div>
                            <p className="text-sm font-semibold text-neutral-900">
                              {post.authorName}
                            </p>
                            <p className="text-xs text-neutral-500">{shortDate(post.createdAt)}</p>
                          </div>
                        </div>
                        <span className="rounded-full border border-neutral-200 bg-neutral-50 px-2.5 py-1 text-[11px] uppercase tracking-wide text-neutral-600">
                          {post.privacyLabel}
                        </span>
                      </header>

                      <p className="mt-3 line-clamp-5 text-sm leading-relaxed text-neutral-700">
                        {post.content}
                      </p>

                      {post.media_path ? (
                        <div className="mt-3 overflow-hidden rounded-2xl border border-neutral-200">
                          {/* eslint-disable-next-line @next/next/no-img-element */}
                          <img
                            src={post.media_path}
                            alt="Post media"
                            className="h-36 w-full object-cover"
                          />
                        </div>
                      ) : null}
                    </div>

                    <footer className="flex items-center justify-between border-t border-neutral-200 bg-neutral-50 px-4 py-3 text-xs text-neutral-600">
                      <span>{post.counts.likes} likes</span>
                      <span>{post.counts.comments} comments</span>
                      <span>{post.counts.dislikes} dislikes</span>
                    </footer>
                  </Link>
                </article>
              ))}
            </div>
          )}
        </div>
      </main>
    </div>
  );
}

