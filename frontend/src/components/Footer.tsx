"use client";

import Link from "next/link";
import { landingData } from "@/lib/data";

export function Footer() {
  const productName = landingData.productName;
  const { productLinks, legalLinks, supportLinks } = landingData.footer;
  const allLinks = [...productLinks, ...legalLinks, ...supportLinks];

  return (
    <footer className="relative border-t border-white/10 bg-black/30 overflow-hidden">
      {/* CSS-only gradient corner */}
      <div
        className="pointer-events-none absolute -bottom-24 -right-24 h-48 w-48 rounded-full opacity-30 blur-3xl"
        style={{
          background:
            "radial-gradient(circle, rgba(168,85,247,0.4) 0%, rgba(59,130,246,0.2) 50%, transparent 70%)",
        }}
        aria-hidden
      />

      <div className="relative z-10 mx-auto flex w-full max-w-6xl flex-col items-center gap-5 px-4 py-8 sm:px-6 md:flex-row md:justify-between md:gap-8">
        <Link
          href="/"
          className="text-lg font-semibold text-white focus:outline-none focus-visible:ring-2 focus-visible:ring-white/50 focus-visible:ring-offset-2 focus-visible:ring-offset-[#2b2929]"
        >
          {productName}
        </Link>

        <nav className="flex flex-wrap items-center justify-center gap-x-6 gap-y-1 text-sm text-white/70">
          {allLinks.map((label) => (
            <Link key={label} href="#" className="transition hover:text-white">
              {label}
            </Link>
          ))}
        </nav>

        <p className="text-xs text-white/50">
          © {new Date().getFullYear()} {productName}
        </p>
      </div>
    </footer>
  );
}
