import Link from "next/link";
import { ArrowRight, CheckCircle2, Sparkles } from "lucide-react";
import { Button } from "@/components/ui/button";

export function HeroSection() {
  return (
    <section className="relative overflow-hidden pb-24 pt-16 lg:pb-32 lg:pt-24">
      <div className="pointer-events-none absolute inset-0 -z-10">
        <div className="absolute inset-x-0 top-0 h-24 bg-gradient-to-b from-black/30 to-transparent" />
        <div className="absolute left-[8%] top-10 h-56 w-56 rounded-full border border-white/10" />
        <div className="absolute left-[10%] top-12 h-56 w-56 rounded-full border border-white/5 animate-float-y-soft" />
        <div className="absolute right-[8%] top-28 h-48 w-48 rounded-full border border-cyan-300/20 animate-drift-x" />
      </div>

      <div className="mx-auto grid max-w-6xl gap-10 px-6 lg:grid-cols-2 lg:items-center">
        <div className="animate-slide-up">
          <div className="mb-7 inline-flex items-center gap-2 rounded-full border border-primary/35 bg-primary/10 px-4 py-1.5 text-sm text-primary shadow-[0_0_30px_rgba(255,150,60,0.25)]">
            <Sparkles className="h-3.5 w-3.5" />
            <span>AI 기반 취업 코칭 플랫폼</span>
          </div>

          <h1 className="font-display text-5xl font-bold leading-[1.06] tracking-tight text-foreground sm:text-6xl lg:text-7xl">
            자소서 작성,
            <br />
            <span className="gradient-text">AI 코치</span>의 속도로
          </h1>

          <p className="mt-6 max-w-xl text-lg leading-relaxed text-muted-foreground sm:text-xl">
            URL 하나로 기업 인사이트를 추출하고, 내 경험을 자동 매칭해 바로 제출 가능한 STAR 초안을 완성합니다.
          </p>

          <div className="mt-8 flex flex-wrap items-center gap-3">
            <Button size="lg" className="glow-md h-13 px-8 text-base" asChild>
              <Link href="/signup">
                무료로 시작하기
                <ArrowRight className="ml-2 h-4 w-4" />
              </Link>
            </Button>
            <Button variant="outline" size="lg" className="h-13 border-white/25 px-8 text-base hover:bg-white/10" asChild>
              <Link href="/login">로그인</Link>
            </Button>
          </div>

          <div className="mt-7 flex flex-col gap-2 text-sm text-muted-foreground/80 sm:flex-row sm:items-center sm:gap-5">
            <span className="inline-flex items-center gap-2">
              <CheckCircle2 className="h-4 w-4 text-primary" />
              가입 즉시 무료 체험
            </span>
            <span className="inline-flex items-center gap-2">
              <CheckCircle2 className="h-4 w-4 text-primary" />
              신용카드 불필요
            </span>
          </div>
        </div>

        <div className="relative animate-fade-in lg:pl-8">
          <div className="landing-panel glass-strong rounded-3xl p-3 shadow-[0_24px_90px_rgba(0,0,0,0.45)]">
            <div className="overflow-hidden rounded-2xl border border-white/10 bg-black/30">
              <div className="flex items-center gap-2 border-b border-white/[0.08] px-4 py-3">
                <div className="flex gap-1.5">
                  <div className="h-2.5 w-2.5 rounded-full bg-white/10" />
                  <div className="h-2.5 w-2.5 rounded-full bg-white/10" />
                  <div className="h-2.5 w-2.5 rounded-full bg-white/10" />
                </div>
                <div className="ml-4 h-5 w-40 rounded bg-white/[0.05]" />
              </div>
              <div className="grid grid-cols-12">
                <div className="col-span-4 border-r border-white/[0.08] p-4">
                  <div className="space-y-3">
                    <div className="h-3.5 w-16 rounded bg-primary/35" />
                    <div className="h-3 w-24 rounded bg-white/[0.08]" />
                    <div className="h-3 w-20 rounded bg-white/[0.08]" />
                    <div className="h-3 w-14 rounded bg-white/[0.08]" />
                  </div>
                </div>
                <div className="col-span-8 p-4 sm:p-5">
                  <div className="mb-4 flex items-center justify-between">
                    <div className="h-5 w-28 rounded bg-white/[0.09]" />
                    <div className="h-6 w-24 rounded-full bg-cyan-400/25" />
                  </div>
                  <div className="space-y-3">
                    {[...Array(3)].map((_, i) => (
                      <div
                        key={i}
                        className="rounded-xl border border-white/[0.08] bg-white/[0.03] p-3 animate-float-y-soft"
                        style={{ animationDelay: `${i * 0.4}s` }}
                      >
                        <div className="mb-2.5 h-2.5 w-20 rounded bg-white/[0.12]" />
                        <div className="space-y-2">
                          <div className="h-2 w-full rounded bg-white/[0.06]" />
                          <div className="h-2 w-4/5 rounded bg-white/[0.06]" />
                        </div>
                        <div className="mt-3 flex gap-2">
                          <div className="h-5 w-12 rounded-full bg-orange-400/25" />
                          <div className="h-5 w-14 rounded-full bg-cyan-400/20" />
                        </div>
                      </div>
                    ))}
                  </div>
                </div>
              </div>
            </div>
          </div>
          <div className="absolute -inset-7 -z-10 rounded-[2.5rem] bg-gradient-to-r from-primary/20 via-cyan-400/20 to-transparent blur-3xl" />
        </div>
      </div>
    </section>
  );
}
