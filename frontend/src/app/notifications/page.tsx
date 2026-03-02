"use client";

import { useEffect, useMemo, useState } from "react";
import { useRouter } from "next/navigation";
import TopNav from "@/components/TopNav";
import LeftNav from "@/components/LeftNav";
import { useAuth } from "@/components/AuthContext";
import { useNotifications } from "@/components/NotificationsContext";

function notificationActorName(item: {
  metadata?: Record<string, unknown>;
  actor_id?: number;
}) {
  const meta = item.metadata ?? {};
  const requester = meta["requester_name"];
  if (typeof requester === "string" && requester.trim()) return requester;
  return "Someone";
}

function notificationGroupName(item: { metadata?: Record<string, unknown> }) {
  const meta = item.metadata ?? {};
  const groupName = meta["group_name"];
  if (typeof groupName === "string" && groupName.trim()) return groupName;
  return "your group";
}

function notificationTitle(item: { type: string }) {
  switch (item.type) {
    case "follow_request":
      return "Follow request";
    case "group_invitation":
      return "Group invitation";
    case "group_join_request":
      return "Join request";
    case "event_created":
      return "New group event";
    default:
      return "Notification";
  }
}

function notificationBody(item: {
  type: string;
  metadata?: Record<string, unknown>;
  actor_id?: number;
}) {
  switch (item.type) {
    case "follow_request":
      return `${notificationActorName(item)} sent you a follow request.`;
    case "group_invitation":
      return `${notificationActorName(item)} invited you to ${notificationGroupName(item)}.`;
    case "group_join_request":
      return `${notificationActorName(item)} requested to join ${notificationGroupName(item)}.`;
    case "event_created":
      return `New event in ${notificationGroupName(item)}.`;
    default:
      return "Notification update.";
  }
}

function notificationHref(item: {
  type: string;
  entity_id: number;
  metadata?: Record<string, unknown>;
}) {
  const meta = item.metadata ?? {};
  switch (item.type) {
    case "follow_request":
      return "/follow-requests";
    case "group_invitation":
      return "/group-invitations";
    case "group_join_request": {
      const groupID = meta["group_id"];
      if (typeof groupID === "number") {
        return `/groups/${groupID}/join-requests`;
      }
      return "/groups";
    }
    case "event_created":
      return `/events/${item.entity_id}`;
    default:
      return null;
  }
}

export default function NotificationsPage() {
  const router = useRouter();
  const { logout } = useAuth();
  const notificationsContext = useNotifications();
  const [viewer, setViewer] = useState<{
    id: number;
    email?: string;
    first_name?: string;
    last_name?: string;
    nickname?: string | null;
    avatar_path?: string | null;
  } | null>(null);
  const [loadingViewer, setLoadingViewer] = useState(true);

  const notifications = notificationsContext?.notifications ?? [];
  const loading = notificationsContext?.loading ?? false;
  const markRead = notificationsContext?.markRead ?? (async () => Promise.resolve());
  const markAllRead = notificationsContext?.markAllRead ?? (async () => Promise.resolve());

  const apiBaseUrl = useMemo(
    () =>
      process.env.NEXT_PUBLIC_API_BASE_URL?.trim().replace(/\/+$/, "") ||
      "http://localhost:8080",
    [],
  );

  useEffect(() => {
    let cancelled = false;
    const loadViewer = async () => {
      try {
        const response = await fetch(`${apiBaseUrl}/auth/me`, { credentials: "include" });
        const result = (await response.json().catch(() => null)) as
          | { success?: boolean; data?: typeof viewer }
          | null;
        if (!cancelled && response.ok && result?.success) {
          setViewer(result.data ?? null);
        } else if (!cancelled) {
          router.replace("/login");
        }
      } catch {
        if (!cancelled) {
          router.replace("/login");
        }
      } finally {
        if (!cancelled) {
          setLoadingViewer(false);
        }
      }
    };
    void loadViewer();
    return () => {
      cancelled = true;
    };
  }, [apiBaseUrl, router]);

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
      <TopNav
        user={viewer ?? undefined}
        onLogout={() => {
          logout();
          router.replace("/login");
        }}
        variant="dark"
      />

      <main className="mx-auto grid w-full max-w-6xl gap-6 px-4 py-6 sm:px-6 lg:grid-cols-[240px_minmax(0,1fr)]">
        <aside className="hidden lg:block">
          <LeftNav user={viewer ?? undefined} activeHref="/notifications" variant="dark" />
        </aside>

        <section className="space-y-5">
          <div className="rounded-3xl border border-white/10 bg-white/5 p-6 backdrop-blur-sm">
            <div className="flex items-center justify-between">
              <div>
                <h1 className="text-xl font-semibold tracking-tight text-white">
                  Notifications
                </h1>
                <p className="text-sm text-neutral-400">
                  All activity that needs your attention.
                </p>
              </div>
              <button
                type="button"
                onClick={markAllRead}
                className="rounded-full border border-white/20 bg-white/5 px-4 py-2 text-xs font-semibold text-neutral-300 transition hover:bg-white/10 hover:text-white"
              >
                Mark all read
              </button>
            </div>
          </div>

          <div className="rounded-3xl border border-white/10 bg-white/5 p-6 backdrop-blur-sm">
            {loading ? (
              <p className="text-sm text-neutral-400">Loading notifications...</p>
            ) : loadingViewer ? (
              <p className="text-sm text-neutral-400">Loading account...</p>
            ) : notifications.length === 0 ? (
              <p className="text-sm text-neutral-400">No notifications yet.</p>
            ) : (
              <div className="space-y-3">
                {notifications.map((item) => {
                  const href = notificationHref(item);
                  return (
                    <button
                      key={item.id}
                      type="button"
                      onClick={async () => {
                        await markRead(item.id);
                        if (href) {
                          router.push(href);
                        }
                      }}
                      className={`flex w-full flex-col rounded-2xl border px-4 py-3 text-left text-sm transition ${
                        item.is_read
                          ? "border-white/10 bg-white/5 text-neutral-400"
                          : "border-emerald-500/30 bg-emerald-500/20 text-emerald-400 shadow-[0_0_0_1px_rgba(16,185,129,0.25)]"
                      }`}
                    >
                      <span className="text-[11px] font-semibold uppercase tracking-wide">
                        {notificationTitle(item)}
                      </span>
                      <span className="mt-1 text-sm">{notificationBody(item)}</span>
                    </button>
                  );
                })}
              </div>
            )}
          </div>
        </section>
      </main>
    </div>
  );
}
