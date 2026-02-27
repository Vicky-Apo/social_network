"use client";

import { useCallback, useEffect, useMemo, useState } from "react";
import Link from "next/link";
import { useParams, useRouter } from "next/navigation";
import { ArrowLeft, Calendar, Check, Pencil, ThumbsDown, ThumbsUp, Trash2 } from "lucide-react";
import { motion } from "framer-motion";
import TopNav from "@/components/TopNav";
import LeftNav from "@/components/LeftNav";
import Avatar from "@/components/Avatar";
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

type EventItem = {
  id: number;
  group_id: number;
  creator_id: number;
  title: string;
  description?: string | null;
  event_time: string;
  created_at: string;
  updated_at: string;
};

type EventResponse = {
  event_id: number;
  user_id: number;
  first_name: string;
  last_name: string;
  nickname?: string | null;
  avatar_path?: string | null;
  response: string;
  responded_at?: string | null;
};

type GroupSummary = {
  id: number;
  title?: string | null;
  name?: string | null;
};

function toMediaUrl(apiBaseUrl: string, path?: string | null) {
  if (!path) return "";
  if (path.startsWith("http://") || path.startsWith("https://")) return path;
  const normalized = path.startsWith("/") ? path : `/${path}`;
  return `${apiBaseUrl}${normalized}`;
}

function formatDateTime(value?: string) {
  if (!value) return "N/A";
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return "N/A";
  return date.toLocaleString(undefined, { month: "short", day: "numeric", hour: "2-digit", minute: "2-digit" });
}

