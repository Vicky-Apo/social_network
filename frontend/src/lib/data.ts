export type NavItem = {
  label: string;
  href: string;
};

export const landingData = {
  productName: "Vybez",
  heroHeadline: "Connect. Share. Belong.",
  heroSubtext: "Your space to share moments, join communities, and stay close with the people who matter.",
  ctaPrimary: "Get started",
  ctaSecondary: "Sign in",
  ctaUrl: "/register",
  navItems: [] satisfies NavItem[],
  footer: {
    description: "A social network where real connections happen.",
    productLinks: ["Features", "About"],
    companyLinks: ["Privacy", "Terms"],
  },
};
