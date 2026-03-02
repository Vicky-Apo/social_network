"use client";

import Link from "next/link";
import Image from "next/image";
import { ArrowRight } from "lucide-react";
import { motion, useReducedMotion } from "framer-motion";
import { Section } from "@/components/Section";
import { fadeUp, staggerContainer } from "@/components/Motion";
import { landingData } from "@/lib/data";

export function Hero() {
  const reducedMotion = useReducedMotion();

  return (
    <Section id="top" className="relative overflow-hidden pt-28 md:pt-36 min-h-[85vh] flex items-center">
      <div className="flex flex-row items-center justify-between gap-8 lg:gap-12">
        <motion.div
          variants={staggerContainer}
          initial={reducedMotion ? false : "hidden"}
          animate={reducedMotion ? undefined : "show"}
          className="relative z-10 min-w-0 flex-1"
        >
          <motion.h1
            variants={fadeUp}
            className="max-w-2xl text-balance text-4xl font-bold tracking-tight text-white sm:text-5xl lg:text-6xl"
          >
            {landingData.heroHeadline}
          </motion.h1>

          <motion.p
            variants={fadeUp}
            className="mt-6 max-w-lg text-base leading-relaxed text-neutral-300 sm:text-lg"
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
            <Link
              href="/login"
              className="inline-flex items-center rounded-full border border-white/30 bg-transparent px-6 py-3 text-sm font-semibold text-neutral-200 transition hover:-translate-y-0.5 hover:border-white/50 hover:bg-white/5 hover:text-white focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-white/50 focus-visible:ring-offset-2 focus-visible:ring-offset-[#2b2929]"
            >
              {landingData.ctaSecondary}
            </Link>
          </motion.div>
        </motion.div>

        <div className="relative flex shrink-0 items-center justify-center">
          <Image
            src="/vybez-logo.png"
            alt={`${landingData.productName} logo`}
            width={320}
            height={140}
            className="w-[240px] sm:w-[280px] lg:w-[320px] h-auto object-contain"
            priority
          />
        </div>
      </div>
    </Section>
  );
}
