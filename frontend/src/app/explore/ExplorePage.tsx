"use client";

import { useEffect, useMemo, useState } from "react";
import Image from "next/image";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { Bell, Search, ArrowLeft } from "lucide-react";
import { motion } from "framer-motion";
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

type Post = {
  id: number;
  author_id: number;
  author_first_name: string;
  author_last_name: string;
  content: string;
  media_path?: string | null;
  privacy: string;
  created_at: string;
  comment_count: number;
  like_count: number;
  dislike_count: number;
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

export default function ExplorePage() {
  const router = useRouter();
  const [user, setUser] = useState<User | null>(null);
  const [posts, setPosts] = useState<Post[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [notificationCount, setNotificationCount] = useState(0);
  const [query, setQuery] = useState("");

  const apiBaseUrl = useMemo(
    () =>
      process.env.NEXT_PUBLIC_API_BASE_URL?.trim().replace(/\/+$/, "") ||
      "http://localhost:8080",
    [],
  );

  const load = async () => {
    setIsLoading(true);
    setError(null);

    try {
      const meRes = await fetch(`${apiBaseUrl}/auth/me`, { credentials: "include" });
      const meJson = (await meRes.json().catch(() => null)) as ApiResponse<User> | null;
      if (!meRes.ok || !meJson?.success) {
        router.replace("/login");
        return;
      }
      setUser(meJson.data ?? null);

      const unreadRes = await fetch(`${apiBaseUrl}/notifications/unread-count`, {
        credentials: "include",
      }).catch(() => null);
      if (unreadRes?.ok) {
        const unreadJson = (await unreadRes.json().catch(() => null)) as
          | ApiResponse<{ count: number }>
          | null;
        if (unreadJson?.success) {
          setNotificationCount(Number(unreadJson.data?.count ?? 0));
        }
      }

      const postsRes = await fetch(`${apiBaseUrl}/posts`, { credentials: "include" });
      const postsJson = (await postsRes.json().catch(() => null)) as ApiResponse<Post[]> | null;
      if (!postsRes.ok || !postsJson?.success) {
        setError(postsJson?.error || "Unable to load explore feed.");
        setPosts([]);
        return;
      }
      setPosts(postsJson.data ?? []);
    } catch {
      setError("Network error. Please try again.");
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    void load();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [apiBaseUrl]);

  useEffect(() => {
    const id = window.setInterval(() => void load(), 8000);
    return () => window.clearInterval(id);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [apiBaseUrl]);

  const filtered = query.trim()
    ? posts.filter((post) => {
        const q = query.trim().toLowerCase();
        const author = `${post.author_first_name} ${post.author_last_name}`.toLowerCase();
        return author.includes(q) || post.content.toLowerCase().includes(q);
      })
    : posts;

  const displayName = user ? `${user.first_name} ${user.last_name}` : "Explore";

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
        <motion.div
          initial="hidden"
          whileInView="show"
          viewport={viewportOnce}
          variants={fadeUp}
          className="rounded-3xl border border-neutral-200 bg-white p-4 shadow-sm sm:p-5"
        >
          <div className="flex flex-wrap items-center justify-between gap-3">
            <div>
              <h1 className="text-xl font-semibold tracking-tight text-neutral-900">Explore</h1>
              <p className="text-sm text-neutral-600">
                Browse the latest posts from the community. You’re signed in as {displayName}.
              </p>
            </div>
            <button
              type="button"
              onClick={() => void load()}
              className="rounded-full border border-neutral-200 bg-white px-3 py-1.5 text-xs font-semibold text-neutral-700 transition hover:border-neutral-300 hover:text-neutral-900"
            >
              Refresh
            </button>
          </div>
        </motion.div>

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
                <motion.article
                  key={post.id}
                  initial="hidden"
                  whileInView="show"
                  viewport={viewportOnce}
                  variants={fadeUp}
                  className="group overflow-hidden rounded-3xl border border-neutral-200 bg-white shadow-sm transition hover:-translate-y-0.5 hover:shadow-md"
                >
                  <div className="p-4">
                    <header className="flex items-start justify-between gap-3">
                      <div className="flex items-center gap-3">
                        <span className="inline-flex h-10 w-10 items-center justify-center rounded-full bg-neutral-900 text-xs font-semibold text-white">
                          {initials(post.author_first_name, post.author_last_name)}
                        </span>
                        <div>
                          <p className="text-sm font-semibold text-neutral-900">
                            {post.author_first_name} {post.author_last_name}
                          </p>
                          <p className="text-xs text-neutral-500">{shortDate(post.created_at)}</p>
                        </div>
                      </div>
                      <span className="rounded-full border border-neutral-200 bg-neutral-50 px-2.5 py-1 text-[11px] uppercase tracking-wide text-neutral-600">
                        {post.privacy}
                      </span>
                    </header>

                    <p className="mt-3 line-clamp-5 text-sm leading-relaxed text-neutral-700">
                      {post.content}
                    </p>

                    {post.media_path ? (
                      <div className="mt-3 overflow-hidden rounded-2xl border border-neutral-200">
                        {/* eslint-disable-next-line @next/next/no-img-element */}
                        <img src={post.media_path} alt="Post media" className="h-36 w-full object-cover" />
                      </div>
                    ) : null}
                  </div>

                  <footer className="flex items-center justify-between border-t border-neutral-200 bg-neutral-50 px-4 py-3 text-xs text-neutral-600">
                    <span>{post.like_count} likes</span>
                    <span>{post.comment_count} comments</span>
                    <span>{post.dislike_count} dislikes</span>
                  </footer>
                </motion.article>
              ))}
            </div>
          )}
        </div>
      </main>
    </div>
  );
}

