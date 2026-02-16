import { CoachingFlow } from "@/components/coaching/coaching-flow";

export default function CoachingPage() {
  return (
    <div className="container mx-auto max-w-3xl space-y-6 px-4 py-8 animate-fade-in">
      <div className="brand-surface relative overflow-hidden rounded-2xl px-6 py-5">
        <div className="pointer-events-none absolute -right-10 -top-5 h-44 w-44 rounded-full bg-primary/18 blur-3xl" />
        <div className="relative">
          <span className="brand-kicker inline-flex rounded-full px-3 py-1 text-[11px] font-semibold tracking-[0.18em] uppercase">
            Writing Co-Pilot
          </span>
          <h1 className="mt-3 text-3xl font-bold font-display text-foreground">
            AI 자소서 코칭
          </h1>
          <p className="mt-2 text-muted-foreground">
            문항 의도를 분석하고, 가장 설득력 있는 경험으로 초안을 시작하세요.
          </p>
        </div>
      </div>

      <CoachingFlow />
    </div>
  );
}
