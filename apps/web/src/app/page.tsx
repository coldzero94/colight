import type { Metadata } from "next";
import { LandingClient } from "@/components/landing/landing-client";

export const metadata: Metadata = {
  title: "Colight - AI 취업 코치 | 경험 정리부터 자소서 코칭까지",
  description:
    "채용공고 URL만 넣으면 기업 분석, 경험 매칭, STAR 구조 초안, 4점 첨삭까지. 당신만의 AI 취업 코치가 합격까지 이끌어드립니다.",
  openGraph: {
    title: "Colight - AI 취업 코치",
    description: "경험 정리부터 기업 분석, 자소서 코칭까지",
    type: "website",
  },
};

export default function LandingPage() {
  return <LandingClient />;
}
