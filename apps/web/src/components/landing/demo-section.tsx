"use client";

import { useEffect, useRef, useState } from "react";
import type { LucideIcon } from "lucide-react";
import {
  Link2,
  Search,
  GitCompare,
  FileEdit,
  CheckCircle,
  Activity,
  BarChart3,
  Radar,
} from "lucide-react";

interface DemoStep {
  icon: LucideIcon;
  title: string;
  description: string;
  stageLabel: string;
  panelTitle: string;
  statCards: Array<{
    label: string;
    value: string;
    delta: string;
    tone: "cyan" | "violet" | "emerald" | "amber";
  }>;
  chart: number[];
  board: Array<{
    status: string;
    count: number;
    accent: "cyan" | "violet" | "emerald" | "amber";
  }>;
  starList: Array<{
    title: string;
    score: number;
    weapons: string[];
  }>;
  feed: Array<{
    company: string;
    note: string;
    tags: string[];
  }>;
}

const toneClass: Record<DemoStep["statCards"][number]["tone"], string> = {
  cyan: "text-cyan-300 bg-cyan-500/15 border-cyan-400/25",
  violet: "text-violet-300 bg-violet-500/15 border-violet-400/25",
  emerald: "text-emerald-300 bg-emerald-500/15 border-emerald-400/25",
  amber: "text-amber-300 bg-amber-500/15 border-amber-400/25",
};

const boardAccentClass: Record<DemoStep["board"][number]["accent"], string> = {
  cyan: "bg-cyan-400/20 text-cyan-200 border-cyan-400/25",
  violet: "bg-violet-400/20 text-violet-200 border-violet-400/25",
  emerald: "bg-emerald-400/20 text-emerald-200 border-emerald-400/25",
  amber: "bg-amber-400/20 text-amber-200 border-amber-400/25",
};

