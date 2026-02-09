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

const navItems = [
  { label: "Home", href: "/home", icon: "home" },
  { label: "Explore", href: "/explore", icon: "compass" },
  { label: "Groups", href: "/groups", icon: "users", badge: 3 },
  { label: "Messages", href: "/messages", icon: "chat", badge: 5 },
  { label: "Notifications", href: "/notifications", icon: "bell", badge: 2 },
  { label: "More", href: "/more", icon: "more" },
];

const groupCards = [
  { name: "Design Crew", members: "1.4k members" },
  { name: "Web Developers Hub", members: "892 members" },
  { name: "Hiking Club", members: "623 members" },
];

const suggestions = [
  { name: "Amanda Lewis", mutual: "12 mutual friends" },
  { name: "Matt Williams", mutual: "8 mutual friends" },
  { name: "Lisa Green", mutual: "6 mutual friends" },
];

const chats = [
  { name: "Alex Brown", message: "That looks awesome!" },
  { name: "Cameron Lee", message: "Can we hop on a call later?" },
  { name: "Morgan Fox", message: "Sent the docs over." },
];

const initialsFromName = (first?: string, last?: string) => {
  const firstInitial = first?.trim().charAt(0) ?? "";
  const lastInitial = last?.trim().charAt(0) ?? "";
  const initials = `${firstInitial}${lastInitial}`.toUpperCase();
  return initials || "U";
};

const formatDateTime = (value: string) => {
  const parsed = new Date(value);
  if (Number.isNaN(parsed.getTime())) {
    return "";
  }
  return parsed.toLocaleDateString(undefined, {
    month: "short",
    day: "numeric",
  });
};

const NavIcon = ({ name }: { name: string }) => {
  switch (name) {
    case "home":
      return (
        <svg viewBox="0 0 24 24" className="h-4 w-4" fill="none" stroke="currentColor" strokeWidth="2">
          <path d="M3 11l9-8 9 8" />
          <path d="M5 10v10h14V10" />
        </svg>
      );
    case "compass":
      return (
        <svg viewBox="0 0 24 24" className="h-4 w-4" fill="none" stroke="currentColor" strokeWidth="2">
          <circle cx="12" cy="12" r="9" />
          <path d="M16 8l-3 8-5 2 3-8 5-2z" />
        </svg>
      );
    case "users":
      return (
        <svg viewBox="0 0 24 24" className="h-4 w-4" fill="none" stroke="currentColor" strokeWidth="2">
          <path d="M16 11a4 4 0 10-8 0 4 4 0 008 0z" />
          <path d="M20 20a8 8 0 00-16 0" />
        </svg>
      );
    case "chat":
      return (
        <svg viewBox="0 0 24 24" className="h-4 w-4" fill="none" stroke="currentColor" strokeWidth="2">
          <path d="M4 6h16v10H8l-4 4V6z" />
        </svg>
      );
    case "bell":
      return (
        <svg viewBox="0 0 24 24" className="h-4 w-4" fill="none" stroke="currentColor" strokeWidth="2">
          <path d="M18 8a6 6 0 10-12 0c0 7-3 7-3 7h18s-3 0-3-7" />
          <path d="M13.73 21a2 2 0 01-3.46 0" />
        </svg>
      );
    default:
      return (
        <svg viewBox="0 0 24 24" className="h-4 w-4" fill="none" stroke="currentColor" strokeWidth="2">
          <circle cx="6" cy="12" r="1.5" />
          <circle cx="12" cy="12" r="1.5" />
          <circle cx="18" cy="12" r="1.5" />
        </svg>
      );
  }
};

