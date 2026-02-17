"use client";
/* eslint-disable @next/next/no-img-element */

import { useEffect, useMemo, useState } from "react";
import Image from "next/image";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { ArrowLeft, MessageCircle, ThumbsDown, ThumbsUp, Send } from "lucide-react";
import { landingData } from "@/lib/data";
import { apiJson, asArray, asNumber, asString, isRecord } from "@/lib/api";

type PostVM = {
  id: number;
  authorName: string;
  content: string;
  mediaUrl?: string | null;
  privacyLabel: string;
  createdAt: string;
  counts: { likes: number; dislikes: number; comments: number };
};

type CommentVM = {
  id: number;
  content: string;
  createdAt: string;
  counts: { likes: number; dislikes: number };
};

type ReactionKind = "like" | "dislike";

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

function toPostVM(value: unknown): PostVM | null {
  if (!isRecord(value)) return null;
  const id = asNumber(value.id);
  if (!id) return null;

  const first = asString(value.author_first_name) ?? "";
  const last = asString(value.author_last_name) ?? "";

  return {
    id,
    authorName: `${first} ${last}`.trim() || "Member",
    content: asString(value.content) ?? "",
    mediaUrl: asString(value.media_path),
    privacyLabel: asString(value.privacy) ?? "public",
    createdAt: asString(value.created_at) ?? "",
    counts: {
      comments: asNumber(value.comment_count) ?? 0,
      likes: asNumber(value.like_count) ?? 0,
      dislikes: asNumber(value.dislike_count) ?? 0,
    },
  };
}

function toCommentVM(value: unknown): CommentVM | null {
  if (!isRecord(value)) return null;
  const id = asNumber(value.id);
  if (!id) return null;
  return {
    id,
    content: asString(value.content) ?? "",
    createdAt: asString(value.created_at) ?? "",
    counts: {
      likes: asNumber(value.like_count) ?? 0,
      dislikes: asNumber(value.dislike_count) ?? 0,
    },
  };
}

