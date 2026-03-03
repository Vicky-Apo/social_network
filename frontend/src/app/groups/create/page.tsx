"use client";

import { useEffect, useMemo, useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { ArrowLeft } from "lucide-react";
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
  avatar_path?: string | null;
};

export default function CreateGroupPage() {
  const router = useRouter();
  const [viewer, setViewer] = useState<User | null>(null);
  const [title, setTitle] = useState("");
  const [description, setDescription] = useState("");
  const [isSaving, setIsSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);

  const apiBaseUrl = useMemo(() => getApiBaseUrl(), []);

  const loadViewer = async () => {
    if (viewer) return;
    const { response, result } = await apiFetchJson<ApiResponse<User>>("/auth/me", {}, apiBaseUrl);
    if (!response.ok || !result?.success || !result.data) {
      router.replace("/login");
      return;
    }
    setViewer(result.data);
  };

  const handleCreate = async () => {
    if (isSaving) return;
    const cleanTitle = title.trim();
    const cleanDescription = description.trim();
    if (!cleanTitle) {
      setError("Group title is required.");
      return;
    }
    setIsSaving(true);
    setError(null);
    setSuccess(null);
    try {
      const { response, result } = await apiFetchJson<ApiResponse<{ id?: number }>>(
        "/groups",
        {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({
            title: cleanTitle,
            description: cleanDescription || undefined,
          }),
        },
        apiBaseUrl,
      );
      if (!response.ok || !result?.success) {
        setError(result?.error || "Could not create group.");
        return;
      }
      setSuccess("Group created.");
      const groupID = result?.data?.id;
      if (typeof groupID === "number") {
        router.replace(`/groups/${groupID}`);
      }
    } catch {
      setError("Network error. Please try again.");
    } finally {
      setIsSaving(false);
    }
  };

  useEffect(() => {
    void loadViewer();
  }, []);

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

      <main className="mx-auto grid w-full max-w-5xl gap-6 px-4 py-6 sm:px-6 lg:grid-cols-[220px_minmax(0,1fr)]">
        <aside className="hidden lg:block">
          <LeftNav user={viewer ?? undefined} activeHref="/groups" variant="dark" />
        </aside>

        <section className="space-y-5">
          <motion.div
            initial="hidden"
            whileInView="show"
            viewport={viewportOnce}
            variants={fadeUp}
            className="rounded-2xl border border-white/10 bg-white/5 p-5 backdrop-blur-sm"
          >
            <div className="flex items-center justify-between gap-3">
              <div>
                <h1 className="text-xl font-semibold tracking-tight text-white">
                  Create a group
                </h1>
                <p className="text-sm text-neutral-400">
                  Start a new space for your community.
                </p>
              </div>
              <Link
                href="/groups"
                className="inline-flex items-center gap-2 rounded-xl border border-white/20 bg-white/5 px-3 py-2 text-xs font-semibold text-neutral-300 transition hover:bg-white/10 hover:text-white"
              >
                <ArrowLeft className="h-3.5 w-3.5" />
                Back to groups
              </Link>
            </div>

            <div className="mt-5 space-y-4">
              <label className="block text-xs font-semibold text-neutral-400">
                Group title
                <input
                  value={title}
                  onChange={(event) => setTitle(event.target.value)}
                  placeholder="e.g., Go Builders"
                  className="mt-2 h-11 w-full rounded-xl border border-white/20 bg-white/5 px-4 text-sm text-white placeholder:text-neutral-500 outline-none transition focus:border-white/40"
                />
              </label>
              <label className="block text-xs font-semibold text-neutral-400">
                Description
                <textarea
                  value={description}
                  onChange={(event) => setDescription(event.target.value)}
                  rows={4}
                  placeholder="What is this group about?"
                  className="mt-2 w-full rounded-xl border border-white/20 bg-white/5 px-4 py-3 text-sm text-white placeholder:text-neutral-500 outline-none transition focus:border-white/40"
                />
              </label>
            </div>

            {error ? <p className="mt-3 text-xs text-rose-400">{error}</p> : null}
            {success ? <p className="mt-3 text-xs text-emerald-400">{success}</p> : null}

            <button
              type="button"
              onClick={handleCreate}
              disabled={isSaving}
              className="mt-5 rounded-xl bg-white px-4 py-2 text-xs font-semibold text-[#2b2929] transition hover:bg-neutral-100 disabled:cursor-not-allowed disabled:opacity-70"
            >
              {isSaving ? "Creating..." : "Create group"}
            </button>
          </motion.div>
        </section>
      </main>
    </div>
  );
}