const steps: DemoStep[] = [
  {
    icon: Link2,
    title: "URL 입력",
    description: "채용공고 링크 붙여넣기",
    stageLabel: "intake stage",
    panelTitle: "채용공고 분석 대기열",
    statCards: [
      { label: "신규 공고", value: "12", delta: "+3", tone: "cyan" },
      { label: "자동 인식률", value: "98%", delta: "+2.1%", tone: "emerald" },
      { label: "처리 대기", value: "4", delta: "-1", tone: "amber" },
    ],
    chart: [42, 48, 56, 61, 68, 73, 80],
    board: [
      { status: "크롤링 대기", count: 4, accent: "amber" },
      { status: "정보 수집중", count: 5, accent: "cyan" },
      { status: "분석 준비 완료", count: 3, accent: "emerald" },
    ],
    starList: [
      { title: "캡스톤 API 성능 개선", score: 81, weapons: ["문제해결", "성장"] },
      { title: "교내 서비스 장애 대응", score: 77, weapons: ["위기극복", "팀워크"] },
    ],
    feed: [
      {
        company: "현대오토에버",
        note: "백엔드 채용공고 등록. 직무 키워드 18개 추출 완료.",
        tags: ["Java", "MSA"],
      },
      {
        company: "토스",
        note: "Data 직무 링크 수집. 기업 트렌드 데이터 연결 대기.",
        tags: ["Data", "Fintech"],
      },
    ],
  },
  {
    icon: Search,
    title: "자동 분석",
    description: "기업 · 직무 정보 파악",
    stageLabel: "analysis stage",
    panelTitle: "기업 인사이트 분석 리포트",
    statCards: [
      { label: "핵심가치 추출", value: "27", delta: "+6", tone: "cyan" },
      { label: "신뢰 점수", value: "91", delta: "+5", tone: "violet" },
      { label: "최신 뉴스 반영", value: "8", delta: "+2", tone: "emerald" },
    ],
    chart: [36, 46, 58, 64, 70, 79, 88],
    board: [
      { status: "기업가치 맵핑", count: 6, accent: "violet" },
      { status: "직무 키워드", count: 9, accent: "cyan" },
      { status: "리스크 점검", count: 3, accent: "amber" },
    ],
    starList: [
      { title: "데이터 파이프라인 최적화", score: 85, weapons: ["문제해결", "리더십"] },
      { title: "신규 기능 A/B 실험", score: 79, weapons: ["도전", "소통"] },
    ],
    feed: [
      {
        company: "카카오모빌리티",
        note: "인재상 키워드 업데이트. 고객집착/실행력 우선순위 상향.",
        tags: ["인재상", "실행력"],
      },
      {
        company: "우아한형제들",
        note: "분석 출처 동기화 완료. 최근 전략 키워드 반영됨.",
        tags: ["전략", "브랜드"],
      },
    ],
  },
  {
    icon: GitCompare,
    title: "소재 매칭",
    description: "내 경험에서 최적 소재",
    stageLabel: "matching stage",
    panelTitle: "경험 매칭 스코어보드",
    statCards: [
      { label: "매칭 경험 수", value: "36", delta: "+8", tone: "violet" },
      { label: "평균 적합도", value: "84%", delta: "+11%", tone: "cyan" },
      { label: "중복 제거", value: "14", delta: "-5", tone: "emerald" },
    ],
    chart: [28, 42, 53, 67, 72, 84, 92],
    board: [
      { status: "상위 추천", count: 7, accent: "emerald" },
      { status: "재가공 필요", count: 5, accent: "amber" },
      { status: "보류", count: 2, accent: "violet" },
    ],
    starList: [
      { title: "인턴 배포 자동화 구축", score: 92, weapons: ["문제해결", "성장"] },
      { title: "프로젝트 커뮤니케이션 리드", score: 88, weapons: ["소통", "팀워크"] },
    ],
    feed: [
      {
        company: "네이버",
        note: "팀 프로젝트 경험이 협업 역량에서 가장 높은 스코어를 기록.",
        tags: ["협업", "문제해결"],
      },
      {
        company: "라인",
        note: "도전정신 무기 비중이 낮아 보완 경험 추가 추천 생성.",
        tags: ["도전", "보완"],
      },
    ],
  },
  {
    icon: FileEdit,
    title: "초안 작성",
    description: "STAR 구조로 자동 생성",
    stageLabel: "draft stage",
    panelTitle: "문항별 STAR 초안 생성",
    statCards: [
      { label: "초안 생성", value: "18", delta: "+4", tone: "cyan" },
      { label: "평균 완성도", value: "87%", delta: "+7%", tone: "emerald" },
      { label: "문항 적합도", value: "A-", delta: "상승", tone: "violet" },
    ],
    chart: [38, 44, 57, 69, 76, 83, 90],
    board: [
      { status: "작성중", count: 6, accent: "cyan" },
      { status: "검토 대기", count: 4, accent: "amber" },
      { status: "완성", count: 8, accent: "emerald" },
    ],
    starList: [
      { title: "운영 자동화로 장애율 감소", score: 90, weapons: ["문제해결", "위기극복"] },
      { title: "학회 서비스 리뉴얼", score: 84, weapons: ["도전", "팀워크"] },
    ],
    feed: [
      {
        company: "쿠팡",
        note: "문항 2번 초안 생성 완료. Action 단락에 수치 근거 추가됨.",
        tags: ["STAR", "정량화"],
      },
      {
        company: "당근",
        note: "초안 길이 자동 최적화. 950자 -> 780자로 압축 적용.",
        tags: ["압축", "가독성"],
      },
    ],
  },
  {
    icon: CheckCircle,
    title: "AI 첨삭",
    description: "4점 코칭으로 완성",
    stageLabel: "coaching stage",
    panelTitle: "실시간 첨삭 및 최종 점검",
    statCards: [
      { label: "코칭 완료", value: "42", delta: "+9", tone: "emerald" },
      { label: "합격 예측 점수", value: "89", delta: "+12", tone: "cyan" },
      { label: "개선 제안", value: "23", delta: "-4", tone: "violet" },
    ],
    chart: [51, 58, 66, 74, 82, 88, 94],
    board: [
      { status: "강점 강조", count: 11, accent: "emerald" },
      { status: "표현 정제", count: 7, accent: "violet" },
      { status: "리스크 문장", count: 3, accent: "amber" },
    ],
    starList: [
      { title: "서비스 전환율 2배 개선", score: 94, weapons: ["문제해결", "리더십"] },
      { title: "협업 프로세스 표준화", score: 89, weapons: ["소통", "팀워크"] },
    ],
    feed: [
      {
        company: "삼성전자",
        note: "직무적합성 문단 개선 제안 반영. 설득력 점수 +13.",
        tags: ["직무적합", "설득력"],
      },
      {
        company: "SK하이닉스",
        note: "최종 검토 완료. 제출 체크리스트 100% 달성.",
        tags: ["최종점검", "완료"],
      },
    ],
  },
];

