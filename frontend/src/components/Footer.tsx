import Link from "next/link";
import { landingData } from "@/lib/data";
import { BrandMark } from "@/components/BrandMark";

export function Footer() {
  return (
    <footer className="border-t border-neutral-200/90 bg-white">
      <div className="mx-auto w-full max-w-6xl px-4 py-14 sm:px-6">
        <div className="grid gap-10 md:grid-cols-[1.2fr_1fr_1fr_1fr]">
          <div>
            <div className="inline-flex items-center gap-2">
              <BrandMark label={landingData.productName} />
              <span className="text-sm font-semibold">{landingData.productName}</span>
            </div>
            <p className="mt-4 max-w-xs text-sm leading-relaxed text-neutral-600">
              {landingData.footer.description}
            </p>
          </div>

          <FooterColumn title="Product" links={landingData.footer.productLinks} />
          <FooterColumn title="Company" links={landingData.footer.companyLinks} />
          <FooterColumn title="Resources" links={landingData.footer.resourceLinks} />
        </div>

        <div className="mt-10 border-t border-neutral-200 pt-6 text-sm text-neutral-500">
          © {new Date().getFullYear()} {landingData.productName}. All rights reserved.
        </div>
      </div>
    </footer>
  );
}

function FooterColumn({ title, links }: { title: string; links: string[] }) {
  return (
    <div>
      <h3 className="text-sm font-semibold text-neutral-900">{title}</h3>
      <ul className="mt-4 space-y-2.5">
        {links.map((link) => (
          <li key={link}>
            <Link href="#" className="text-sm text-neutral-600 transition hover:text-neutral-900">
              {link}
            </Link>
          </li>
        ))}
      </ul>
    </div>
  );
}
