"use client";

import { useEffect, useMemo } from "react";
import { useRouter } from "next/navigation";
import { Navbar } from "@/components/Navbar";
import { Hero } from "@/components/Hero";
import { apiFetchJson, getApiBaseUrl } from "@/lib/api";

export default function HomePage() {
  const router = useRouter();
  const apiBaseUrl = useMemo(() => getApiBaseUrl(), []);

  useEffect(() => {
    let cancelled = false;
    const checkSession = async () => {
      try {
        const { response, result } = await apiFetchJson<{ success?: boolean }>(
          "/auth/me",
          {},
          apiBaseUrl,
        );
        if (!cancelled && response.ok && result?.success) {
          router.replace("/dashboard");
        }
      } catch {
        // If the check fails, keep the landing page.
      }
    };

    void checkSession();
    return () => {
      cancelled = true;
    };
  }, [apiBaseUrl, router]);

  return (
    <div className="min-h-screen bg-[#2b2929] text-neutral-100">
      <Navbar />
      <main>
        <Hero />
      </main>
    </div>
  );
}
