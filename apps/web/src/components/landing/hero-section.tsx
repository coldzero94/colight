import Link from "next/link";
import { ArrowRight, Sparkles } from "lucide-react";
import { Button } from "@/components/ui/button";

export function HeroSection() {
  return (
    <section className="relative overflow-hidden">
      {/* Background effects */}
      <div className="absolute inset-0 -z-10">
        <div className="dot-grid absolute inset-0 opacity-40" />
        <div className="absolute top-0 left-1/4 w-[600px] h-[600px] bg-primary/[0.12] rounded-full blur-[150px] animate-pulse-glow" />
        <div className="absolute bottom-0 right-1/4 w-[500px] h-[500px] bg-blue-500/[0.08] rounded-full blur-[150px] animate-pulse-glow" />
        {/* Top gradient fade from header */}
        <div className="absolute inset-x-0 top-0 h-32 bg-gradient-to-b from-background to-transparent" />
      </div>

      <div className="mx-auto max-w-5xl px-6 pb-24 pt-20 lg:pb-32 lg:pt-28">
        <div className="animate-slide-up text-center">
          {/* Badge */}
          <div className="mb-8 inline-flex items-center gap-2 rounded-full border border-primary/20 bg-primary/[0.06] px-4 py-1.5 text-sm text-primary">
            <Sparkles className="h-3.5 w-3.5" />
            <span>AI 기반 취업 코칭 플랫폼</span>
          </div>

          {/* Heading */}
          <h1 className="font-display text-5xl font-bold leading-[1.1] tracking-tight text-foreground sm:text-6xl lg:text-7xl">
            자소서 작성,
            <br />
            <span className="gradient-text">AI 코치</span>에게 맡기세요
          </h1>

          {/* Subheading */}
          <p className="mx-auto mt-6 max-w-2xl text-lg leading-relaxed text-muted-foreground sm:text-xl">
            채용공고 URL 하나면 끝. 기업 분석부터 경험 매칭, STAR 초안, 4점 첨삭까지.
            <br className="hidden sm:block" />
            합격을 위한 모든 과정을 AI가 코칭합니다.
          </p>

          {/* CTA */}
          <div className="mt-10 flex flex-col items-center gap-4 sm:flex-row sm:justify-center">
            <Button size="lg" className="glow-md h-13 px-8 text-base" asChild>
              <Link href="/signup">
                무료로 시작하기
                <ArrowRight className="ml-2 h-4 w-4" />
              </Link>
            </Button>
            <Button variant="outline" size="lg" className="h-13 px-8 text-base" asChild>
              <Link href="/login">로그인</Link>
            </Button>
          </div>

          <p className="mt-5 text-sm text-muted-foreground/50">
            가입 즉시 무료 체험 — 신용카드 불필요
          </p>
        </div>

        {/* Product preview mockup */}
        <div className="relative mx-auto mt-16 max-w-4xl animate-fade-in lg:mt-20">
          <div className="rounded-xl border border-white/[0.08] bg-card/80 p-2 shadow-2xl shadow-black/40">
            <div className="rounded-lg border border-white/[0.06] bg-background">
              {/* Fake app header */}
              <div className="flex items-center gap-2 border-b border-white/[0.06] px-4 py-3">
                <div className="flex gap-1.5">
                  <div className="h-2.5 w-2.5 rounded-full bg-white/10" />
                  <div className="h-2.5 w-2.5 rounded-full bg-white/10" />
                  <div className="h-2.5 w-2.5 rounded-full bg-white/10" />
                </div>
                <div className="ml-4 h-5 w-48 rounded bg-white/[0.04]" />
              </div>
              {/* Fake app content */}
              <div className="grid grid-cols-12 gap-0">
                {/* Fake sidebar */}
                <div className="col-span-3 border-r border-white/[0.06] p-4 lg:col-span-2">
                  <div className="space-y-3">
                    <div className="h-3 w-16 rounded bg-primary/20" />
                    <div className="h-3 w-20 rounded bg-white/[0.06]" />
                    <div className="h-3 w-14 rounded bg-white/[0.06]" />
                    <div className="h-3 w-18 rounded bg-white/[0.06]" />
                    <div className="h-3 w-12 rounded bg-white/[0.06]" />
                  </div>
                </div>
                {/* Fake main content */}
                <div className="col-span-9 p-6 lg:col-span-10">
                  <div className="flex items-center justify-between">
                    <div className="h-5 w-32 rounded bg-white/[0.08]" />
                    <div className="h-7 w-24 rounded-md bg-primary/20" />
                  </div>
                  <div className="mt-6 grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
                    {[...Array(3)].map((_, i) => (
                      <div key={i} className="rounded-lg border border-white/[0.06] p-4">
                        <div className="mb-3 h-3 w-20 rounded bg-white/[0.08]" />
                        <div className="space-y-2">
                          <div className="h-2 w-full rounded bg-white/[0.04]" />
                          <div className="h-2 w-3/4 rounded bg-white/[0.04]" />
                          <div className="h-2 w-1/2 rounded bg-white/[0.04]" />
                        </div>
                        <div className="mt-4 flex gap-2">
                          <div className="h-5 w-12 rounded-full bg-primary/10" />
                          <div className="h-5 w-14 rounded-full bg-blue-500/10" />
                        </div>
                      </div>
                    ))}
                  </div>
                </div>
              </div>
            </div>
          </div>
          {/* Glow behind the mockup */}
          <div className="absolute -inset-4 -z-10 rounded-2xl bg-primary/[0.06] blur-3xl" />
        </div>
      </div>
    </section>
  );
}
