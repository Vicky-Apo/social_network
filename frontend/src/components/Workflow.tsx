"use client";

import Image from "next/image";
import { MessageSquareText, Rocket, Sparkles, Users } from "lucide-react";
import { motion } from "framer-motion";
import { Section } from "@/components/Section";
import { fadeUp, staggerContainer, viewportOnce } from "@/components/Motion";
import { landingData, type WorkflowStep } from "@/lib/data";

const iconByKey = {
  sparkles: Sparkles,
  users: Users,
  message: MessageSquareText,
  rocket: Rocket,
};

function createThumb(step: WorkflowStep) {
  const svg = `<svg xmlns='http://www.w3.org/2000/svg' width='640' height='360' viewBox='0 0 640 360' fill='none'>
  <defs>
    <linearGradient id='g' x1='0' y1='0' x2='640' y2='360' gradientUnits='userSpaceOnUse'>
      <stop stop-color='#E0E7FF'/>
      <stop offset='1' stop-color='#D1FAE5'/>
    </linearGradient>
  </defs>
  <rect width='640' height='360' rx='32' fill='url(#g)'/>
  <rect x='40' y='40' width='260' height='24' rx='12' fill='white' fill-opacity='0.75'/>
  <rect x='40' y='82' width='560' height='18' rx='9' fill='white' fill-opacity='0.6'/>
  <rect x='40' y='120' width='240' height='160' rx='20' fill='white' fill-opacity='0.7'/>
  <rect x='300' y='120' width='300' height='72' rx='20' fill='white' fill-opacity='0.7'/>
  <rect x='300' y='208' width='300' height='72' rx='20' fill='white' fill-opacity='0.7'/>
  <text x='42' y='332' fill='#334155' font-family='Arial' font-size='24' font-weight='700'>${step.thumbnailLabel}</text>
  </svg>`;
  return `data:image/svg+xml;utf8,${encodeURIComponent(svg)}`;
}

export function Workflow() {
  return (
    <Section id="workflow">
      <motion.div
        initial="hidden"
        whileInView="show"
        viewport={viewportOnce}
        variants={staggerContainer}
        className="mx-auto max-w-3xl text-center"
      >
        <motion.p
          variants={fadeUp}
          className="text-xs font-semibold uppercase tracking-[0.18em] text-neutral-500"
        >
          {landingData.workflow.label}
        </motion.p>
        <motion.h2
          id="workflow-heading"
          variants={fadeUp}
          className="mt-4 text-balance text-3xl font-semibold tracking-tight text-neutral-900 sm:text-4xl"
        >
          {landingData.workflow.title}
        </motion.h2>
      </motion.div>

      <motion.div
        initial="hidden"
        whileInView="show"
        viewport={viewportOnce}
        variants={staggerContainer}
        className="mt-12 grid gap-6 md:grid-cols-2"
      >
        {landingData.workflow.steps.map((step) => {
          const Icon = iconByKey[step.icon];
          return (
            <motion.article
              key={step.title}
              variants={fadeUp}
              className="group rounded-3xl border border-neutral-200 bg-white p-6 shadow-sm transition duration-300 hover:-translate-y-1 hover:shadow-xl"
            >
              <div className="mb-4 inline-flex h-11 w-11 items-center justify-center rounded-2xl bg-neutral-900 text-white">
                <Icon className="h-5 w-5" />
              </div>
              <h3 className="text-xl font-semibold tracking-tight text-neutral-900">{step.title}</h3>
              <p className="mt-2 text-sm leading-relaxed text-neutral-600">{step.description}</p>
              <div className="mt-5 overflow-hidden rounded-2xl border border-neutral-200/80">
                <Image
                  src={createThumb(step)}
                  alt={`${step.title} illustration`}
                  width={640}
                  height={360}
                  className="h-32 w-full object-cover transition duration-300 group-hover:scale-[1.03]"
                />
              </div>
            </motion.article>
          );
        })}
      </motion.div>
    </Section>
  );
}
