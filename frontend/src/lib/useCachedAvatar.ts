"use client";

import { useEffect, useState } from "react";

const keyFor = (userId: number) => `vybez.avatar.${userId}`;

export function useCachedAvatar(
  userId?: number | null,
  avatarPath?: string | null,
): string | null {
  const [cached, setCached] = useState<string | null>(null);

  useEffect(() => {
    if (typeof window === "undefined" || !userId) {
      setCached(null);
      return;
    }

    if (avatarPath) {
      localStorage.setItem(keyFor(userId), avatarPath);
      setCached(avatarPath);
      return;
    }

    const stored = localStorage.getItem(keyFor(userId));
    setCached(stored || null);
  }, [userId, avatarPath]);

  return avatarPath || cached;
}
