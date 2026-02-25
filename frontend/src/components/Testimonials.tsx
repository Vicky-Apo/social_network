"use client";

import { useRef } from "react";
import { ChevronLeft, ChevronRight } from "lucide-react";
import { motion } from "framer-motion";
import { Section } from "@/components/Section";
import { fadeUp, staggerContainer, viewportOnce } from "@/components/Motion";
import { landingData } from "@/lib/data";

function initials(name: string) {
  return name
    .split(" ")
    .map((part) => part[0])
    .join("")
    .slice(0, 2)
    .toUpperCase();
}

export function Testimonials() {
  const scrollerRef = useRef<HTMLDivElement>(null);

  const scrollByAmount = (direction: "prev" | "next") => {
    const node = scrollerRef.current;
    if (!node) {
      return;
    }
    const amount = Math.max(node.clientWidth * 0.9, 280);
    node.scrollBy({
      left: direction === "next" ? amount : -amount,
      behavior: "smooth",
    });
  };

  return (
    <Section id="testimonials">
      <motion.div
        initial="hidden"
        whileInView="show"
        viewport={viewportOnce}
        variants={staggerContainer}
        className="mb-8 flex flex-wrap items-end justify-between gap-4"
      >
        <div className="max-w-2xl">
          <motion.h2
            id="testimonials-heading"
            variants={fadeUp}
            className="text-balance text-3xl font-semibold tracking-tight text-neutral-900 sm:text-4xl"
          >
            {landingData.testimonials.title}
          </motion.h2>
          <motion.p variants={fadeUp} className="mt-3 text-neutral-600">
            {landingData.testimonials.subtitle}
          </motion.p>
        </div>

        <motion.div variants={fadeUp} className="flex items-center gap-2">
          <button
            type="button"
            onClick={() => scrollByAmount("prev")}
            aria-label="Scroll testimonials left"
            className="inline-flex h-10 w-10 items-center justify-center rounded-full border border-neutral-300 bg-white text-neutral-700 transition hover:border-neutral-400 hover:text-neutral-900 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-neutral-900/70 focus-visible:ring-offset-2"
          >
            <ChevronLeft className="h-4 w-4" />
          </button>
          <button
            type="button"
            onClick={() => scrollByAmount("next")}
            aria-label="Scroll testimonials right"
            className="inline-flex h-10 w-10 items-center justify-center rounded-full border border-neutral-300 bg-white text-neutral-700 transition hover:border-neutral-400 hover:text-neutral-900 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-neutral-900/70 focus-visible:ring-offset-2"
          >
            <ChevronRight className="h-4 w-4" />
          </button>
        </motion.div>
      </motion.div>

      <motion.div
        ref={scrollerRef}
        initial="hidden"
        whileInView="show"
        viewport={viewportOnce}
        variants={staggerContainer}
        className="testimonials-scroll grid auto-cols-[88%] grid-flow-col gap-4 overflow-x-auto pb-3 [scrollbar-width:none] sm:auto-cols-[60%] lg:auto-cols-[33.333%]"
        style={{ scrollSnapType: "x mandatory" }}
      >
        {landingData.testimonials.items.map((item, index) => (
          <motion.figure
            key={`${item.name}-${index}`}
            variants={fadeUp}
            className="group min-h-64 rounded-3xl border border-neutral-200 bg-white p-6 shadow-sm transition duration-300 hover:-translate-y-1 hover:shadow-lg"
            style={{ scrollSnapAlign: "start" }}
          >
            <blockquote className="text-base leading-relaxed text-neutral-700">{item.quote}</blockquote>
            <figcaption className="mt-8 flex items-center gap-3">
              <span className="inline-flex h-10 w-10 items-center justify-center rounded-full bg-neutral-900 text-xs font-semibold text-white">
                {initials(item.name)}
              </span>
              <span className="flex flex-col">
                <span className="text-sm font-semibold text-neutral-900">{item.name}</span>
                <span className="text-xs text-neutral-500">{item.role}</span>
              </span>
            </figcaption>
          </motion.figure>
        ))}
      </motion.div>
    </Section>
  );
}
