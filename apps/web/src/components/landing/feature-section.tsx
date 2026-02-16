import { Building2, Target, PenTool, Mic, BarChart3, Zap } from "lucide-react";

const features = [
  {
    icon: Building2,
    number: "01",
    title: "기업 분석",
    description:
      "채용공고 URL만 입력하면 기업 재무, 뉴스, 인재상까지 자동 분석합니다.",
    color: "from-emerald-500/20 to-emerald-500/5",
    iconColor: "text-emerald-400",
  },
  {
    icon: Target,
    number: "02",
    title: "경험 매칭",
    description:
      "등록한 경험과 채용 요건을 AI가 매칭하여 최적의 소재를 추천합니다.",
    color: "from-blue-500/20 to-blue-500/5",
    iconColor: "text-blue-400",
  },
  {
    icon: PenTool,
    number: "03",
    title: "AI 코칭",
    description:
      "STAR 구조 초안 생성부터 구체성·직무적합성·기업적합성·진정성 4점 첨삭까지.",
    color: "from-violet-500/20 to-violet-500/5",
    iconColor: "text-violet-400",
  },
  {
    icon: Mic,
    number: "04",
    title: "경험 인터뷰",
    description:
      "AI가 질문하고, 답변에서 STAR 구조를 자동으로 추출합니다.",
    color: "from-amber-500/20 to-amber-500/5",
    iconColor: "text-amber-400",
  },
  {
    icon: BarChart3,
    number: "05",
    title: "대시보드",
    description:
      "모든 지원 현황을 한눈에. 칸반 보드로 진행 상태를 관리합니다.",
    color: "from-rose-500/20 to-rose-500/5",
    iconColor: "text-rose-400",
  },
  {
    icon: Zap,
    number: "06",
    title: "실시간 피드백",
    description:
      "AI가 작성 중인 자소서를 실시간으로 분석하고 개선점을 제안합니다.",
    color: "from-cyan-500/20 to-cyan-500/5",
    iconColor: "text-cyan-400",
  },
];

export function FeatureSection() {
  return (
    <section className="relative py-28 lg:py-36">
      <div className="mx-auto max-w-6xl px-6">
        {/* Section header */}
        <div className="mx-auto max-w-2xl text-center">
          <p className="mb-3 text-sm font-medium uppercase tracking-widest text-primary">
            Features
          </p>
          <h2 className="font-display text-3xl font-bold tracking-tight text-foreground sm:text-4xl lg:text-5xl">
            이런 걸 할 수 있어요
          </h2>
          <p className="mt-4 text-lg text-muted-foreground">
            자소서 작성의 모든 과정을 AI가 도와드립니다
          </p>
        </div>

        {/* Feature grid - bento style */}
        <div className="mt-16 grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {features.map((feature) => (
            <div
              key={feature.title}
              className="group relative rounded-2xl border border-white/[0.06] bg-white/[0.02] p-6 transition-all duration-300 hover:border-white/[0.12] hover:bg-white/[0.04]"
            >
              {/* Number */}
              <span className="font-display text-xs font-medium text-muted-foreground/40">
                {feature.number}
              </span>

              {/* Icon */}
              <div className={`mt-4 inline-flex h-11 w-11 items-center justify-center rounded-xl bg-gradient-to-br ${feature.color}`}>
                <feature.icon className={`h-5 w-5 ${feature.iconColor}`} />
              </div>

              {/* Text */}
              <h3 className="mt-4 text-lg font-semibold text-foreground">
                {feature.title}
              </h3>
              <p className="mt-2 text-sm leading-relaxed text-muted-foreground">
                {feature.description}
              </p>

              {/* Hover glow */}
              <div className={`absolute -inset-px rounded-2xl bg-gradient-to-br ${feature.color} opacity-0 transition-opacity duration-300 group-hover:opacity-100 -z-10 blur-xl`} />
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}
