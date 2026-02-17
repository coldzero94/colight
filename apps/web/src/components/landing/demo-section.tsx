"use client";

import { useState } from "react";
import type { LucideIcon } from "lucide-react";
import {
  Link2,
  Search,
  GitCompare,
  FileEdit,
  CheckCircle,
  Sparkles,
  Radar,
  Target,
} from "lucide-react";

interface StepMeta {
  key: "url" | "analysis" | "matching" | "draft" | "review";
  icon: LucideIcon;
  title: string;
  description: string;
}

const steps: StepMeta[] = [
  {
    key: "url",
    icon: Link2,
    title: "URL 입력",
    description: "채용공고 링크 입력",
  },
  {
    key: "analysis",
    icon: Search,
    title: "자동 분석",
    description: "기업 인사이트 추출",
  },
  {
    key: "matching",
    icon: GitCompare,
    title: "소재 매칭",
    description: "STAR 경험 추천",
  },
  {
    key: "draft",
    icon: FileEdit,
    title: "초안 작성",
    description: "문항별 구조화 작성",
  },
  {
    key: "review",
    icon: CheckCircle,
    title: "AI 첨삭",
    description: "최종 개선 가이드",
  },
];

function UrlInputPanel() {
  return (
    <div className="grid gap-3 lg:grid-cols-[1.1fr_0.9fr]">
      <div className="rounded-xl border border-white/10 bg-white/[0.03] p-3">
        <p className="text-[11px] uppercase tracking-[0.18em] text-muted-foreground/70">
          URL 입력 센터
        </p>
        <p className="mt-1 text-sm font-semibold text-foreground">
          채용공고 URL 붙여넣기
        </p>
        <div className="mt-3 space-y-2">
          <div className="flex items-center gap-2 rounded-lg border border-white/12 bg-black/30 px-3 py-2">
            <span className="text-[10px] text-muted-foreground">https://</span>
            <span className="text-xs text-foreground/90">www.jobkorea.co.kr/Recruit/...</span>
          </div>
          <div className="flex items-center gap-2">
            <button className="rounded-lg bg-primary px-3 py-1.5 text-xs font-medium text-primary-foreground">
              분석 시작
            </button>
            <span className="rounded-full border border-cyan-400/25 bg-cyan-400/12 px-2 py-1 text-[10px] text-cyan-200">
              예상 12초
            </span>
          </div>
        </div>
        <div className="mt-3 flex flex-wrap gap-1.5">
          {["잡코리아", "원티드", "사람인", "캐치"].map((site) => (
            <span
              key={site}
              className="rounded-full border border-white/12 bg-white/[0.03] px-2 py-0.5 text-[10px] text-muted-foreground"
            >
              {site}
            </span>
          ))}
        </div>
      </div>

      <div className="rounded-xl border border-white/10 bg-white/[0.03] p-3">
        <p className="text-[11px] uppercase tracking-[0.18em] text-muted-foreground/70">
          최근 등록
        </p>
        <div className="mt-3 space-y-2">
          {[
            ["카카오모빌리티", "Backend Engineer", "수집 중"],
            ["토스", "Data Analyst", "분석 대기"],
            ["네이버", "Frontend", "완료"],
          ].map(([company, role, state]) => (
            <div
              key={company}
              className="rounded-lg border border-white/10 bg-black/20 px-2.5 py-2"
            >
              <p className="text-xs font-medium text-foreground">{company}</p>
              <p className="mt-0.5 text-[11px] text-muted-foreground">{role}</p>
              <p className="mt-1 text-[10px] text-cyan-200">{state}</p>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}

function AnalysisPanel() {
  return (
    <div className="grid gap-3 lg:grid-cols-[1fr_1fr]">
      <div className="rounded-xl border border-white/10 bg-white/[0.03] p-3">
        <p className="text-[11px] uppercase tracking-[0.18em] text-muted-foreground/70">
          기업 분석 리포트
        </p>
        <p className="mt-1 text-sm font-semibold text-foreground">카카오모빌리티</p>
        <div className="mt-3 flex flex-wrap gap-1.5">
          {["고객집착", "빠른실행", "협업문화", "도전정신"].map((keyword) => (
            <span
              key={keyword}
              className="rounded-full border border-cyan-400/25 bg-cyan-400/12 px-2 py-0.5 text-[10px] text-cyan-200"
            >
              {keyword}
            </span>
          ))}
        </div>
        <div className="mt-3 space-y-2">
          {[
            ["핵심가치 추출", "27"],
            ["직무 키워드", "18"],
            ["최근 트렌드", "6"],
          ].map(([label, value]) => (
            <div key={label} className="rounded-lg border border-white/10 bg-black/20 px-2.5 py-2">
              <p className="text-[11px] text-muted-foreground">{label}</p>
              <p className="mt-1 text-base font-semibold text-foreground">{value}</p>
            </div>
          ))}
        </div>
      </div>

      <div className="rounded-xl border border-white/10 bg-white/[0.03] p-3">
        <p className="text-[11px] uppercase tracking-[0.18em] text-muted-foreground/70">
          인재상 적합도
        </p>
        <div className="mt-3 space-y-2.5">
          {[
            ["문제 해결 능력", 89],
            ["협업/소통", 84],
            ["실행력", 92],
            ["성장 가능성", 86],
          ].map(([name, score]) => (
            <div key={name} className="space-y-1">
              <div className="flex items-center justify-between">
                <p className="text-[11px] text-muted-foreground">{name}</p>
                <p className="text-[11px] text-cyan-200">{score}</p>
              </div>
              <div className="h-1.5 rounded-full bg-white/10">
                <div
                  className="h-1.5 rounded-full bg-gradient-to-r from-primary to-cyan-300"
                  style={{ width: `${score}%` }}
                />
              </div>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}

function MatchingPanel() {
  return (
    <div className="grid gap-3 lg:grid-cols-[1.1fr_0.9fr]">
      <div className="rounded-xl border border-white/10 bg-white/[0.03] p-3">
        <p className="text-[11px] uppercase tracking-[0.18em] text-muted-foreground/70">
          STAR 경험 목록
        </p>
        <div className="mt-3 space-y-2">
          {[
            {
              title: "인턴 배포 자동화 구축",
              score: 92,
              weapons: ["문제해결", "성장"],
            },
            {
              title: "서비스 장애 대응 리드",
              score: 88,
              weapons: ["위기극복", "소통"],
            },
            {
              title: "학회 플랫폼 리뉴얼",
              score: 83,
              weapons: ["도전", "팀워크"],
            },
          ].map((item) => (
            <div
              key={item.title}
              className="rounded-lg border border-white/10 bg-black/20 px-2.5 py-2"
            >
              <div className="flex items-center justify-between gap-2">
                <p className="text-xs font-medium text-foreground">{item.title}</p>
                <span className="rounded-full border border-emerald-400/30 bg-emerald-400/12 px-2 py-0.5 text-[10px] text-emerald-200">
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

      <div className="rounded-xl border border-white/10 bg-white/[0.03] p-3">
        <p className="inline-flex items-center gap-1 text-[11px] uppercase tracking-[0.18em] text-muted-foreground/70">
          <Radar className="h-3.5 w-3.5 text-cyan-300" />
          무기 커버리지
        </p>
        <div className="mt-3 rounded-lg border border-white/10 bg-black/20 p-2.5">
          <p className="text-[11px] text-muted-foreground">필수 역량 충족률</p>
          <p className="mt-1 text-lg font-semibold text-foreground">84%</p>
          <div className="mt-2 h-2 rounded-full bg-white/10">
            <div className="h-2 w-[84%] rounded-full bg-gradient-to-r from-primary to-cyan-300" />
          </div>
        </div>
        <div className="mt-3 space-y-2">
          {[
            ["강점", "문제해결, 팀워크"],
            ["보완 필요", "소통/설득"],
          ].map(([label, text]) => (
            <div key={label} className="rounded-lg border border-white/10 bg-black/20 px-2.5 py-2">
              <p className="text-[11px] text-muted-foreground">{label}</p>
              <p className="mt-1 text-xs text-foreground">{text}</p>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}

function DraftPanel() {
  return (
    <div className="grid gap-3 lg:grid-cols-[1fr_1fr]">
      <div className="rounded-xl border border-white/10 bg-white/[0.03] p-3">
        <p className="text-[11px] uppercase tracking-[0.18em] text-muted-foreground/70">
          문항 입력
        </p>
        <div className="mt-2 rounded-lg border border-white/10 bg-black/20 p-2.5">
          <p className="text-xs text-foreground">
            본인이 팀 프로젝트에서 문제를 해결한 경험을 구체적으로 작성해 주세요.
          </p>
        </div>

        <p className="mt-3 text-[11px] uppercase tracking-[0.18em] text-muted-foreground/70">
          STAR 초안
        </p>
        <div className="mt-2 space-y-2">
          {[
            ["S", "모바일 서비스 트래픽 급증으로 장애 발생"],
            ["T", "30분 내 복구, 재발 방지 체계 수립"],
            ["A", "로그 분석 자동화 + 핫픽스 배포 프로세스 구축"],
            ["R", "복구 시간 48% 단축, 재발률 0건"],
          ].map(([label, text]) => (
            <div key={label} className="rounded-lg border border-white/10 bg-black/20 px-2.5 py-2">
              <p className="text-[10px] text-cyan-200">{label}</p>
              <p className="mt-1 text-[11px] text-foreground/90">{text}</p>
            </div>
          ))}
        </div>
      </div>

      <div className="rounded-xl border border-white/10 bg-white/[0.03] p-3">
        <p className="text-[11px] uppercase tracking-[0.18em] text-muted-foreground/70">
          초안 프리뷰
        </p>
        <div className="mt-2 rounded-lg border border-white/10 bg-black/20 p-3">
          <p className="text-xs leading-relaxed text-foreground/90">
            인턴 기간 중 서비스 장애 상황에서 로그 기반 원인 분석 자동화를 설계하고
            배포 프로세스를 개선했습니다. 그 결과 복구 시간은 48% 단축되었고
            이후 동일 장애의 재발은 발생하지 않았습니다...
          </p>
        </div>
        <div className="mt-3 grid grid-cols-2 gap-2">
          {[
            ["글자수", "742 / 800"],
            ["직무 적합", "A"],
            ["구체성", "A-"],
            ["기업 적합", "B+"],
          ].map(([label, value]) => (
            <div key={label} className="rounded-lg border border-white/10 bg-black/20 px-2.5 py-2">
              <p className="text-[10px] text-muted-foreground">{label}</p>
              <p className="mt-1 text-xs font-medium text-foreground">{value}</p>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}

function ReviewPanel() {
  return (
    <div className="grid gap-3 lg:grid-cols-[1fr_1fr]">
      <div className="rounded-xl border border-white/10 bg-white/[0.03] p-3">
        <p className="text-[11px] uppercase tracking-[0.18em] text-muted-foreground/70">
          최종 첨삭 리포트
        </p>
        <div className="mt-3 grid grid-cols-2 gap-2">
          {[
            ["직무적합성", "91"],
            ["구체성", "88"],
            ["진정성", "86"],
            ["기업적합", "84"],
          ].map(([label, score]) => (
            <div key={label} className="rounded-lg border border-white/10 bg-black/20 px-2.5 py-2">
              <p className="text-[10px] text-muted-foreground">{label}</p>
              <p className="mt-1 text-base font-semibold text-foreground">{score}</p>
            </div>
          ))}
        </div>
        <div className="mt-3 rounded-lg border border-emerald-400/25 bg-emerald-400/10 px-2.5 py-2">
          <p className="text-[11px] text-emerald-200">합격 예측 점수 89</p>
        </div>
      </div>

      <div className="rounded-xl border border-white/10 bg-white/[0.03] p-3">
        <p className="inline-flex items-center gap-1 text-[11px] uppercase tracking-[0.18em] text-muted-foreground/70">
          <Target className="h-3.5 w-3.5 text-primary" />
          개선 제안
        </p>
        <div className="mt-3 space-y-2">
          {[
            "성과 수치를 문장 첫머리에 배치해 임팩트를 강화하세요.",
            "기업 핵심가치(고객집착)와 경험 연결 문장을 1줄 추가하세요.",
            "행동(Action) 파트의 의사결정 근거를 더 명확히 드러내세요.",
          ].map((item) => (
            <div key={item} className="rounded-lg border border-white/10 bg-black/20 px-2.5 py-2">
              <p className="text-[11px] leading-relaxed text-foreground/90">{item}</p>
            </div>
          ))}
        </div>
        <div className="mt-3 flex items-center gap-2 rounded-lg border border-cyan-400/25 bg-cyan-400/10 px-2.5 py-2">
          <Sparkles className="h-4 w-4 text-cyan-300" />
          <p className="text-[11px] text-cyan-100">
            클릭 한 번으로 개선 문장을 본문에 자동 반영할 수 있습니다.
          </p>
        </div>
      </div>
    </div>
  );
}

function renderPanel(key: StepMeta["key"]) {
  switch (key) {
    case "url":
      return <UrlInputPanel />;
    case "analysis":
      return <AnalysisPanel />;
    case "matching":
      return <MatchingPanel />;
    case "draft":
      return <DraftPanel />;
    case "review":
      return <ReviewPanel />;
    default:
      return null;
  }
}

export function DemoSection() {
  const [activeIndex, setActiveIndex] = useState(0);

  const activeStep = steps[activeIndex];
  const progress = ((activeIndex + 1) / steps.length) * 100;

  const handleStepChange = (nextIndex: number) => {
    const targetIndex = Math.max(0, Math.min(steps.length - 1, nextIndex));
    setActiveIndex(targetIndex);
  };

  return (
    <section className="relative overflow-hidden py-28 lg:py-36">
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

        <div className="mt-16 grid gap-8 lg:grid-cols-[minmax(0,1.12fr)_minmax(0,0.88fr)] lg:items-start">
          <div>
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
                    {activeStep.title}
                  </span>
                </div>

                <div className="border-b border-white/[0.08] px-4 py-3">
                  <div className="flex flex-wrap gap-1.5">
                    {steps.map((step, index) => (
                      <button
                        key={step.title}
                        type="button"
                        onClick={() => handleStepChange(index)}
                        className={`rounded-full border px-2.5 py-1 text-[10px] transition-colors ${
                          index === activeIndex
                            ? "border-primary/35 bg-primary/12 text-primary"
                            : "border-white/12 bg-white/[0.03] text-muted-foreground hover:border-white/25 hover:text-foreground"
                        }`}
                      >
                        {step.title}
                      </button>
                    ))}
                  </div>
                </div>

                <div key={activeStep.key} className="animate-fade-in p-4 sm:p-5">
                  {renderPanel(activeStep.key)}
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
                        <h3 className="text-base font-semibold text-foreground">
                          {step.title}
                        </h3>
                        <p className="mt-1 text-sm text-muted-foreground">
                          {step.description}
                        </p>
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