export function DemoSection() {
  const sectionRef = useRef<HTMLElement | null>(null);
  const [activeIndex, setActiveIndex] = useState(0);
  const [progress, setProgress] = useState(0);

  useEffect(() => {
    let rafId = 0;

    const updateByScroll = () => {
      if (!sectionRef.current) return;

      const rect = sectionRef.current.getBoundingClientRect();
      const viewport = window.innerHeight || 1;
      const scrollRange = rect.height - viewport * 0.5;
      const traveled = viewport * 0.32 - rect.top;
      const nextProgress =
        scrollRange <= 0
          ? rect.top < viewport * 0.32
            ? 1
            : 0
          : Math.max(0, Math.min(1, traveled / scrollRange));

      setProgress((prev) => (Math.abs(prev - nextProgress) > 0.001 ? nextProgress : prev));

      const nextIndex = Math.round(nextProgress * (steps.length - 1));
      setActiveIndex((prev) => (prev !== nextIndex ? nextIndex : prev));
    };

    const onScroll = () => {
      cancelAnimationFrame(rafId);
      rafId = window.requestAnimationFrame(updateByScroll);
    };

    onScroll();
    window.addEventListener("scroll", onScroll, { passive: true });
    window.addEventListener("resize", onScroll);

    return () => {
      cancelAnimationFrame(rafId);
      window.removeEventListener("scroll", onScroll);
      window.removeEventListener("resize", onScroll);
    };
  }, []);

  const activeStep = steps[activeIndex];

  const handleStepChange = (nextIndex: number) => {
    if (!sectionRef.current) return;

    const targetIndex = Math.max(0, Math.min(steps.length - 1, nextIndex));
    const nextProgress = targetIndex / (steps.length - 1);
    setActiveIndex(targetIndex);
    setProgress(nextProgress);

    const rect = sectionRef.current.getBoundingClientRect();
    const sectionTop = window.scrollY + rect.top;
    const viewport = window.innerHeight || 1;
    const scrollRange = sectionRef.current.offsetHeight - viewport * 0.5;
    const targetY =
      sectionTop - viewport * 0.32 + Math.max(0, scrollRange) * nextProgress;

    window.scrollTo({
      top: targetY,
      behavior: "smooth",
    });
  };

  return (
    <section ref={sectionRef} className="relative overflow-hidden py-28 lg:py-36">
      <div className="pointer-events-none absolute inset-0 -z-10">
        <div className="absolute left-1/2 top-1/2 h-[460px] w-[900px] -translate-x-1/2 -translate-y-1/2 rounded-full bg-primary/10 blur-[140px]" />
        <div className="absolute left-[10%] top-[26%] h-28 w-28 rounded-full border border-white/10 animate-float-y-soft" />
        <div className="absolute right-[10%] top-[58%] h-24 w-24 rounded-full border border-cyan-300/20 animate-float-y" />
      </div>

      <div className="mx-auto max-w-6xl px-6">
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

        <div className="mt-16 grid gap-8 lg:grid-cols-[minmax(0,1.1fr)_minmax(0,0.9fr)] lg:items-start">
          <div className="lg:sticky lg:top-24">
            <div className="glass landing-panel rounded-3xl border border-white/12 p-3 shadow-[0_24px_90px_rgba(0,0,0,0.45)]">
              <div className="overflow-hidden rounded-2xl border border-white/10 bg-black/35">
                <div className="flex items-center gap-3 border-b border-white/[0.08] px-4 py-3">
                  <div className="flex gap-1.5">
                    <div className="h-2.5 w-2.5 rounded-full bg-white/10" />
                    <div className="h-2.5 w-2.5 rounded-full bg-white/10" />
                    <div className="h-2.5 w-2.5 rounded-full bg-white/10" />
                  </div>
                  <div className="h-5 flex-1 rounded bg-white/[0.05]" />
                  <span className="rounded-full border border-cyan-400/25 bg-cyan-400/12 px-2.5 py-1 text-[10px] uppercase tracking-[0.16em] text-cyan-200">
                    {activeStep.stageLabel}
                  </span>
                </div>

                <div className="grid grid-cols-12">
                  <div className="col-span-4 hidden border-r border-white/[0.08] p-4 sm:block">
                    <div className="space-y-2">
                      <p className="text-[10px] uppercase tracking-[0.2em] text-muted-foreground/70">
                        Workspace
                      </p>
                      {steps.map((step, index) => (
                        <button
                          key={step.title}
                          type="button"
                          onClick={() => handleStepChange(index)}
                          className={`w-full rounded-lg border px-2.5 py-2 text-left text-xs transition-colors ${
                            index === activeIndex
                              ? "border-primary/35 bg-primary/12 text-primary"
                              : "border-white/10 bg-white/[0.03] text-muted-foreground hover:border-white/20 hover:text-foreground"
                          }`}
                        >
                          {step.title}
                        </button>
                      ))}
                    </div>
                  </div>

                  <div className="col-span-12 p-4 sm:col-span-8 sm:p-5">
                    <div key={activeStep.title} className="animate-fade-in space-y-4">
                      <div className="flex items-start justify-between gap-2">
                        <div>
                          <p className="text-[11px] uppercase tracking-[0.2em] text-muted-foreground/70">
                            Live Dashboard
                          </p>
                          <h3 className="mt-1 text-sm font-semibold text-foreground">
                            {activeStep.panelTitle}
                          </h3>
                        </div>
                        <span className="inline-flex items-center gap-1 rounded-full border border-emerald-400/30 bg-emerald-400/12 px-2.5 py-1 text-[10px] text-emerald-200">
                          <Activity className="h-3 w-3" />
                          LIVE
                        </span>
                      </div>

                      <div className="grid grid-cols-1 gap-2 sm:grid-cols-3">
                        {activeStep.statCards.map((card) => (
                          <div
                            key={card.label}
                            className={`rounded-lg border px-2.5 py-2 ${toneClass[card.tone]}`}
                          >
                            <p className="text-[10px] text-muted-foreground/80">{card.label}</p>
                            <p className="mt-1 text-base font-semibold leading-none">{card.value}</p>
                            <p className="mt-1 text-[10px]">{card.delta}</p>
                          </div>
                        ))}
                      </div>

                      <div className="rounded-xl border border-white/10 bg-white/[0.03] p-3">
                        <div className="mb-2 flex items-center justify-between">
                          <p className="inline-flex items-center gap-1 text-[11px] text-muted-foreground">
                            <BarChart3 className="h-3.5 w-3.5 text-primary" />
                            지원 적합도 추이
                          </p>
                          <span className="text-[10px] text-primary">7-day snapshot</span>
                        </div>
                        <div className="flex h-24 items-end gap-1.5">
                          {activeStep.chart.map((value, index) => (
                            <div
                              key={`${activeStep.title}-chart-${index}`}
                              className="group relative flex-1 rounded-t-md bg-gradient-to-t from-primary/40 to-cyan-300/70 transition-all"
                              style={{
                                height: `${Math.max(14, value)}%`,
                                opacity: index <= activeIndex + 2 ? 1 : 0.45,
                              }}
                            >
                              <span className="absolute -top-5 left-1/2 -translate-x-1/2 text-[9px] text-muted-foreground/70 opacity-0 transition-opacity group-hover:opacity-100">
                                {value}
                              </span>
                            </div>
                          ))}
                        </div>
                      </div>

                      <div className="grid grid-cols-1 gap-2 sm:grid-cols-3">
                        {activeStep.board.map((item) => (
                          <div
                            key={item.status}
                            className={`rounded-lg border px-2.5 py-2 ${boardAccentClass[item.accent]}`}
                          >
                            <p className="text-[10px]">{item.status}</p>
                            <p className="mt-1 text-base font-semibold leading-none">{item.count}</p>
                          </div>
                        ))}
                      </div>

                      <div className="rounded-xl border border-white/10 bg-white/[0.03] p-3">
                        <div className="mb-2 flex items-center justify-between">
                          <p className="text-[11px] text-muted-foreground">STAR 경험 목록</p>
                          <span className="text-[10px] text-cyan-200">
                            TOP MATCHED EXPERIENCES
                          </span>
                        </div>
                        <div className="space-y-2">
                          {activeStep.starList.map((item) => (
                            <div
                              key={item.title}
                              className="rounded-lg border border-white/10 bg-black/20 px-2.5 py-2"
                            >
                              <div className="flex items-center justify-between gap-2">
                                <p className="text-xs font-medium text-foreground">{item.title}</p>
                                <span className="rounded-full border border-cyan-400/30 bg-cyan-400/15 px-2 py-0.5 text-[10px] text-cyan-200">
                                  {item.score}
                                </span>
                              </div>
                              <div className="mt-1.5 flex flex-wrap gap-1">
                                {item.weapons.map((weapon) => (
                                  <span
                                    key={weapon}
                                    className="rounded-full border border-violet-400/25 bg-violet-400/12 px-1.5 py-0.5 text-[10px] text-violet-200"
                                  >
                                    {weapon}
                                  </span>
                                ))}
                              </div>
                            </div>
                          ))}
                        </div>
                      </div>

                      <div className="space-y-2">
                        {activeStep.feed.map((item) => (
                          <div
                            key={`${activeStep.title}-${item.company}`}
                            className="rounded-lg border border-white/10 bg-white/[0.02] p-2.5"
                          >
                            <div className="flex items-center justify-between gap-2">
                              <p className="text-xs font-medium text-foreground">{item.company}</p>
                              <span className="inline-flex items-center gap-1 text-[10px] text-cyan-200">
                                <Radar className="h-3 w-3" />
                                synced
                              </span>
                            </div>
                            <p className="mt-1 text-[11px] leading-relaxed text-muted-foreground">
                              {item.note}
                            </p>
                            <div className="mt-2 flex flex-wrap gap-1">
                              {item.tags.map((tag) => (
                                <span
                                  key={tag}
                                  className="rounded-full border border-white/12 bg-white/[0.04] px-1.5 py-0.5 text-[10px] text-muted-foreground"
                                >
                                  {tag}
                                </span>
                              ))}
                            </div>
                          </div>
                        ))}
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>

          <div className="relative pl-0 lg:pl-9">
            <div className="pointer-events-none absolute left-2 top-2 hidden h-[calc(100%-0.5rem)] w-px bg-white/10 lg:block">
              <span
                className="absolute left-0 top-0 w-px rounded-full bg-gradient-to-b from-primary via-cyan-300 to-transparent transition-all duration-500"
                style={{ height: `${Math.max(8, progress * 100)}%` }}
              />
            </div>

            <div className="space-y-4">
              {steps.map((step, index) => {
                const isActive = index === activeIndex;
                return (
                  <button
                    key={step.title}
                    type="button"
                    onClick={() => handleStepChange(index)}
                    className={`w-full rounded-2xl border px-4 py-4 text-left transition-all duration-300 ${
                      isActive
                        ? "border-primary/35 bg-primary/10 shadow-[0_14px_34px_rgba(56,189,248,0.15)]"
                        : "border-white/10 bg-white/[0.02] hover:border-white/20"
                    }`}
                  >
                    <div className="flex items-start gap-3">
                      <span
                        className={`mt-0.5 inline-flex h-9 w-9 items-center justify-center rounded-xl border ${
                          isActive
                            ? "border-primary/35 bg-primary/15 text-primary"
                            : "border-white/12 bg-white/[0.03] text-muted-foreground"
                        }`}
                      >
                        <step.icon className="h-4 w-4" />
                      </span>
                      <div className="flex-1">
                        <h3 className="text-left text-base font-semibold text-foreground">{step.title}</h3>
                        <p className="mt-1 text-sm text-muted-foreground">{step.description}</p>
                      </div>
                    </div>
                  </button>
                );
              })}
            </div>
          </div>
        </div>
      </div>
    </section>
  );
}