export default function PostPage({ postId }: { postId: number }) {
  const router = useRouter();

  const [post, setPost] = useState<PostVM | null>(null);
  const [comments, setComments] = useState<CommentVM[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [draft, setDraft] = useState("");
  const [isCommenting, setIsCommenting] = useState(false);
  const [commentError, setCommentError] = useState<string | null>(null);
  const [postReaction, setPostReaction] = useState<ReactionKind | null>(null);

  const apiBaseUrl = useMemo(
    () =>
      process.env.NEXT_PUBLIC_API_BASE_URL?.trim().replace(/\/+$/, "") ||
      "http://localhost:8080",
    [],
  );

  useEffect(() => {
    if (!Number.isFinite(postId) || postId <= 0) {
      setIsLoading(false);
      setError("Invalid post.");
      return;
    }

    let cancelled = false;

    const load = async () => {
      setIsLoading(true);
      setError(null);
      setCommentError(null);

      try {
        const [postRes, commentsRes] = await Promise.all([
          apiJson(apiBaseUrl, `/posts/${postId}`),
          apiJson(apiBaseUrl, `/posts/${postId}/comments`).catch(() => null),
        ]);

        if (postRes.status === 401) {
          router.replace("/login");
          return;
        }

        if (!postRes.ok || !postRes.json?.success || !postRes.json.data) {
          setError(postRes.json?.error || "Post not found.");
          setPost(null);
          setComments([]);
          return;
        }

        if (cancelled) return;
        setPost(toPostVM(postRes.json.data));

        if (commentsRes?.ok && commentsRes.json?.success) {
          const raw = asArray(commentsRes.json.data) ?? [];
          setComments(raw.map(toCommentVM).filter(Boolean) as CommentVM[]);
        } else {
          setComments([]);
        }
      } catch {
        if (!cancelled) {
          setError("Network error. Please try again.");
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
  }, [apiBaseUrl, postId, router]);

  const submitComment = async () => {
    if (isCommenting) return;
    const content = draft.trim();
    if (!content) {
      setCommentError("Write a comment first.");
      return;
    }

    setIsCommenting(true);
    setCommentError(null);

    try {
      const res = await apiJson(apiBaseUrl, `/posts/${postId}/comments`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ content }),
      });
      if (res.status === 401) {
        router.replace("/login");
        return;
      }
      if (!res.ok || !res.json?.success || !res.json.data) {
        setCommentError(res.json?.error || "Could not post comment.");
        return;
      }

      const created = toCommentVM(res.json.data);
      if (created) {
        setComments((prev) => [created, ...prev]);
      }
      setDraft("");
      setPost((prev) =>
        prev ? { ...prev, counts: { ...prev.counts, comments: prev.counts.comments + 1 } } : prev,
      );
    } catch {
      setCommentError("Network error. Please try again.");
    } finally {
      setIsCommenting(false);
    }
  };

  const reactToPost = async (reaction: ReactionKind) => {
    if (!post) return;
    const previous = postReaction;
    const next = previous === reaction ? null : reaction;

    setPostReaction(next);
    setPost((prev) => {
      if (!prev) return prev;
      let like = prev.counts.likes;
      let dislike = prev.counts.dislikes;
      if (previous === "like") like -= 1;
      if (previous === "dislike") dislike -= 1;
      if (next === "like") like += 1;
      if (next === "dislike") dislike += 1;
      return {
        ...prev,
        counts: { ...prev.counts, likes: Math.max(0, like), dislikes: Math.max(0, dislike) },
      };
    });

    await apiJson(apiBaseUrl, `/posts/${post.id}/reactions`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ reaction }),
    }).catch(() => undefined);
  };

  return (
    <div className="min-h-screen bg-neutral-50 text-neutral-900">
      <header className="sticky top-0 z-40 border-b border-neutral-200/80 bg-white/85 backdrop-blur-md">
        <div className="mx-auto flex w-full max-w-6xl items-center gap-3 px-4 py-3 sm:px-6">
          <button
            type="button"
            onClick={() => router.back()}
            className="inline-flex items-center gap-2 text-sm font-semibold text-neutral-700"
          >
            <ArrowLeft className="h-4 w-4" />
            Back
          </button>
          <Link href="/explore" className="ml-auto inline-flex items-center gap-2">
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
        </div>
      </header>

      <main className="mx-auto w-full max-w-6xl px-4 py-6 sm:px-6">
        {isLoading ? (
          <article className="rounded-3xl border border-neutral-200 bg-white p-6 text-sm text-neutral-600 shadow-sm">
            Loading post...
          </article>
        ) : error ? (
          <article className="rounded-3xl border border-rose-200 bg-rose-50 p-6 text-sm text-rose-700">
            {error}
          </article>
        ) : post ? (
          <div className="space-y-5">
            <article className="rounded-3xl border border-neutral-200 bg-white p-5 shadow-sm">
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

              <p className="mt-4 text-sm leading-relaxed text-neutral-700">{post.content}</p>

              {post.mediaUrl ? (
                <div className="mt-4 overflow-hidden rounded-2xl border border-neutral-200">
                  <img src={post.mediaUrl} alt="Post media" className="h-72 w-full object-cover" />
                </div>
              ) : null}

              <footer className="mt-4 flex flex-wrap items-center gap-3 text-xs text-neutral-500">
                <button
                  type="button"
                  onClick={() => void reactToPost("like")}
                  className={`inline-flex items-center gap-1 rounded-full px-2 py-1 transition ${
                    postReaction === "like"
                      ? "bg-emerald-100 text-emerald-800"
                      : "bg-neutral-100 text-neutral-600 hover:bg-neutral-200"
                  }`}
                >
                  <ThumbsUp className="h-3.5 w-3.5" />
                  {post.counts.likes}
                </button>
                <button
                  type="button"
                  onClick={() => void reactToPost("dislike")}
                  className={`inline-flex items-center gap-1 rounded-full px-2 py-1 transition ${
                    postReaction === "dislike"
                      ? "bg-rose-100 text-rose-800"
                      : "bg-neutral-100 text-neutral-600 hover:bg-neutral-200"
                  }`}
                >
                  <ThumbsDown className="h-3.5 w-3.5" />
                  {post.counts.dislikes}
                </button>
                <span className="inline-flex items-center gap-1 rounded-full bg-neutral-100 px-2 py-1 text-neutral-600">
                  <MessageCircle className="h-3.5 w-3.5" />
                  {post.counts.comments} comments
                </span>
              </footer>
            </article>

            <section className="rounded-3xl border border-neutral-200 bg-white p-5 shadow-sm">
              <h2 className="text-sm font-semibold text-neutral-900">Comments</h2>

              <div className="mt-4 flex gap-2">
                <input
                  value={draft}
                  onChange={(event) => setDraft(event.target.value)}
                  placeholder="Write a comment..."
                  className="h-10 flex-1 rounded-xl border border-neutral-200 bg-neutral-50 px-3 text-sm outline-none transition focus:border-neutral-400"
                />
                <button
                  type="button"
                  onClick={submitComment}
                  disabled={isCommenting}
                  className="brand-gradient inline-flex items-center gap-2 rounded-xl px-4 text-sm font-semibold text-white disabled:opacity-70"
                >
                  <Send className="h-4 w-4" />
                  {isCommenting ? "Sending" : "Post"}
                </button>
              </div>
              {commentError ? <p className="mt-2 text-xs text-rose-600">{commentError}</p> : null}

              <div className="mt-4 space-y-2">
                {comments.length === 0 ? (
                  <p className="text-sm text-neutral-600">No comments yet.</p>
                ) : (
                  comments.map((comment) => (
                    <article key={comment.id} className="rounded-2xl border border-neutral-200 bg-neutral-50 p-3">
                      <p className="text-sm text-neutral-700">{comment.content}</p>
                      <p className="mt-2 text-xs text-neutral-500">{shortDate(comment.createdAt)}</p>
                    </article>
                  ))
                )}
              </div>
            </section>
          </div>
        ) : null}
      </main>
    </div>
  );
}

