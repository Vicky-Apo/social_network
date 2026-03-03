"use client";
/* eslint-disable @next/next/no-img-element */

import { useCallback, useEffect, useMemo, useState } from "react";
import Link from "next/link";
import { useParams, useRouter } from "next/navigation";
import {
  ArrowLeft,
  ArrowRight,
  Calendar,
  MessageCircle,
  RefreshCw,
  Send,
  ThumbsDown,
  ThumbsUp,
  Users,
} from "lucide-react";
import { motion } from "framer-motion";
import TopNav from "@/components/TopNav";
import LeftNav from "@/components/LeftNav";
import Avatar from "@/components/Avatar";
import Pagination from "@/components/Pagination";
import { fadeUp, viewportOnce } from "@/components/Motion";
import { formatApiError } from "@/lib/formatApiError";
import { shortDate } from "@/lib/date";
import { toMediaUrl } from "@/lib/media";
import { apiFetch, apiFetchJson, getApiBaseUrl } from "@/lib/api";
import { ApiResponse } from "@/lib/types";

type GroupDetail = {
  id: number;
  name: string;
  description: string;
  creatorID?: number;
  memberCount: number;
  createdAt?: string;
  updatedAt?: string;
};

type ProfileSummary = {
  id: number;
  first_name: string;
  last_name: string;
  nickname?: string | null;
};

type UserSearchItem = {
  id: number;
  first_name: string;
  last_name: string;
  nickname?: string | null;
  avatar_path?: string | null;
};

type Post = {
  id: number;
  author_id: number;
  author_first_name: string;
  author_last_name: string;
  content: string;
  media_path?: string | null;
  created_at: string;
  comment_count: number;
  like_count: number;
  dislike_count: number;
};

type Comment = {
  id: number;
  post_id: number;
  author_id: number;
  content: string;
  media_path?: string;
  like_count: number;
  dislike_count: number;
  created_at: string;
};

type Reaction = {
  user_id: number;
  reaction: "like" | "dislike";
};

type ReactionKind = "like" | "dislike";
type ReactionMap = Record<number, ReactionKind | null>;