function toInputDate(value?: string) {
  if (!value) return "";
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return "";
  const pad = (n: number) => String(n).padStart(2, "0");
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())}T${pad(date.getHours())}:${pad(date.getMinutes())}`;
}

export default function EventDetailPage() {
  const router = useRouter();
  const params = useParams<{ id: string }>();
  const eventID = typeof params?.id === "string" ? Number(params.id) : NaN;

  const [viewer, setViewer] = useState<User | null>(null);
  const [event, setEvent] = useState<EventItem | null>(null);
  const [responses, setResponses] = useState<EventResponse[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [responsesError, setResponsesError] = useState<string | null>(null);
  const [responseAction, setResponseAction] = useState<string | null>(null);
  const [isEditing, setIsEditing] = useState(false);
  const [editTitle, setEditTitle] = useState("");
  const [editDescription, setEditDescription] = useState("");
  const [editTime, setEditTime] = useState("");
  const [editError, setEditError] = useState<string | null>(null);
  const [isSaving, setIsSaving] = useState(false);
  const [deleteError, setDeleteError] = useState<string | null>(null);
  const [groupName, setGroupName] = useState<string | null>(null);

  const apiBaseUrl = useMemo(
    () =>
      process.env.NEXT_PUBLIC_API_BASE_URL?.trim().replace(/\/+$/, "") ||
      "http://localhost:8080",
    [],
  );

  const loadEvent = useCallback(async () => {
    if (!Number.isFinite(eventID) || eventID <= 0) {
      setError("Invalid event id.");
      setIsLoading(false);
      return;
    }

    setIsLoading(true);
    setError(null);
    try {
      const meResponse = await fetch(`${apiBaseUrl}/auth/me`, { credentials: "include" });
      const meResult = (await meResponse.json().catch(() => null)) as ApiResponse<User> | null;
      if (!meResponse.ok || !meResult?.success || !meResult.data) {
        router.replace("/login");
        return;
      }
      setViewer(meResult.data);

      const eventResponse = await fetch(`${apiBaseUrl}/events/${eventID}`, { credentials: "include" });
      const eventResult = (await eventResponse.json().catch(() => null)) as ApiResponse<EventItem> | null;
      if (!eventResponse.ok || !eventResult?.success || !eventResult.data) {
        setError(eventResult?.error || "Could not load event.");
        setEvent(null);
        return;
      }
      setEvent(eventResult.data);
      setEditTitle(eventResult.data.title);
      setEditDescription(eventResult.data.description || "");
      setEditTime(toInputDate(eventResult.data.event_time));

      const groupResponse = await fetch(`${apiBaseUrl}/groups/${eventResult.data.group_id}`, {
        credentials: "include",
      });
      const groupResult = (await groupResponse.json().catch(() => null)) as
        | ApiResponse<GroupSummary>
        | null;
      if (groupResponse.ok && groupResult?.success && groupResult.data) {
        const title = groupResult.data.title || groupResult.data.name;
        if (title) setGroupName(title);
      }

      const responsesResponse = await fetch(`${apiBaseUrl}/events/${eventID}/responses`, {
        credentials: "include",
      });
      const responsesResult = (await responsesResponse.json().catch(() => null)) as
        | ApiResponse<EventResponse[]>
        | null;
      if (!responsesResponse.ok || !responsesResult?.success) {
        setResponsesError(responsesResult?.error || "Could not load responses.");
        setResponses([]);
        return;
      }
      setResponses(responsesResult.data ?? []);
    } catch {
      setError("Network error. Please try again.");
      setEvent(null);
    } finally {
      setIsLoading(false);
    }
  }, [apiBaseUrl, eventID, router]);

  useEffect(() => {
    void loadEvent();
  }, [loadEvent]);

  const handleRespond = async (response: "going" | "not_going") => {
    setResponseAction(null);
    try {
      const res = await fetch(`${apiBaseUrl}/events/${eventID}/responses`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        credentials: "include",
        body: JSON.stringify({ response }),
      });
      const json = (await res.json().catch(() => null)) as ApiResponse<unknown> | null;
      if (!res.ok || !json?.success) {
        setResponseAction(json?.error || "Could not submit response.");
        return;
      }
      await loadEvent();
    } catch {
      setResponseAction("Network error. Please try again.");
    }
  };

  const handleSave = async () => {
    if (!event) return;
    setEditError(null);
    if (!editTitle.trim() || !editTime) {
      setEditError("Title and date/time are required.");
      return;
    }
    setIsSaving(true);
    try {
      const payload = {
        title: editTitle.trim(),
        description: editDescription.trim() || undefined,
        event_time: new Date(editTime).toISOString(),
      };
      const response = await fetch(`${apiBaseUrl}/events/${event.id}`, {
        method: "PATCH",
        headers: { "Content-Type": "application/json" },
        credentials: "include",
        body: JSON.stringify(payload),
      });
      const result = (await response.json().catch(() => null)) as ApiResponse<EventItem> | null;
      if (!response.ok || !result?.success || !result.data) {
        setEditError(result?.error || "Could not update event.");
        return;
      }
      setEvent(result.data);
      setIsEditing(false);
    } catch {
      setEditError("Network error. Please try again.");
    } finally {
      setIsSaving(false);
    }
  };

  const handleDelete = async () => {
    if (!event) return;
    setDeleteError(null);
    try {
      const response = await fetch(`${apiBaseUrl}/events/${event.id}`, {
        method: "DELETE",
        credentials: "include",
      });
      if (!response.ok) {
        const result = (await response.json().catch(() => null)) as ApiResponse<unknown> | null;
        setDeleteError(result?.error || "Could not delete event.");
        return;
      }
      router.replace(`/groups/${event.group_id}/events`);
    } catch {
      setDeleteError("Network error. Please try again.");
    }
  };

  const isCreator = Boolean(event && viewer && event.creator_id === viewer.id);

  return (
    <div className="min-h-screen bg-neutral-50 text-neutral-900">
      <TopNav user={viewer ?? undefined} onLogout={() => router.replace("/login")} />

      <main className="mx-auto grid w-full max-w-6xl gap-6 px-4 py-6 sm:px-6 lg:grid-cols-[240px_minmax(0,1fr)]">
        <aside className="hidden lg:block">
          <LeftNav user={viewer ?? undefined} activeHref="/groups" />
        </aside>

        <section>
          <motion.section
            initial="hidden"
            whileInView="show"
            viewport={viewportOnce}
            variants={fadeUp}
            className="rounded-3xl border border-neutral-200 bg-white p-6 shadow-sm"
          >
            <div className="flex flex-wrap items-center justify-between gap-3">
              <div>
                <h1 className="text-xl font-semibold tracking-tight text-neutral-900">Event details</h1>
                <p className="text-sm text-neutral-600">
                  View event info and RSVP {groupName ? `· ${groupName}` : ""}.
                </p>
              </div>
              {event ? (
                <Link
                  href={`/groups/${event.group_id}/events`}
                  className="inline-flex items-center gap-2 rounded-full border border-neutral-200 bg-white px-3 py-2 text-xs font-semibold text-neutral-700 transition hover:border-neutral-400 hover:text-neutral-900"
                >
                  <ArrowLeft className="h-3.5 w-3.5" />
                  Back to events
                </Link>
              ) : null}
            </div>
          </motion.section>

          {isLoading ? (
          <motion.div
            initial="hidden"
            whileInView="show"
            viewport={viewportOnce}
            variants={fadeUp}
            className="mt-5 rounded-3xl border border-neutral-200 bg-white p-6 text-sm text-neutral-600 shadow-sm"
          >
            Loading event...
          </motion.div>
        ) : error ? (
          <motion.div
            initial="hidden"
            whileInView="show"
            viewport={viewportOnce}
            variants={fadeUp}
            className="mt-5 rounded-3xl border border-rose-200 bg-rose-50 p-6 text-sm text-rose-700"
          >
            {error}
          </motion.div>
        ) : event ? (
          <div className="mt-5 space-y-5">
            <motion.div
              initial="hidden"
              whileInView="show"
              viewport={viewportOnce}
              variants={fadeUp}
              className="rounded-3xl border border-neutral-200 bg-white p-6 shadow-sm"
            >
              <div className="flex flex-wrap items-start justify-between gap-3">
                <div>
                  <h2 className="text-lg font-semibold text-neutral-900">{event.title}</h2>
                  <p className="mt-1 text-sm text-neutral-600">
                    {event.description || "No description."}
                  </p>
                </div>
                <span className="rounded-full border border-neutral-200 bg-neutral-50 px-3 py-1 text-xs text-neutral-600">
                  {formatDateTime(event.event_time)}
                </span>
              </div>
              <div className="mt-4 flex flex-wrap gap-2">
                <button
                  type="button"
                  onClick={() => handleRespond("going")}
                  className="inline-flex items-center gap-2 rounded-full bg-emerald-600 px-3 py-2 text-xs font-semibold text-white transition hover:bg-emerald-700"
                >
                  <ThumbsUp className="h-3.5 w-3.5" />
                  Going
                </button>
                <button
                  type="button"
                  onClick={() => handleRespond("not_going")}
                  className="inline-flex items-center gap-2 rounded-full border border-neutral-200 bg-white px-3 py-2 text-xs font-semibold text-neutral-700 transition hover:border-neutral-400"
                >
                  <ThumbsDown className="h-3.5 w-3.5" />
                  Not going
                </button>
                {responseAction ? (
                  <span className="text-xs text-rose-600">{responseAction}</span>
                ) : null}
              </div>
            </motion.div>

            {isCreator ? (
              <motion.div
                initial="hidden"
                whileInView="show"
                viewport={viewportOnce}
                variants={fadeUp}
                className="rounded-3xl border border-neutral-200 bg-white p-6 shadow-sm"
              >
                <div className="flex flex-wrap items-center justify-between gap-3">
                  <div>
                    <h3 className="text-sm font-semibold text-neutral-900">Manage event</h3>
                    <p className="text-xs text-neutral-500">Only the creator can edit or delete.</p>
                  </div>
                  <div className="flex items-center gap-2">
                    <button
                      type="button"
                      onClick={() => setIsEditing((prev) => !prev)}
                      className="inline-flex items-center gap-2 rounded-full border border-neutral-200 bg-white px-3 py-2 text-xs font-semibold text-neutral-700 transition hover:border-neutral-400"
                    >
                      <Pencil className="h-3.5 w-3.5" />
                      {isEditing ? "Cancel" : "Edit"}
                    </button>
                    <button
                      type="button"
                      onClick={handleDelete}
                      className="inline-flex items-center gap-2 rounded-full border border-rose-200 bg-rose-50 px-3 py-2 text-xs font-semibold text-rose-700 transition hover:border-rose-300"
                    >
                      <Trash2 className="h-3.5 w-3.5" />
                      Delete
                    </button>
                  </div>
                </div>

                {isEditing ? (
                  <div className="mt-4 grid gap-3">
                    <input
                      value={editTitle}
                      onChange={(event) => setEditTitle(event.target.value)}
                      className="h-10 w-full rounded-2xl border border-neutral-200 bg-white px-3 text-xs outline-none focus:border-neutral-400"
                    />
                    <textarea
                      value={editDescription}
                      onChange={(event) => setEditDescription(event.target.value)}
                      rows={3}
                      className="w-full resize-none rounded-2xl border border-neutral-200 bg-white px-3 py-2 text-xs outline-none focus:border-neutral-400"
                    />
                    <input
                      type="datetime-local"
                      value={editTime}
                      onChange={(event) => setEditTime(event.target.value)}
                      className="h-10 w-full rounded-2xl border border-neutral-200 bg-white px-3 text-xs outline-none focus:border-neutral-400"
                    />
                    <button
                      type="button"
                      onClick={handleSave}
                      disabled={isSaving}
                      className="inline-flex w-fit items-center gap-2 rounded-full bg-neutral-900 px-4 py-2 text-xs font-semibold text-white transition hover:bg-neutral-800 disabled:cursor-not-allowed disabled:opacity-60"
                    >
                      <Check className="h-3.5 w-3.5" />
                      Save changes
                    </button>
                    {editError ? <p className="text-xs text-rose-600">{editError}</p> : null}
                  </div>
                ) : null}
                {deleteError ? <p className="mt-2 text-xs text-rose-600">{deleteError}</p> : null}
              </motion.div>
            ) : null}

            <motion.div
              initial="hidden"
              whileInView="show"
              viewport={viewportOnce}
              variants={fadeUp}
              className="rounded-3xl border border-neutral-200 bg-white p-6 shadow-sm"
            >
              <div className="flex items-center gap-2 text-sm font-semibold text-neutral-900">
                <Calendar className="h-4 w-4" />
                Responses
              </div>
              {responsesError ? (
                <p className="mt-2 text-xs text-rose-600">{responsesError}</p>
              ) : responses.length === 0 ? (
                <p className="mt-3 text-xs text-neutral-500">No responses yet.</p>
              ) : (
                <div className="mt-3 space-y-2">
                  {responses.map((resp) => (
                    <div
                      key={`${resp.user_id}-${resp.response}`}
                      className="flex items-center justify-between gap-3 rounded-2xl border border-neutral-200 bg-neutral-50 px-3 py-2"
                    >
                      <div className="flex items-center gap-3">
                        <Avatar
                          src={resp.avatar_path ? toMediaUrl(apiBaseUrl, resp.avatar_path) : null}
                          name={`${resp.first_name} ${resp.last_name}`}
                          size={36}
                          textClassName="text-[11px]"
                        />
                        <div>
                          <p className="text-xs font-semibold text-neutral-800">
                            {resp.first_name} {resp.last_name}
                          </p>
                          <p className="text-[11px] text-neutral-500">
                            @{resp.nickname || "user"}
                          </p>
                        </div>
                      </div>
                      <span
                        className={`rounded-full px-3 py-1 text-[11px] font-semibold ${
                          resp.response === "going"
                            ? "bg-emerald-100 text-emerald-700"
                            : "bg-neutral-200 text-neutral-700"
                        }`}
                      >
                        {resp.response === "going" ? "Going" : "Not going"}
                      </span>
                    </div>
                  ))}
                </div>
              )}
            </motion.div>
          </div>
        ) : null}
        </section>
      </main>
    </div>
  );
}
