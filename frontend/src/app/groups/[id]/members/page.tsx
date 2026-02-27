"use client";

import { useEffect, useMemo, useState } from "react";
import Link from "next/link";
import { useParams, useRouter } from "next/navigation";
import { ArrowLeft } from "lucide-react";
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
  avatar_path?: string | null;
};

type GroupMember = {
  id: number;
  first_name: string;
  last_name: string;
  nickname?: string | null;
  avatar_path?: string | null;
};

type UserSearchItem = {
  id: number;
  first_name: string;
  last_name: string;
  nickname?: string | null;
  avatar_path?: string | null;
};

type SentInvite = {
  id: number;
  invited_at: string;
};

function toMediaUrl(apiBaseUrl: string, path?: string | null) {
  if (!path) return "";
  if (path.startsWith("http://") || path.startsWith("https://")) return path;
  const normalized = path.startsWith("/") ? path : `/${path}`;
  return `${apiBaseUrl}${normalized}`;
}

async function fetchProfileSummary(
  apiBaseUrl: string,
  id: number,
): Promise<UserSearchItem | null> {
  try {
    const response = await fetch(`${apiBaseUrl}/profiles/${id}`, { credentials: "include" });
    const result = (await response.json().catch(() => null)) as
      | ApiResponse<{ user?: UserSearchItem }>
      | null;
    if (!response.ok || !result?.success || !result.data?.user) {
      return null;
    }
    return result.data.user;
  } catch {
    return null;
  }
}

