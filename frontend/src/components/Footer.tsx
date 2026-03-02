import Link from "next/link";
import Image from "next/image";
import { landingData } from "@/lib/data";

export function Footer() {
  return (
    <footer className="border-t border-white/[0.06] bg-[#2b2929]">
      <div className="mx-auto w-full max-w-6xl px-5 py-5 sm:px-8 sm:py-6">
        <div className="flex flex-wrap items-center justify-between gap-6">
          {/* Brand + description */}
          <div className="flex items-center gap-6">
            <Link href="/" className="shrink-0">
              <Image
                src="/vybez-logo.png"
                alt={`${landingData.productName} logo`}
                width={80}
                height={28}
                className="h-7 w-auto object-contain opacity-90"
              />
            </Link>
            <p className="text-sm text-neutral-500">
              {landingData.footer.description}
            </p>
          </div>

          {/* Links + copyright in one row */}
          <div className="flex flex-wrap items-center gap-6">
            {landingData.footer.productLinks.map((link) => (
              <Link
                key={link}
                href="#"
                className="text-sm text-neutral-400 transition-colors hover:text-white"
              >
                {link}
              </Link>
            ))}
            {landingData.footer.companyLinks.map((link) => (
              <Link
                key={link}
                href="#"
                className="text-sm text-neutral-400 transition-colors hover:text-white"
              >
                {link}
              </Link>
            ))}
            <span className="text-sm text-neutral-600">
              © {new Date().getFullYear()} {landingData.productName}
            </span>
          </div>
        </div>
      </div>
    </footer>
  );
}
