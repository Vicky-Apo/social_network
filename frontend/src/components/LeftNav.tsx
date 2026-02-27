"use client";

import Link from "next/link";
import { Compass, MessageSquare, UserPlus, Users } from "lucide-react";
import { useCachedAvatar } from "@/lib/useCachedAvatar";

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

function toMediaUrl(apiBaseUrl: string, path?: string | null) {
  if (!path) return "";
  if (path.startsWith("http://") || path.startsWith("https://")) return path;
  const normalized = path.startsWith("/") ? path : `/${path}`;
  return `${apiBaseUrl}${normalized}`;
}

export default function LeftNav({ user, activeHref }: Props) {
  const apiBaseUrl =
    process.env.NEXT_PUBLIC_API_BASE_URL?.trim().replace(/\/+$/, "") ||
    "http://localhost:8080";

  const displayName = user ? `${user.first_name ?? ""} ${user.last_name ?? ""}`.trim() : "Loading";
  const userTag = user?.nickname || (user?.email ? user.email.split("@")[0] : "member");
  const avatarPath = useCachedAvatar(user?.id ?? null, user?.avatar_path ?? null);

  return (
    <div className="rounded-3xl border border-neutral-200 bg-white p-5 shadow-sm">
      <div className="flex items-center gap-3">
        {avatarPath ? (
          <div className="h-11 w-11 overflow-hidden rounded-full border border-neutral-200 bg-white">
            <img
              src={toMediaUrl(apiBaseUrl, avatarPath)}
              alt={displayName}
              className="h-full w-full object-contain"
            />
          </div>
        ) : (
          <div className="inline-flex h-11 w-11 items-center justify-center rounded-full bg-neutral-900 text-sm font-semibold text-white">
            {initials(user?.first_name, user?.last_name)}
          </div>
        )}
        <div>
          <p className="text-sm font-semibold text-neutral-900">{displayName || "Loading"}</p>
          <p className="text-xs text-neutral-500">@{userTag}</p>
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
              className={`flex items-center gap-2 rounded-2xl border px-3 py-2 text-sm transition ${
                isActive
                  ? "brand-gradient border-transparent text-white"
                  : "border-neutral-200 bg-neutral-50 text-neutral-700 hover:border-neutral-400 hover:text-neutral-900"
              }`}
            >
              <Icon className="h-4 w-4" />
              <span>{item.label}</span>
            </Link>
          );
        })}
      </nav>
    </div>
  );
}