export default function GroupMembersPage() {
  const router = useRouter();
  const params = useParams<{ id: string }>();
  const groupID = typeof params?.id === "string" ? Number(params.id) : NaN;

  const [viewer, setViewer] = useState<User | null>(null);
  const [members, setMembers] = useState<GroupMember[]>([]);
  const [error, setError] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [inviteQuery, setInviteQuery] = useState("");
  const [inviteResults, setInviteResults] = useState<UserSearchItem[]>([]);
  const [inviteLoading, setInviteLoading] = useState(false);
  const [selectedInvitee, setSelectedInvitee] = useState<UserSearchItem | null>(null);
  const [sentInvites, setSentInvites] = useState<SentInvite[]>([]);
  const [sentProfiles, setSentProfiles] = useState<Record<number, UserSearchItem>>({});
  const [inviteError, setInviteError] = useState<string | null>(null);
  const [inviteSuccess, setInviteSuccess] = useState<string | null>(null);
  const [leaveError, setLeaveError] = useState<string | null>(null);

  const apiBaseUrl = useMemo(
    () =>
      process.env.NEXT_PUBLIC_API_BASE_URL?.trim().replace(/\/+$/, "") ||
      "http://localhost:8080",
    [],
  );

  useEffect(() => {
    if (!Number.isFinite(groupID) || groupID <= 0) {
      setError("Invalid group id.");
      setIsLoading(false);
      return;
    }

    let cancelled = false;
    const load = async () => {
      setIsLoading(true);
      setError(null);
      try {
        const meResponse = await fetch(`${apiBaseUrl}/auth/me`, { credentials: "include" });
        const meResult = (await meResponse.json().catch(() => null)) as ApiResponse<User> | null;
        if (!meResponse.ok || !meResult?.success || !meResult.data) {
          router.replace("/login");
          return;
        }
        if (!cancelled) {
          setViewer(meResult.data);
        }

        const membersResponse = await fetch(`${apiBaseUrl}/groups/${groupID}/members`, {
          credentials: "include",
        });
        const membersResult = (await membersResponse.json().catch(() => null)) as
          | ApiResponse<GroupMember[]>
          | null;
        if (!membersResponse.ok || !membersResult?.success) {
          setError(membersResult?.error || "Could not load group members.");
          setMembers([]);
          return;
        }
        setMembers(membersResult.data ?? []);
      } catch {
        setError("Network error. Please try again.");
      } finally {
        setIsLoading(false);
      }
    };
    void load();
    return () => {
      cancelled = true;
    };
  }, [apiBaseUrl, groupID, router]);

  useEffect(() => {
    if (!Number.isFinite(groupID) || groupID <= 0) return;
    const key = `group-sent-invites:${groupID}`;
    const raw = localStorage.getItem(key);
    if (!raw) return;
    try {
      const parsed = JSON.parse(raw) as SentInvite[];
      if (Array.isArray(parsed)) {
        setSentInvites(parsed.filter((item) => Number.isFinite(item.id)));
      }
    } catch {
      // ignore
    }
  }, [groupID]);

  useEffect(() => {
    if (!inviteQuery.trim()) {
      setInviteResults([]);
      setInviteLoading(false);
      return;
    }

    let cancelled = false;
    const controller = new AbortController();
    const timeoutID = window.setTimeout(async () => {
      setInviteLoading(true);
      try {
        const response = await fetch(
          `${apiBaseUrl}/users?q=${encodeURIComponent(inviteQuery.trim())}&limit=6&offset=0`,
          { credentials: "include", signal: controller.signal },
        );
        const result = (await response.json().catch(() => null)) as
          | ApiResponse<UserSearchItem[]>
          | null;
        if (!cancelled && response.ok && result?.success) {
          setInviteResults(result.data ?? []);
        } else if (!cancelled) {
          setInviteResults([]);
        }
      } catch {
        if (!cancelled) {
          setInviteResults([]);
        }
      } finally {
        if (!cancelled) {
          setInviteLoading(false);
        }
      }
    }, 400);

    return () => {
      cancelled = true;
      window.clearTimeout(timeoutID);
      controller.abort();
    };
  }, [apiBaseUrl, inviteQuery]);

  useEffect(() => {
    if (sentInvites.length === 0) return;
    const missing = sentInvites
      .map((item) => item.id)
      .filter((id) => !sentProfiles[id]);
    if (missing.length === 0) return;

    let cancelled = false;
    Promise.all(missing.map((id) => fetchProfileSummary(apiBaseUrl, id).then((user) => [id, user] as const)))
      .then((entries) => {
        if (cancelled) return;
        const updates: Record<number, UserSearchItem> = {};
        for (const [id, user] of entries) {
          if (user) {
            updates[id] = user;
          }
        }
        if (Object.keys(updates).length > 0) {
          setSentProfiles((prev) => ({ ...prev, ...updates }));
        }
      })
      .catch(() => undefined);

    return () => {
      cancelled = true;
    };
  }, [apiBaseUrl, sentInvites, sentProfiles]);

  const handleInvite = async () => {
    setInviteError(null);
    setInviteSuccess(null);
    if (!selectedInvitee) {
      setInviteError("Pick a user to invite.");
      return;
    }
    try {
      const response = await fetch(`${apiBaseUrl}/groups/${groupID}/invitations`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        credentials: "include",
        body: JSON.stringify({ invitee_id: selectedInvitee.id }),
      });
      const result = (await response.json().catch(() => null)) as ApiResponse<unknown> | null;
      if (!response.ok || !result?.success) {
        setInviteError(result?.error || "Could not send invitation.");
        return;
      }
      setInviteSuccess("Invitation sent.");
      const nextInvite: SentInvite = {
        id: selectedInvitee.id,
        invited_at: new Date().toISOString(),
      };
      setSentInvites((prev) => {
        if (prev.some((item) => item.id === nextInvite.id)) return prev;
        const updated = [nextInvite, ...prev];
        localStorage.setItem(`group-sent-invites:${groupID}`, JSON.stringify(updated));
        return updated;
      });
      setSelectedInvitee(null);
      setInviteQuery("");
      setInviteResults([]);
    } catch {
      setInviteError("Network error. Please try again.");
    }
  };

  const handleLeave = async () => {
    setLeaveError(null);
    try {
      const response = await fetch(`${apiBaseUrl}/groups/${groupID}/members/me`, {
        method: "DELETE",
        credentials: "include",
      });
      if (!response.ok) {
        const result = (await response.json().catch(() => null)) as ApiResponse<unknown> | null;
        setLeaveError(result?.error || "Could not leave group.");
        return;
      }
      router.replace("/groups");
    } catch {
      setLeaveError("Network error. Please try again.");
    }
  };

  return (
    <div className="min-h-screen bg-neutral-50 text-neutral-900">
      <TopNav user={viewer ?? undefined} onLogout={() => router.replace("/login")} />

      <main className="mx-auto grid w-full max-w-6xl gap-6 px-4 py-6 sm:px-6 lg:grid-cols-[240px_minmax(0,1fr)]">
        <aside className="hidden lg:block">
          <LeftNav user={viewer ?? undefined} activeHref="/groups" />
        </aside>

        <section className="space-y-5">
          <motion.div
            initial="hidden"
            whileInView="show"
            viewport={viewportOnce}
            variants={fadeUp}
            className="rounded-3xl border border-neutral-200 bg-white p-5 shadow-sm"
          >
            <div className="flex flex-wrap items-center justify-between gap-3">
              <div>
                <h1 className="text-xl font-semibold tracking-tight text-neutral-900">
                  Group members
                </h1>
                <p className="text-sm text-neutral-600">
                  Manage members and invite new people.
                </p>
              </div>
              <Link
                href={`/groups/${groupID}`}
                className="inline-flex items-center gap-2 rounded-full border border-neutral-200 bg-white px-3 py-2 text-xs font-semibold text-neutral-700 transition hover:border-neutral-400 hover:text-neutral-900"
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
            className="rounded-3xl border border-neutral-200 bg-white p-5 shadow-sm"
          >
            <h2 className="text-sm font-semibold text-neutral-900">Invite someone</h2>
            <div className="mt-3 space-y-3">
              <div className="relative">
                <input
                  value={inviteQuery}
                  onChange={(event) => {
                    setInviteQuery(event.target.value);
                    if (selectedInvitee) {
                      setSelectedInvitee(null);
                    }
                  }}
                  placeholder="Search people by name or nickname"
                  className="h-10 w-full rounded-2xl border border-neutral-200 bg-white px-3 text-xs outline-none focus:border-neutral-400"
                />
                {inviteQuery.trim() ? (
                  <div className="absolute z-20 mt-2 w-full rounded-2xl border border-neutral-200 bg-white p-2 shadow-xl">
                    {inviteLoading ? (
                      <p className="px-2 py-2 text-xs text-neutral-500">Searching...</p>
                    ) : inviteResults.length === 0 ? (
                      <p className="px-2 py-2 text-xs text-neutral-500">No users found.</p>
                    ) : (
                      <div className="space-y-2">
                        {inviteResults.map((person) => (
                          <button
                            key={person.id}
                            type="button"
                            onClick={() => {
                              setSelectedInvitee(person);
                              setInviteQuery("");
                              setInviteResults([]);
                            }}
                            className="flex w-full items-center gap-3 rounded-2xl border border-neutral-200 bg-neutral-50 px-3 py-2 text-left text-xs text-neutral-700 transition hover:border-neutral-400 hover:bg-white"
                          >
                            <Avatar
                              src={
                                person.avatar_path
                                  ? toMediaUrl(apiBaseUrl, person.avatar_path)
                                  : null
                              }
                              name={`${person.first_name} ${person.last_name}`}
                              size={32}
                              textClassName="text-[10px]"
                            />
                            <div>
                              <p className="text-xs font-semibold text-neutral-900">
                                {person.first_name} {person.last_name}
                              </p>
                              <p className="text-[11px] text-neutral-500">
                                @{person.nickname || "user"}
                              </p>
                            </div>
                          </button>
                        ))}
                      </div>
                    )}
                  </div>
                ) : null}
              </div>
              {selectedInvitee ? (
                <div className="flex flex-wrap items-center justify-between gap-2 rounded-2xl border border-neutral-200 bg-neutral-50 px-3 py-2 text-xs text-neutral-700">
                  <div className="flex items-center gap-2">
                    <span className="font-semibold">
                      {selectedInvitee.first_name} {selectedInvitee.last_name}
                    </span>
                    <span className="text-[11px] text-neutral-500">
                      @{selectedInvitee.nickname || "user"}
                    </span>
                  </div>
                  <button
                    type="button"
                    onClick={() => setSelectedInvitee(null)}
                    className="rounded-full border border-neutral-200 bg-white px-3 py-1 text-[11px] font-semibold text-neutral-600 transition hover:border-neutral-400 hover:text-neutral-900"
                  >
                    Clear
                  </button>
                </div>
              ) : null}
              <button
                type="button"
                onClick={handleInvite}
                disabled={!selectedInvitee}
                className="rounded-full bg-neutral-900 px-4 py-2 text-xs font-semibold text-white transition hover:bg-neutral-800 disabled:cursor-not-allowed disabled:opacity-60"
              >
                Send invite
              </button>
            </div>
            {inviteError ? <p className="mt-2 text-xs text-rose-600">{inviteError}</p> : null}
            {inviteSuccess ? (
              <p className="mt-2 text-xs text-emerald-600">{inviteSuccess}</p>
            ) : null}
          </motion.div>

          <motion.div
            initial="hidden"
            whileInView="show"
            viewport={viewportOnce}
            variants={fadeUp}
            className="rounded-3xl border border-neutral-200 bg-white p-5 shadow-sm"
          >
            <div className="flex items-center justify-between gap-3">
              <div>
                <h2 className="text-sm font-semibold text-neutral-900">Sent invitations</h2>
                <p className="text-xs text-neutral-500">
                  Stored locally for this browser.
                </p>
              </div>
              <Link
                href="/group-invitations"
                className="rounded-full border border-neutral-200 bg-white px-3 py-1.5 text-[11px] font-semibold text-neutral-600 transition hover:border-neutral-400 hover:text-neutral-900"
              >
                Incoming invites
              </Link>
            </div>

            <div className="mt-4 space-y-2">
              {sentInvites.length === 0 ? (
                <p className="text-xs text-neutral-500">No sent invitations yet.</p>
              ) : (
                sentInvites.map((invite) => {
                  const profile = sentProfiles[invite.id];
                  return (
                    <div
                      key={invite.id}
                      className="flex items-center justify-between gap-3 rounded-2xl border border-neutral-200 bg-neutral-50 px-3 py-2 text-xs text-neutral-700"
                    >
                      <div>
                        <p className="font-semibold">
                          {profile
                            ? `${profile.first_name} ${profile.last_name}`
                            : "User"}
                        </p>
                        <p className="text-[11px] text-neutral-500">
                          {profile?.nickname ? `@${profile.nickname}` : "Pending"}
                        </p>
                      </div>
                      <span className="rounded-full border border-neutral-200 bg-white px-2 py-1 text-[10px] text-neutral-500">
                        {new Date(invite.invited_at).toLocaleDateString(undefined, {
                          month: "short",
                          day: "numeric",
                        })}
                      </span>
                    </div>
                  );
                })
              )}
            </div>
          </motion.div>

          {isLoading ? (
            <motion.div
              initial="hidden"
              whileInView="show"
              viewport={viewportOnce}
              variants={fadeUp}
              className="rounded-3xl border border-neutral-200 bg-white p-5 text-sm text-neutral-600 shadow-sm"
            >
              Loading members...
            </motion.div>
          ) : error ? (
            <motion.div
              initial="hidden"
              whileInView="show"
              viewport={viewportOnce}
              variants={fadeUp}
              className="rounded-3xl border border-rose-200 bg-rose-50 p-5 text-sm text-rose-700"
            >
              {error}
            </motion.div>
          ) : (
            <motion.div
              initial="hidden"
              whileInView="show"
              viewport={viewportOnce}
              variants={fadeUp}
              className="rounded-3xl border border-neutral-200 bg-white p-5 shadow-sm"
            >
              <div className="space-y-3">
                {members.map((member) => (
                  <div
                    key={member.id}
                    className="flex items-center gap-3 rounded-2xl border border-neutral-200 bg-neutral-50 px-4 py-3"
                  >
                    <Avatar
                      src={
                        member.avatar_path ? toMediaUrl(apiBaseUrl, member.avatar_path) : null
                      }
                      name={`${member.first_name} ${member.last_name}`}
                      size={40}
                      textClassName="text-xs"
                    />
                    <div>
                      <p className="text-sm font-semibold text-neutral-900">
                        {member.first_name} {member.last_name}
                      </p>
                      <p className="text-xs text-neutral-500">
                        @{member.nickname || "user"}
                      </p>
                    </div>
                  </div>
                ))}
              </div>
            </motion.div>
          )}

          <motion.div
            initial="hidden"
            whileInView="show"
            viewport={viewportOnce}
            variants={fadeUp}
            className="rounded-3xl border border-neutral-200 bg-white p-5 shadow-sm"
          >
            <div className="flex items-center justify-between gap-3">
              <div>
                <h2 className="text-sm font-semibold text-neutral-900">Leave group</h2>
                <p className="text-xs text-neutral-500">
                  Group creators cannot leave their own group.
                </p>
              </div>
              <button
                type="button"
                onClick={handleLeave}
                className="rounded-full border border-neutral-300 bg-white px-3 py-2 text-xs font-semibold text-neutral-700 transition hover:border-neutral-400"
              >
                Leave group
              </button>
            </div>
            {leaveError ? <p className="mt-2 text-xs text-rose-600">{leaveError}</p> : null}
          </motion.div>
        </section>
      </main>
    </div>
  );
}
