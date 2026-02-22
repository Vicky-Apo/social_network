"use client";

import Link from "next/link";
import Image from "next/image";
import { landingData } from "@/lib/data";

export function Footer() {
  const productName = landingData.productName;
  const { productLinks, legalLinks, supportLinks } = landingData.footer;
  const allLinks = [...productLinks, ...legalLinks, ...supportLinks];

  return (
    <footer className="relative border-t border-white/10 bg-black/30 overflow-hidden">
      {/* Gradient köşe */}
      <div className="pointer-events-none absolute -bottom-20 -right-20 h-64 w-64 opacity-40">
        <Image
          src="/gradient-2.png"
          alt=""
          fill
          className="object-cover"
        />
      </div>

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
