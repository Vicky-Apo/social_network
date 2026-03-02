"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import Image from "next/image";
import { Menu, X, ArrowRight } from "lucide-react";
import { AnimatePresence, motion, useReducedMotion } from "framer-motion";
import clsx from "clsx";
import { landingData } from "@/lib/data";

export function Navbar() {
  const [isOpen, setIsOpen] = useState(false);
  const [isScrolled, setIsScrolled] = useState(false);
  const reducedMotion = useReducedMotion();

  useEffect(() => {
    const handleScroll = () => setIsScrolled(window.scrollY > 8);
    handleScroll();
    window.addEventListener("scroll", handleScroll, { passive: true });
    return () => window.removeEventListener("scroll", handleScroll);
  }, []);

  useEffect(() => {
    const closeOnResize = () => {
      if (window.innerWidth >= 768) setIsOpen(false);
    };
    window.addEventListener("resize", closeOnResize);
    return () => window.removeEventListener("resize", closeOnResize);
  }, []);

  return (
    <header
      className={clsx(
        "fixed inset-x-0 top-0 z-50 border-b border-white/[0.06] transition-all duration-300 ease-out",
        isScrolled
          ? "bg-[#2b2929]/90 shadow-[0_1px_0_0_rgba(255,255,255,0.03)_inset] backdrop-blur-xl"
          : "bg-transparent",
      )}
    >
      <div className="mx-auto flex h-[72px] w-full max-w-6xl items-center justify-between px-5 sm:px-8">
        <Link
          href="/"
          className="group relative inline-flex items-center p-2 -m-2 transition-opacity hover:opacity-90"
        >
          <Image
            src="/vybez-logo.png"
            alt={`${landingData.productName} logo`}
            width={110}
            height={40}
            className="h-10 w-auto object-contain"
            priority
          />
        </Link>

        <nav className="hidden items-center gap-3 md:flex">
          <Link
            href="/login"
            className="rounded-xl px-5 py-2.5 text-sm font-medium text-neutral-300 transition-colors hover:bg-white/[0.04] hover:text-white"
          >
            Sign in
          </Link>
          <Link
            href="/register"
            className="group inline-flex items-center gap-2 rounded-xl bg-white px-5 py-2.5 text-sm font-medium text-black transition-all hover:bg-neutral-100 hover:shadow-lg hover:shadow-white/5"
          >
            <span className="text-black">Get started</span>
            <ArrowRight className="h-4 w-4 text-black transition-transform duration-200 group-hover:translate-x-0.5" />
          </Link>
        </nav>

        <button
          type="button"
          className="inline-flex items-center justify-center rounded-xl p-2.5 text-neutral-300 transition-colors hover:bg-white/[0.04] hover:text-white md:hidden"
          onClick={() => setIsOpen((value) => !value)}
          aria-expanded={isOpen}
          aria-controls="mobile-menu"
          aria-label={isOpen ? "Close menu" : "Open menu"}
        >
          {isOpen ? <X className="h-5 w-5" /> : <Menu className="h-5 w-5" />}
        </button>
      </div>

      <AnimatePresence initial={false}>
        {isOpen && (
          <motion.div
            id="mobile-menu"
            initial={reducedMotion ? false : { opacity: 0, y: -8 }}
            animate={reducedMotion ? { opacity: 1 } : { opacity: 1, y: 0 }}
            exit={reducedMotion ? { opacity: 0 } : { opacity: 0, y: -8 }}
            transition={{ duration: reducedMotion ? 0.01 : 0.2 }}
            className="border-t border-white/[0.06] bg-[#2b2929] px-5 pb-6 pt-4 md:hidden"
          >
            <nav className="flex flex-col gap-2" aria-label="Mobile navigation">
              <Link
                href="/login"
                onClick={() => setIsOpen(false)}
                className="rounded-xl px-4 py-3 text-sm font-medium text-neutral-300 transition-colors hover:bg-white/[0.04] hover:text-white"
              >
                Sign in
              </Link>
              <Link
                href="/register"
                onClick={() => setIsOpen(false)}
                className="inline-flex items-center justify-center gap-2 rounded-xl bg-white px-4 py-3 text-sm font-medium text-black transition-colors hover:bg-neutral-100"
              >
                <span className="text-black">Get started</span>
                <ArrowRight className="h-4 w-4 text-black" />
              </Link>
            </nav>
          </motion.div>
        )}
      </AnimatePresence>
    </header>
  );
}
