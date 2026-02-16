"use client";

import { useEffect } from "react";
import { useRouter } from "next/navigation";
import { useAuthStore } from "@/stores/auth-store";
import { LandingHeader } from "./landing-header";
import { HeroSection } from "./hero-section";
import { FeatureSection } from "./feature-section";
import { DemoSection } from "./demo-section";
import { CTASection } from "./cta-section";
import { LandingFooter } from "./landing-footer";

export function LandingClient() {
  const accessToken = useAuthStore((s) => s.accessToken);
  const router = useRouter();

  useEffect(() => {
    if (accessToken) {
      router.replace("/experiences");
    }
  }, [accessToken, router]);

  if (accessToken) return null;

  return (
    <div className="landing-shell relative min-h-screen overflow-x-clip">
      <div className="landing-noise pointer-events-none absolute inset-0 -z-10" />
      <div className="dot-grid animate-grid-pan pointer-events-none absolute inset-0 -z-10 opacity-25" />
      <div className="pointer-events-none absolute left-[-8rem] top-32 -z-10 h-[28rem] w-[28rem] rounded-full bg-orange-400/15 blur-[120px] animate-float-y" />
      <div className="pointer-events-none absolute right-[-10rem] top-[24rem] -z-10 h-[34rem] w-[34rem] rounded-full bg-cyan-400/10 blur-[130px] animate-drift-x" />
      <LandingHeader />
      <HeroSection />
      <FeatureSection />
      <DemoSection />
      <CTASection />
      <LandingFooter />
    </div>
  );
}
