import { Link2, Search, GitCompare, FileEdit, CheckCircle } from "lucide-react";

const steps = [
  {
    icon: Link2,
    title: "URL 입력",
    description: "채용공고 링크 붙여넣기",
  },
  {
    icon: Search,
    title: "자동 분석",
    description: "기업 · 직무 정보 파악",
  },
  {
    icon: GitCompare,
    title: "소재 매칭",
    description: "내 경험에서 최적 소재",
  },
  {
    icon: FileEdit,
    title: "초안 작성",
    description: "STAR 구조로 자동 생성",
  },
  {
    icon: CheckCircle,
    title: "AI 첨삭",
    description: "4점 코칭으로 완성",
  },
];

export function DemoSection() {
  return (
    <section className="relative py-28 lg:py-36 overflow-hidden">
      {/* Background accent */}
      <div className="absolute inset-0 -z-10">
        <div className="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-[800px] h-[400px] bg-primary/[0.04] rounded-full blur-[120px]" />
      </div>

      <div className="mx-auto max-w-6xl px-6">
        {/* Section header */}
        <div className="mx-auto max-w-2xl text-center">
          <p className="mb-3 text-sm font-medium uppercase tracking-widest text-primary">
            Process
          </p>
          <h2 className="font-display text-3xl font-bold tracking-tight text-foreground sm:text-4xl lg:text-5xl">
            한눈에 보는 코칭 프로세스
          </h2>
          <p className="mt-4 text-lg text-muted-foreground">
            채용공고 입력부터 최종 자소서까지, 5단계 AI 코칭
          </p>
        </div>

        {/* Steps */}
        <div className="glass relative mx-auto mt-16 max-w-4xl rounded-2xl p-8 lg:p-12">
          <div className="grid gap-8 sm:grid-cols-5">
            {steps.map((step, index) => (
              <div key={step.title} className="relative text-center">
                {/* Connector line (hidden on last item and on mobile) */}
                {index < steps.length - 1 && (
                  <div className="absolute top-6 left-[calc(50%+24px)] right-[calc(-50%+24px)] hidden h-px bg-gradient-to-r from-primary/30 to-primary/10 sm:block" />
                )}

                {/* Icon circle */}
                <div className="relative mx-auto flex h-12 w-12 items-center justify-center rounded-full border border-primary/20 bg-primary/[0.08]">
                  <step.icon className="h-5 w-5 text-primary" />
                  {/* Step number */}
                  <span className="absolute -right-1 -top-1 flex h-5 w-5 items-center justify-center rounded-full bg-primary text-[10px] font-bold text-primary-foreground">
                    {index + 1}
                  </span>
                </div>

                {/* Text */}
                <h3 className="mt-4 text-sm font-semibold text-foreground">
                  {step.title}
                </h3>
                <p className="mt-1 text-xs text-muted-foreground">
                  {step.description}
                </p>
              </div>
            ))}
          </div>
        </div>
      </div>
    </section>
  );
}
