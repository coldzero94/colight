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
          <div className="mb-7 inline-flex items-center gap-2 rounded-full border border-primary/35 bg-primary/10 px-4 py-1.5 text-sm text-primary shadow-[0_0_30px_rgba(56,189,248,0.25)]">
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
            <div className="overflow-hidden rounded-2xl border border-white/10 bg-black/35">
              <div className="flex items-center gap-2 border-b border-white/[0.08] px-4 py-3">
                <div className="flex gap-1.5">
                  <div className="h-2.5 w-2.5 rounded-full bg-white/10" />
                  <div className="h-2.5 w-2.5 rounded-full bg-white/10" />
                  <div className="h-2.5 w-2.5 rounded-full bg-white/10" />
                </div>
                <div className="ml-3 rounded-md border border-white/10 bg-white/[0.03] px-2.5 py-1 text-[11px] text-muted-foreground">
                  Colight Workspace
                </div>
                <div className="ml-auto rounded-full border border-cyan-400/25 bg-cyan-400/12 px-2.5 py-1 text-[10px] uppercase tracking-[0.14em] text-cyan-200">
                  live preview
                </div>
              </div>
              <div className="grid grid-cols-12">
                <div className="col-span-4 border-r border-white/[0.08] p-4">
                  <p className="mb-3 text-[10px] uppercase tracking-[0.2em] text-muted-foreground/70">
                    메뉴
                  </p>
                  <div className="space-y-2.5">
                    <div className="rounded-lg border border-primary/35 bg-primary/12 px-2.5 py-2 text-xs text-primary">
                      STAR 경험 목록
                    </div>
                    <div className="rounded-lg border border-white/10 bg-white/[0.03] px-2.5 py-2 text-xs text-muted-foreground">
                      기업 분석 리포트
                    </div>
                    <div className="rounded-lg border border-white/10 bg-white/[0.03] px-2.5 py-2 text-xs text-muted-foreground">
                      경험 매칭 보드
                    </div>
                    <div className="rounded-lg border border-white/10 bg-white/[0.03] px-2.5 py-2 text-xs text-muted-foreground">
                      코칭 히스토리
                    </div>
                  </div>
                </div>
                <div className="col-span-8 p-4 sm:p-5">
                  <div className="mb-4 flex items-center justify-between">
                    <div>
                      <p className="text-[10px] uppercase tracking-[0.18em] text-muted-foreground/70">
                        Live Dashboard
                      </p>
                      <p className="mt-1 text-sm font-semibold text-foreground">
                        STAR 경험 목록
                      </p>
                    </div>
                    <div className="rounded-full border border-emerald-400/30 bg-emerald-400/12 px-2.5 py-1 text-[10px] text-emerald-200">
                      3개 추천 완료
                    </div>
                  </div>
                  <div className="space-y-3">
                    {[
                      {
                        title: "교내 서비스 장애 대응 경험",
                        note: "문제해결 / 팀워크 무기와 매칭률 92%",
                        tags: ["S", "A", "R"],
                      },
                      {
                        title: "인턴 기간 배치 배포 자동화",
                        note: "직무적합성 키워드와 연결. STAR 구조 자동 정리 완료",
                        tags: ["T", "A", "R"],
                      },
                    ].map((item, i) => (
                      <div
                        key={item.title}
                        className="rounded-xl border border-white/[0.08] bg-white/[0.03] p-3 animate-float-y-soft"
                        style={{ animationDelay: `${i * 0.45}s` }}
                      >
                        <p className="text-xs font-medium text-foreground">{item.title}</p>
                        <p className="mt-1.5 text-[11px] text-muted-foreground">{item.note}</p>
                        <div className="mt-2.5 flex gap-1.5">
                          {item.tags.map((tag) => (
                            <span
                              key={tag}
                              className="rounded-full border border-violet-400/25 bg-violet-400/15 px-2 py-0.5 text-[10px] text-violet-200"
                            >
                              {tag}
                            </span>
                          ))}
                          <span className="rounded-full border border-cyan-400/25 bg-cyan-400/15 px-2 py-0.5 text-[10px] text-cyan-200">
                            MATCH
                          </span>
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
