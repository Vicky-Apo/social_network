"use client";

import { useMemo } from "react";
import { Globe2, ShieldCheck, TrendingUp, Zap } from "lucide-react";
import { Section } from "@/components/Section";
import { landingData, type StatMetric } from "@/lib/data";

const iconByKey = {
  trending: TrendingUp,
  shield: ShieldCheck,
  zap: Zap,
  globe: Globe2,
};

function parseMetricValue(input: string) {
  const match = input.trim().match(/^([^0-9-+]*)([-+]?\d*\.?\d+)(.*)$/);
  if (!match) {
    return null;
  }
  return {
    prefix: match[1],
    value: Number(match[2]),
    suffix: match[3],
  };
}

function CountUp({ metric }: { metric: StatMetric }) {
  // Show the final value directly; no animated counting.
  const parsed = useMemo(() => parseMetricValue(metric.value), [metric.value]);
  if (!parsed) return <span>{metric.value}</span>;
  return <span>{metric.value}</span>;
}

export function Stats() {
  return (
    <Section id="statistics" className="pt-10">
      <div className="mx-auto max-w-3xl text-center">
        <h2
          id="statistics-heading"
          className="text-balance text-3xl font-semibold tracking-tight text-neutral-900 sm:text-4xl"
        >
          {landingData.statistics.title}
        </h2>
        <p className="mt-4 text-neutral-600">
          {landingData.statistics.subtitle}
        </p>
      </div>

      <div className="mt-12 grid gap-5 sm:grid-cols-2 lg:grid-cols-4">
        {landingData.statistics.metrics.map((metric) => {
          const Icon = iconByKey[metric.icon];
          return (
            <article
              key={metric.label}
              className="rounded-3xl border border-neutral-200 bg-white p-6 shadow-sm transition duration-300 hover:-translate-y-1 hover:shadow-lg"
            >
              <div className="mb-4 inline-flex h-10 w-10 items-center justify-center rounded-2xl bg-neutral-100 text-neutral-800">
                <Icon className="h-5 w-5" />
              </div>
              <p className="text-4xl font-semibold tracking-tight text-neutral-900">
                <CountUp metric={metric} />
              </p>
              <h3 className="mt-2 text-sm font-semibold uppercase tracking-wide text-neutral-700">
                {metric.label}
              </h3>
              <p className="mt-2 text-sm leading-relaxed text-neutral-600">{metric.description}</p>
            </article>
          );
        })}
      </div>
    </Section>
  );
}
