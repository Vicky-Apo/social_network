"use client";

import { useEffect, useMemo, useState } from "react";
import { useRouter } from "next/navigation";
import TopNav from "@/components/TopNav";
import LeftNav from "@/components/LeftNav";
import { useNotifications } from "@/components/NotificationsContext";
import { apiFetchJson, getApiBaseUrl } from "@/lib/api";
import {
  allowedNotificationTypes,
  getNotificationBody,
  getNotificationHref,
  getNotificationTitle,
} from "@/lib/notifications";
import { ApiResponse } from "@/lib/types";

export default function NotificationsPage() {
  const router = useRouter();
  const notificationsContext = useNotifications();
  const [viewer, setViewer] = useState<{
    id: number;
    email: string;
    first_name: string;
    last_name: string;
  } | null>(null);

  const notifications = (notificationsContext?.notifications ?? []).filter((item) =>
    allowedNotificationTypes.has(item.type),
  );
  const loading = notificationsContext?.loading ?? false;
  const markRead = notificationsContext?.markRead ?? (async () => Promise.resolve());
  const markAllRead = notificationsContext?.markAllRead ?? (async () => Promise.resolve());

  const apiBaseUrl = useMemo(() => getApiBaseUrl(), []);

  useEffect(() => {
    let cancelled = false;
    const loadViewer = async () => {
      const { response, result } = await apiFetchJson<ApiResponse<{
        id: number;
        email: string;
        first_name: string;
        last_name: string;
      }>>("/auth/me", {}, apiBaseUrl);
      if (!response.ok || !result?.success || !result.data) {
        router.replace("/login");
        return;
      }
      if (!cancelled) {
        setViewer(result.data);
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
        onLogout={() => router.replace("/login")}
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
            ) : notifications.length === 0 ? (
              <p className="text-sm text-neutral-400">No notifications yet.</p>
            ) : (
              <div className="space-y-3">
                {notifications.map((item) => {
                  const href = getNotificationHref(item);
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
                        {getNotificationTitle(item.type)}
                      </span>
                      <span className="mt-1 text-sm">{getNotificationBody(item.type, item.metadata)}</span>
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
