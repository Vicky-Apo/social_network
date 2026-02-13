import { Navbar } from "@/components/Navbar";
import { Hero } from "@/components/Hero";
import { Workflow } from "@/components/Workflow";
import { Stats } from "@/components/Stats";
import { CTA } from "@/components/CTA";
import { Testimonials } from "@/components/Testimonials";
import { Articles } from "@/components/Articles";
import { Footer } from "@/components/Footer";

export default function HomePage() {
  return (
    <div className="min-h-screen bg-neutral-50 text-neutral-900">
      <Navbar />
      <main>
        <Hero />
        <Workflow />
        <Stats />
        <CTA />
        <Testimonials />
        <Articles />
      </main>
      <Footer />
    </div>
  );
}
