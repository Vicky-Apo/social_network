"use client";

import Image from "next/image";
import Link from "next/link";
import { ArrowUpRight } from "lucide-react";
import { Section } from "@/components/Section";
import { landingData, type ArticleItem } from "@/lib/data";

function articleThumb(item: ArticleItem) {
  const svg = `<svg xmlns='http://www.w3.org/2000/svg' width='680' height='420' viewBox='0 0 680 420' fill='none'>
  <defs>
    <linearGradient id='g' x1='0' y1='0' x2='680' y2='420'>
      <stop stop-color='#F5F3FF'/>
      <stop offset='1' stop-color='#E0F2FE'/>
    </linearGradient>
  </defs>
  <rect width='680' height='420' rx='34' fill='url(#g)'/>
  <rect x='46' y='54' width='118' height='36' rx='18' fill='white' fill-opacity='0.9'/>
  <rect x='46' y='112' width='500' height='26' rx='13' fill='white' fill-opacity='0.8'/>
  <rect x='46' y='150' width='430' height='18' rx='9' fill='white' fill-opacity='0.65'/>
  <rect x='46' y='200' width='588' height='160' rx='24' fill='white' fill-opacity='0.75'/>
  <text x='58' y='78' fill='#0F172A' font-family='Arial' font-size='17' font-weight='700'>${item.tag}</text>
  </svg>`;
  return `data:image/svg+xml;utf8,${encodeURIComponent(svg)}`;
}

export function Articles() {
  return (
    <Section id="articles">
      <div className="mx-auto max-w-3xl text-center">
        <h2
          id="articles-heading"
          className="text-balance text-3xl font-semibold tracking-tight text-neutral-900 sm:text-4xl"
        >
          {landingData.articles.title}
        </h2>
        <p className="mt-4 text-neutral-600">
          {landingData.articles.subtitle}
        </p>
      </div>

      <div className="mt-12 grid gap-6 md:grid-cols-2">
        {landingData.articles.items.map((article) => (
          <article
            key={article.title}
            className="group overflow-hidden rounded-sm border border-neutral-200 bg-white shadow-sm transition duration-300 hover:-translate-y-1 hover:shadow-xl"
          >
            <Image
              src={articleThumb(article)}
              alt={`${article.title} cover`}
              width={680}
              height={420}
              className="h-44 w-full object-cover transition duration-300 group-hover:scale-[1.02]"
            />
            <div className="p-6">
              <p className="text-xs font-semibold uppercase tracking-[0.16em] text-neutral-500">{article.tag}</p>
              <h3 className="mt-3 text-xl font-semibold tracking-tight text-neutral-900">{article.title}</h3>
              <p className="mt-3 text-sm leading-relaxed text-neutral-600">{article.excerpt}</p>
              <Link
                href={article.href}
                className="mt-5 inline-flex items-center gap-1 text-sm font-semibold text-neutral-900 transition hover:text-neutral-700 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-neutral-900/70 focus-visible:ring-offset-2"
              >
                <span>Read more</span>
                <ArrowUpRight className="h-4 w-4" />
              </Link>
            </div>
          </article>
        ))}
      </div>
    </Section>
  );
}
