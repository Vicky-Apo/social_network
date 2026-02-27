"use client";

import { useMemo, useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { ArrowLeft } from "lucide-react";
import { motion } from "framer-motion";
import TopNav from "@/components/TopNav";
import LeftNav from "@/components/LeftNav";
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

export default function CreateGroupPage() {
  const router = useRouter();
  const [viewer, setViewer] = useState<User | null>(null);
  const [title, setTitle] = useState("");
  const [description, setDescription] = useState("");
  const [isSaving, setIsSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);

  const apiBaseUrl = useMemo(
    () =>
      process.env.NEXT_PUBLIC_API_BASE_URL?.trim().replace(/\/+$/, "") ||
      "http://localhost:8080",
    [],
  );

  const loadViewer = async () => {
    if (viewer) return;
    const response = await fetch(`${apiBaseUrl}/auth/me`, { credentials: "include" });
    const result = (await response.json().catch(() => null)) as ApiResponse<User> | null;
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
      const response = await fetch(`${apiBaseUrl}/groups`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        credentials: "include",
        body: JSON.stringify({
          title: cleanTitle,
          description: cleanDescription || undefined,
        }),
      });
      const result = (await response.json().catch(() => null)) as ApiResponse<{ id?: number }> | null;
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

  if (!viewer) {
    void loadViewer();
  }

  return (
    <div className="min-h-screen bg-neutral-50 text-neutral-900">
      <TopNav user={viewer ?? undefined} onLogout={() => router.replace("/login")} />

      <main className="mx-auto grid w-full max-w-5xl gap-6 px-4 py-6 sm:px-6 lg:grid-cols-[240px_minmax(0,1fr)]">
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
            <div className="flex items-center justify-between gap-3">
              <div>
                <h1 className="text-xl font-semibold tracking-tight text-neutral-900">
                  Create a group
                </h1>
                <p className="text-sm text-neutral-600">
                  Start a new space for your community.
                </p>
              </div>
              <Link
                href="/groups"
                className="inline-flex items-center gap-2 rounded-full border border-neutral-200 bg-white px-3 py-2 text-xs font-semibold text-neutral-700 transition hover:border-neutral-400 hover:text-neutral-900"
              >
                <ArrowLeft className="h-3.5 w-3.5" />
                Back to groups
              </Link>
            </div>

            <div className="mt-5 space-y-4">
              <label className="block text-xs font-semibold text-neutral-600">
                Group title
                <input
                  value={title}
                  onChange={(event) => setTitle(event.target.value)}
                  placeholder="e.g., Go Builders"
                  className="mt-2 h-11 w-full rounded-2xl border border-neutral-200 bg-white px-4 text-sm text-neutral-900 outline-none transition focus:border-neutral-400"
                />
              </label>
              <label className="block text-xs font-semibold text-neutral-600">
                Description
                <textarea
                  value={description}
                  onChange={(event) => setDescription(event.target.value)}
                  rows={4}
                  placeholder="What is this group about?"
                  className="mt-2 w-full rounded-2xl border border-neutral-200 bg-white px-4 py-3 text-sm text-neutral-900 outline-none transition focus:border-neutral-400"
                />
              </label>
            </div>

            {error ? <p className="mt-3 text-xs text-rose-600">{error}</p> : null}
            {success ? <p className="mt-3 text-xs text-emerald-600">{success}</p> : null}

            <button
              type="button"
              onClick={handleCreate}
              disabled={isSaving}
              className="mt-5 rounded-full bg-neutral-900 px-4 py-2 text-xs font-semibold text-white transition hover:bg-neutral-800 disabled:cursor-not-allowed disabled:opacity-70"
            >
              {isSaving ? "Creating..." : "Create group"}
            </button>
          </motion.div>
        </section>
      </main>
    </div>
  );
}
