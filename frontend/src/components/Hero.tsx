"use client";

import Link from "next/link";
import { ArrowRight } from "lucide-react";
import { motion, useReducedMotion } from "framer-motion";
import { Section } from "@/components/Section";
import { fadeUp, staggerContainer, MotionFloat } from "@/components/Motion";
import { landingData } from "@/lib/data";

export function Hero() {
  const reducedMotion = useReducedMotion();
  const scrollToWorkflow = () => {
    const section = document.querySelector("#workflow");
    if (!(section instanceof HTMLElement)) return;

    const offset = 88;
    const nextTop = section.getBoundingClientRect().top + window.scrollY - offset;
    window.scrollTo({
      top: Math.max(0, nextTop),
      behavior: reducedMotion ? "auto" : "smooth",
    });
    const cleanUrl = `${window.location.pathname}${window.location.search}`;
    window.history.replaceState(null, "", cleanUrl);
  };

  return (
    <Section id="top" className="relative overflow-hidden pt-28 md:pt-36">
      <div className="pointer-events-none absolute -left-28 top-20 h-72 w-72 rounded-full bg-indigo-200/35 blur-3xl" />
      <div className="pointer-events-none absolute -right-28 top-6 h-72 w-72 rounded-full bg-cyan-200/35 blur-3xl" />

      <div className="grid items-center gap-12 lg:grid-cols-[1.05fr_0.95fr]">
        <motion.div
          variants={staggerContainer}
          initial={reducedMotion ? false : "hidden"}
          animate={reducedMotion ? undefined : "show"}
          className="relative z-10"
        >
          <motion.p
            variants={fadeUp}
            className="mb-4 inline-flex rounded-full border border-neutral-200 bg-white px-4 py-1.5 text-xs font-semibold uppercase tracking-[0.16em] text-neutral-600"
          >
            {landingData.taglineSmall}
          </motion.p>

          <motion.h1
            variants={fadeUp}
            className="max-w-2xl text-balance text-4xl font-semibold tracking-tight text-neutral-900 sm:text-5xl lg:text-6xl"
          >
            {landingData.heroHeadline}
          </motion.h1>

          <motion.p
            variants={fadeUp}
            className="mt-6 max-w-xl text-base leading-relaxed text-neutral-600 sm:text-lg"
          >
            {landingData.heroSubtext}
          </motion.p>

          <motion.div variants={fadeUp} className="mt-10 flex flex-wrap items-center gap-3">
            <Link
              href={landingData.ctaUrl}
              className="brand-gradient group inline-flex items-center gap-2 rounded-full px-6 py-3 text-sm font-semibold text-white shadow-sm transition hover:-translate-y-0.5 hover:shadow-lg focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-neutral-900/70 focus-visible:ring-offset-2"
            >
              <span>{landingData.ctaPrimary}</span>
              <ArrowRight className="h-4 w-4 transition-transform duration-200 group-hover:translate-x-0.5" />
            </Link>
            <button
              type="button"
              onClick={scrollToWorkflow}
              className="inline-flex items-center rounded-full border border-neutral-300 bg-white px-6 py-3 text-sm font-semibold text-neutral-800 transition hover:-translate-y-0.5 hover:border-neutral-400 hover:shadow-md focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-neutral-900/70 focus-visible:ring-offset-2"
            >
              {landingData.ctaSecondary}
            </button>
          </motion.div>
        </motion.div>

        <MotionFloat className="relative">
          <div className="brand-gradient-soft rounded-[2rem] border border-white/60 p-3 shadow-[0_30px_70px_-35px_rgba(15,23,42,0.45)]">
            <div className="rounded-[1.6rem] border border-neutral-200/80 bg-white p-5">
              <div className="mb-5 flex items-center justify-between">
                <span className="h-2.5 w-20 rounded-full bg-neutral-200" />
                <span className="h-2.5 w-10 rounded-full bg-neutral-200" />
              </div>
              <div className="space-y-3">
                <div className="h-24 rounded-2xl bg-gradient-to-r from-indigo-100 via-sky-100 to-cyan-100" />
                <div className="grid grid-cols-2 gap-3">
                  <div className="h-20 rounded-2xl bg-neutral-100" />
                  <div className="h-20 rounded-2xl bg-neutral-100" />
                </div>
                <div className="h-12 rounded-2xl bg-neutral-100" />
              </div>
            </div>
          </div>
        </MotionFloat>
      </div>
    </Section>
  );
}
