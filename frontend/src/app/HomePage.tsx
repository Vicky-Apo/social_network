"use client";

import { useEffect, useMemo } from "react";
import { useRouter } from "next/navigation";
import { Navbar } from "@/components/Navbar";
import { Hero } from "@/components/Hero";
import { Workflow } from "@/components/Workflow";
import { Stats } from "@/components/Stats";
import { CTA } from "@/components/CTA";
import { Testimonials } from "@/components/Testimonials";
import { Articles } from "@/components/Articles";
import { Footer } from "@/components/Footer";

export default function HomePage() {
  const router = useRouter();
  const apiBaseUrl = useMemo(
    () =>
      process.env.NEXT_PUBLIC_API_BASE_URL?.trim().replace(/\/+$/, "") ||
      "http://localhost:8080",
    [],
  );

  useEffect(() => {
    let cancelled = false;
    const checkSession = async () => {
      try {
        const response = await fetch(`${apiBaseUrl}/auth/me`, {
          credentials: "include",
        });
        const result = (await response.json().catch(() => null)) as
          | { success?: boolean }
          | null;
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
    <div className="min-h-screen bg-neutral-50 text-neutral-900">
      <Navbar />
      <main>
        <Hero />
        <Workflow />
        <Stats />
        <CTA />
        <Testimonials />
        <Articles />
      </main>
      <Footer />
    </div>
  );
}
