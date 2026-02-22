import type { Metadata } from "next";
import { Space_Grotesk, Plus_Jakarta_Sans } from "next/font/google";
import "./globals.css";
import { AuthProvider } from "./component/AuthContext";
import { landingData } from "@/lib/data";

const spaceGrotesk = Space_Grotesk({
  subsets: ["latin"],
  variable: "--font-heading",
  display: "swap",
  weight: ["600", "700"],
});

const plusJakartaSans = Plus_Jakarta_Sans({
  subsets: ["latin"],
  variable: "--font-sans",
  display: "swap",
  weight: ["400", "500"],
});

export const metadata: Metadata = {
  title: `${landingData.productName} | Community Platform`,
  description: landingData.heroSubtext,
};

const bodyClassName = [
  spaceGrotesk.variable,
  plusJakartaSans.variable,
  "font-sans text-neutral-900 antialiased",
].join(" ");

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en">
      <body className={bodyClassName}>
        <div className="relative z-[1]">
          <AuthProvider>{children}</AuthProvider>
        </div>
      </body>
    </html>
  );
}
