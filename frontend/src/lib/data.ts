export type NavItem = {
  label: string;
  href: string;
};

export type ArticleItem = {
  title: string;
  excerpt: string;
  tag: string;
  href: string;
};

export type StatMetric = {
  label: string;
  value: string;
  description: string;
  icon: "trending" | "shield" | "zap" | "globe";
};

export type TestimonialItem = {
  name: string;
  role: string;
  quote: string;
};

export type WorkflowStep = {
  title: string;
  description: string;
  icon: "sparkles" | "users" | "message" | "rocket";
  thumbnailLabel: string;
};

export const landingData = {
  productName: "Vybez",
  heroHeadline: "Connect. Share. Belong.",
  heroSubtext: "Your space to share moments, join communities, and stay close with the people who matter.",
  ctaPrimary: "Get started",
  ctaSecondary: "Sign in",
  ctaUrl: "/register",
  navItems: [] satisfies NavItem[],
  articles: {
    title: "Latest updates",
    subtitle: "Project highlights and community notes.",
    items: [] as ArticleItem[],
  },
  statistics: {
    title: "Built for real communities",
    subtitle: "Measurable growth once you are live.",
    metrics: [] as StatMetric[],
  },
  testimonials: {
    title: "Loved by early members",
    subtitle: "Hear how people use Vybez to stay connected.",
    items: [] as TestimonialItem[],
  },
  workflow: {
    label: "How it works",
    title: "Create, connect, and grow together",
    steps: [] as WorkflowStep[],
  },
  ctaBand: {
    title: "Ready to join?",
    description: "Create an account and start sharing with your communities.",
    href: "/register",
    buttonLabel: "Create account",
  },
  footer: {
    description: "A social network where real connections happen.",
    productLinks: ["Features", "About"],
    companyLinks: ["Privacy", "Terms"],
  },
};