export default function HomePage() {
  const router = useRouter();
  const { logout } = useAuth();
  const [user, setUser] = useState<User | null>(null);
  const [posts, setPosts] = useState<Post[]>([]);
  const [loading, setLoading] = useState(true);
  const [feedError, setFeedError] = useState<string | null>(null);
  const [composerText, setComposerText] = useState("");
  const [composerError, setComposerError] = useState<string | null>(null);
  const [isPosting, setIsPosting] = useState(false);
  const [groupsOnly, setGroupsOnly] = useState(false);

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
      setFeedError(null);

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

        const feedPath = groupsOnly ? "/posts?groups_only=true" : "/posts";
        const feed = await fetchJson<Post[]>(feedPath);
        if (!feed.response.ok || !feed.result?.success) {
          if (!cancelled) {
            setFeedError(feed.result?.error || "Failed to load posts.");
            setPosts([]);
          }
          return;
        }

        if (!cancelled) {
          setPosts(feed.result.data ?? []);
        }
      } catch {
        if (!cancelled) {
          setFeedError("Network error. Please try again.");
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
  }, [apiBaseUrl, router, groupsOnly]);

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
      setComposerError("Write something before posting.");
      return;
    }

    setIsPosting(true);
    setComposerError(null);

    try {
      const response = await fetch(`${apiBaseUrl}/posts`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        credentials: "include",
        body: JSON.stringify({
          content,
          privacy: "public",
        }),
      });

      const result = (await response.json().catch(() => null)) as
        | ApiResponse<Post>
        | null;

      if (!response.ok || !result?.success || !result.data) {
        setComposerError(result?.error || "Failed to create post.");
        return;
      }

      setPosts((prev) => [result.data as Post, ...prev]);
      setComposerText("");
    } catch {
      setComposerError("Network error. Please try again.");
    } finally {
      setIsPosting(false);
    }
  };

  const displayName = user
    ? `${user.first_name} ${user.last_name}`
    : "Loading...";
  const handle =
    user?.nickname ||
    (user?.email ? user.email.split("@")[0] : undefined) ||
    "user";
  const initials = initialsFromName(user?.first_name, user?.last_name);

  return (
    <div className="min-h-screen bg-gradient-to-b from-[#2b2e34] via-[#1f2228] to-[#14171c] text-slate-100">
      <header className="sticky top-0 z-30 border-b border-white/10 bg-[#2a2d33]/90 backdrop-blur">
        <div className="mx-auto flex max-w-7xl items-center gap-6 px-6 py-4">
          <Link href="/home" className="text-lg font-semibold tracking-tight text-white">
            SocialNetwork
          </Link>
          <div className="flex flex-1 items-center">
            <div className="relative w-full max-w-md">
              <span className="pointer-events-none absolute left-3 top-1/2 -translate-y-1/2 text-slate-400">
                <svg
                  aria-hidden="true"
                  viewBox="0 0 24 24"
                  className="h-4 w-4"
                  fill="none"
                  stroke="currentColor"
                  strokeWidth="2"
                >
                  <circle cx="11" cy="11" r="7" />
                  <path d="M20 20l-3.5-3.5" />
                </svg>
              </span>
              <input
                type="search"
                placeholder="Search..."
                className="w-full rounded-full border border-white/10 bg-[#1d2026] py-2 pl-9 pr-4 text-sm text-slate-200 placeholder:text-slate-500 focus:border-white/30 focus:outline-none"
              />
            </div>
          </div>
          <div className="flex items-center gap-4">
            <button
              type="button"
              className="relative rounded-full border border-white/10 bg-[#1d2026] p-2 text-slate-200"
              aria-label="Messages"
            >
              <svg
                aria-hidden="true"
                viewBox="0 0 24 24"
                className="h-4 w-4"
                fill="none"
                stroke="currentColor"
                strokeWidth="2"
              >
                <path d="M4 5h16v10H8l-4 4V5z" />
              </svg>
              <span className="absolute -right-1 -top-1 flex h-4 w-4 items-center justify-center rounded-full bg-rose-500 text-[10px] font-semibold text-white">
                2
              </span>
            </button>
            <button
              type="button"
              className="relative rounded-full border border-white/10 bg-[#1d2026] p-2 text-slate-200"
              aria-label="Notifications"
            >
              <svg
                aria-hidden="true"
                viewBox="0 0 24 24"
                className="h-4 w-4"
                fill="none"
                stroke="currentColor"
                strokeWidth="2"
              >
                <path d="M18 8a6 6 0 10-12 0c0 7-3 7-3 7h18s-3 0-3-7" />
                <path d="M13.73 21a2 2 0 01-3.46 0" />
              </svg>
              <span className="absolute -right-1 -top-1 flex h-4 w-4 items-center justify-center rounded-full bg-rose-500 text-[10px] font-semibold text-white">
                6
              </span>
            </button>
            <div className="flex items-center gap-3">
              <div className="flex h-9 w-9 items-center justify-center rounded-full bg-gradient-to-br from-amber-300 to-pink-400 text-sm font-semibold text-slate-900">
                {initials}
              </div>
              <div className="hidden text-sm leading-tight text-slate-200 sm:block">
                <p className="font-medium text-white">{displayName}</p>
                <p className="text-xs text-slate-400">@{handle}</p>
              </div>
            </div>
            <button
              type="button"
              className="rounded-full border border-white/10 bg-[#1d2026] px-4 py-2 text-xs font-semibold uppercase tracking-wide text-slate-200"
              onClick={handleLogout}
            >
              Log out
            </button>
          </div>
        </div>
      </header>

      <main className="mx-auto grid max-w-7xl gap-6 px-6 py-8 lg:grid-cols-[250px_minmax(0,1fr)_280px]">
        <aside className="space-y-6">
          <div className="rounded-3xl border border-white/70 bg-white/90 p-5 text-slate-700 shadow-[0_20px_40px_rgba(0,0,0,0.2)]">
            <div className="flex items-center gap-3">
              <div className="flex h-12 w-12 items-center justify-center rounded-2xl bg-gradient-to-br from-sky-400 to-emerald-400 text-lg font-semibold text-slate-900">
                {initials}
              </div>
              <div>
                <p className="text-sm font-semibold text-slate-900">{displayName}</p>
                <p className="text-xs text-slate-500">@{handle}</p>
              </div>
            </div>
            <div className="mt-4 grid grid-cols-3 gap-2 text-center text-xs">
              <div className="rounded-2xl border border-slate-200 bg-slate-50 px-2 py-2">
                <p className="text-sm font-semibold text-slate-900">875</p>
                <p className="text-slate-500">Followers</p>
              </div>
              <div className="rounded-2xl border border-slate-200 bg-slate-50 px-2 py-2">
                <p className="text-sm font-semibold text-slate-900">499</p>
                <p className="text-slate-500">Following</p>
              </div>
              <div className="rounded-2xl border border-slate-200 bg-slate-50 px-2 py-2">
                <p className="text-sm font-semibold text-slate-900">392</p>
                <p className="text-slate-500">Groups</p>
              </div>
            </div>
            <nav className="mt-5 space-y-2 text-sm">
              {navItems.map((item) => (
                <Link
                  key={item.label}
                  href={item.href}
                  className="flex items-center justify-between rounded-2xl border border-slate-200 bg-slate-100 px-3 py-2 text-slate-700 transition hover:border-slate-300 hover:text-slate-900"
                >
                  <span className="flex items-center gap-2">
                    <span className="text-slate-600">
                      <NavIcon name={item.icon} />
                    </span>
                    {item.label}
                  </span>
                  {item.badge ? (
                    <span className="rounded-full bg-slate-200 px-2 py-0.5 text-[10px] text-slate-700">
                      {item.badge}
                    </span>
                  ) : null}
                </Link>
              ))}
            </nav>
            <button className="mt-5 w-full rounded-2xl bg-gradient-to-b from-[#2d3138] to-[#1c1f26] px-4 py-2 text-xs font-semibold uppercase tracking-wide text-white shadow-inner">
              Create Post
            </button>
          </div>
        </aside>

        <section className="space-y-6">
          <div className="rounded-3xl border border-white/10 bg-[#2a2d33] p-4 shadow-[0_20px_40px_rgba(0,0,0,0.3)]">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-xs uppercase tracking-[0.4em] text-slate-300">
                  Post Feed
                </p>
                <p className="text-sm text-slate-400">
                  Share updates or browse the latest activity.
                </p>
              </div>
              <label className="flex items-center gap-3 text-xs text-slate-200">
                <span>Groups Only</span>
                <span className="relative inline-flex h-5 w-10 items-center">
                  <input
                    type="checkbox"
                    className="peer sr-only"
                    checked={groupsOnly}
                    onChange={(event) => setGroupsOnly(event.target.checked)}
                  />
                  <span className="h-5 w-10 rounded-full bg-slate-700 transition peer-checked:bg-slate-300" />
                  <span className="absolute left-1 top-1 h-3 w-3 rounded-full bg-white shadow transition peer-checked:translate-x-5 peer-checked:bg-slate-900" />
                </span>
              </label>
            </div>
          </div>

          <div className="rounded-3xl border border-white/80 bg-white/95 p-4 text-slate-700 shadow-[0_20px_40px_rgba(0,0,0,0.2)]">
            <div className="flex flex-col gap-4 sm:flex-row sm:items-start">
              <div className="flex h-11 w-11 items-center justify-center rounded-full bg-gradient-to-br from-amber-300 to-pink-400 text-sm font-semibold text-slate-900">
                {initials}
              </div>
              <div className="flex-1">
                <textarea
                  value={composerText}
                  onChange={(event) => setComposerText(event.target.value)}
                  placeholder="What's on your mind?"
                  className="min-h-[90px] w-full resize-none rounded-2xl border border-slate-200 bg-white px-4 py-3 text-sm text-slate-700 placeholder:text-slate-400 focus:border-slate-300 focus:outline-none"
                />
                <div className="mt-3 flex flex-wrap items-center justify-between gap-3 text-xs text-slate-500">
                  <div className="flex flex-wrap items-center gap-3">
                    <button className="flex items-center gap-2 rounded-full border border-slate-200 bg-slate-50 px-3 py-1.5 text-xs text-slate-600">
                      <span>Photo</span>
                    </button>
                    <button className="flex items-center gap-2 rounded-full border border-slate-200 bg-slate-50 px-3 py-1.5 text-xs text-slate-600">
                      <span>GIF</span>
                    </button>
                    <button className="flex items-center gap-2 rounded-full border border-slate-200 bg-slate-50 px-3 py-1.5 text-xs text-slate-600">
                      <span>Location</span>
                    </button>
                    <button className="flex items-center gap-2 rounded-full border border-slate-200 bg-slate-50 px-3 py-1.5 text-xs text-slate-600">
                      <span>Emoji</span>
                    </button>
                  </div>
                  <button
                    type="button"
                    onClick={handleCreatePost}
                    disabled={isPosting}
                    className="rounded-xl bg-gradient-to-b from-[#2d3138] to-[#1c1f26] px-5 py-2 text-xs font-semibold uppercase tracking-wide text-white shadow-inner disabled:cursor-not-allowed disabled:opacity-70"
                  >
                    {isPosting ? "Posting..." : "Post"}
                  </button>
                </div>
                {composerError ? (
                  <p className="mt-3 text-xs text-rose-500">{composerError}</p>
                ) : null}
              </div>
            </div>
          </div>

          <div className="space-y-6">
            {loading ? (
              <div className="rounded-3xl border border-white/10 bg-[#2a2d33] p-6 text-center text-sm text-slate-300">
                Loading your feed...
              </div>
            ) : feedError ? (
              <div className="rounded-3xl border border-rose-400/30 bg-rose-500/10 p-6 text-sm text-rose-200">
                {feedError}
              </div>
            ) : posts.length === 0 ? (
              <div className="rounded-3xl border border-white/10 bg-[#2a2d33] p-6 text-center text-sm text-slate-300">
                No posts yet. Follow people or create your first post.
              </div>
            ) : (
              posts.map((post) => (
                <article
                  key={post.id}
                  className="rounded-3xl border border-white/10 bg-[#2a2d33] p-5 shadow-[0_18px_35px_rgba(0,0,0,0.35)]"
                >
                  <header className="flex items-start justify-between text-xs text-slate-300">
                    <div className="flex items-center gap-3">
                      <div className="flex h-10 w-10 items-center justify-center rounded-full bg-gradient-to-br from-amber-300 to-pink-400 text-sm font-semibold text-slate-900">
                        {initialsFromName(post.author_first_name, post.author_last_name)}
                      </div>
                      <div>
                        <p className="text-sm font-semibold text-white">
                          {post.author_first_name} {post.author_last_name}
                        </p>
                        <p className="text-xs text-slate-400">
                          {formatDateTime(post.created_at)}
                        </p>
                      </div>
                    </div>
                    <button className="rounded-full border border-white/10 bg-[#1c1f26] px-3 py-1 text-[10px] uppercase tracking-wide text-slate-200">
                      {post.privacy}
                    </button>
                  </header>
                  <p className="mt-4 text-sm leading-relaxed text-slate-100">
                    {post.content}
                  </p>
                  {post.media_path ? (
                    <div className="mt-4 overflow-hidden rounded-2xl border border-white/10">
                      <img
                        src={post.media_path}
                        alt="Post media"
                        className="h-64 w-full object-cover"
                      />
                    </div>
                  ) : null}
                  <footer className="mt-4 flex items-center gap-5 text-xs text-slate-300">
                    <span>{post.like_count} likes</span>
                    <span>{post.comment_count} comments</span>
                    <span>{post.dislike_count} dislikes</span>
                  </footer>
                </article>
              ))
            )}
          </div>
        </section>

        <aside className="space-y-6">
          <div className="rounded-3xl border border-white/70 bg-white/95 p-5 text-slate-700 shadow-[0_20px_40px_rgba(0,0,0,0.2)]">
            <div className="flex items-center justify-between">
              <p className="text-sm font-semibold text-slate-900">Groups</p>
              <button className="text-xs text-slate-400">View all</button>
            </div>
            <div className="mt-3">
              <input
                type="search"
                placeholder="Search groups..."
                className="w-full rounded-2xl border border-slate-200 bg-white px-3 py-2 text-xs text-slate-600 placeholder:text-slate-400 focus:border-slate-300 focus:outline-none"
              />
            </div>
            <button className="mt-3 w-full rounded-2xl bg-gradient-to-b from-[#2d3138] to-[#1c1f26] px-4 py-2 text-xs font-semibold uppercase tracking-wide text-white shadow-inner">
              Create Group
            </button>
            <div className="mt-4 space-y-3 text-xs text-slate-600">
              {groupCards.map((group) => (
                <div
                  key={group.name}
                  className="rounded-2xl border border-slate-200 bg-slate-50 px-3 py-3"
                >
                  <p className="text-sm font-semibold text-slate-800">{group.name}</p>
                  <p className="text-xs text-slate-500">{group.members}</p>
                </div>
              ))}
            </div>

            <div className="mt-6 border-t border-slate-200 pt-5">
              <p className="text-sm font-semibold text-slate-900">Suggestions for You</p>
              <div className="mt-4 space-y-3">
                {suggestions.map((suggestion) => (
                  <div
                    key={suggestion.name}
                    className="flex items-center justify-between rounded-2xl border border-slate-200 bg-slate-50 px-3 py-3"
                  >
                    <div className="flex items-center gap-3">
                      <div className="flex h-9 w-9 items-center justify-center rounded-full bg-gradient-to-br from-indigo-400 to-sky-400 text-xs font-semibold text-white">
                        {suggestion.name
                          .split(" ")
                          .map((part) => part.charAt(0))
                          .join("")
                          .slice(0, 2)
                          .toUpperCase()}
                      </div>
                      <div>
                        <p className="text-xs font-semibold text-slate-800">{suggestion.name}</p>
                        <p className="text-[10px] text-slate-500">{suggestion.mutual}</p>
                      </div>
                    </div>
                    <button className="rounded-full border border-slate-200 bg-white px-3 py-1 text-[10px] font-semibold uppercase tracking-wide text-slate-700">
                      Follow
                    </button>
                  </div>
                ))}
              </div>
            </div>

            <div className="mt-6 border-t border-slate-200 pt-5">
              <div className="flex items-center justify-between">
                <p className="text-sm font-semibold text-slate-900">Chat (3)</p>
                <button className="text-xs text-slate-400">Open</button>
              </div>
              <div className="mt-4 space-y-3">
                {chats.map((chat) => (
                  <div
                    key={chat.name}
                    className="flex items-center justify-between rounded-2xl border border-slate-200 bg-slate-50 px-3 py-3"
                  >
                    <div>
                      <p className="text-xs font-semibold text-slate-800">{chat.name}</p>
                      <p className="text-[10px] text-slate-500">{chat.message}</p>
                    </div>
                    <button className="rounded-full border border-slate-200 bg-white px-2 py-1 text-[10px] text-slate-600">
                      Reply
                    </button>
                  </div>
                ))}
              </div>
            </div>
          </div>
        </aside>
      </main>
    </div>
  );
}