type Viewer = {
  id: number;
  email?: string;
  first_name?: string;
  last_name?: string;
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

function toNumber(value: unknown): number | null {
  const parsed = Number(value);
  return Number.isFinite(parsed) ? parsed : null;
}

function formatDate(value?: string) {
  if (!value) return "N/A";
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return "N/A";
  return date.toLocaleDateString(undefined, { month: "short", day: "numeric", year: "numeric" });
}

function formatDateTime(value?: string) {
  if (!value) return "N/A";
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return "N/A";
  return date.toLocaleString(undefined, { month: "short", day: "numeric", hour: "2-digit", minute: "2-digit" });
}

function parseGroup(data: unknown): GroupDetail | null {
  if (!data || typeof data !== "object") {
    return null;
  }

  const root = data as Record<string, unknown>;
  const source =
    root.group && typeof root.group === "object"
      ? (root.group as Record<string, unknown>)
      : root;

  const id = toNumber(source.id);
  if (!id || id <= 0) {
    return null;
  }

  const nameRaw = source.title ?? source.name;
  const name = typeof nameRaw === "string" && nameRaw.trim() ? nameRaw.trim() : `Group ${id}`;
  const descriptionRaw = source.description ?? source.about;
  const description =
    typeof descriptionRaw === "string" && descriptionRaw.trim()
      ? descriptionRaw.trim()
      : "No group description yet.";
  const creatorID = toNumber(source.creator_id ?? source.creatorID) ?? undefined;
  const memberCount =
    toNumber(source.members_count ?? source.member_count ?? source.membersCount) ?? 0;
  const createdAtRaw = source.created_at ?? source.createdAt;
  const updatedAtRaw = source.updated_at ?? source.updatedAt;

  return {
    id,
    name,
    description,
    creatorID,
    memberCount: Math.max(0, memberCount),
    createdAt: typeof createdAtRaw === "string" ? createdAtRaw : undefined,
    updatedAt: typeof updatedAtRaw === "string" ? updatedAtRaw : undefined,
  };
}

export default function GroupDetailsPage() {
  const router = useRouter();
  const params = useParams<{ id: string }>();
  const groupID = typeof params?.id === "string" ? params.id : "";
  const groupIDNumber = Number(groupID);
  const [group, setGroup] = useState<GroupDetail | null>(null);
  const [viewer, setViewer] = useState<Viewer | null>(null);
  const [creatorProfile, setCreatorProfile] = useState<ProfileSummary | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [userID, setUserID] = useState<number | null>(null);
  const [isMember, setIsMember] = useState(false);
  const [joinStatus, setJoinStatus] = useState<"idle" | "requested" | "error">("idle");
  const [joinError, setJoinError] = useState<string | null>(null);
  const [leaveError, setLeaveError] = useState<string | null>(null);
  const [posts, setPosts] = useState<Post[]>([]);
  const [postsLoading, setPostsLoading] = useState(true);
  const [postsError, setPostsError] = useState<string | null>(null);
  const [totalPosts, setTotalPosts] = useState(0);
  const [currentPage, setCurrentPage] = useState(1);
  const [pageSize] = useState(10);
  const totalPages = Math.max(1, Math.ceil(totalPosts / pageSize));
  const commentLimit = 10;
  const [composerText, setComposerText] = useState("");
  const [mediaUrl, setMediaUrl] = useState("");
  const [composerFile, setComposerFile] = useState<File | null>(null);
  const [composerFileName, setComposerFileName] = useState("");
  const [composerError, setComposerError] = useState<string | null>(null);
  const [isPosting, setIsPosting] = useState(false);
  const [postReactionMap, setPostReactionMap] = useState<ReactionMap>({});
  const [commentReactionMap, setCommentReactionMap] = useState<ReactionMap>({});
  const [commentsByPost, setCommentsByPost] = useState<Record<number, Comment[]>>({});
  const [commentsOpenByPost, setCommentsOpenByPost] = useState<Record<number, boolean>>({});
  const [commentsLoadingByPost, setCommentsLoadingByPost] = useState<Record<number, boolean>>({});
  const [commentDraftByPost, setCommentDraftByPost] = useState<Record<number, string>>({});
  const [commentFileByPost, setCommentFileByPost] = useState<Record<number, File | null>>({});
  const [commentFileNameByPost, setCommentFileNameByPost] = useState<Record<number, string>>({});
  const [commentErrorByPost, setCommentErrorByPost] = useState<Record<number, string>>({});
  const [commentPageByPost, setCommentPageByPost] = useState<Record<number, number>>({});
  const [commentTotalByPost, setCommentTotalByPost] = useState<Record<number, number>>({});
  const [editingPostID, setEditingPostID] = useState<number | null>(null);
  const [editPostText, setEditPostText] = useState("");
  const [editPostFile, setEditPostFile] = useState<File | null>(null);
  const [editPostFileName, setEditPostFileName] = useState("");
  const [editPostClearMedia, setEditPostClearMedia] = useState(false);
  const [editPostError, setEditPostError] = useState<string | null>(null);
  const [editingCommentID, setEditingCommentID] = useState<number | null>(null);
  const [editCommentText, setEditCommentText] = useState("");
  const [editCommentFile, setEditCommentFile] = useState<File | null>(null);
  const [editCommentFileName, setEditCommentFileName] = useState("");
  const [editCommentClearMedia, setEditCommentClearMedia] = useState(false);
  const [editCommentError, setEditCommentError] = useState<string | null>(null);
  const [inviteQuery, setInviteQuery] = useState("");
  const [inviteResults, setInviteResults] = useState<UserSearchItem[]>([]);
  const [inviteLoading, setInviteLoading] = useState(false);
  const [selectedInvitee, setSelectedInvitee] = useState<UserSearchItem | null>(null);
  const [inviteError, setInviteError] = useState<string | null>(null);
  const [inviteSuccess, setInviteSuccess] = useState<string | null>(null);
  const [activeTab, setActiveTab] = useState<"posts" | "events" | "members">("posts");
  const [members, setMembers] = useState<GroupMember[]>([]);
  const [membersLoading, setMembersLoading] = useState(false);
  const [membersError, setMembersError] = useState<string | null>(null);
  const [events, setEvents] = useState<EventItem[]>([]);
  const [eventsLoading, setEventsLoading] = useState(false);
  const [eventsError, setEventsError] = useState<string | null>(null);
  const [eventsHasMore, setEventsHasMore] = useState(true);
  const [eventsLoadingMore, setEventsLoadingMore] = useState(false);
  const [eventTitle, setEventTitle] = useState("");
  const [eventDescription, setEventDescription] = useState("");
  const [eventTime, setEventTime] = useState("");
  const [eventCreateError, setEventCreateError] = useState<string | null>(null);
  const [eventCreating, setEventCreating] = useState(false);

  const apiBaseUrl = useMemo(() => getApiBaseUrl(), []);
  const fetchJson = useCallback(
    async <T,>(path: string, options: RequestInit = {}) =>
      apiFetchJson<ApiResponse<T>>(path, options, apiBaseUrl),
    [apiBaseUrl],
  );

  useEffect(() => {
    setCurrentPage(1);
  }, [groupIDNumber]);

  useEffect(() => {
    if (!Number.isFinite(groupIDNumber) || groupIDNumber <= 0) {
      setError("Invalid group id.");
      setIsLoading(false);
      setPostsLoading(false);
      return;
    }

    let cancelled = false;
    const load = async () => {
      setIsLoading(true);
      setError(null);
      setPostsLoading(true);
      setPostsError(null);

      try {
        const { response: meResponse, result: meResult } = await fetchJson<unknown>("/auth/me");
        if (!meResponse.ok || !meResult?.success) {
          if (!cancelled) {
            router.replace("/login");
          }
          return;
        }
        const meUser = meResult.data as Viewer | null;
        if (!cancelled) {
          setViewer(meUser);
          setUserID(typeof meUser?.id === "number" ? meUser.id : null);
        }

        const offset = (currentPage - 1) * pageSize;
        const [groupResponse, postsResponse] = await Promise.all([
          fetchJson<unknown>(`/groups/${groupIDNumber}`),
          fetchJson<unknown>(`/groups/${groupIDNumber}/posts?limit=${pageSize}&offset=${offset}`),
        ]);

        const groupResult = groupResponse.result;
        if (!groupResponse.response.ok || !groupResult?.success) {
          if (!cancelled) {
            if (groupResponse.response.status === 404) {
              setError("Group endpoint is not available yet or this group does not exist.");
            } else {
              setError(groupResult?.error || "Could not load this group.");
            }
            setGroup(null);
          }
        } else {
          const normalized = parseGroup(groupResult.data);
          if (!normalized) {
            if (!cancelled) {
              setError("Received an unexpected group response format.");
              setGroup(null);
            }
          } else if (!cancelled) {
            setGroup(normalized);
          }
        }

        const postsResult = postsResponse.result as ApiResponse<Post[]> | null;
        if (!postsResponse.response.ok || !postsResult?.success) {
          if (!cancelled) {
            if (postsResponse.response.status === 404) {
              setPostsError("Group posts endpoint is not available yet.");
            } else {
              setPostsError(postsResult?.error || "Could not load group posts.");
            }
            setPosts([]);
          }
        } else if (!cancelled) {
          const nextPosts = postsResult.data ?? [];
          setPosts(nextPosts);
          const totalHeader = postsResponse.response.headers.get("X-Total-Count");
          const parsedTotal = totalHeader ? Number(totalHeader) : Number.NaN;
          setTotalPosts(Number.isFinite(parsedTotal) ? parsedTotal : nextPosts.length);

          setPostReactionMap({});
        }

        if (!cancelled) {
          try {
            const membersResponse = await apiFetch(
              `/groups/${groupIDNumber}/members`,
              {},
              apiBaseUrl,
            );
            setIsMember(membersResponse.ok);
          } catch {
            setIsMember(false);
          }
        }
      } catch {
        if (!cancelled) {
          setError("Network error while loading group details.");
          setGroup(null);
          setPostsError("Network error while loading group posts.");
          setPosts([]);
          setIsMember(false);
        }
      } finally {
        if (!cancelled) {
          setIsLoading(false);
          setPostsLoading(false);
        }
      }
    };

    void load();
    return () => {
      cancelled = true;
    };
  }, [apiBaseUrl, fetchJson, groupIDNumber, pageSize, router, currentPage]);

  useEffect(() => {
    const creatorID = group?.creatorID;
    if (!creatorID) {
      setCreatorProfile(null);
      return;
    }
    let cancelled = false;
    const loadCreator = async () => {
      try {
        const { response, result } = await fetchJson<{ user?: ProfileSummary }>(
          `/profiles/${creatorID}`,
        );
        if (!cancelled && response.ok && result?.success && result.data?.user) {
          setCreatorProfile(result.data.user);
        }
      } catch {
        if (!cancelled) {
          setCreatorProfile(null);
        }
      }
    };
    void loadCreator();
    return () => {
      cancelled = true;
    };
  }, [fetchJson, group?.creatorID]);

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
        const { response, result } = await fetchJson<UserSearchItem[]>(
          `/users?q=${encodeURIComponent(inviteQuery.trim())}&limit=6&offset=0`,
          { signal: controller.signal },
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
  }, [fetchJson, inviteQuery]);

  useEffect(() => {
    if (!userID || !Number.isFinite(groupIDNumber) || groupIDNumber <= 0) return;
    const key = `group-join-request:${groupIDNumber}:${userID}`;
    if (isMember) {
      localStorage.removeItem(key);
      setJoinStatus("idle");
      return;
    }
    const cached = localStorage.getItem(key);
    if (cached) {
      setJoinStatus("requested");
    }
  }, [groupIDNumber, isMember, userID]);

  useEffect(() => {
    if (!isMember) return;
    if (activeTab === "members" && members.length === 0 && !membersLoading) {
      void loadMembers();
    }
    if (activeTab === "events" && events.length === 0 && !eventsLoading) {
      void loadEvents(0, false);
    }
  }, [activeTab, events.length, eventsLoading, isMember, members.length, membersLoading]);

  const requestToJoin = async () => {
    setJoinError(null);
    setJoinStatus("idle");
    try {
      const response = await apiFetch(
        `/groups/${groupIDNumber}/join-requests`,
        { method: "POST" },
        apiBaseUrl,
      );
      if (!response.ok) {
        const result = (await response.json().catch(() => null)) as ApiResponse<unknown> | null;
        if (response.status === 409) {
          setJoinStatus("requested");
          if (userID) {
            localStorage.setItem(`group-join-request:${groupIDNumber}:${userID}`, "requested");
          }
          setJoinError("Join request already sent.");
        } else {
          setJoinError(result?.error || "Could not send join request.");
          setJoinStatus("error");
        }
        return;
      }
      setJoinStatus("requested");
      if (userID) {
        localStorage.setItem(`group-join-request:${groupIDNumber}:${userID}`, "requested");
      }
    } catch {
      setJoinError("Network error. Please try again.");
      setJoinStatus("error");
    }
  };

  const loadMembers = async () => {
    if (!Number.isFinite(groupIDNumber) || groupIDNumber <= 0) return;
    setMembersLoading(true);
    setMembersError(null);
    try {
      const { response, result } = await fetchJson<GroupMember[]>(`/groups/${groupIDNumber}/members`);
      if (!response.ok || !result?.success) {
        setMembersError(result?.error || "Could not load members.");
        setMembers([]);
        return;
      }
      setMembers(result.data ?? []);
    } catch {
      setMembersError("Network error. Please try again.");
      setMembers([]);
    } finally {
      setMembersLoading(false);
    }
  };

  const loadEvents = async (offset = 0, append = false) => {
    if (!Number.isFinite(groupIDNumber) || groupIDNumber <= 0) return;
    if (append) {
      setEventsLoadingMore(true);
    } else {
      setEventsLoading(true);
      setEventsError(null);
    }
    try {
      const { response, result } = await fetchJson<EventItem[]>(
        `/groups/${groupIDNumber}/events?limit=${pageSize}&offset=${offset}`,
      );
      if (!response.ok || !result?.success) {
        if (!append) {
          setEventsError(result?.error || "Could not load events.");
          setEvents([]);
        }
        return;
      }
      const items = result.data ?? [];
      setEvents((prev) => (append ? [...prev, ...items] : items));
      setEventsHasMore(items.length >= pageSize);
    } catch {
      if (!append) {
        setEventsError("Network error. Please try again.");
        setEvents([]);
      }
    } finally {
      if (append) {
        setEventsLoadingMore(false);
      } else {
        setEventsLoading(false);
      }
    }
  };

  const handleCreateEvent = async () => {
    setEventCreateError(null);
    if (!eventTitle.trim() || !eventTime) {
      setEventCreateError("Title and date/time are required.");
      return;
    }
    setEventCreating(true);
    try {
      const payload = {
        title: eventTitle.trim(),
        description: eventDescription.trim() || undefined,
        event_time: new Date(eventTime).toISOString(),
      };
      const { response, result } = await fetchJson<EventItem>(`/groups/${groupIDNumber}/events`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(payload),
      });
      if (!response.ok || !result?.success || !result.data) {
        setEventCreateError(result?.error || "Could not create event.");
        return;
      }
      setEvents((prev) => [result.data as EventItem, ...prev]);
      setEventTitle("");
      setEventDescription("");
      setEventTime("");
      setActiveTab("events");
    } catch {
      setEventCreateError("Network error. Please try again.");
    } finally {
      setEventCreating(false);
    }
  };

  const leaveGroup = async () => {
    setLeaveError(null);
    try {
      const response = await apiFetch(
        `/groups/${groupIDNumber}/members/me`,
        { method: "DELETE" },
        apiBaseUrl,
      );
      if (!response.ok) {
        const result = (await response.json().catch(() => null)) as ApiResponse<unknown> | null;
        setLeaveError(result?.error || "Could not leave group.");
        return;
      }
      setIsMember(false);
      setPosts([]);
      setPostsError("You left this group. Re-join to view posts.");
    } catch {
      setLeaveError("Network error. Please try again.");
    }
  };

  const sendInvite = async () => {
    setInviteError(null);
    setInviteSuccess(null);
    if (!selectedInvitee) {
      setInviteError("Pick a user to invite.");
      return;
    }
    try {
      const { response, result } = await fetchJson<unknown>(
        `/groups/${groupIDNumber}/invitations`,
        {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ invitee_id: selectedInvitee.id }),
        },
      );
      if (!response.ok || !result?.success) {
        setInviteError(result?.error || "Could not send invitation.");
        return;
      }
      setInviteSuccess("Invitation sent.");
      setSelectedInvitee(null);
      setInviteQuery("");
      setInviteResults([]);
    } catch {
      setInviteError("Network error. Please try again.");
    }
  };

  const uploadMedia = async (file: File, kind: "post" | "comment") => {
    const formData = new FormData();
    formData.append("file", file);
    formData.append("kind", kind);
    const { response: uploadRes, result: uploadJson } = await fetchJson<{ path?: string }>(
      "/uploads",
      {
        method: "POST",
        body: formData,
      },
    );
    if (!uploadRes.ok || !uploadJson?.success || !uploadJson.data?.path) {
      throw new Error(uploadJson?.error || "Could not upload media.");
    }
    return uploadJson.data.path;
  };

  const handleCreatePost = async () => {
    if (isPosting) return;
    const content = composerText.trim();
    const media = mediaUrl.trim();
    if (!content && !media && !composerFile) {
      setComposerError("Add a message or media before posting.");
      return;
    }

    setIsPosting(true);
    setComposerError(null);

    try {
      let mediaPath: string | undefined;
      if (composerFile) {
        mediaPath = await uploadMedia(composerFile, "post");
      }

      const { response, result } = await fetchJson<Post>(`/groups/${groupIDNumber}/posts`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          content: content || undefined,
          media_path: mediaPath || media || undefined,
        }),
      });
      if (!response.ok || !result?.success || !result.data) {
        setComposerError(result?.error || "Could not publish your post.");
        return;
      }
      setPosts((prev) => [result.data as Post, ...prev]);
      setComposerText("");
      setMediaUrl("");
      setComposerFile(null);
      setComposerFileName("");
    } catch {
      setComposerError("Network error. Please try again.");
    } finally {
      setIsPosting(false);
    }
  };

  const loadCommentsForPost = async (postID: number, page: number) => {
    setCommentsLoadingByPost((prev) => ({ ...prev, [postID]: true }));
    setCommentErrorByPost((prev) => ({ ...prev, [postID]: "" }));

    try {
      const offset = (page - 1) * commentLimit;
      const { response, result } = await fetchJson<Comment[]>(
        `/posts/${postID}/comments?limit=${commentLimit}&offset=${offset}`,
      );

      if (!response.ok || !result?.success) {
        setCommentErrorByPost((prev) => ({
          ...prev,
          [postID]: result?.error || "Could not load comments.",
        }));
        return;
      }

      const comments = result.data ?? [];
      setCommentsByPost((prev) => ({ ...prev, [postID]: comments }));
      setCommentPageByPost((prev) => ({ ...prev, [postID]: page }));
      const totalHeader = response.headers.get("X-Total-Count");
      const parsedTotal = totalHeader ? Number(totalHeader) : Number.NaN;
      setCommentTotalByPost((prev) => ({
        ...prev,
        [postID]: Number.isFinite(parsedTotal) ? parsedTotal : comments.length,
      }));

    } catch {
      setCommentErrorByPost((prev) => ({
        ...prev,
        [postID]: "Network error while loading comments.",
      }));
    } finally {
      setCommentsLoadingByPost((prev) => ({ ...prev, [postID]: false }));
    }
  };

  const toggleComments = (postID: number) => {
    const isOpen = commentsOpenByPost[postID] ?? false;
    const nextOpen = !isOpen;
    setCommentsOpenByPost((prev) => ({ ...prev, [postID]: nextOpen }));
    if (nextOpen) {
      const page = commentPageByPost[postID] ?? 1;
      void loadCommentsForPost(postID, page);
    }
  };

  const handleCreateComment = async (postID: number) => {
    const draft = (commentDraftByPost[postID] ?? "").trim();
    const attachment = commentFileByPost[postID] ?? null;
    if (!draft && !attachment) {
      setCommentErrorByPost((prev) => ({
        ...prev,
        [postID]: "Write a comment or attach media before posting.",
      }));
      return;
    }

    try {
      let mediaPath: string | undefined;
      if (attachment) {
        try {
          mediaPath = await uploadMedia(attachment, "comment");
        } catch (err) {
          setCommentErrorByPost((prev) => ({
            ...prev,
            [postID]: err instanceof Error ? err.message : "Could not upload comment media.",
          }));
          return;
        }
      }

      const { response, result } = await fetchJson<Comment>(`/posts/${postID}/comments`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ content: draft || undefined, media_path: mediaPath }),
      });

      if (!response.ok || !result?.success || !result.data) {
        setCommentErrorByPost((prev) => ({
          ...prev,
          [postID]: result?.error || "Could not post comment.",
        }));
        return;
      }

      const nextPage = commentPageByPost[postID] ?? 1;
      setCommentDraftByPost((prev) => ({ ...prev, [postID]: "" }));
      setCommentFileByPost((prev) => ({ ...prev, [postID]: null }));
      setCommentFileNameByPost((prev) => ({ ...prev, [postID]: "" }));
      setCommentErrorByPost((prev) => ({ ...prev, [postID]: "" }));
      setPosts((prev) =>
        prev.map((post) =>
          post.id === postID ? { ...post, comment_count: post.comment_count + 1 } : post,
        ),
      );
      setCommentsOpenByPost((prev) => ({ ...prev, [postID]: true }));
      setCommentTotalByPost((prev) => ({
        ...prev,
        [postID]: (prev[postID] ?? 0) + 1,
      }));
      await loadCommentsForPost(postID, nextPage);
    } catch {
      setCommentErrorByPost((prev) => ({
        ...prev,
        [postID]: "Network error while posting comment.",
      }));
    }
  };

  const startEditPost = (post: Post) => {
    setEditingPostID(post.id);
    setEditPostText(post.content || "");
    setEditPostFile(null);
    setEditPostFileName("");
    setEditPostClearMedia(false);
    setEditPostError(null);
  };

  const cancelEditPost = () => {
    setEditingPostID(null);
    setEditPostText("");
    setEditPostFile(null);
    setEditPostFileName("");
    setEditPostClearMedia(false);
    setEditPostError(null);
  };

  const saveEditPost = async (post: Post) => {
    const content = editPostText.trim();
    if (!content && !editPostFile && !post.media_path && !editPostClearMedia) {
      setEditPostError("Content or media is required.");
      return;
    }
    setEditPostError(null);
    try {
      let mediaPath: string | undefined;
      if (editPostClearMedia && !editPostFile) {
        mediaPath = "";
      }
      if (editPostFile) {
        mediaPath = await uploadMedia(editPostFile, "post");
      }
      const { response, result } = await fetchJson<Post>(`/posts/${post.id}`, {
        method: "PATCH",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          content: content || undefined,
          media_path: mediaPath ?? undefined,
        }),
      });
      if (!response.ok || !result?.success || !result.data) {
        setEditPostError(result?.error || "Could not update post.");
        return;
      }
      setPosts((prev) => prev.map((item) => (item.id === post.id ? result.data as Post : item)));
      cancelEditPost();
    } catch (err) {
      setEditPostError(err instanceof Error ? err.message : "Network error.");
    }
  };

  const deletePost = async (postID: number) => {
    try {
      const response = await apiFetch(`/posts/${postID}`, { method: "DELETE" }, apiBaseUrl);
      if (!response.ok) {
        return;
      }
      setPosts((prev) => prev.filter((post) => post.id !== postID));
    } catch {
      // ignore
    }
  };

  const startEditComment = (comment: Comment) => {
    setEditingCommentID(comment.id);
    setEditCommentText(comment.content || "");
    setEditCommentFile(null);
    setEditCommentFileName("");
    setEditCommentClearMedia(false);
    setEditCommentError(null);
  };

  const cancelEditComment = () => {
    setEditingCommentID(null);
    setEditCommentText("");
    setEditCommentFile(null);
    setEditCommentFileName("");
    setEditCommentClearMedia(false);
    setEditCommentError(null);
  };

  const saveEditComment = async (postID: number, comment: Comment) => {
    const content = editCommentText.trim();
    if (!content && !editCommentFile && !comment.media_path && !editCommentClearMedia) {
      setEditCommentError("Content or media is required.");
      return;
    }
    setEditCommentError(null);
    try {
      let mediaPath: string | undefined;
      if (editCommentClearMedia && !editCommentFile) {
        mediaPath = "";
      }
      if (editCommentFile) {
        mediaPath = await uploadMedia(editCommentFile, "comment");
      }
      const { response, result } = await fetchJson<Comment>(`/comments/${comment.id}`, {
        method: "PATCH",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          content: content || undefined,
          media_path: mediaPath ?? undefined,
        }),
      });
      if (!response.ok || !result?.success || !result.data) {
        setEditCommentError(result?.error || "Could not update comment.");
        return;
      }
      setCommentsByPost((prev) => ({
        ...prev,
        [postID]: (prev[postID] ?? []).map((item) =>
          item.id === comment.id ? (result.data as Comment) : item,
        ),
      }));
      cancelEditComment();
    } catch (err) {
      setEditCommentError(err instanceof Error ? err.message : "Network error.");
    }
  };

  const deleteComment = async (postID: number, commentID: number) => {
    try {
      const response = await apiFetch(`/comments/${commentID}`, { method: "DELETE" }, apiBaseUrl);
      if (!response.ok) {
        return;
      }
      setCommentsByPost((prev) => ({
        ...prev,
        [postID]: (prev[postID] ?? []).filter((item) => item.id !== commentID),
      }));
      setPosts((prev) =>
        prev.map((post) =>
          post.id === postID ? { ...post, comment_count: Math.max(0, post.comment_count - 1) } : post,
        ),
      );
      setCommentTotalByPost((prev) => ({
        ...prev,
        [postID]: Math.max(0, (prev[postID] ?? 0) - 1),
      }));
      const page = commentPageByPost[postID] ?? 1;
      await loadCommentsForPost(postID, page);
    } catch {
      // ignore
    }
  };

  const handlePostReaction = async (postID: number, reaction: ReactionKind) => {
    const previous = postReactionMap[postID] ?? null;
    const next = previous === reaction ? null : reaction;
    setPostReactionMap((prev) => ({ ...prev, [postID]: next }));
    setPosts((prev) =>
      prev.map((post) => {
        if (post.id !== postID) return post;
        let like = post.like_count;
        let dislike = post.dislike_count;
        if (previous === "like") like -= 1;
        if (previous === "dislike") dislike -= 1;
        if (next === "like") like += 1;
        if (next === "dislike") dislike += 1;
        return {
          ...post,
          like_count: Math.max(0, like),
          dislike_count: Math.max(0, dislike),
        };
      }),
    );

    try {
      await apiFetch(
        `/posts/${postID}/reactions`,
        {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({
            reaction,
          }),
        },
        apiBaseUrl,
      );
    } catch {
      setPostReactionMap((prev) => ({ ...prev, [postID]: previous }));
    }
  };

  const handleCommentReaction = async (
    postID: number,
    commentID: number,
    reaction: ReactionKind,
  ) => {
    const previous = commentReactionMap[commentID] ?? null;
    const next = previous === reaction ? null : reaction;
    setCommentReactionMap((prev) => ({ ...prev, [commentID]: next }));
    setCommentsByPost((prev) => ({
      ...prev,
      [postID]: (prev[postID] ?? []).map((comment) => {
        if (comment.id !== commentID) return comment;
        let like = comment.like_count;
        let dislike = comment.dislike_count;
        if (previous === "like") like -= 1;
        if (previous === "dislike") dislike -= 1;
        if (next === "like") like += 1;
        if (next === "dislike") dislike += 1;
        return {
          ...comment,
          like_count: Math.max(0, like),
          dislike_count: Math.max(0, dislike),
        };
      }),
    }));

    try {
      await apiFetch(
        `/comments/${commentID}/reactions`,
        {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({
            reaction,
          }),
        },
        apiBaseUrl,
      );
    } catch {
      setCommentReactionMap((prev) => ({ ...prev, [commentID]: previous }));
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

        <section>
        <motion.section
          initial="hidden"
          whileInView="show"
          viewport={viewportOnce}
          variants={fadeUp}
          className="rounded-3xl border border-white/10 bg-white/5 p-6 backdrop-blur-sm"
        >
          <div className="flex flex-wrap items-center justify-between gap-2">
            <span className="inline-flex items-center gap-2 rounded-full border border-white/20 bg-white/5 px-3 py-1 text-xs font-semibold text-white">
              <Users className="h-3.5 w-3.5" />
              {group?.name || "Group"}
            </span>
            <button
              type="button"
              onClick={() => window.location.reload()}
              className="inline-flex items-center gap-2 rounded-full border border-white/20 bg-white/5 px-3 py-1.5 text-xs font-semibold text-neutral-300 transition hover:bg-white/10 hover:text-white"
            >
              <RefreshCw className="h-3.5 w-3.5" />
              Refresh
            </button>
          </div>

          {isLoading ? (
            <p className="mt-4 text-sm text-neutral-400">Loading group details...</p>
          ) : error ? (
            <p className="mt-4 rounded-2xl border border-rose-500/30 bg-rose-500/10 px-4 py-3 text-sm text-rose-400">
              {formatApiError(error)}
            </p>
          ) : group ? (
            <>
              <h1 className="mt-3 text-2xl font-semibold tracking-tight text-white">{group.name}</h1>
              <p className="mt-2 text-sm text-neutral-400">{group.description}</p>

              <div className="mt-5 grid grid-cols-1 gap-3 sm:grid-cols-2">
                <div className="rounded-2xl border border-white/10 bg-white/5 p-3">
                  <p className="text-[11px] uppercase tracking-wide text-neutral-400">Members</p>
                  <p className="mt-1 text-sm font-semibold text-white">{group.memberCount}</p>
                </div>
                <div className="rounded-2xl border border-white/10 bg-white/5 p-3">
                  <p className="text-[11px] uppercase tracking-wide text-neutral-400">Creator</p>
                  {group.creatorID ? (
                    <Link
                      href={`/profile/${group.creatorID}`}
                      className="mt-1 inline-flex text-sm font-semibold text-white transition hover:text-neutral-300"
                    >
                      {creatorProfile
                        ? `${creatorProfile.first_name} ${creatorProfile.last_name}`
                        : "Group creator"}
                    </Link>
                  ) : (
                    <p className="mt-1 text-sm font-semibold text-white">N/A</p>
                  )}
                </div>
              </div>

              <div className="mt-4 grid grid-cols-1 gap-3 sm:grid-cols-2">
                <div className="rounded-2xl border border-white/10 bg-white/5 p-3">
                  <p className="inline-flex items-center gap-1 text-[11px] uppercase tracking-wide text-neutral-400">
                    <Calendar className="h-3.5 w-3.5" />
                    Created
                  </p>
                  <p className="mt-1 text-sm font-semibold text-white">{formatDate(group.createdAt)}</p>
                </div>
                <div className="rounded-2xl border border-white/10 bg-white/5 p-3">
                  <p className="inline-flex items-center gap-1 text-[11px] uppercase tracking-wide text-neutral-400">
                    <Calendar className="h-3.5 w-3.5" />
                    Updated
                  </p>
                  <p className="mt-1 text-sm font-semibold text-white">{formatDate(group.updatedAt)}</p>
                </div>
              </div>
            </>
          ) : (
            <p className="mt-4 text-sm text-neutral-400">Group details are not available.</p>
          )}

          <div className="mt-6 flex flex-wrap gap-3">
            <Link
              href="/groups"
              className="inline-flex items-center gap-2 rounded-full border border-white/20 bg-white/5 px-4 py-2 text-sm font-semibold text-neutral-300 transition hover:bg-white/10 hover:text-white"
            >
              <ArrowLeft className="h-4 w-4" />
              Back to groups
            </Link>
            <Link
              href="/messages"
              className="brand-gradient inline-flex items-center gap-2 rounded-full px-4 py-2 text-sm font-semibold text-white transition hover:-translate-y-0.5 hover:shadow-md"
            >
              Open group chat
              <ArrowRight className="h-4 w-4" />
            </Link>
          </div>
        </motion.section>

        <motion.section
          initial="hidden"
          whileInView="show"
          viewport={viewportOnce}
          variants={fadeUp}
          className="mt-6 rounded-3xl border border-white/10 bg-white/5 p-6 backdrop-blur-sm"
        >
          <div className="flex flex-wrap items-center justify-between gap-3">
            <div>
              <h2 className="text-lg font-semibold text-white">Group posts</h2>
              <p className="text-sm text-neutral-400">
                Latest posts shared with this group.
              </p>
            </div>
            <span className="rounded-full border border-white/20 bg-white/5 px-3 py-1 text-xs text-white">
              {posts.length} post(s)
            </span>
          </div>

          <div className="mt-4 flex flex-wrap items-center gap-3">
            <div className="flex items-center gap-2 rounded-full border border-white/20 bg-white/5 p-1 text-[11px] font-semibold text-neutral-300">
              <button
                type="button"
                onClick={() => setActiveTab("posts")}
                className={`rounded-full px-3 py-1 transition ${
                  activeTab === "posts"
                    ? "bg-neutral-900 text-white"
                    : "text-neutral-400 hover:text-white"
                }`}
              >
                Posts
              </button>
              <button
                type="button"
                onClick={() => setActiveTab("events")}
                className={`rounded-full px-3 py-1 transition ${
                  activeTab === "events"
                    ? "bg-neutral-900 text-white"
                    : "text-neutral-400 hover:text-white"
                }`}
              >
                Events
              </button>
              <button
                type="button"
                onClick={() => setActiveTab("members")}
                className={`rounded-full px-3 py-1 transition ${
                  activeTab === "members"
                    ? "bg-neutral-900 text-white"
                    : "text-neutral-400 hover:text-white"
                }`}
              >
                Members
              </button>
            </div>
            {group?.creatorID && userID === group.creatorID ? (
              <Link
                href={`/groups/${groupIDNumber}/join-requests`}
                className="inline-flex items-center gap-2 rounded-full border border-white/20 bg-white/5 px-3 py-2 text-xs font-semibold text-neutral-300 transition hover:bg-white/10 hover:text-white"
              >
                Join requests
              </Link>
            ) : null}
            {!isMember ? (
              <button
                type="button"
                onClick={requestToJoin}
                disabled={joinStatus === "requested"}
                className="inline-flex items-center gap-2 rounded-full bg-neutral-900 px-3 py-2 text-xs font-semibold text-white transition hover:bg-neutral-800 disabled:cursor-not-allowed disabled:opacity-60"
              >
                Request to join
              </button>
            ) : group?.creatorID && userID === group.creatorID ? (
              <span className="rounded-full bg-emerald-50 px-3 py-2 text-xs font-semibold text-emerald-700">
                You are the creator
              </span>
            ) : (
              <button
                type="button"
                onClick={leaveGroup}
                className="inline-flex items-center gap-2 rounded-full border border-white/20 bg-white/5 px-3 py-2 text-xs font-semibold text-neutral-300 transition hover:bg-white/10 hover:text-white"
              >
                Leave group
              </button>
            )}
            {joinStatus === "requested" ? (
              <span className="rounded-full bg-emerald-50 px-3 py-2 text-xs font-semibold text-emerald-700">
                Join request pending
              </span>
            ) : null}
          </div>
          {joinError ? <p className="mt-2 text-xs text-rose-600">{formatApiError(joinError)}</p> : null}
          {leaveError ? <p className="mt-2 text-xs text-rose-600">{formatApiError(leaveError)}</p> : null}

          {activeTab === "posts" ? (
            <>
              <div className="mt-5 rounded-3xl border border-white/10 bg-white/5 p-4 backdrop-blur-sm">
                <textarea
                  value={composerText}
                  onChange={(event) => setComposerText(event.target.value)}
                  rows={4}
                  placeholder="Share an update with this group..."
                  className="w-full resize-none rounded-2xl border border-white/20 bg-white/5 px-4 py-3 text-sm text-white placeholder:text-neutral-500 outline-none transition focus:border-white/40"
                />
                <div className="mt-3 flex flex-wrap items-center gap-3">
                  <label className="inline-flex items-center gap-2 rounded-full border border-white/20 bg-white/5 px-3 py-2 text-xs font-semibold text-neutral-300 transition hover:bg-white/10 hover:text-white">
                    <input
                      type="file"
                      accept="image/png,image/jpeg,image/gif"
                      className="hidden"
                      onChange={(event) => {
                        const file = event.target.files?.[0] ?? null;
                        setComposerFile(file);
                        setComposerFileName(file?.name ?? "");
                      }}
                    />
                    Add media
                  </label>
                  {composerFileName ? (
                    <span className="text-xs text-neutral-400">{composerFileName}</span>
                  ) : null}
                  <input
                    value={mediaUrl}
                    onChange={(event) => setMediaUrl(event.target.value)}
                    placeholder="Or paste media URL"
                    className="h-10 flex-1 rounded-2xl border border-white/20 bg-white/5 px-4 text-sm text-white placeholder:text-neutral-500 outline-none transition focus:border-white/40"
                  />
                </div>
                <div className="mt-3 flex flex-wrap items-center justify-between gap-3">
                  <button
                    type="button"
                    onClick={handleCreatePost}
                    disabled={isPosting}
                    className="brand-gradient inline-flex items-center gap-2 rounded-full px-4 py-2 text-xs font-semibold text-white transition hover:-translate-y-0.5 hover:shadow-md disabled:cursor-not-allowed disabled:opacity-70"
                  >
                    <Send className="h-3.5 w-3.5" />
                    {isPosting ? "Posting..." : "Publish"}
                  </button>
                </div>
                {composerError ? (
                  <p className="mt-3 text-xs text-rose-600">{formatApiError(composerError)}</p>
                ) : null}
              </div>

              {postsLoading ? (
                <p className="mt-4 text-sm text-neutral-400">Loading group posts...</p>
              ) : postsError ? (
                <p className="mt-4 rounded-2xl border border-rose-500/30 bg-rose-500/10 px-4 py-3 text-sm text-rose-400">
                  {formatApiError(postsError)}
                </p>
              ) : posts.length === 0 ? (
                <p className="mt-4 text-sm text-neutral-400">
                  No posts yet. Be the first to share something.
                </p>
              ) : (
                <div className="mt-4 space-y-4">
                  {posts.map((post) => (
                    <article
                      key={post.id}
                      className="rounded-3xl border border-white/10 bg-white/5 p-5 backdrop-blur-sm"
                    >
                  <header className="flex items-start justify-between gap-3">
                    <div>
                      <p className="text-sm font-semibold text-white">
                        {post.author_first_name} {post.author_last_name}
                      </p>
                      <p className="text-xs text-neutral-400">{shortDate(post.created_at)}</p>
                    </div>
                    <button
                      type="button"
                      onClick={() => toggleComments(post.id)}
                      className="inline-flex items-center gap-1 rounded-full border border-white/20 bg-white/5 px-2.5 py-1 text-[11px] text-neutral-400"
                    >
                      <MessageCircle className="h-3.5 w-3.5" />
                      {post.comment_count} comments
                    </button>
                  </header>

                  {editingPostID === post.id ? (
                    <div className="mt-3 space-y-3">
                      <textarea
                        value={editPostText}
                        onChange={(event) => setEditPostText(event.target.value)}
                        rows={3}
                        className="w-full rounded-2xl border border-white/20 bg-white/5 px-4 py-3 text-sm outline-none focus:border-neutral-400"
                      />
                        <div className="flex flex-wrap items-center gap-2">
                          <label className="inline-flex items-center gap-2 rounded-full border border-white/20 bg-white/5 px-3 py-2 text-xs font-semibold text-neutral-300 transition hover:bg-white/10 hover:text-white">
                            <input
                              type="file"
                              accept="image/png,image/jpeg,image/gif"
                              className="hidden"
                              onChange={(event) => {
                                const file = event.target.files?.[0] ?? null;
                                setEditPostFile(file);
                                setEditPostFileName(file?.name ?? "");
                                setEditPostClearMedia(false);
                              }}
                            />
                            Change media
                          </label>
                          {editPostFileName ? (
                            <span className="text-xs text-neutral-400">{editPostFileName}</span>
                          ) : null}
                          {post.media_path ? (
                            <button
                              type="button"
                              onClick={() => {
                                setEditPostClearMedia(true);
                                setEditPostFile(null);
                                setEditPostFileName("");
                              }}
                              className={`rounded-full border px-3 py-2 text-xs font-semibold transition ${
                                editPostClearMedia
                                  ? "border-rose-200 bg-rose-50 text-rose-700"
                                  : "border-white/20 bg-white/5 text-neutral-300 hover:bg-white/10 hover:text-white"
                              }`}
                            >
                              {editPostClearMedia ? "Media removed" : "Remove media"}
                            </button>
                          ) : null}
                          <button
                            type="button"
                            onClick={() => saveEditPost(post)}
                            className="rounded-full bg-neutral-900 px-3 py-2 text-xs font-semibold text-white"
                        >
                          Save
                        </button>
                        <button
                          type="button"
                          onClick={cancelEditPost}
                          className="rounded-full border border-white/20 bg-white/5 px-3 py-2 text-xs font-semibold text-neutral-300"
                        >
                          Cancel
                        </button>
                      </div>
                      {editPostError ? (
                        <p className="text-xs text-rose-600">{formatApiError(editPostError)}</p>
                      ) : null}
                    </div>
                  ) : (
                    <p className="mt-3 text-sm leading-relaxed text-neutral-300">{post.content}</p>
                  )}

                  {post.media_path ? (
                    <div className="mt-4 overflow-hidden rounded-2xl border border-white/20 bg-white/5">
                      <img
                        src={toMediaUrl(apiBaseUrl, post.media_path)}
                        alt="Post media"
                        className="max-h-[520px] w-full object-contain bg-white"
                      />
                    </div>
                  ) : null}

                  <footer className="mt-4 flex items-center gap-3 text-xs text-neutral-400">
                    <button
                      type="button"
                      onClick={() => handlePostReaction(post.id, "like")}
                      className={`inline-flex items-center gap-1 rounded-full px-2 py-1 transition ${
                        postReactionMap[post.id] === "like"
                          ? "bg-emerald-100 text-emerald-800"
                          : "bg-white/10 text-neutral-400 hover:bg-white/20 hover:text-white"
                      }`}
                    >
                      <ThumbsUp className="h-3.5 w-3.5" />
                      {post.like_count}
                    </button>
                    <button
                      type="button"
                      onClick={() => handlePostReaction(post.id, "dislike")}
                      className={`inline-flex items-center gap-1 rounded-full px-2 py-1 transition ${
                        postReactionMap[post.id] === "dislike"
                          ? "bg-rose-100 text-rose-800"
                          : "bg-white/10 text-neutral-400 hover:bg-white/20 hover:text-white"
                      }`}
                    >
                      <ThumbsDown className="h-3.5 w-3.5" />
                      {post.dislike_count}
                    </button>
                    <button
                      type="button"
                      onClick={() => toggleComments(post.id)}
                      className="inline-flex items-center gap-1 rounded-full bg-white/10 px-2 py-1 text-neutral-300 transition hover:bg-white/20"
                    >
                      <MessageCircle className="h-3.5 w-3.5" />
                      {post.comment_count}
                    </button>
                    {userID === post.author_id ? (
                      <>
                        <button
                          type="button"
                          onClick={() => startEditPost(post)}
                          className="inline-flex items-center gap-1 rounded-full bg-white/10 px-2 py-1 text-neutral-300 transition hover:bg-white/20"
                        >
                          Edit
                        </button>
                        <button
                          type="button"
                          onClick={() => deletePost(post.id)}
                          className="inline-flex items-center gap-1 rounded-full bg-white/10 px-2 py-1 text-neutral-300 transition hover:bg-white/20"
                        >
                          Delete
                        </button>
                      </>
                    ) : null}
                  </footer>

                  {commentsOpenByPost[post.id] ? (
                    <section className="mt-4 rounded-2xl border border-white/20 bg-white/5 p-3">
                      <div className="space-y-2">
                        {(commentsByPost[post.id] ?? []).map((comment) => (
                          <article key={comment.id} className="rounded-xl bg-white/5 p-3">
                            {editingCommentID === comment.id ? (
                              <div className="space-y-2">
                                <textarea
                                  value={editCommentText}
                                  onChange={(event) => setEditCommentText(event.target.value)}
                                  rows={2}
                                  className="w-full rounded-xl border border-white/20 bg-white/5 px-3 py-2 text-xs outline-none focus:border-neutral-400"
                                />
                                <div className="flex flex-wrap items-center gap-2">
                                  <label className="inline-flex items-center gap-2 rounded-full border border-white/20 bg-white/5 px-3 py-1 text-[11px] font-semibold text-neutral-300 transition hover:bg-white/10 hover:text-white">
                                    <input
                                      type="file"
                                      accept="image/png,image/jpeg,image/gif"
                                      className="hidden"
                                      onChange={(event) => {
                                        const file = event.target.files?.[0] ?? null;
                                        setEditCommentFile(file);
                                        setEditCommentFileName(file?.name ?? "");
                                        setEditCommentClearMedia(false);
                                      }}
                                    />
                                    Change media
                                  </label>
                                  {editCommentFileName ? (
                                    <span className="text-[11px] text-neutral-400">
                                      {editCommentFileName}
                                    </span>
                                  ) : null}
                                  {comment.media_path ? (
                                    <button
                                      type="button"
                                      onClick={() => {
                                        setEditCommentClearMedia(true);
                                        setEditCommentFile(null);
                                        setEditCommentFileName("");
                                      }}
                                      className={`rounded-full border px-3 py-1 text-[11px] font-semibold transition ${
                                        editCommentClearMedia
                                          ? "border-rose-200 bg-rose-50 text-rose-700"
                                          : "border-white/20 bg-white/5 text-neutral-300 hover:bg-white/10 hover:text-white"
                                      }`}
                                    >
                                      {editCommentClearMedia ? "Media removed" : "Remove media"}
                                    </button>
                                  ) : null}
                                  <button
                                    type="button"
                                    onClick={() => saveEditComment(post.id, comment)}
                                    className="rounded-full bg-neutral-900 px-3 py-1 text-[11px] font-semibold text-white"
                                  >
                                    Save
                                  </button>
                                  <button
                                    type="button"
                                    onClick={cancelEditComment}
                                    className="rounded-full border border-white/20 bg-white/5 px-3 py-1 text-[11px] font-semibold text-neutral-300"
                                  >
                                    Cancel
                                  </button>
                                </div>
                                {editCommentError ? (
                                  <p className="text-[11px] text-rose-600">{formatApiError(editCommentError)}</p>
                                ) : null}
                              </div>
                            ) : (
                              <>
                                <p className="text-sm text-neutral-300">{comment.content}</p>
                                {comment.media_path ? (
                                  <div className="mt-2 overflow-hidden rounded-xl border border-white/20 bg-white/5">
                                    <img
                                      src={toMediaUrl(apiBaseUrl, comment.media_path)}
                                      alt="Comment media"
                                      className="max-h-64 w-full object-contain bg-white"
                                    />
                                  </div>
                                ) : null}
                              </>
                            )}
                            <div className="mt-2 flex items-center gap-2 text-xs">
                              <button
                                type="button"
                                onClick={() =>
                                  handleCommentReaction(post.id, comment.id, "like")
                                }
                                className={`inline-flex items-center gap-1 rounded-full px-2 py-1 ${
                                  commentReactionMap[comment.id] === "like"
                                    ? "bg-emerald-100 text-emerald-800"
                                    : "bg-white/10 text-neutral-400 hover:text-white"
                                }`}
                              >
                                <ThumbsUp className="h-3 w-3" />
                                {comment.like_count}
                              </button>
                              <button
                                type="button"
                                onClick={() =>
                                  handleCommentReaction(post.id, comment.id, "dislike")
                                }
                                className={`inline-flex items-center gap-1 rounded-full px-2 py-1 ${
                                  commentReactionMap[comment.id] === "dislike"
                                    ? "bg-rose-100 text-rose-800"
                                    : "bg-white/10 text-neutral-400 hover:text-white"
                                }`}
                              >
                                <ThumbsDown className="h-3 w-3" />
                                {comment.dislike_count}
                              </button>
                              {userID === comment.author_id ? (
                                <>
                                  <button
                                    type="button"
                                    onClick={() => startEditComment(comment)}
                                    className="inline-flex items-center gap-1 rounded-full bg-white/10 px-2 py-1 text-neutral-300 transition hover:bg-white/20"
                                  >
                                    Edit
                                  </button>
                                  <button
                                    type="button"
                                    onClick={() => deleteComment(post.id, comment.id)}
                                    className="inline-flex items-center gap-1 rounded-full bg-white/10 px-2 py-1 text-neutral-300 transition hover:bg-white/20"
                                  >
                                    Delete
                                  </button>
                                </>
                              ) : null}
                            </div>
                          </article>
                        ))}

                        {commentsLoadingByPost[post.id] ? (
                          <p className="text-xs text-neutral-400">Loading comments...</p>
                        ) : null}
                        {commentErrorByPost[post.id] ? (
                          <p className="text-xs text-rose-600">{commentErrorByPost[post.id]}</p>
                        ) : null}
                      </div>

                      <Pagination
                        currentPage={commentPageByPost[post.id] ?? 1}
                        totalPages={Math.max(
                          1,
                          Math.ceil(
                            (commentTotalByPost[post.id] ??
                              (commentsByPost[post.id]?.length ?? 0)) / commentLimit,
                          ),
                        )}
                        onPageChange={(page) => loadCommentsForPost(post.id, page)}
                        className="mt-3"
                      />

                      <div className="mt-3 flex gap-2">
                        <input
                          value={commentDraftByPost[post.id] ?? ""}
                          onChange={(event) =>
                            setCommentDraftByPost((prev) => ({
                              ...prev,
                              [post.id]: event.target.value,
                            }))
                          }
                          placeholder="Write a comment..."
                          className="h-9 flex-1 rounded-xl border border-white/20 bg-white/5 px-3 text-xs text-white outline-none focus:border-white/40"
                        />
                        <label className="inline-flex h-9 items-center gap-2 rounded-xl border border-white/20 bg-white/5 px-3 text-xs font-semibold text-neutral-700 transition hover:bg-white/10 hover:text-white">
                          <input
                            type="file"
                            accept="image/png,image/jpeg,image/gif"
                            className="hidden"
                            onChange={(event) => {
                              const file = event.target.files?.[0] ?? null;
                              setCommentFileByPost((prev) => ({ ...prev, [post.id]: file }));
                              setCommentFileNameByPost((prev) => ({
                                ...prev,
                                [post.id]: file?.name ?? "",
                              }));
                            }}
                          />
                          Add media
                        </label>
                        <button
                          type="button"
                          onClick={() => handleCreateComment(post.id)}
                          className="rounded-xl bg-neutral-900 px-3 text-xs font-semibold text-white"
                        >
                          Comment
                        </button>
                      </div>
                      {commentFileNameByPost[post.id] ? (
                        <p className="mt-2 text-[11px] text-neutral-400">
                          Attached: {commentFileNameByPost[post.id]}
                        </p>
                      ) : null}
                    </section>
                  ) : null}
                    </article>
                  ))}
                </div>
              )}
            </>
          ) : null}

          {!isMember && activeTab !== "posts" ? (
            <div className="mt-5 rounded-2xl border border-amber-200 bg-amber-50 px-4 py-3 text-sm text-amber-800">
              Join this group to access {activeTab === "events" ? "events" : "members"}.
            </div>
          ) : null}

          {activeTab === "events" && isMember ? (
            <div className="mt-5 space-y-4">
              <div className="rounded-3xl border border-white/10 bg-white/5 p-4 backdrop-blur-sm">
                <h3 className="text-sm font-semibold text-white">Create event</h3>
                <div className="mt-3 grid gap-3">
                  <input
                    value={eventTitle}
                    onChange={(event) => setEventTitle(event.target.value)}
                    placeholder="Event title"
                    className="h-10 w-full rounded-2xl border border-white/20 bg-white/5 px-3 text-xs outline-none focus:border-neutral-400"
                  />
                  <textarea
                    value={eventDescription}
                    onChange={(event) => setEventDescription(event.target.value)}
                    rows={3}
                    placeholder="Description"
                    className="w-full resize-none rounded-2xl border border-white/20 bg-white/5 px-3 py-2 text-xs outline-none focus:border-neutral-400"
                  />
                  <input
                    type="datetime-local"
                    value={eventTime}
                    onChange={(event) => setEventTime(event.target.value)}
                    className="h-10 w-full rounded-2xl border border-white/20 bg-white/5 px-3 text-xs outline-none focus:border-neutral-400"
                  />
                  <button
                    type="button"
                    onClick={handleCreateEvent}
                    disabled={eventCreating}
                    className="inline-flex w-fit items-center gap-2 rounded-full bg-neutral-900 px-4 py-2 text-xs font-semibold text-white transition hover:bg-neutral-800 disabled:cursor-not-allowed disabled:opacity-60"
                  >
                    Create event
                  </button>
                  {eventCreateError ? (
                    <p className="text-xs text-rose-600">{formatApiError(eventCreateError)}</p>
                  ) : null}
                </div>
              </div>

              {eventsLoading ? (
                <p className="text-sm text-neutral-400">Loading events...</p>
              ) : eventsError ? (
                <p className="rounded-2xl border border-rose-500/30 bg-rose-500/10 px-4 py-3 text-sm text-rose-400">
                  {formatApiError(eventsError)}
                </p>
              ) : events.length === 0 ? (
                <p className="text-sm text-neutral-400">No events yet.</p>
              ) : (
                <div className="space-y-3">
                  {events.map((event) => (
                    <article
                      key={event.id}
                      className="rounded-3xl border border-white/10 bg-white/5 p-5 backdrop-blur-sm"
                    >
                      <div className="flex flex-wrap items-start justify-between gap-3">
                        <div>
                          <h4 className="text-base font-semibold text-white">
                            {event.title}
                          </h4>
                          <p className="mt-1 text-sm text-neutral-400">
                            {event.description || "No description."}
                          </p>
                        </div>
                        <span className="rounded-full border border-white/20 bg-white/5 px-3 py-1 text-xs text-white">
                          {formatDateTime(event.event_time)}
                        </span>
                      </div>
                      <div className="mt-4 flex items-center justify-between gap-2">
                        <span className="text-xs text-neutral-400">Event</span>
                        <Link
                          href={`/events/${event.id}`}
                          className="inline-flex items-center gap-2 rounded-full border border-white/20 bg-white/5 px-3 py-2 text-xs font-semibold text-neutral-300 transition hover:bg-white/10 hover:text-white"
                        >
                          View details
                        </Link>
                      </div>
                    </article>
                  ))}
                  {eventsHasMore ? (
                    <button
                      type="button"
                      onClick={() => loadEvents(events.length, true)}
                      disabled={eventsLoadingMore}
                      className="w-full rounded-2xl border border-white/20 bg-white/5 px-4 py-2 text-xs font-semibold text-neutral-300 transition hover:bg-white/10 disabled:cursor-not-allowed disabled:opacity-60"
                    >
                      {eventsLoadingMore ? "Loading..." : "Load more"}
                    </button>
                  ) : null}
                </div>
              )}
            </div>
          ) : null}

          {activeTab === "members" && isMember ? (
            <div className="mt-5 space-y-4">
              {group?.creatorID && userID === group.creatorID ? (
                <div className="rounded-3xl border border-white/10 bg-white/5 p-4 backdrop-blur-sm">
                  <h3 className="text-sm font-semibold text-white">Invite members</h3>
                  <p className="mt-1 text-xs text-neutral-400">
                    Search users and send an invitation to join this group.
                  </p>
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
                        placeholder="Search by name or nickname"
                        className="h-10 w-full rounded-2xl border border-white/20 bg-white/5 px-3 text-xs outline-none focus:border-neutral-400"
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
                      <div className="flex flex-wrap items-center justify-between gap-2 rounded-2xl border border-white/20 bg-white/5 px-3 py-2 text-xs text-neutral-300">
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
                          className="rounded-full border border-neutral-200 bg-white px-3 py-1 text-[11px] font-semibold text-neutral-600 transition hover:bg-white/10 hover:text-white"
                        >
                          Clear
                        </button>
                      </div>
                    ) : null}
                    <button
                      type="button"
                      onClick={sendInvite}
                      disabled={!selectedInvitee}
                      className="rounded-full bg-neutral-900 px-4 py-2 text-xs font-semibold text-white transition hover:bg-neutral-800 disabled:cursor-not-allowed disabled:opacity-60"
                    >
                      Send invitation
                    </button>
                  </div>
                  {inviteError ? <p className="mt-2 text-xs text-rose-600">{formatApiError(inviteError)}</p> : null}
                  {inviteSuccess ? (
                    <p className="mt-2 text-xs text-emerald-600">{inviteSuccess}</p>
                  ) : null}
                </div>
              ) : null}

              {membersLoading ? (
                <p className="text-sm text-neutral-400">Loading members...</p>
              ) : membersError ? (
                <p className="rounded-2xl border border-rose-500/30 bg-rose-500/10 px-4 py-3 text-sm text-rose-400">
                  {formatApiError(membersError)}
                </p>
              ) : members.length === 0 ? (
                <p className="text-sm text-neutral-400">No members found.</p>
              ) : (
                <div className="space-y-2">
                  {members.map((member, index) => (
                    <div
                      key={`${member.id}-${index}`}
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
              )}
            </div>
          ) : null}

          <Pagination
            currentPage={currentPage}
            totalPages={totalPages}
            onPageChange={setCurrentPage}
            className="mt-5"
          />
        </motion.section>
        </section>
      </main>
    </div>
  );
}
