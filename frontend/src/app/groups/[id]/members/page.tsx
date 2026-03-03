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
import { toMediaUrl } from "@/lib/media";
import { apiFetch, apiFetchJson, getApiBaseUrl } from "@/lib/api";
import { ApiResponse } from "@/lib/types";

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
  first_name?: string;
  last_name?: string;
  nickname?: string | null;
  avatar_path?: string | null;
};

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
  const [inviteError, setInviteError] = useState<string | null>(null);
  const [inviteSuccess, setInviteSuccess] = useState<string | null>(null);
  const [leaveError, setLeaveError] = useState<string | null>(null);

  const apiBaseUrl = useMemo(() => getApiBaseUrl(), []);

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
        const { response: meResponse, result: meResult } = await apiFetchJson<ApiResponse<User>>(
          "/auth/me",
          {},
          apiBaseUrl,
        );
        if (!meResponse.ok || !meResult?.success || !meResult.data) {
          router.replace("/login");
          return;
        }
        if (!cancelled) {
          setViewer(meResult.data);
        }

        const { response: membersResponse, result: membersResult } = await apiFetchJson<
          ApiResponse<GroupMember[]>
        >(`/groups/${groupID}/members`, {}, apiBaseUrl);
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
        const { response, result } = await apiFetchJson<ApiResponse<UserSearchItem[]>>(
          `/users?q=${encodeURIComponent(inviteQuery.trim())}&limit=6&offset=0`,
          { signal: controller.signal },
          apiBaseUrl,
        );
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

  const handleInvite = async () => {
    setInviteError(null);
    setInviteSuccess(null);
    if (!selectedInvitee) {
      setInviteError("Pick a user to invite.");
      return;
    }
    try {
      const { response, result } = await apiFetchJson<ApiResponse<unknown>>(
        `/groups/${groupID}/invitations`,
        {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ invitee_id: selectedInvitee.id }),
        },
        apiBaseUrl,
      );
      if (!response.ok || !result?.success) {
        setInviteError(result?.error || "Could not send invitation.");
        return;
      }
      setInviteSuccess("Invitation sent.");
      const nextInvite: SentInvite = {
        id: selectedInvitee.id,
        invited_at: new Date().toISOString(),
        first_name: selectedInvitee.first_name,
        last_name: selectedInvitee.last_name,
        nickname: selectedInvitee.nickname,
        avatar_path: selectedInvitee.avatar_path,
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
      const response = await apiFetch(`/groups/${groupID}/members/me`, { method: "DELETE" }, apiBaseUrl);
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
                <h1 className="text-xl font-semibold tracking-tight text-white">
                  Group members
                </h1>
                <p className="text-sm text-neutral-400">
                  Manage members and invite new people.
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
            <h2 className="text-sm font-semibold text-white">Invite someone</h2>
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
                  className="h-10 w-full rounded-2xl border border-white/20 bg-white/5 px-3 text-xs text-white outline-none focus:border-white/40"
                />
                {inviteQuery.trim() ? (
                  <div className="absolute z-20 mt-2 w-full rounded-2xl border border-white/20 bg-white/5 p-2 shadow-xl">
                    {inviteLoading ? (
                      <p className="px-2 py-2 text-xs text-neutral-400">Searching...</p>
                    ) : inviteResults.length === 0 ? (
                      <p className="px-2 py-2 text-xs text-neutral-400">No users found.</p>
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
                            className="flex w-full items-center gap-3 rounded-2xl border border-white/10 bg-white/5 px-3 py-2 text-left text-xs text-neutral-300 transition hover:bg-white/10 hover:text-white"
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
                              <p className="text-xs font-semibold text-white">
                                {person.first_name} {person.last_name}
                              </p>
                              <p className="text-[11px] text-neutral-400">
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
                <div className="flex flex-wrap items-center justify-between gap-2 rounded-2xl border border-white/10 bg-white/5 px-3 py-2 text-xs text-neutral-300">
                  <div className="flex items-center gap-2">
                    <span className="font-semibold">
                      {selectedInvitee.first_name} {selectedInvitee.last_name}
                    </span>
                    <span className="text-[11px] text-neutral-400">
                      @{selectedInvitee.nickname || "user"}
                    </span>
                  </div>
                  <button
                    type="button"
                    onClick={() => setSelectedInvitee(null)}
                    className="rounded-full border border-white/20 bg-white/5 px-3 py-1 text-[11px] font-semibold text-neutral-300 transition hover:bg-white/10 hover:text-white"
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
            {inviteError ? <p className="mt-2 text-xs text-rose-400">{inviteError}</p> : null}
            {inviteSuccess ? (
              <p className="mt-2 text-xs text-emerald-400">{inviteSuccess}</p>
            ) : null}
          </motion.div>

          <motion.div
            initial="hidden"
            whileInView="show"
            viewport={viewportOnce}
            variants={fadeUp}
            className="rounded-3xl border border-white/10 bg-white/5 p-5 backdrop-blur-sm"
          >
            <div className="flex items-center justify-between gap-3">
              <div>
                <h2 className="text-sm font-semibold text-white">Sent invitations</h2>
                <p className="text-xs text-neutral-400">
                  Stored locally for this browser.
                </p>
              </div>
              <Link
                href="/group-invitations"
                className="rounded-full border border-white/20 bg-white/5 px-3 py-1.5 text-[11px] font-semibold text-neutral-300 transition hover:bg-white/10 hover:text-white"
              >
                Incoming invites
              </Link>
            </div>

            <div className="mt-4 space-y-2">
              {sentInvites.length === 0 ? (
                <p className="text-xs text-neutral-400">No sent invitations yet.</p>
              ) : (
                sentInvites.map((invite) => {
                  const profile = invite;
                  return (
                    <div
                      key={invite.id}
                      className="flex items-center justify-between gap-3 rounded-2xl border border-white/10 bg-white/5 px-3 py-2 text-xs text-neutral-300"
                    >
                      <div>
                        <p className="font-semibold">
                          {profile
                            ? `${profile.first_name} ${profile.last_name}`
                            : "User"}
                        </p>
                        <p className="text-[11px] text-neutral-400">
                          {profile?.nickname ? `@${profile.nickname}` : "Pending"}
                        </p>
                      </div>
                      <span className="rounded-full border border-white/20 bg-white/5 px-2 py-1 text-[10px] text-neutral-400">
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
              className="rounded-3xl border border-neutral-200 bg-white p-5 text-sm text-neutral-400 shadow-sm"
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
              className="rounded-3xl border border-white/10 bg-white/5 p-5 backdrop-blur-sm"
            >
              <div className="space-y-3">
                {members.map((member) => (
                  <div
                    key={member.id}
                    className="flex items-center gap-3 rounded-2xl border border-white/10 bg-white/5 px-4 py-3"
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
                      <p className="text-sm font-semibold text-white">
                        {member.first_name} {member.last_name}
                      </p>
                      <p className="text-xs text-neutral-400">
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
            className="rounded-3xl border border-white/10 bg-white/5 p-5 backdrop-blur-sm"
          >
            <div className="flex items-center justify-between gap-3">
              <div>
                <h2 className="text-sm font-semibold text-white">Leave group</h2>
                <p className="text-xs text-neutral-400">
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
            {leaveError ? <p className="mt-2 text-xs text-rose-400">{leaveError}</p> : null}
          </motion.div>
        </section>
      </main>
    </div>
  );
}
