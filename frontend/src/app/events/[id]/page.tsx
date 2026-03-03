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
import { toMediaUrl } from "@/lib/media";
import { apiFetch, apiFetchJson, getApiBaseUrl } from "@/lib/api";
import { ApiResponse } from "@/lib/types";

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
  group_title?: string | null;
  creator_id: number;
  title: string;
  description?: string | null;
  event_time: string;
  created_at: string;
  updated_at: string;
  responses_count?: number;
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
  const [responsesLoaded, setResponsesLoaded] = useState(false);
  const [isEditing, setIsEditing] = useState(false);
  const [editTitle, setEditTitle] = useState("");
  const [editDescription, setEditDescription] = useState("");
  const [editTime, setEditTime] = useState("");
  const [editError, setEditError] = useState<string | null>(null);
  const [isSaving, setIsSaving] = useState(false);
  const [deleteError, setDeleteError] = useState<string | null>(null);

  const apiBaseUrl = useMemo(() => getApiBaseUrl(), []);

  const loadEvent = useCallback(async () => {
    if (!Number.isFinite(eventID) || eventID <= 0) {
      setError("Invalid event id.");
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

      const { response: eventResponse, result: eventResult } = await apiFetchJson<
        ApiResponse<EventItem>
      >(`/events/${eventID}`, {}, apiBaseUrl);
      if (!eventResponse.ok || !eventResult?.success || !eventResult.data) {
        setError(eventResult?.error || "Could not load event.");
        setEvent(null);
        return;
      }
      setEvent(eventResult.data);
      setEditTitle(eventResult.data.title);
      setEditDescription(eventResult.data.description || "");
      setEditTime(toInputDate(eventResult.data.event_time));

      setResponses([]);
      setResponsesError(null);
      setResponsesLoaded(false);
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
      const { response: res, result: json } = await apiFetchJson<ApiResponse<unknown>>(
        `/events/${eventID}/responses`,
        {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ response }),
        },
        apiBaseUrl,
      );
      if (!res.ok || !json?.success) {
        setResponseAction(json?.error || "Could not submit response.");
        return;
      }
      await loadEvent();
      if (responsesLoaded) {
        await loadResponses();
      }
    } catch {
      setResponseAction("Network error. Please try again.");
    }
  };

  const loadResponses = useCallback(async () => {
    setResponsesError(null);
    try {
      const { response: responsesResponse, result: responsesResult } = await apiFetchJson<
        ApiResponse<EventResponse[]>
      >(`/events/${eventID}/responses`, {}, apiBaseUrl);
      if (!responsesResponse.ok || !responsesResult?.success) {
        setResponsesError(responsesResult?.error || "Could not load responses.");
        setResponses([]);
        return;
      }
      setResponses(responsesResult.data ?? []);
      setResponsesLoaded(true);
    } catch {
      setResponsesError("Network error. Please try again.");
    }
  }, [apiBaseUrl, eventID]);

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
      const { response, result } = await apiFetchJson<ApiResponse<EventItem>>(
        `/events/${event.id}`,
        {
          method: "PATCH",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify(payload),
        },
        apiBaseUrl,
      );
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
      const response = await apiFetch(`/events/${event.id}`, { method: "DELETE" }, apiBaseUrl);
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

  const groupName = event?.group_title?.trim();
  const responsesCount =
    typeof event?.responses_count === "number" ? event.responses_count : responses.length;

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

      <main className="mx-auto grid w-full max-w-6xl gap-6 px-4 py-6 sm:px-6 md:grid-cols-[1fr_280px] lg:grid-cols-[240px_minmax(0,1fr)_280px]">
        <aside className="hidden lg:block">
          <LeftNav user={viewer ?? undefined} activeHref="/groups" variant="dark" />
        </aside>

        <section>
          <motion.section
            initial="hidden"
            whileInView="show"
            viewport={viewportOnce}
            variants={fadeUp}
            className="rounded-3xl border border-white/10 bg-white/5 p-6 backdrop-blur-sm"
          >
            <div className="flex flex-wrap items-center justify-between gap-3">
              <div>
                <h1 className="text-xl font-semibold tracking-tight text-white">Event details</h1>
                <p className="text-sm text-neutral-400">
                  View event info and RSVP {groupName ? `· ${groupName}` : ""}.
                </p>
              </div>
              {event ? (
                <Link
                  href={`/groups/${event.group_id}/events`}
                  className="inline-flex items-center gap-2 rounded-full border border-white/20 bg-white/5 px-3 py-2 text-xs font-semibold text-neutral-300 transition hover:bg-white/10 hover:text-white"
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
            className="mt-5 rounded-3xl border border-white/10 bg-white/5 p-6 text-sm text-neutral-400 backdrop-blur-sm"
          >
            Loading event...
          </motion.div>
        ) : error ? (
          <motion.div
            initial="hidden"
            whileInView="show"
            viewport={viewportOnce}
            variants={fadeUp}
            className="mt-5 rounded-3xl border border-rose-500/30 bg-rose-500/10 p-6 text-sm text-rose-400"
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
              className="rounded-3xl border border-white/10 bg-white/5 p-6 backdrop-blur-sm"
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
                  className="inline-flex items-center gap-2 rounded-full border border-white/20 bg-white/5 px-3 py-2 text-xs font-semibold text-neutral-300 transition hover:bg-white/10 hover:text-white"
                >
                  <ThumbsDown className="h-3.5 w-3.5" />
                  Not going
                </button>
                {responseAction ? (
                  <span className="text-xs text-rose-400">{responseAction}</span>
                ) : null}
              </div>
            </motion.div>

            {isCreator ? (
              <motion.div
                initial="hidden"
                whileInView="show"
                viewport={viewportOnce}
                variants={fadeUp}
                className="rounded-3xl border border-white/10 bg-white/5 p-6 backdrop-blur-sm"
              >
                <div className="flex flex-wrap items-center justify-between gap-3">
                  <div>
                    <h3 className="text-sm font-semibold text-white">Manage event</h3>
                    <p className="text-xs text-neutral-400">Only the creator can edit or delete.</p>
                  </div>
                  <div className="flex items-center gap-2">
                    <button
                      type="button"
                      onClick={() => setIsEditing((prev) => !prev)}
                      className="inline-flex items-center gap-2 rounded-full border border-white/20 bg-white/5 px-3 py-2 text-xs font-semibold text-neutral-300 transition hover:bg-white/10 hover:text-white"
                    >
                      <Pencil className="h-3.5 w-3.5" />
                      {isEditing ? "Cancel" : "Edit"}
                    </button>
                    <button
                      type="button"
                      onClick={handleDelete}
                      className="inline-flex items-center gap-2 rounded-full border border-rose-500/30 bg-rose-500/10 px-3 py-2 text-xs font-semibold text-rose-400 transition hover:bg-rose-500/20"
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
                      className="h-10 w-full rounded-2xl border border-neutral-200 bg-white px-3 text-xs text-black outline-none focus:border-neutral-400 placeholder:text-neutral-500"
                    />
                    <textarea
                      value={editDescription}
                      onChange={(event) => setEditDescription(event.target.value)}
                      rows={3}
                      className="w-full resize-none rounded-2xl border border-neutral-200 bg-white px-3 py-2 text-xs text-black outline-none focus:border-neutral-400 placeholder:text-neutral-500"
                    />
                    <input
                      type="datetime-local"
                      value={editTime}
                      onChange={(event) => setEditTime(event.target.value)}
                      className="h-10 w-full rounded-2xl border border-neutral-200 bg-white px-3 text-xs text-black outline-none focus:border-neutral-400 placeholder:text-neutral-500"
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
                    {editError ? <p className="text-xs text-rose-400">{editError}</p> : null}
                  </div>
                ) : null}
                {deleteError ? <p className="mt-2 text-xs text-rose-400">{deleteError}</p> : null}
              </motion.div>
            ) : null}

            {/* Responses - mobile only (md+ shows in right sidebar) */}
            {event ? (
              <motion.div
                initial="hidden"
                whileInView="show"
                viewport={viewportOnce}
                variants={fadeUp}
                className="mt-5 rounded-3xl border border-white/10 bg-white/5 p-6 backdrop-blur-sm md:hidden"
              >
                <div className="flex items-center gap-2 text-sm font-semibold text-white">
                  <Calendar className="h-4 w-4" />
                  Responses
                </div>
                {!responsesLoaded ? (
                  <button
                    type="button"
                    onClick={() => void loadResponses()}
                    className="mt-3 inline-flex items-center gap-2 rounded-full border border-white/20 bg-white/5 px-3 py-1 text-xs text-white transition hover:bg-white/10"
                  >
                    View responses ({responsesCount})
                  </button>
                ) : responsesError ? (
                  <p className="mt-2 text-xs text-rose-400">{responsesError}</p>
                ) : responses.length === 0 ? (
                  <p className="mt-3 text-xs text-neutral-400">No responses yet.</p>
                ) : (
                  <div className="mt-3 space-y-2">
                    {responses.map((resp) => (
                      <div
                        key={`${resp.user_id}-${resp.response}`}
                        className="flex items-center justify-between gap-3 rounded-2xl border border-white/10 bg-white/5 px-3 py-2"
                      >
                        <div className="flex items-center gap-3">
                          <Avatar
                            src={resp.avatar_path ? toMediaUrl(apiBaseUrl, resp.avatar_path) : null}
                            name={`${resp.first_name} ${resp.last_name}`}
                            size={36}
                            textClassName="text-[11px]"
                          />
                          <div>
                            <p className="text-xs font-semibold text-white">
                              {resp.first_name} {resp.last_name}
                            </p>
                            <p className="text-[11px] text-neutral-400">
                              @{resp.nickname || "user"}
                            </p>
                          </div>
                        </div>
                        <span
                          className={`shrink-0 rounded-full px-3 py-1 text-[11px] font-semibold ${
                            resp.response === "going"
                              ? "bg-emerald-500/20 text-emerald-400"
                              : "bg-white/10 text-neutral-400"
                          }`}
                        >
                          {resp.response === "going" ? "Going" : "Not going"}
                        </span>
                      </div>
                    ))}
                  </div>
                )}
              </motion.div>
            ) : null}
          </div>
        ) : null}
        </section>

        {event && !isLoading && !error ? (
          <aside className="hidden space-y-4 md:block md:col-start-2 lg:col-start-3">
            <motion.div
              initial="hidden"
              whileInView="show"
              viewport={viewportOnce}
              variants={fadeUp}
              className="rounded-3xl border border-white/10 bg-white/5 p-6 backdrop-blur-sm"
            >
              <div className="flex items-center gap-2 text-sm font-semibold text-white">
                <Calendar className="h-4 w-4" />
                Responses
              </div>
              {!responsesLoaded ? (
                <button
                  type="button"
                  onClick={() => void loadResponses()}
                  className="mt-3 inline-flex items-center gap-2 rounded-full border border-white/20 bg-white/5 px-3 py-1 text-xs text-white transition hover:bg-white/10"
                >
                  View responses ({responsesCount})
                </button>
              ) : responsesError ? (
                <p className="mt-2 text-xs text-rose-400">{responsesError}</p>
              ) : responses.length === 0 ? (
                <p className="mt-3 text-xs text-neutral-400">No responses yet.</p>
              ) : (
                <div className="mt-3 space-y-2">
                  {responses.map((resp) => (
                    <div
                      key={`${resp.user_id}-${resp.response}`}
                      className="flex items-center justify-between gap-3 rounded-2xl border border-white/10 bg-white/5 px-3 py-2"
                    >
                      <div className="flex items-center gap-3">
                        <Avatar
                          src={resp.avatar_path ? toMediaUrl(apiBaseUrl, resp.avatar_path) : null}
                          name={`${resp.first_name} ${resp.last_name}`}
                          size={36}
                          textClassName="text-[11px]"
                        />
                        <div>
                          <p className="text-xs font-semibold text-white">
                            {resp.first_name} {resp.last_name}
                          </p>
                          <p className="text-[11px] text-neutral-400">
                            @{resp.nickname || "user"}
                          </p>
                        </div>
                      </div>
                      <span
                        className={`shrink-0 rounded-full px-3 py-1 text-[11px] font-semibold ${
                          resp.response === "going"
                            ? "bg-emerald-500/20 text-emerald-400"
                            : "bg-white/10 text-neutral-400"
                        }`}
                      >
                        {resp.response === "going" ? "Going" : "Not going"}
                      </span>
                    </div>
                  ))}
                </div>
              )}
            </motion.div>
          </aside>
        ) : null}
      </main>
    </div>
  );
}
