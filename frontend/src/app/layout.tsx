import type { Metadata } from "next";
import "./globals.css";
import { AuthProvider } from "@/components/AuthContext";
import { NotificationsProvider } from "@/components/NotificationsContext";
import { MessagesProvider } from "@/components/MessagesContext";
import { Footer } from "@/components/Footer";
import { landingData } from "@/lib/data";

export const metadata: Metadata = {
  title: `${landingData.productName} | Social Network`,
  description: landingData.heroSubtext,
};

const bodyClassName = "font-body text-neutral-100 antialiased";

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
