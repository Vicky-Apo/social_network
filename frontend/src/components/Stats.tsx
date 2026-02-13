"use client";

import { useEffect, useMemo, useRef, useState } from "react";
import { Globe2, ShieldCheck, TrendingUp, Zap } from "lucide-react";
import { motion, useReducedMotion } from "framer-motion";
import { Section } from "@/components/Section";
import { fadeUp, staggerContainer, viewportOnce } from "@/components/Motion";
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

function CountUp({ metric, start }: { metric: StatMetric; start: boolean }) {
  const reducedMotion = useReducedMotion();
  const parsed = useMemo(() => parseMetricValue(metric.value), [metric.value]);
  const [display, setDisplay] = useState(metric.value);
  const hasAnimatedRef = useRef(false);

  useEffect(() => {
    if (!start || hasAnimatedRef.current || !parsed || reducedMotion) {
      return;
    }

    const duration = 1200;
    let frame = 0;
    const begin = performance.now();
    hasAnimatedRef.current = true;

    const tick = (timestamp: number) => {
      const elapsed = timestamp - begin;
      const progress = Math.min(elapsed / duration, 1);
      const eased = 1 - (1 - progress) ** 3;
      const next = parsed.value * eased;
      const decimals = parsed.value % 1 === 0 ? 0 : 1;
      setDisplay(`${parsed.prefix}${next.toFixed(decimals)}${parsed.suffix}`);
      if (progress < 1) {
        frame = requestAnimationFrame(tick);
      }
    };

    frame = requestAnimationFrame(tick);
    return () => cancelAnimationFrame(frame);
  }, [metric.value, parsed, reducedMotion, start]);

  return <span>{display}</span>;
}

export function Stats() {
  const [inView, setInView] = useState(false);

  return (
    <Section id="statistics" className="pt-10">
      <motion.div
        initial="hidden"
        whileInView="show"
        viewport={viewportOnce}
        variants={staggerContainer}
        onViewportEnter={() => setInView(true)}
        className="mx-auto max-w-3xl text-center"
      >
        <motion.h2
          id="statistics-heading"
          variants={fadeUp}
          className="text-balance text-3xl font-semibold tracking-tight text-neutral-900 sm:text-4xl"
        >
          {landingData.statistics.title}
        </motion.h2>
        <motion.p variants={fadeUp} className="mt-4 text-neutral-600">
          {landingData.statistics.subtitle}
        </motion.p>
      </motion.div>

      <motion.div
        initial="hidden"
        whileInView="show"
        viewport={viewportOnce}
        variants={staggerContainer}
        className="mt-12 grid gap-5 sm:grid-cols-2 lg:grid-cols-4"
      >
        {landingData.statistics.metrics.map((metric) => {
          const Icon = iconByKey[metric.icon];
          return (
            <motion.article
              key={metric.label}
              variants={fadeUp}
              className="rounded-3xl border border-neutral-200 bg-white p-6 shadow-sm transition duration-300 hover:-translate-y-1 hover:shadow-lg"
            >
              <div className="mb-4 inline-flex h-10 w-10 items-center justify-center rounded-2xl bg-neutral-100 text-neutral-800">
                <Icon className="h-5 w-5" />
              </div>
              <p className="text-4xl font-semibold tracking-tight text-neutral-900">
                <CountUp metric={metric} start={inView} />
              </p>
              <h3 className="mt-2 text-sm font-semibold uppercase tracking-wide text-neutral-700">
                {metric.label}
              </h3>
              <p className="mt-2 text-sm leading-relaxed text-neutral-600">{metric.description}</p>
            </motion.article>
          );
        })}
      </motion.div>
    </Section>
  );
}
