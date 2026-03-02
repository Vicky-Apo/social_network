"use client";

import Link from "next/link";
import { Compass, Home, MessageSquare, UserPlus, Users } from "lucide-react";
import { useCachedAvatar } from "@/lib/useCachedAvatar";
import Avatar from "@/components/Avatar";

type LeftNavUser = {
  id: number;
  email?: string;
  first_name?: string;
  last_name?: string;
  nickname?: string | null;
  avatar_path?: string | null;
};

type Props = {
  user?: LeftNavUser | null;
  activeHref?: string;
  variant?: "light" | "dark";
};

const quickLinks = [
  { label: "Dashboard", href: "/dashboard", icon: Home },
  { label: "Explore", href: "/explore", icon: Compass },
  { label: "Groups", href: "/groups", icon: Users },
  { label: "Messages", href: "/messages", icon: MessageSquare },
  { label: "Requests", href: "/follow-requests", icon: UserPlus },
];

function toMediaUrl(apiBaseUrl: string, path?: string | null) {
  if (!path) return "";
  if (path.startsWith("http://") || path.startsWith("https://")) return path;
  const normalized = path.startsWith("/") ? path : `/${path}`;
  return `${apiBaseUrl}${normalized}`;
}

export default function LeftNav({ user, activeHref, variant = "light" }: Props) {
  const apiBaseUrl =
    process.env.NEXT_PUBLIC_API_BASE_URL?.trim().replace(/\/+$/, "") ||
    "http://localhost:8080";

  const displayName = user ? `${user.first_name ?? ""} ${user.last_name ?? ""}`.trim() : "Loading";
  const userTag = user?.nickname || (user?.email ? user.email.split("@")[0] : "member");
  const avatarPath = useCachedAvatar(user?.id ?? null, user?.avatar_path ?? null);

  const isDark = variant === "dark";

  return (
    <div
      className={
        isDark
          ? "rounded-2xl border border-white/10 bg-white/5 p-4 backdrop-blur-sm"
          : "rounded-3xl border border-neutral-200 bg-white p-5 shadow-sm"
      }
    >
      <div className="flex items-center gap-3">
        <Avatar
          src={avatarPath ? toMediaUrl(apiBaseUrl, avatarPath) : null}
          name={displayName}
          size={44}
          textClassName="text-sm"
        />
        <div>
          <p className={`text-sm font-semibold ${isDark ? "text-white" : "text-neutral-900"}`}>
            {displayName || "Loading"}
          </p>
          <p className={`text-xs ${isDark ? "text-white" : "text-neutral-500"}`}>@{userTag}</p>
        </div>
      </div>
      <nav className="mt-5 space-y-2">
        {quickLinks.map((item) => {
          const Icon = item.icon;
          const isActive = activeHref === item.href;
          return (
            <Link
              key={item.label}
              href={item.href}
              className={
                isDark
                  ? "flex items-center gap-2 rounded-xl border border-neutral-200 bg-white px-3 py-2 text-sm font-medium !text-black transition hover:bg-neutral-100 [&_svg]:!text-black"
                  : `flex items-center gap-2 rounded-2xl border px-3 py-2 text-sm transition ${
                      isActive
                        ? "brand-gradient border-transparent text-white"
                        : "border-neutral-200 bg-neutral-50 text-neutral-700 hover:border-neutral-400 hover:text-neutral-900"
                    }`
              }
            >
              <Icon className="h-4 w-4 shrink-0" />
              <span className={isDark ? "!text-black" : undefined}>{item.label}</span>
            </Link>
          );
        })}
      </nav>
    </div>
  );
}
