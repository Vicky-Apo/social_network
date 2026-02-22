import { Navbar } from "@/components/Navbar";
import { Hero } from "@/components/Hero";
import { LandingFeatures } from "@/components/LandingFeatures";
import { Footer } from "@/components/Footer";

export default function LandingPage() {
  return (
    <div className="min-h-screen bg-[#2b2929] text-neutral-900">
      <Navbar />
      <main>
        <Hero />
        <LandingFeatures />
      </main>
      <Footer />
    </div>
  );
}

