"use client";

import { useEffect, useMemo, useState } from "react";
import Link from "next/link";
import type { MouseEvent } from "react";
import { Menu, X, ArrowRight } from "lucide-react";
import { useReducedMotion } from "framer-motion";
import clsx from "clsx";
import { landingData } from "@/lib/data";
import { BrandMark } from "@/components/BrandMark";

export function Navbar() {
  const [isOpen, setIsOpen] = useState(false);
  const [isScrolled, setIsScrolled] = useState(false);
  const [activeSection, setActiveSection] = useState("");
  const reducedMotion = useReducedMotion();
  const navItems = useMemo(() => landingData.navItems, []);

  useEffect(() => {
    const handleScroll = () => setIsScrolled(window.scrollY > 8);
    handleScroll();
    window.addEventListener("scroll", handleScroll, { passive: true });
    return () => window.removeEventListener("scroll", handleScroll);
  }, []);

  useEffect(() => {
    const sections = navItems
      .map((item) => document.querySelector(item.href))
      .filter((node): node is HTMLElement => Boolean(node));

    const observer = new IntersectionObserver(
      (entries) => {
        const visible = entries
          .filter((entry) => entry.isIntersecting)
          .sort((a, b) => b.intersectionRatio - a.intersectionRatio);

        if (visible[0]?.target.id) {
          setActiveSection(`#${visible[0].target.id}`);
        }
      },
      { rootMargin: "-30% 0px -55% 0px", threshold: [0.15, 0.4, 0.7] },
    );

    sections.forEach((section) => observer.observe(section));
    return () => observer.disconnect();
  }, [navItems]);

  useEffect(() => {
    const closeOnResize = () => {
      if (window.innerWidth >= 768) {
        setIsOpen(false);
      }
    };
    window.addEventListener("resize", closeOnResize);
    return () => window.removeEventListener("resize", closeOnResize);
  }, []);

  const scrollToSection = (target: string) => {
    const element = document.querySelector(target);
    if (!(element instanceof HTMLElement)) return;

    const offset = 88;
    const nextTop = element.getBoundingClientRect().top + window.scrollY - offset;
    window.scrollTo({
      top: Math.max(0, nextTop),
      behavior: reducedMotion ? "auto" : "smooth",
    });

    const cleanUrl = `${window.location.pathname}${window.location.search}`;
    window.history.replaceState(null, "", cleanUrl);
  };

  const handleBrandClick = (event: MouseEvent<HTMLAnchorElement>) => {
    event.preventDefault();
    scrollToSection("#top");
  };

  return (
    <header
      className={clsx(
        "fixed inset-x-0 top-0 z-50 transition-all duration-300",
        isScrolled ? "border-b border-neutral-200/80 bg-white/80 backdrop-blur-md" : "bg-transparent",
      )}
    >
      <div className="mx-auto flex h-20 w-full max-w-6xl items-center justify-between px-4 sm:px-6">
        <Link
          href="/"
          onClick={handleBrandClick}
          className="group inline-flex items-center gap-2 rounded-full p-1"
        >
          <BrandMark label={landingData.productName} size="lg" />
          <span className="text-sm font-semibold tracking-tight text-neutral-900">
            {landingData.productName}
          </span>
        </Link>

        <nav className="hidden items-center gap-2 md:flex" aria-label="Primary navigation">
          {navItems.map((item) => (
            <button
              type="button"
              key={item.href}
              onClick={() => scrollToSection(item.href)}
              className={clsx(
                "rounded-full px-4 py-2 text-sm font-medium transition",
                activeSection === item.href
                  ? "brand-gradient text-white"
                  : "text-neutral-600 hover:bg-neutral-100 hover:text-neutral-900",
              )}
            >
              {item.label}
            </button>
          ))}
        </nav>

        <div className="hidden items-center gap-2 md:flex">
          <Link
            href="/login"
            className="inline-flex items-center rounded-full border border-neutral-300 bg-white px-4 py-2.5 text-sm font-semibold text-neutral-700 transition hover:border-neutral-400 hover:text-neutral-900 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-neutral-900/70 focus-visible:ring-offset-2"
          >
            Login
          </Link>
          <Link
            href="/register"
            className="brand-gradient group inline-flex items-center gap-2 rounded-full px-5 py-2.5 text-sm font-semibold text-white shadow-sm transition hover:-translate-y-0.5 hover:shadow-lg focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-neutral-900/70 focus-visible:ring-offset-2"
          >
            <span>Register</span>
            <ArrowRight className="h-4 w-4 transition-transform duration-200 group-hover:translate-x-0.5" />
          </Link>
        </div>

        <button
          type="button"
          className="inline-flex items-center justify-center rounded-full border border-neutral-200 p-2.5 text-neutral-700 transition hover:border-neutral-300 hover:text-neutral-900 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-neutral-900/70 focus-visible:ring-offset-2 md:hidden"
          onClick={() => setIsOpen((value) => !value)}
          aria-expanded={isOpen}
          aria-controls="mobile-menu"
          aria-label={isOpen ? "Close menu" : "Open menu"}
        >
          {isOpen ? <X className="h-5 w-5" /> : <Menu className="h-5 w-5" />}
        </button>
      </div>

      {isOpen ? (
        <div
          id="mobile-menu"
          className="border-b border-neutral-200 bg-white/95 px-4 pb-5 pt-2 backdrop-blur md:hidden"
        >
          <nav className="mx-auto flex w-full max-w-6xl flex-col gap-2" aria-label="Mobile navigation">
            {navItems.map((item) => (
              <button
                type="button"
                key={item.href}
                onClick={() => {
                  scrollToSection(item.href);
                  setIsOpen(false);
                }}
                className={clsx(
                  "rounded-xl px-4 py-3 text-sm font-medium transition",
                  activeSection === item.href
                    ? "brand-gradient text-white"
                    : "bg-neutral-100/70 text-neutral-700 hover:bg-neutral-200 hover:text-neutral-900",
                )}
              >
                {item.label}
              </button>
            ))}
            <Link
              href="/login"
              onClick={() => setIsOpen(false)}
              className="mt-2 inline-flex items-center justify-center rounded-xl border border-neutral-300 bg-white px-4 py-3 text-sm font-semibold text-neutral-700 transition hover:border-neutral-400 hover:text-neutral-900 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-neutral-900/70 focus-visible:ring-offset-2"
            >
              Login
            </Link>
            <Link
              href="/register"
              onClick={() => setIsOpen(false)}
              className="brand-gradient inline-flex items-center justify-center gap-2 rounded-xl px-4 py-3 text-sm font-semibold text-white shadow-sm transition hover:shadow-md focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-neutral-900/70 focus-visible:ring-offset-2"
            >
              <span>Register</span>
              <ArrowRight className="h-4 w-4" />
            </Link>
          </nav>
        </div>
      ) : null}
    </header>
  );
}
