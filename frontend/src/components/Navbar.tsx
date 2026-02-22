"use client";

import { useEffect, useMemo, useState } from "react";
import Link from "next/link";
import Image from "next/image";
import { usePathname } from "next/navigation";
import type { MouseEvent } from "react";
import { Menu, X } from "lucide-react";
import { useReducedMotion } from "framer-motion";
import clsx from "clsx";
import { landingData, type NavItem } from "@/lib/data";
import { Button3D } from "@/components/Button3D";

export function Navbar() {
  const pathname = usePathname();
  const [isOpen, setIsOpen] = useState(false);
  const [activeSection, setActiveSection] = useState("");
  const reducedMotion = useReducedMotion();
  const navItems = useMemo<NavItem[]>(() => landingData.navItems, []);
  const showAuthButtons = pathname !== "/login" && pathname !== "/register";

  useEffect(() => {
    if (navItems.length === 0) return;
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
    <header className="fixed inset-x-0 top-0 z-50 border-b border-white/10 bg-black/30 backdrop-blur-md">
      <div className="relative z-10 mx-auto flex w-full max-w-6xl flex-col items-center gap-5 px-4 py-6 sm:px-6 md:flex-row md:justify-between md:gap-8">
        <Link
          href="/"
          onClick={handleBrandClick}
          className="inline-flex items-center focus:outline-none focus-visible:ring-2 focus-visible:ring-white/50 focus-visible:ring-offset-2 focus-visible:ring-offset-[#2b2929] rounded-sm"
        >
          <Image
            src="/vybez-logo-v2.png"
            alt={landingData.productName}
            width={120}
            height={48}
            className="h-8 w-auto sm:h-9"
          />
        </Link>

        <nav className="hidden flex-wrap items-center justify-center gap-x-6 gap-y-1 text-sm text-white/70 md:flex">
          {navItems.map((item) => (
            <button
              type="button"
              key={item.href}
              onClick={() => scrollToSection(item.href)}
              className={clsx(
                "transition hover:text-white",
                activeSection === item.href && "text-white",
              )}
            >
              {item.label}
            </button>
          ))}
        </nav>

        {showAuthButtons ? (
          <div className="hidden items-center gap-2 [&_.btn-3d-inner]:h-9 [&_.btn-3d-inner]:min-w-0 [&_.btn-3d-inner]:px-4 md:flex">
            <Button3D href="/login">Login</Button3D>
            <Button3D href={landingData.ctaUrl}>{landingData.ctaPrimary}</Button3D>
          </div>
        ) : null}

        <button
          type="button"
          className="absolute right-4 top-1/2 -translate-y-1/2 inline-flex items-center justify-center rounded-sm border border-white/30 p-2.5 text-white transition hover:border-white/50 hover:bg-white/10 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-white/50 focus-visible:ring-offset-2 focus-visible:ring-offset-[#2b2929] md:hidden"
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
          className="border-b border-white/10 bg-black/30 px-4 pb-5 pt-2 backdrop-blur md:hidden"
        >
          <nav className="mx-auto flex w-full max-w-6xl flex-col gap-2 pt-2" aria-label="Mobile navigation">
            {navItems.map((item) => (
              <button
                type="button"
                key={item.href}
                onClick={() => {
                  scrollToSection(item.href);
                  setIsOpen(false);
                }}
                className={clsx(
                  "rounded-sm px-4 py-3 text-sm font-medium transition text-left",
                  activeSection === item.href ? "bg-white/20 text-white" : "text-white/90 hover:bg-white/10",
                )}
              >
                {item.label}
              </button>
            ))}
            {showAuthButtons ? (
              <div className="mt-2 flex flex-col gap-2 [&_.btn-3d-inner]:h-10 [&_.btn-3d-inner]:min-w-0 [&_.btn-3d-inner]:px-5" onClick={() => setIsOpen(false)}>
                <Button3D href="/login">Login</Button3D>
                <Button3D href={landingData.ctaUrl}>{landingData.ctaPrimary}</Button3D>
              </div>
            ) : null}
          </nav>
        </div>
      ) : null}
    </header>
  );
}
