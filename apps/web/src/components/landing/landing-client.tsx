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
    <div className="min-h-screen bg-white">
      <LandingHeader />
      <HeroSection />
      <FeatureSection />
      <DemoSection />
      <CTASection />
      <LandingFooter />
    </div>
  );
}
