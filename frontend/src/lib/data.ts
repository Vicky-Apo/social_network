export type NavItem = {
  label: string;
  href: string;
};

export type WorkflowStep = {
  title: string;
  description: string;
  thumbnailLabel: string;
  icon: "sparkles" | "users" | "message" | "rocket";
};

export type StatMetric = {
  value: string;
  label: string;
  description: string;
  icon: "trending" | "shield" | "zap" | "globe";
};

export type TestimonialItem = {
  quote: string;
  name: string;
  role: string;
};

export type ArticleItem = {
  tag: string;
  title: string;
  excerpt: string;
  href: string;
};

/** Short feature item for landing "what you can do" strip */
export type LandingFeature = {
  label: string;
  description: string;
  icon: "feed" | "profile" | "groups" | "chat" | "notifications";
};

export const landingData = {
  productName: "Vybez",
  taglineSmall: "EST. 2026",
  heroHeadline: "Connect. Share. Stay close.",
  heroSubtext:
    "Your feed, your people, your groups. Post, chat, and keep the conversation going—all in one place.",
  /** Hero: Your Circle + value prop */
  heroSlogan: {
    line1: "Your Circle",
    line2: "This isn’t just another social app.",
    line3: "It’s your digital house, where you can share, chat, connect, and actually feel seen.",
  },
  ctaPrimary: "Create account",
  ctaSecondary: "See what’s inside",
  ctaUrl: "/register",
  /** Empty = no middle nav (social network: just logo + Login + Register) */
  navItems: [] satisfies NavItem[],
  /** Strip shown under hero on landing (social network features) */
  landingFeatures: [
    { label: "Feed", description: "Posts from people you follow", icon: "feed" as const },
    { label: "Profile", description: "Your page, your way", icon: "profile" as const },
    { label: "Groups", description: "Communities and events", icon: "groups" as const },
    { label: "Chat", description: "DMs and group chats", icon: "chat" as const },
    { label: "Notifications", description: "Never miss a beat", icon: "notifications" as const },
  ] satisfies LandingFeature[],
  ctaBand: {
    title: "Join Vybez today",
    description: "Sign up in seconds. Start sharing, following, and chatting with people who matter.",
    buttonLabel: "Create account",
    href: "/register",
  },
  footer: {
    description: "A place to connect, share, and stay close. Your social network, your way.",
    productLinks: ["Feed", "Groups", "Chat", "Explore"],
    legalLinks: ["Privacy", "Terms"],
    supportLinks: ["Help", "Contact"],
  },
  // Legacy keys kept for any remaining references (can remove when unused)
  workflow: {
    label: "Our workflow",
    title: "Build stronger discussions with a clear collaboration flow",
    steps: [
      {
        title: "Set up your space in minutes",
        description:
          "Create categories, add posting guidelines, and configure roles so every discussion starts with clarity.",
        thumbnailLabel: "Onboarding",
        icon: "sparkles",
      },
      {
        title: "Surface the right conversations",
        description:
          "Highlight trending threads and recommend relevant topics to help members find what matters fast.",
        thumbnailLabel: "Discovery",
        icon: "users",
      },
      {
        title: "Keep engagement meaningful",
        description:
          "Use moderation tools, reactions, and structured replies to keep discussions healthy and useful.",
        thumbnailLabel: "Conversation",
        icon: "message",
      },
      {
        title: "Turn activity into growth",
        description:
          "Track outcomes, refine community rituals, and scale high-signal engagement over time.",
        thumbnailLabel: "Momentum",
        icon: "rocket",
      },
    ] satisfies WorkflowStep[],
  },
  statistics: {
    title: "The numbers that define our success",
    subtitle:
      "Trusted by communities that care about quality engagement and sustainable growth.",
    metrics: [
      {
        value: "94%",
        label: "Weekly Member Retention",
        description: "Members who come back and participate in discussions every week.",
        icon: "trending",
      },
      {
        value: "99.9%",
        label: "Platform Uptime",
        description: "Reliable performance to keep conversations available around the clock.",
        icon: "shield",
      },
      {
        value: "42k",
        label: "Monthly Active Contributors",
        description: "People posting, replying, and helping others across community channels.",
        icon: "zap",
      },
      {
        value: "18h",
        label: "Average Thread Lifespan",
        description: "Sustained conversation windows that encourage deeper participation.",
        icon: "globe",
      },
    ] satisfies StatMetric[],
  },
  testimonials: {
    title: "What community teams say",
    subtitle:
      "Real feedback from moderators and creators building meaningful spaces every day.",
    items: [
      {
        quote:
          "Vybez helped us organize chaotic chat threads into focused discussions people actually return to.",
        name: "Ariana Patel",
        role: "Community Lead",
      },
      {
        quote:
          "The moderation workflow is simple, fast, and transparent. Our team finally feels in control.",
        name: "Marcus Chen",
        role: "Head of Operations",
      },
      {
        quote:
          "We launched in one afternoon and immediately saw better-quality replies from day one.",
        name: "Nora Williams",
        role: "Founder",
      },
      {
        quote:
          "The analytics gave us a clear picture of what keeps members engaged and what we should improve next.",
        name: "David Romero",
        role: "Growth Manager",
      },
      {
        quote:
          "Our onboarding to first-post conversion improved dramatically after moving our forum to Vybez.",
        name: "Leila Hassan",
        role: "Product Marketing Lead",
      },
      {
        quote:
          "Clean UI, great mobile experience, and all the controls we need for a healthy member culture.",
        name: "Jonah Brooks",
        role: "Community Moderator",
      },
    ] satisfies TestimonialItem[],
  },
  articles: {
    title: "Maximizing the value of every discussion",
    subtitle:
      "Learn practical strategies to grow a high-signal community and keep conversations productive.",
    items: [
      {
        tag: "Community",
        title: "How to design onboarding that invites thoughtful participation",
        excerpt:
          "Set clear expectations, guide first-time contributors, and create a smooth first post experience.",
        href: "#",
      },
      {
        tag: "Moderation",
        title: "A practical framework for moderation without killing momentum",
        excerpt:
          "Balance safety and openness by defining response rules that scale with your Vybez community.",
        href: "#",
      },
      {
        tag: "Growth",
        title: "Turning recurring threads into long-term retention loops",
        excerpt:
          "Convert repeat topics into dependable engagement patterns your members return to weekly.",
        href: "#",
      },
      {
        tag: "Insights",
        title: "Measuring the Vybez metrics that actually indicate community health",
        excerpt:
          "Track behavior that reflects real value, not vanity numbers that hide churn.",
        href: "#",
      },
    ] satisfies ArticleItem[],
  },
};
