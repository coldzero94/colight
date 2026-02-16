import { Building2, Target, PenTool } from "lucide-react";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";

const features = [
  {
    icon: Building2,
    title: "기업 분석",
    description:
      "채용공고 URL만 입력하면 기업 재무, 뉴스, 인재상까지 자동 분석합니다.",
  },
  {
    icon: Target,
    title: "경험 매칭",
    description:
      "등록한 경험과 채용 요건을 AI가 매칭하여 최적의 소재를 추천합니다.",
  },
  {
    icon: PenTool,
    title: "AI 코칭",
    description:
      "STAR 구조 초안 생성부터 구체성·직무적합성·기업적합성·진정성 4점 첨삭까지.",
  },
];

export function FeatureSection() {
  return (
    <section className="border-t border-border py-20">
      <div className="mx-auto max-w-6xl px-4">
        <h2 className="text-center text-2xl font-bold text-foreground sm:text-3xl">
          이런 걸 할 수 있어요
        </h2>
        <p className="mx-auto mt-3 max-w-xl text-center text-muted-foreground">
          자소서 작성의 모든 과정을 AI가 도와드립니다
        </p>
        <div className="mt-12 grid gap-6 sm:grid-cols-2 lg:grid-cols-3">
          {features.map((feature) => (
            <Card key={feature.title} className="border-border hover:border-primary/20 transition-all duration-300">
              <CardHeader>
                <div className="mb-2 flex h-10 w-10 items-center justify-center rounded-lg bg-primary/10">
                  <feature.icon className="h-5 w-5 text-primary" />
                </div>
                <CardTitle className="text-lg">{feature.title}</CardTitle>
              </CardHeader>
              <CardContent>
                <CardDescription className="text-sm leading-relaxed">
                  {feature.description}
                </CardDescription>
              </CardContent>
            </Card>
          ))}
        </div>
      </div>
    </section>
  );
}
