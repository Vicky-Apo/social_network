"use client";

import Link from "next/link";
import Image from "next/image";
import { Section } from "@/components/Section";
import { Button3D } from "@/components/Button3D";
import { landingData } from "@/lib/data";

export function Hero() {
  const scrollToFeatures = () => {
    const section = document.querySelector("#features");
    if (!(section instanceof HTMLElement)) return;
    const offset = 88;
    const nextTop = section.getBoundingClientRect().top + window.scrollY - offset;
    window.scrollTo({ top: Math.max(0, nextTop), behavior: "smooth" });
    window.history.replaceState(null, "", `${window.location.pathname}${window.location.search}`);
  };

  const { heroSlogan } = landingData;

  return (
    <Section id="top" className="relative overflow-hidden pt-28 md:pt-36">
      <div className="relative z-10 mx-auto grid max-w-5xl gap-10 md:grid-cols-2 md:items-center md:gap-12">
        {/* Sol: slogan + See what's inside */}
        <div className="text-left">
          <p className="text-2xl font-semibold tracking-tight text-white sm:text-3xl md:text-4xl">
            {heroSlogan.line2}
          </p>
          <p className="mt-3 max-w-lg text-base text-white/80 sm:text-lg md:text-xl leading-relaxed">
            {heroSlogan.line3}
          </p>
          <div className="mt-8 flex flex-wrap items-center gap-3">
            <Button3D variant="secondary" onClick={scrollToFeatures}>
              See what’s inside
            </Button3D>
            <Button3D variant="primary" href={landingData.ctaUrl}>
              {landingData.ctaPrimary}
            </Button3D>
          </div>
        </div>

        {/* Sağ: logo (section'ın ~yarısı) */}
        <div className="flex flex-col items-center justify-center">
          <Link
            href="/"
            className="flex items-center justify-center rounded-sm focus:outline-none focus-visible:ring-2 focus-visible:ring-white/50 focus-visible:ring-offset-2 focus-visible:ring-offset-[#2b2929] h-[clamp(10rem,38vh,14rem)] w-full max-w-[18rem] md:h-[clamp(12rem,42vh,16rem)] md:max-w-[22rem]"
          >
            <Image
              src="/vybez-logo-v2.png"
              alt={landingData.productName}
              width={352}
              height={224}
              className="h-full w-full object-contain object-center"
              priority
            />
          </Link>
        </div>
      </div>
    </Section>
  );
}
