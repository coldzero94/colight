"use client";

import { toast } from "sonner";
import { PricingCard } from "@/components/paywall/pricing-card";
import { useUsage } from "@/hooks/use-usage";

const plans = [
  {
    name: "Free",
    price: "₩0",
    description: "시작하기 좋은 무료 플랜",
    features: [
      "경험 등록 3건",
      "기업 분석 일 1회",
      "문항 분석 일 1회",
      "초안 생성 일 1회",
      "첨삭 코칭 일 1회",
    ],
    cta: "현재 플랜",
    key: "free" as const,
  },
  {
    name: "Starter",
    price: "₩9,900/월",
    description: "취업 준비에 집중하는 분을 위해",
    features: [
      "경험 등록 무제한",
      "기업 분석 일 5회",
      "문항 분석 일 5회",
      "초안 생성 일 5회",
      "첨삭 코칭 일 5회",
    ],
    cta: "시작하기",
    key: "starter" as const,
  },
  {
    name: "Pro",
    price: "₩19,900/월",
    description: "합격률을 높이고 싶은 분을 위해",
    features: [
      "모든 기능 무제한",
      "우선 AI 처리",
      "고급 분석 리포트",
      "이메일 지원",
    ],
    cta: "시작하기",
    key: "pro" as const,
    highlighted: true,
  },
  {
    name: "Season",
    price: "₩49,900/3개월",
    description: "채용 시즌에 몰아서 준비하는 분을 위해",
    features: [
      "Pro 플랜의 모든 기능",
      "3개월 집중 패키지",
      "월 대비 17% 할인",
    ],
    cta: "시작하기",
    key: "season" as const,
  },
];

export default function PricingPage() {
  const { data } = useUsage();
  const currentPlan = data?.plan ?? "free";

  function handleCtaClick(planKey: string) {
    if (planKey === currentPlan) return;
    if (planKey === "free") return;
    toast.info("결제 기능은 곧 출시됩니다!");
  }

  return (
    <div className="mx-auto max-w-5xl px-4 py-10">
      <div className="mb-10 text-center">
        <h1 className="text-2xl font-bold">플랜 및 가격</h1>
        <p className="mt-2 text-sm text-muted-foreground">
          나에게 맞는 플랜을 선택하세요
        </p>
      </div>

      <div className="grid gap-6 sm:grid-cols-2 lg:grid-cols-4">
        {plans.map((plan) => (
          <PricingCard
            key={plan.key}
            name={plan.name}
            price={plan.price}
            description={plan.description}
            features={plan.features}
            cta={plan.cta}
            onCtaClick={() => handleCtaClick(plan.key)}
            highlighted={plan.highlighted}
            current={plan.key === currentPlan}
          />
        ))}
      </div>
    </div>
  );
}
