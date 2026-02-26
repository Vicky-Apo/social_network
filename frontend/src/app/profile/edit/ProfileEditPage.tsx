"use client";
/* eslint-disable @next/next/no-img-element */

import { useCallback, useEffect, useMemo, useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { Compass, MessageSquare, UserPlus, Users } from "lucide-react";
import { motion } from "framer-motion";
import TopNav from "../../component/TopNav";
import { fadeUp, viewportOnce } from "@/components/Motion";

type ApiResponse<T> = {
  success?: boolean;
  data?: T;
  error?: string;
};

type MeUser = {
  id: number;
  email: string;
  first_name: string;
  last_name: string;
  nickname?: string | null;
};

type ProfileDTO = {
  user: {
    id: number;
    email?: string | null;
    first_name: string;
    last_name: string;
    date_of_birth?: string | null;
    avatar_path?: string | null;
    nickname?: string | null;
    about?: string | null;
    is_public: boolean;
  };
};

const quickLinks = [
  { label: "Explore", href: "/dashboard", icon: Compass },
  { label: "Groups", href: "/groups", icon: Users },
  { label: "Messages", href: "/messages", icon: MessageSquare },
  { label: "Requests", href: "/follow-requests", icon: UserPlus },
];

function initials(first?: string, last?: string) {
  const left = first?.trim().charAt(0) ?? "";
  const right = last?.trim().charAt(0) ?? "";
  return `${left}${right}`.toUpperCase() || "U";
}

export default function ProfileEditPage() {
  const router = useRouter();
  const [viewer, setViewer] = useState<MeUser | null>(null);
  const [profile, setProfile] = useState<ProfileDTO | null>(null);
  const [nickname, setNickname] = useState("");
  const [about, setAbout] = useState("");
  const [avatarPath, setAvatarPath] = useState("");
  const [avatarFile, setAvatarFile] = useState<File | null>(null);
  const [isPublic, setIsPublic] = useState(false);
  const [isLoading, setIsLoading] = useState(true);
  const [isSaving, setIsSaving] = useState(false);
  const [isUploading, setIsUploading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);

  const apiBaseUrl = useMemo(
    () =>
      process.env.NEXT_PUBLIC_API_BASE_URL?.trim().replace(/\/+$/, "") ||
      "http://localhost:8080",
    [],
  );

  const loadProfile = useCallback(async () => {
    setIsLoading(true);
    setError(null);
    setSuccess(null);

    try {
      const meResponse = await fetch(`${apiBaseUrl}/auth/me`, {
        credentials: "include",
      });
      const meResult = (await meResponse.json().catch(() => null)) as ApiResponse<MeUser> | null;
      if (!meResponse.ok || !meResult?.success || !meResult.data) {
        router.replace("/login");
        return;
      }
      setViewer(meResult.data);

      const profileResponse = await fetch(`${apiBaseUrl}/profiles/${meResult.data.id}`, {
        credentials: "include",
      });
      const profileResult = (await profileResponse.json().catch(() => null)) as
        | ApiResponse<ProfileDTO>
        | null;
      if (!profileResponse.ok || !profileResult?.success || !profileResult.data) {
        setError(profileResult?.error || "Could not load profile.");
        return;
      }

      setProfile(profileResult.data);
      setNickname(profileResult.data.user.nickname ?? "");
      setAbout(profileResult.data.user.about ?? "");
      setAvatarPath(profileResult.data.user.avatar_path ?? "");
      setIsPublic(Boolean(profileResult.data.user.is_public));
    } catch {
      setError("Network error. Please try again.");
    } finally {
      setIsLoading(false);
    }
  }, [apiBaseUrl, router]);

  useEffect(() => {
    void loadProfile();
  }, [loadProfile]);

  const handleSave = async () => {
    if (!viewer) return;
    setIsSaving(true);
    setError(null);
    setSuccess(null);

    try {
      let nextAvatarPath = avatarPath.trim();
      if (avatarFile) {
        setIsUploading(true);
        const formData = new FormData();
        formData.append("file", avatarFile);
        formData.append("kind", "avatar");
        const uploadRes = await fetch(`${apiBaseUrl}/uploads`, {
          method: "POST",
          credentials: "include",
          body: formData,
        });
        const uploadJson = (await uploadRes.json().catch(() => null)) as
          | ApiResponse<{ path?: string }>
          | null;
        if (!uploadRes.ok || !uploadJson?.success || !uploadJson.data?.path) {
          setError(uploadJson?.error || "Could not upload avatar.");
          return;
        }
        nextAvatarPath = uploadJson.data.path;
        setAvatarPath(nextAvatarPath);
      }

      const updateRes = await fetch(`${apiBaseUrl}/profiles/${viewer.id}`, {
        method: "PATCH",
        headers: { "Content-Type": "application/json" },
        credentials: "include",
        body: JSON.stringify({
          nickname: nickname.trim() ? nickname.trim() : "",
          about: about.trim() ? about.trim() : "",
          avatar_path: nextAvatarPath ? nextAvatarPath : "",
        }),
      });
      const updateJson = (await updateRes.json().catch(() => null)) as ApiResponse<unknown> | null;
      if (!updateRes.ok || !updateJson?.success) {
        setError(updateJson?.error || "Could not update profile.");
        return;
      }

      const visibilityRes = await fetch(`${apiBaseUrl}/profiles/${viewer.id}/visibility`, {
        method: "PATCH",
        headers: { "Content-Type": "application/json" },
        credentials: "include",
        body: JSON.stringify({ is_public: isPublic }),
      });
      const visibilityJson = (await visibilityRes.json().catch(() => null)) as
        | ApiResponse<unknown>
        | null;
      if (!visibilityRes.ok || !visibilityJson?.success) {
        setError(visibilityJson?.error || "Could not update visibility.");
        return;
      }

      setSuccess("Profile updated successfully.");
      setTimeout(() => {
        router.replace(`/profile/${viewer.id}`);
      }, 300);
    } catch {
      setError("Network error. Please try again.");
    } finally {
      setIsSaving(false);
      setIsUploading(false);
      setAvatarFile(null);
    }
  };

  const displayName = viewer ? `${viewer.first_name} ${viewer.last_name}` : "Loading";
  const userTag =
    viewer?.nickname || (viewer?.email ? viewer.email.split("@")[0] : "member");

  return (
    <div className="min-h-screen bg-neutral-50 text-neutral-900">
      <TopNav user={viewer ?? undefined} onLogout={() => router.replace("/login")} />

      <main className="mx-auto grid w-full max-w-6xl gap-6 px-4 py-6 sm:px-6 lg:grid-cols-[240px_minmax(0,1fr)_280px]">
        <aside className="hidden lg:block">
          <div className="rounded-3xl border border-neutral-200 bg-white p-5 shadow-sm">
            <div className="flex items-center gap-3">
              <div className="inline-flex h-11 w-11 items-center justify-center rounded-full bg-neutral-900 text-sm font-semibold text-white">
                {initials(viewer?.first_name, viewer?.last_name)}
              </div>
              <div>
                <p className="text-sm font-semibold text-neutral-900">{displayName}</p>
                <p className="text-xs text-neutral-500">@{userTag}</p>
              </div>
            </div>
            <nav className="mt-5 space-y-2">
              {quickLinks.map((item) => {
                const Icon = item.icon;
                return (
                  <Link
                    key={item.label}
                    href={item.href}
                    className="flex items-center gap-2 rounded-2xl border border-neutral-200 bg-neutral-50 px-3 py-2 text-sm text-neutral-700 transition hover:border-neutral-400 hover:text-neutral-900"
                  >
                    <Icon className="h-4 w-4" />
                    <span>{item.label}</span>
                  </Link>
                );
              })}
            </nav>
          </div>
        </aside>

        <section className="space-y-5">
          <motion.div
            initial="hidden"
            whileInView="show"
            viewport={viewportOnce}
            variants={fadeUp}
            className="rounded-3xl border border-neutral-200 bg-white p-5 shadow-sm"
          >
            <h1 className="text-xl font-semibold tracking-tight text-neutral-900">Edit profile</h1>
            <p className="text-sm text-neutral-600">
              Update your profile details and visibility.
            </p>
          </motion.div>

          {isLoading ? (
            <motion.div
              initial="hidden"
              whileInView="show"
              viewport={viewportOnce}
              variants={fadeUp}
              className="rounded-3xl border border-neutral-200 bg-white p-5 text-sm text-neutral-600 shadow-sm"
            >
              Loading profile...
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
              <div className="grid gap-4 sm:grid-cols-2">
                <div className="space-y-2">
                  <label className="text-sm font-semibold text-neutral-700">Nickname</label>
                  <input
                    value={nickname}
                    onChange={(event) => setNickname(event.target.value)}
                    placeholder="Nickname"
                    className="h-12 w-full rounded-2xl border border-neutral-200 bg-white px-4 text-sm text-neutral-900 placeholder:text-neutral-400 focus:border-neutral-400 focus:outline-none"
                  />
                </div>
                <div className="space-y-2">
                  <label className="text-sm font-semibold text-neutral-700">Avatar</label>
                  <div className="flex flex-wrap items-center gap-3">
                    <label className="inline-flex items-center gap-2 rounded-full border border-neutral-200 bg-white px-3 py-2 text-xs font-semibold text-neutral-700 transition hover:border-neutral-400 hover:text-neutral-900">
                      <input
                        type="file"
                        accept="image/png,image/jpeg,image/gif"
                        className="hidden"
                        onChange={(event) => {
                          const file = event.target.files?.[0] ?? null;
                          setAvatarFile(file);
                          if (!file) return;
                          setAvatarPath(file.name);
                        }}
                      />
                      Choose file
                    </label>
                    <span className="text-xs text-neutral-500">
                      {avatarFile ? avatarFile.name : avatarPath || "No file selected"}
                    </span>
                  </div>
                </div>
              </div>

              <div className="mt-4 space-y-2">
                <label className="text-sm font-semibold text-neutral-700">About</label>
                <textarea
                  value={about}
                  onChange={(event) => setAbout(event.target.value)}
                  rows={4}
                  placeholder="Tell people about yourself"
                  className="w-full rounded-2xl border border-neutral-200 bg-white px-4 py-3 text-sm text-neutral-900 placeholder:text-neutral-400 focus:border-neutral-400 focus:outline-none"
                />
              </div>

              <div className="mt-4 flex items-center justify-between rounded-2xl border border-neutral-200 bg-neutral-50 px-4 py-3">
                <div>
                  <p className="text-sm font-semibold text-neutral-900">Profile visibility</p>
                  <p className="text-xs text-neutral-500">
                    {isPublic ? "Public profile" : "Private profile"}
                  </p>
                </div>
                <label className="inline-flex items-center gap-2 text-xs text-neutral-600">
                  Public
                  <input
                    type="checkbox"
                    checked={isPublic}
                    onChange={(event) => setIsPublic(event.target.checked)}
                    className="h-4 w-4 rounded border-neutral-300 text-neutral-900 focus:ring-neutral-900"
                  />
                </label>
              </div>

              <div className="mt-5 flex flex-wrap items-center gap-3">
                <button
                  type="button"
                  onClick={handleSave}
                  disabled={isSaving || isUploading}
                  className="brand-gradient inline-flex items-center gap-2 rounded-full px-4 py-2 text-xs font-semibold text-white transition hover:-translate-y-0.5 hover:shadow-md disabled:cursor-not-allowed disabled:opacity-70"
                >
                  {isSaving || isUploading ? "Saving..." : "Save changes"}
                </button>
                <Link
                  href={`/profile/${viewer?.id ?? ""}`}
                  className="inline-flex items-center gap-2 rounded-full border border-neutral-200 bg-white px-3 py-2 text-xs font-semibold text-neutral-700 transition hover:border-neutral-400 hover:text-neutral-900"
                >
                  Back to profile
                </Link>
                {success ? <span className="text-xs text-emerald-600">{success}</span> : null}
              </div>
            </motion.div>
          )}
        </section>

        <aside className="hidden lg:block">
          <div className="rounded-3xl border border-neutral-200 bg-white p-5 shadow-sm">
            <h3 className="text-sm font-semibold text-neutral-900">Profile tips</h3>
            <p className="mt-2 text-xs text-neutral-500">
              Use a friendly nickname and a short bio to help others recognize you.
            </p>
          </div>
        </aside>
      </main>
    </div>
  );
}
