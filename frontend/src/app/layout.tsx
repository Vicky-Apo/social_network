import type { Metadata } from "next";
import { Space_Grotesk, Plus_Jakarta_Sans } from "next/font/google";
import "./globals.css";
import { AuthProvider } from "@/components/AuthContext";
import { NotificationsProvider } from "@/components/NotificationsContext";
import { MessagesProvider } from "@/components/MessagesContext";
import { Footer } from "@/components/Footer";
import { landingData } from "@/lib/data";

const spaceGrotesk = Space_Grotesk({
  subsets: ["latin"],
  variable: "--font-heading-family",
  display: "swap",
});

const plusJakartaSans = Plus_Jakarta_Sans({
  subsets: ["latin"],
  variable: "--font-body-family",
  display: "swap",
});

export const metadata: Metadata = {
  title: `${landingData.productName} | Social Network`,
  description: landingData.heroSubtext,
};

const bodyClassName = [
  spaceGrotesk.variable,
  plusJakartaSans.variable,
  "font-body text-neutral-100 antialiased",
].join(" ");

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en">
      <body className={bodyClassName}>
        <AuthProvider>
          <NotificationsProvider>
            <MessagesProvider>
              {children}
              <Footer />
            </MessagesProvider>
          </NotificationsProvider>
        </AuthProvider>
      </body>
    </html>
  );
}
