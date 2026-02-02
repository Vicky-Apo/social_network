"use client";

import { useEffect, useMemo, useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { useAuth } from "../component/AuthContext";

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
  about?: string | null;
  is_public?: boolean;
};

type Post = {
  id: number;
  author_id: number;
  author_first_name: string;
  author_last_name: string;
  author_nickname?: string | null;
  author_avatar_path?: string | null;
  content: string;
  media_path?: string | null;
  privacy: string;
  created_at: string;
  comment_count: number;
  like_count: number;
  dislike_count: number;
};

export default function HomePage() {
  const router = useRouter();
  const { logout } = useAuth();
  const [user, setUser] = useState<User | null>(null);
  const [posts, setPosts] = useState<Post[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const apiBaseUrl = useMemo(
    () =>
      process.env.NEXT_PUBLIC_API_BASE_URL?.trim().replace(/\/+$/, "") ||
      "http://localhost:8080",
    []
  );

  useEffect(() => {
    let cancelled = false;

    const fetchJson = async <T,>(path: string) => {
      const response = await fetch(`${apiBaseUrl}${path}`, {
        credentials: "include",
      });
      const result = (await response.json().catch(() => null)) as ApiResponse<T> | null;
      return { response, result };
    };

    const load = async () => {
      setLoading(true);
      setError(null);

      try {
        const me = await fetchJson<User>("/auth/me");
        if (!me.response.ok || !me.result?.success) {
          if (!cancelled) {
            router.replace("/login");
          }
          return;
        }

        if (!cancelled) {
          setUser(me.result.data ?? null);
        }

        const feed = await fetchJson<Post[]>("/posts");
        if (!feed.response.ok || !feed.result?.success) {
          if (!cancelled) {
            setError(feed.result?.error || "Failed to load posts.");
            setPosts([]);
          }
          return;
        }

        if (!cancelled) {
          setPosts(feed.result.data ?? []);
        }
      } catch {
        if (!cancelled) {
          setError("Network error. Please try again.");
        }
      } finally {
        if (!cancelled) {
          setLoading(false);
        }
      }
    };

    load();

    return () => {
      cancelled = true;
    };
  }, [apiBaseUrl, router]);

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

  return (
    <div className="min-h-screen bg-[radial-gradient(circle_at_top,_#f5f3ff,_#ffffff_45%)] text-slate-900">
      <header className="border-b border-slate-200/70 bg-white/70 backdrop-blur">
        <div className="mx-auto flex max-w-6xl items-center justify-between px-5 py-5">
          <Link href="/home" className="text-xl font-semibold tracking-tight">
            ConNextioN
          </Link>
          <div className="flex items-center gap-6 text-sm">
            <Link
              href="/profile"
              className="text-slate-600 transition hover:text-slate-900"
            >
              Profile
            </Link>
            <button
              type="button"
              className="rounded-full border border-slate-200 px-4 py-2 text-slate-700 transition hover:border-slate-300 hover:text-slate-900"
              onClick={handleLogout}
            >
              Log out
            </button>
          </div>
        </div>
      </header>

      <main className="mx-auto grid max-w-6xl gap-8 px-5 py-10 lg:grid-cols-[220px_1fr_260px]">
        <aside className="hidden lg:flex lg:flex-col lg:gap-4">
          <div className="rounded-3xl border border-slate-200 bg-white p-5 shadow-sm">
            <p className="text-xs uppercase tracking-[0.2em] text-slate-400">
              Navigation
            </p>
            <div className="mt-4 flex flex-col gap-3 text-sm">
              <Link href="/home" className="text-slate-900">
                Feed
              </Link>
              <Link href="/groups" className="text-slate-500">
                Groups
              </Link>
              <Link href="/messages" className="text-slate-500">
                Chats
              </Link>
              <Link href="/notifications" className="text-slate-500">
                Notifications
              </Link>
            </div>
          </div>
        </aside>

        <section className="space-y-6">
          <div className="rounded-3xl border border-slate-200 bg-white p-6 shadow-sm">
            <p className="text-sm text-slate-500">Welcome back</p>
            <h1 className="mt-2 text-3xl font-semibold tracking-tight">
              {user
                ? `${user.first_name} ${user.last_name}`
                : loading
                ? "Loading..."
                : "Your Feed"}
            </h1>
            <p className="mt-2 text-sm text-slate-500">
              This is your personalized feed. Posts from people you follow will
              appear here.
            </p>
          </div>

          <div className="space-y-4">
            {loading ? (
              <div className="rounded-3xl border border-dashed border-slate-200 bg-white/70 p-8 text-center text-sm text-slate-500">
                Loading your feed...
              </div>
            ) : error ? (
              <div className="rounded-3xl border border-red-200 bg-red-50 p-6 text-sm text-red-700">
                {error}
              </div>
            ) : posts.length === 0 ? (
              <div className="rounded-3xl border border-slate-200 bg-white p-8 text-center text-sm text-slate-500">
                No posts yet. Follow people or create your first post.
              </div>
            ) : (
              posts.map((post) => (
                <article
                  key={post.id}
                  className="rounded-3xl border border-slate-200 bg-white p-6 shadow-sm"
                >
                  <header className="flex items-center justify-between text-xs text-slate-500">
                    <span>
                      {post.author_first_name} {post.author_last_name}
                      {post.author_nickname ? ` · ${post.author_nickname}` : ""}
                    </span>
                    <span className="uppercase tracking-wide">{post.privacy}</span>
                  </header>
                  <p className="mt-4 text-sm leading-relaxed text-slate-700">
                    {post.content}
                  </p>
                  <footer className="mt-6 flex items-center gap-5 text-xs text-slate-400">
                    <span>{post.comment_count} comments</span>
                    <span>{post.like_count} likes</span>
                    <span>{post.dislike_count} dislikes</span>
                  </footer>
                </article>
              ))
            )}
          </div>
        </section>

        <aside className="space-y-6">
          <div className="rounded-3xl border border-slate-200 bg-white p-6 shadow-sm">
            <p className="text-xs uppercase tracking-[0.2em] text-slate-400">
              Quick Actions
            </p>
            <div className="mt-4 space-y-3 text-sm text-slate-600">
              <button className="w-full rounded-2xl border border-slate-200 px-4 py-2 text-left transition hover:border-slate-300 hover:text-slate-900">
                Create a post
              </button>
              <button className="w-full rounded-2xl border border-slate-200 px-4 py-2 text-left transition hover:border-slate-300 hover:text-slate-900">
                Find new people
              </button>
              <button className="w-full rounded-2xl border border-slate-200 px-4 py-2 text-left transition hover:border-slate-300 hover:text-slate-900">
                Explore groups
              </button>
            </div>
          </div>
          <div className="rounded-3xl border border-slate-200 bg-white p-6 text-sm text-slate-500 shadow-sm">
            Notifications will appear here when the backend is connected.
          </div>
        </aside>
      </main>
    </div>
  );
}
