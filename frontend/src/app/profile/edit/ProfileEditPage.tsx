"use client";
/* eslint-disable @next/next/no-img-element */

import { useCallback, useEffect, useMemo, useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { motion } from "framer-motion";
import TopNav from "@/components/TopNav";
import LeftNav from "@/components/LeftNav";
import { fadeUp, viewportOnce } from "@/components/Motion";
import { apiFetchJson, getApiBaseUrl } from "@/lib/api";
import { ApiResponse } from "@/lib/types";

type MeUser = {
  id: number;
  email: string;
  first_name: string;
  last_name: string;
  nickname?: string | null;
  avatar_path?: string | null;
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

  const apiBaseUrl = useMemo(() => getApiBaseUrl(), []);

  const loadProfile = useCallback(async () => {
    setIsLoading(true);
    setError(null);
    setSuccess(null);

    try {
      const { response: meResponse, result: meResult } = await apiFetchJson<ApiResponse<MeUser>>(
        "/auth/me",
        {},
        apiBaseUrl,
      );
      if (!meResponse.ok || !meResult?.success || !meResult.data) {
        router.replace("/login");
        return;
      }
      setViewer(meResult.data);
      const viewerID = meResult.data.id;
      const { response: profileResponse, result: profileResult } = await apiFetchJson<
        ApiResponse<ProfileDTO>
      >(`/profiles/${viewerID}`, {}, apiBaseUrl);
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
        const { response: uploadRes, result: uploadJson } = await apiFetchJson<
          ApiResponse<{ path?: string }>
        >(
          "/uploads",
          {
            method: "POST",
            body: formData,
          },
          apiBaseUrl,
        );
        if (!uploadRes.ok || !uploadJson?.success || !uploadJson.data?.path) {
          setError(uploadJson?.error || "Could not upload avatar.");
          return;
        }
        nextAvatarPath = uploadJson.data.path;
        setAvatarPath(nextAvatarPath);
      }

      const { response: updateRes, result: updateJson } = await apiFetchJson<ApiResponse<unknown>>(
        `/profiles/${viewer.id}`,
        {
          method: "PATCH",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({
            nickname: nickname.trim() ? nickname.trim() : "",
            about: about.trim() ? about.trim() : "",
            avatar_path: nextAvatarPath ? nextAvatarPath : "",
          }),
        },
        apiBaseUrl,
      );
      if (!updateRes.ok || !updateJson?.success) {
        setError(updateJson?.error || "Could not update profile.");
        return;
      }

      if (profile?.user.is_public !== isPublic) {
        const { response: visibilityRes, result: visibilityJson } = await apiFetchJson<
          ApiResponse<unknown>
        >(
          `/profiles/${viewer.id}/visibility`,
          {
            method: "PATCH",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({ is_public: isPublic }),
          },
          apiBaseUrl,
        );
        if (!visibilityRes.ok || !visibilityJson?.success) {
          setError(visibilityJson?.error || "Could not update visibility.");
          return;
        }
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
          <LeftNav user={viewer ?? undefined} activeHref="/dashboard" variant="dark" />
        </aside>

        <section className="space-y-5">
          <motion.div
            initial="hidden"
            whileInView="show"
            viewport={viewportOnce}
            variants={fadeUp}
            className="rounded-3xl border border-white/10 bg-white/5 p-6 backdrop-blur-sm"
          >
            <h1 className="text-xl font-semibold tracking-tight text-white">Edit profile</h1>
            <p className="text-sm text-neutral-400">
              Update your profile details and visibility.
            </p>
          </motion.div>

          {isLoading ? (
            <motion.div
              initial="hidden"
              whileInView="show"
              viewport={viewportOnce}
              variants={fadeUp}
              className="rounded-3xl border border-white/10 bg-white/5 p-6 text-sm text-neutral-400 backdrop-blur-sm"
            >
              Loading profile...
            </motion.div>
          ) : error ? (
            <motion.div
              initial="hidden"
              whileInView="show"
              viewport={viewportOnce}
              variants={fadeUp}
              className="rounded-3xl border border-rose-500/30 bg-rose-500/10 p-6 text-sm text-rose-400"
            >
              {error}
            </motion.div>
          ) : (
            <motion.div
              initial="hidden"
              whileInView="show"
              viewport={viewportOnce}
              variants={fadeUp}
              className="rounded-3xl border border-white/10 bg-white/5 p-6 backdrop-blur-sm"
            >
              <div className="grid gap-4 sm:grid-cols-2">
                <div className="space-y-2">
                  <label className="text-sm font-semibold text-white">Nickname</label>
                  <input
                    value={nickname}
                    onChange={(event) => setNickname(event.target.value)}
                    placeholder="Nickname"
                    className="h-12 w-full rounded-2xl border border-neutral-200 bg-white px-4 text-sm text-black placeholder:text-neutral-500 focus:border-neutral-400 focus:outline-none"
                  />
                </div>
                <div className="space-y-2">
                  <label className="text-sm font-semibold text-white">Avatar</label>
                  <div className="flex flex-wrap items-center gap-3">
                    <label className="inline-flex items-center gap-2 rounded-full border border-white/20 bg-white/5 px-3 py-2 text-xs font-semibold text-neutral-300 transition hover:bg-white/10 hover:text-white">
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
                    <span className="text-xs text-neutral-400">
                      {avatarFile ? avatarFile.name : avatarPath || "No file selected"}
                    </span>
                  </div>
                </div>
              </div>

              <div className="mt-4 space-y-2">
                <label className="text-sm font-semibold text-white">About</label>
                <textarea
                  value={about}
                  onChange={(event) => setAbout(event.target.value)}
                  rows={4}
                  placeholder="Tell people about yourself"
                  className="w-full rounded-2xl border border-neutral-200 bg-white px-4 py-3 text-sm text-black placeholder:text-neutral-500 focus:border-neutral-400 focus:outline-none"
                />
              </div>

              <div className="mt-4 flex items-center justify-between rounded-2xl border border-white/10 bg-white/5 px-4 py-3">
                <div>
                  <p className="text-sm font-semibold text-white">Profile visibility</p>
                  <p className="text-xs text-neutral-400">
                    {isPublic ? "Public profile" : "Private profile"}
                  </p>
                </div>
                <label className="inline-flex items-center gap-2 text-xs text-neutral-300">
                  Public
                  <input
                    type="checkbox"
                    checked={isPublic}
                    onChange={(event) => setIsPublic(event.target.checked)}
                    className="h-4 w-4 rounded border-neutral-400 text-neutral-900 focus:ring-neutral-900"
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
                  className="inline-flex items-center gap-2 rounded-full border-2 border-white bg-white px-3 py-2 text-xs font-semibold transition hover:bg-neutral-100"
                  style={{ color: "#000" }}
                >
                  Back to profile
                </Link>
                {success ? <span className="text-xs text-emerald-400">{success}</span> : null}
              </div>
            </motion.div>
          )}
        </section>
      </main>
    </div>
  );
}
