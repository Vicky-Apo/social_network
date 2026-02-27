import type { Metadata } from "next";
import { Manrope } from "next/font/google";
import "./globals.css";
import { AuthProvider } from "@/components/AuthContext";
import { landingData } from "@/lib/data";

const manrope = Manrope({
  subsets: ["latin"],
  variable: "--font-sans",
  display: "swap",
});

export const metadata: Metadata = {
  title: `${landingData.productName} | Community Platform`,
  description: landingData.heroSubtext,
};

const bodyClassName = [
  manrope.variable,
  "bg-neutral-50 text-neutral-900 antialiased",
].join(" ");

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en">
      <body className={bodyClassName}>
        <AuthProvider>{children}</AuthProvider>
      </body>
    </html>
  );
}
