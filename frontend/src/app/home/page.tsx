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

type CreatePostPayload = {
  content: string;
  privacy: "public" | "followers" | "custom";
};

export default function HomePage() {
  const router = useRouter();
  const { logout } = useAuth();
  const [user, setUser] = useState<User | null>(null);
  const [posts, setPosts] = useState<Post[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [composerText, setComposerText] = useState("");
  const [composerPrivacy, setComposerPrivacy] =
    useState<CreatePostPayload["privacy"]>("public");
  const [isPosting, setIsPosting] = useState(false);

  const apiBaseUrl = useMemo(
    () =>
      process.env.NEXT_PUBLIC_API_BASE_URL?.trim().replace(/\/+$/, "") ||
      "http://localhost:8080",
    []
  );

  useEffect(() => {
    let cancelled = false;

    const fetchJson = async <T,>(path: string, init?: RequestInit) => {
      const response = await fetch(`${apiBaseUrl}${path}`, {
        credentials: "include",
        ...init,
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

  const handleCreatePost = async () => {
    if (isPosting) return;
    const content = composerText.trim();
    if (!content) {
      setError("Write something before posting.");
      return;
    }

    setIsPosting(true);
    setError(null);

    try {
      const response = await fetch(`${apiBaseUrl}/posts`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        credentials: "include",
        body: JSON.stringify({
          content,
          privacy: composerPrivacy,
        } satisfies CreatePostPayload),
      });

      const result = (await response.json().catch(() => null)) as
        | ApiResponse<Post>
        | null;

      if (!response.ok || !result?.success || !result.data) {
        setError(result?.error || "Failed to create post.");
        return;
      }

      setPosts((prev) => [result.data as Post, ...prev]);
      setComposerText("");
      setComposerPrivacy("public");
    } catch {
      setError("Network error. Please try again.");
    } finally {
      setIsPosting(false);
    }
  };

  return (
    <div className="min-h-screen bg-[radial-gradient(circle_at_top,_#fef3c7,_#ffffff_45%)] text-slate-900">
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

          <div className="rounded-3xl border border-slate-200 bg-white p-6 shadow-sm">
            <div className="flex items-center justify-between">
              <p className="text-sm font-medium text-slate-700">
                Create a post
              </p>
              <span className="text-xs uppercase tracking-[0.2em] text-slate-400">
                {composerPrivacy}
              </span>
            </div>
            <textarea
              value={composerText}
              onChange={(event) => setComposerText(event.target.value)}
              placeholder="Share something with your network..."
              className="mt-4 min-h-[120px] w-full resize-none rounded-2xl border border-slate-200 bg-slate-50 px-4 py-3 text-sm text-slate-700 focus:outline-none focus:ring-2 focus:ring-amber-400"
            />
            <div className="mt-4 flex flex-wrap items-center gap-3">
              <select
                value={composerPrivacy}
                onChange={(event) =>
                  setComposerPrivacy(event.target.value as CreatePostPayload["privacy"])
                }
                className="rounded-full border border-slate-200 bg-white px-4 py-2 text-xs text-slate-600"
              >
                <option value="public">Public</option>
                <option value="followers">Followers</option>
                <option value="custom">Custom</option>
              </select>
              <button
                type="button"
                disabled={isPosting}
                onClick={handleCreatePost}
                className="rounded-full bg-amber-500 px-5 py-2 text-xs font-semibold uppercase tracking-wide text-white transition hover:bg-amber-600 disabled:cursor-not-allowed disabled:bg-amber-300"
              >
                {isPosting ? "Posting..." : "Post"}
              </button>
              {error ? (
                <span className="text-xs text-red-600">{error}</span>
              ) : null}
            </div>
          </div>

          <div className="space-y-4">
            {loading ? (
              <div className="rounded-3xl border border-dashed border-slate-200 bg-white/70 p-8 text-center text-sm text-slate-500">
                Loading your feed...
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
                      {post.author_nickname ? ` - ${post.author_nickname}` : ""}
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
                Find new people
              </button>
              <button className="w-full rounded-2xl border border-slate-200 px-4 py-2 text-left transition hover:border-slate-300 hover:text-slate-900">
                Explore groups
              </button>
              <button className="w-full rounded-2xl border border-slate-200 px-4 py-2 text-left transition hover:border-slate-300 hover:text-slate-900">
                Start a group event
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
