"use client";

import Link from "next/link";
import { ArrowRight } from "lucide-react";
import { Section } from "@/components/Section";
import { landingData } from "@/lib/data";

export function CTA() {
  return (
    <Section id="cta" className="pb-10 pt-10">
      <div className="brand-gradient-soft rounded-sm border brand-border p-8 shadow-[0_30px_60px_-40px_rgba(2,6,23,0.45)] md:p-12">
        <div className="mx-auto max-w-3xl text-center">
          <h2 className="text-balance text-3xl font-semibold tracking-tight text-neutral-900 sm:text-4xl">
            {landingData.ctaBand.title}
          </h2>
          <p className="mx-auto mt-4 max-w-2xl text-neutral-600">{landingData.ctaBand.description}</p>
          <Link
            href={landingData.ctaBand.href}
            className="brand-gradient group mt-8 inline-flex items-center gap-2 rounded-sm px-6 py-3 text-sm font-semibold text-white shadow-sm transition hover:-translate-y-0.5 hover:shadow-lg focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-neutral-900/70 focus-visible:ring-offset-2"
          >
            <span>{landingData.ctaBand.buttonLabel}</span>
            <ArrowRight className="h-4 w-4 transition-transform duration-200 group-hover:translate-x-0.5" />
          </Link>
        </div>
      </div>
    </Section>
  );
}
