"use client";

import { useCallback, useEffect, useMemo, useState } from "react";
import Link from "next/link";
import { useParams, useRouter } from "next/navigation";
import { ArrowLeft, Calendar } from "lucide-react";
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
};

type GroupSummary = {
  id: number;
  name: string;
};

type EventItem = {
  id: number;
  group_id: number;
  group_title?: string | null;
  creator_id: number;
  title: string;
  description?: string | null;
  event_time: string;
  created_at: string;
  updated_at: string;
};

function formatDateTime(value?: string) {
  if (!value) return "N/A";
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return "N/A";
  return date.toLocaleString(undefined, { month: "short", day: "numeric", hour: "2-digit", minute: "2-digit" });
}

export default function GroupEventsPage() {
  const router = useRouter();
  const params = useParams<{ id: string }>();
  const groupID = typeof params?.id === "string" ? Number(params.id) : NaN;

  const [viewer, setViewer] = useState<User | null>(null);
  const [group, setGroup] = useState<GroupSummary | null>(null);
  const [events, setEvents] = useState<EventItem[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [isLoadingMore, setIsLoadingMore] = useState(false);
  const [hasMore, setHasMore] = useState(true);

  const [title, setTitle] = useState("");
  const [description, setDescription] = useState("");
  const [eventTime, setEventTime] = useState("");
  const [createError, setCreateError] = useState<string | null>(null);
  const [isCreating, setIsCreating] = useState(false);

  const pageSize = 8;

  const apiBaseUrl = useMemo(() => getApiBaseUrl(), []);

  const loadGroupAndEvents = useCallback(async () => {
    if (!Number.isFinite(groupID) || groupID <= 0) {
      setError("Invalid group id.");
      setIsLoading(false);
      return;
    }

    setIsLoading(true);
    setError(null);
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
      setViewer(meResult.data);

      const { response: eventsResponse, result: eventsResult } = await apiFetchJson<
        ApiResponse<EventItem[]>
      >(`/groups/${groupID}/events?limit=${pageSize}&offset=0`, {}, apiBaseUrl);
      if (!eventsResponse.ok || !eventsResult?.success) {
        if (eventsResponse.status === 403) {
          setError("Join this group to view events.");
        } else {
          setError(eventsResult?.error || "Could not load group events.");
        }
        setEvents([]);
        setHasMore(false);
        return;
      }
      const items = eventsResult.data ?? [];
      setEvents(items);
      if (items.length > 0) {
        const title = items[0].group_title?.trim();
        if (title) {
          setGroup({ id: groupID, name: title });
        }
      } else if (!group) {
        setGroup({ id: groupID, name: `Group ${groupID}` });
      }
      setHasMore(items.length >= pageSize);
    } catch {
      setError("Network error. Please try again.");
      setEvents([]);
      setHasMore(false);
    } finally {
      setIsLoading(false);
    }
  }, [apiBaseUrl, groupID, pageSize, router]);

  useEffect(() => {
    void loadGroupAndEvents();
  }, [loadGroupAndEvents]);

  const loadMore = async () => {
    if (isLoadingMore || !hasMore) return;
    setIsLoadingMore(true);
    try {
      const offset = events.length;
      const { response, result } = await apiFetchJson<ApiResponse<EventItem[]>>(
        `/groups/${groupID}/events?limit=${pageSize}&offset=${offset}`,
        {},
        apiBaseUrl,
      );
      if (!response.ok || !result?.success) return;
      const next = result.data ?? [];
      setEvents((prev) => [...prev, ...next]);
      setHasMore(next.length >= pageSize);
    } finally {
      setIsLoadingMore(false);
    }
  };

  const handleCreate = async () => {
    setCreateError(null);
    if (!title.trim() || !eventTime) {
      setCreateError("Title and date/time are required.");
      return;
    }
    setIsCreating(true);
    try {
      const payload = {
        title: title.trim(),
        description: description.trim() || undefined,
        event_time: new Date(eventTime).toISOString(),
      };
      const { response, result } = await apiFetchJson<ApiResponse<EventItem>>(
        `/groups/${groupID}/events`,
        {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify(payload),
        },
        apiBaseUrl,
      );
      if (!response.ok || !result?.success || !result.data) {
        setCreateError(result?.error || "Could not create event.");
        return;
      }
      setEvents((prev) => [result.data as EventItem, ...prev]);
      setTitle("");
      setDescription("");
      setEventTime("");
    } catch {
      setCreateError("Network error. Please try again.");
    } finally {
      setIsCreating(false);
    }
  };

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
      <TopNav user={viewer ?? undefined} onLogout={() => router.replace("/login")} variant="dark" />

      <main className="mx-auto grid w-full max-w-6xl gap-6 px-4 py-6 sm:px-6 lg:grid-cols-[240px_minmax(0,1fr)]">
        <aside className="hidden lg:block">
          <LeftNav user={viewer ?? undefined} activeHref="/groups" variant="dark" />
        </aside>

        <section className="space-y-5">
          <motion.div
            initial="hidden"
            whileInView="show"
            viewport={viewportOnce}
            variants={fadeUp}
            className="rounded-3xl border border-white/10 bg-white/5 p-5 backdrop-blur-sm"
          >
            <div className="flex flex-wrap items-center justify-between gap-3">
              <div>
                <h1 className="text-xl font-semibold tracking-tight text-white">Group events</h1>
                <p className="text-sm text-neutral-400">
                  {group ? `Events for ${group.name}.` : "Plan activities for this group."}
                </p>
              </div>
              <Link
                href={`/groups/${groupID}`}
                className="inline-flex items-center gap-2 rounded-full border border-white/20 bg-white/5 px-3 py-2 text-xs font-semibold text-neutral-300 transition hover:bg-white/10 hover:text-white"
              >
                <ArrowLeft className="h-3.5 w-3.5" />
                Back to group
              </Link>
            </div>
          </motion.div>

          <motion.div
            initial="hidden"
            whileInView="show"
            viewport={viewportOnce}
            variants={fadeUp}
            className="rounded-3xl border border-white/10 bg-white/5 p-5 backdrop-blur-sm"
          >
            <div className="flex items-center gap-2 text-sm font-semibold text-white">
              <Calendar className="h-4 w-4" />
              Create event
            </div>
            <div className="mt-4 grid gap-3">
              <input
                value={title}
                onChange={(event) => setTitle(event.target.value)}
                placeholder="Event title"
                className="h-10 w-full rounded-2xl border border-white/20 bg-white/5 px-3 text-xs text-white outline-none focus:border-white/40"
              />
              <textarea
                value={description}
                onChange={(event) => setDescription(event.target.value)}
                rows={3}
                placeholder="Description"
                className="w-full resize-none rounded-2xl border border-white/20 bg-white/5 px-3 py-2 text-xs text-white outline-none focus:border-white/40"
              />
              <input
                type="datetime-local"
                value={eventTime}
                onChange={(event) => setEventTime(event.target.value)}
                className="h-10 w-full rounded-2xl border border-white/20 bg-white/5 px-3 text-xs text-white outline-none focus:border-white/40"
              />
              <button
                type="button"
                onClick={handleCreate}
                disabled={isCreating}
                className="inline-flex w-fit items-center gap-2 rounded-full bg-neutral-900 px-4 py-2 text-xs font-semibold text-white transition hover:bg-neutral-800 disabled:cursor-not-allowed disabled:opacity-60"
              >
                Create event
              </button>
              {createError ? <p className="text-xs text-rose-400">{createError}</p> : null}
            </div>
          </motion.div>

          {isLoading ? (
            <motion.div
              initial="hidden"
              whileInView="show"
              viewport={viewportOnce}
              variants={fadeUp}
              className="rounded-3xl border border-white/10 bg-white/5 p-5 text-sm text-neutral-400 backdrop-blur-sm"
            >
              Loading events...
            </motion.div>
          ) : error ? (
            <motion.div
              initial="hidden"
              whileInView="show"
              viewport={viewportOnce}
              variants={fadeUp}
              className="rounded-3xl border border-rose-500/30 bg-rose-500/10 p-5 text-sm text-rose-400"
            >
              {error}
            </motion.div>
          ) : events.length === 0 ? (
            <motion.div
              initial="hidden"
              whileInView="show"
              viewport={viewportOnce}
              variants={fadeUp}
              className="rounded-3xl border border-white/10 bg-white/5 p-5 text-sm text-neutral-400 backdrop-blur-sm"
            >
              No events yet.
            </motion.div>
          ) : (
            <div className="space-y-3">
              {events.map((event) => (
                <motion.article
                  key={event.id}
                  initial="hidden"
                  whileInView="show"
                  viewport={viewportOnce}
                  variants={fadeUp}
                  className="rounded-3xl border border-white/10 bg-white/5 p-5 backdrop-blur-sm"
                >
                  <div className="flex flex-wrap items-start justify-between gap-3">
                    <div>
                      <h2 className="text-lg font-semibold text-white">{event.title}</h2>
                      <p className="mt-1 text-sm text-neutral-400">
                        {event.description || "No description."}
                      </p>
                    </div>
                    <span className="rounded-full border border-white/20 bg-white/5 px-3 py-1 text-xs text-white">
                      {formatDateTime(event.event_time)}
                    </span>
                  </div>
                  <div className="mt-4 flex flex-wrap items-center justify-between gap-2">
                    <span className="text-xs text-neutral-400">Event #{event.id}</span>
                    <Link
                      href={`/events/${event.id}`}
                      className="inline-flex items-center gap-2 rounded-full border border-white/20 bg-white/5 px-3 py-2 text-xs font-semibold text-neutral-300 transition hover:bg-white/10 hover:text-white"
                    >
                      View details
                    </Link>
                  </div>
                </motion.article>
              ))}
              {hasMore ? (
                <button
                  type="button"
                  onClick={loadMore}
                  disabled={isLoadingMore}
                  className="w-full rounded-2xl border border-white/20 bg-white/5 px-4 py-2 text-xs font-semibold text-neutral-300 transition hover:bg-white/10 disabled:cursor-not-allowed disabled:opacity-60"
                >
                  {isLoadingMore ? "Loading..." : "Load more"}
                </button>
              ) : null}
            </div>
          )}
        </section>
      </main>
    </div>
  );
}
