"use client";

import Image from "next/image";
import { Section } from "@/components/Section";
import { motion } from "framer-motion";

export function LandingFeatures() {
  return (
    <Section id="features" className="bg-[#2b2929] py-16 md:py-24">
      <div className="mx-auto max-w-5xl">
        <motion.div
          className="relative overflow-hidden rounded-sm bg-[#2b2929]"
          initial={{ opacity: 0, y: 24 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true, margin: "-40px" }}
          transition={{ duration: 0.5 }}
        >
          <div className="relative aspect-[16/10] w-full sm:aspect-[2/1]">
            <Image
              src="/gradient-4.png"
              alt=""
              fill
              className="object-cover object-center opacity-70"
              sizes="(max-width: 768px) 100vw, 896px"
            />
            <div className="absolute inset-0 bg-[#2b2929]/50" />
            <div className="absolute inset-0 bg-gradient-to-t from-[#2b2929] via-[#2b2929]/15 to-transparent" />
            <div className="absolute inset-0 flex flex-col items-center justify-center gap-5 p-8 text-center sm:gap-6 sm:p-12">
              <p className="text-xs font-medium uppercase tracking-widest text-white/50 sm:text-sm">
                What’s inside
              </p>
              <h2 className="text-2xl font-semibold tracking-tight text-white sm:text-3xl md:text-4xl">
                One feed. Your circle. Groups and chat.
              </h2>
              <p className="mx-auto max-w-xl text-sm leading-relaxed text-white/85 sm:text-base">
                Everything you need to post, share, and stay close—without the noise. Create an account and see for yourself.
              </p>
            </div>
          </div>
        </motion.div>
      </div>
    </Section>
  );
}
