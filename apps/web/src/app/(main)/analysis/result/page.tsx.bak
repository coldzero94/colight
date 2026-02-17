"use client";

import { useRouter, useSearchParams } from "next/navigation";
import { useEffect, useState } from "react";
import { toast } from "sonner";
import { LoadingSpinner } from "@/components/common/loading-spinner";
import { useAnalyzeCompany } from "@/hooks/use-company-analysis";
import { crawlJobPosting } from "@/lib/api/crawl";
import { fetchExperiences, type Experience } from "@/lib/api/experiences";
import type { CompanyAnalysis } from "@/lib/api/analysis";
import { isUsageLimitError, getUsageLimitInfo } from "@/lib/api/errors";
import { matchExperiencesHeuristic, type HeuristicMatch } from "@/lib/utils/heuristic-matching";

type AnalysisStep = "crawling" | "analyzing" | "loading_experiences" | "matching" | "completed";

export default function AnalysisResultPage() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const url = searchParams.get("url");
  const companyParam = searchParams.get("company");
  const [companyName, setCompanyName] = useState<string | null>(companyParam);
  const [analysis, setAnalysis] = useState<CompanyAnalysis | null>(null);
  const [experiences, setExperiences] = useState<Experience[]>([]);
  const [crawlError, setCrawlError] = useState<string | null>(null);
  const [currentStep, setCurrentStep] = useState<AnalysisStep>(companyParam ? "analyzing" : "crawling");
  const [matches, setMatches] = useState<HeuristicMatch[]>([]);

  const analyzeMutation = useAnalyzeCompany();

  // Step 1: Crawl job posting to extract company name
  useEffect(() => {
    if (!url || companyName) return;

    setCurrentStep("crawling");
    (async () => {
      try {
        const jobPosting = await crawlJobPosting(url);
        const extractedName = jobPosting.company_name;
        setCompanyName(extractedName);
        setCurrentStep("analyzing");

        // Update URL to persist company name
        const newUrl = new URL(window.location.href);
        newUrl.searchParams.set("company", extractedName);
        window.history.replaceState({}, "", newUrl.toString());
      } catch {
        setCrawlError("채용공고에서 회사명을 추출하지 못했습니다.");
        toast.error("크롤링에 실패했습니다. URL을 확인해주세요.");
      }
    })();
  }, [url, companyName]);

  // Step 2: Analyze company once we have the name
  useEffect(() => {
    if (!companyName || analysis || analyzeMutation.isPending) return;

    setCurrentStep("analyzing");
    analyzeMutation.mutate({ companyName, url: url || undefined }, {
      onSuccess: (data) => {
        setAnalysis(data);
        setCurrentStep("loading_experiences");
      },
      onError: (error) => {
        if (isUsageLimitError(error)) {
          const info = getUsageLimitInfo(error);
          toast.error(info.message);
          router.push("/pricing");
        } else {
          toast.error("기업 분석에 실패했습니다.");
        }
      },
    });
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [companyName]);

  // Step 3: Load experiences and calculate heuristic matching
  useEffect(() => {
    if (!analysis) return;

    (async () => {
      try {
        const data = await fetchExperiences();
        setExperiences(data);

        // Calculate heuristic matching (instant, no AI)
        const matchResults = matchExperiencesHeuristic(data, analysis);
        setMatches(matchResults);
      } catch {
        // Non-critical: just show analysis without matching
      } finally {
        setCurrentStep("completed");
      }
    })();
  }, [analysis]);

  // Prevent accidental page close during analysis
  useEffect(() => {
    if (currentStep === "completed") return;

    const handleBeforeUnload = (e: BeforeUnloadEvent) => {
      e.preventDefault();
      e.returnValue = "";
    };

    window.addEventListener("beforeunload", handleBeforeUnload);
    return () => window.removeEventListener("beforeunload", handleBeforeUnload);
  }, [currentStep]);

  if (crawlError) {
    return (
      <div className="mx-auto max-w-4xl px-4 py-8">
        <div className="brand-surface flex flex-col items-center justify-center rounded-2xl py-20">
          <p className="text-red-400">{crawlError}</p>
          <button
            onClick={() => window.history.back()}
            className="mt-4 text-sm text-primary hover:text-primary/80"
          >
            ← 돌아가기
          </button>
        </div>
      </div>
    );
  }

  if (!analysis) {
    return (
      <div className="mx-auto max-w-4xl px-4 py-8">
        <div className="brand-surface flex flex-col items-center justify-center rounded-2xl py-20">
          <LoadingSpinner />
          <div className="mt-6 space-y-3 text-center">
            <p className="text-lg font-medium text-foreground">
              {getStepMessage(currentStep)}
            </p>
            <div className="flex items-center justify-center gap-2 text-sm text-muted-foreground">
              <div className="flex items-center gap-1.5">
                <StepIndicator active={currentStep === "crawling"} completed={currentStep !== "crawling"} />
                <span className={currentStep === "crawling" ? "text-primary" : ""}>1. 크롤링</span>
              </div>
              <span className="text-muted-foreground/40">→</span>
              <div className="flex items-center gap-1.5">
                <StepIndicator active={currentStep === "analyzing"} completed={currentStep === "loading_experiences" || currentStep === "completed"} />
                <span className={currentStep === "analyzing" ? "text-primary" : ""}>2. 기업 분석</span>
              </div>
              <span className="text-muted-foreground/40">→</span>
              <div className="flex items-center gap-1.5">
                <StepIndicator active={currentStep === "loading_experiences"} completed={currentStep === "completed"} />
                <span className={currentStep === "loading_experiences" ? "text-primary" : ""}>3. 경험 로드</span>
              </div>
            </div>
            <p className="text-xs text-muted-foreground/60">평균 10-15초 소요</p>
            <div className="mt-4 rounded-lg border border-amber-500/30 bg-amber-500/10 px-4 py-2.5">
              <p className="text-sm text-amber-200/90">
                ⚠️ 잠시만 기다려주세요. 페이지를 벗어나지 마세요.
              </p>
            </div>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="mx-auto max-w-4xl space-y-8 px-4 py-8 animate-fade-in">
      <div className="brand-surface relative overflow-hidden rounded-2xl px-6 py-5">
        <div className="pointer-events-none absolute -right-10 -top-10 h-48 w-48 rounded-full bg-primary/18 blur-3xl" />
        <div className="flex items-center justify-between">
          <div>
            <span className="brand-kicker inline-flex rounded-full px-3 py-1 text-[11px] font-semibold tracking-[0.18em] uppercase">
              Analysis Report
            </span>
            <h1 className="mt-3 text-3xl font-bold font-display text-foreground">
              {analysis.company_name}
            </h1>
            <div className="mt-2 flex items-center gap-3 text-sm text-muted-foreground">
              <span>분석 출처: {getSourceLabel(analysis.source)}</span>
              {analysis.view_count && analysis.view_count > 1 && (
                <span className="text-muted-foreground/60">
                  · 조회 {analysis.view_count}회
                </span>
              )}
              {analysis.cached_at && (
                <span className="text-muted-foreground/60">
                  · {formatCacheAge(analysis.cached_at)}
                </span>
              )}
            </div>
          </div>
          <button
            onClick={() => window.history.back()}
            className="brand-outline-btn rounded-xl px-3 py-1.5 text-sm text-muted-foreground hover:text-foreground"
          >
            ← 돌아가기
          </button>
        </div>
      </div>

      {analysis.core_values.length > 0 && (
        <section>
          <h2 className="mb-4 text-xl font-semibold text-foreground">
            핵심가치 ({analysis.core_values.length})
          </h2>
          <div className="grid gap-4 md:grid-cols-2">
            {analysis.core_values.map((value, i) => (
            <div
              key={i}
              className="brand-surface-soft rounded-xl p-4"
            >
              <h3 className="font-medium text-foreground">{value.keyword}</h3>
              <p className="mt-1 text-sm text-muted-foreground">{value.description}</p>
            </div>
            ))}
          </div>
        </section>
      )}

      {analysis.talent_traits.length > 0 && (
        <section>
          <h2 className="mb-4 text-xl font-semibold text-foreground">
            인재상 ({analysis.talent_traits.length})
        </h2>
        <div className="space-y-4">
          {analysis.talent_traits.map((trait, i) => (
            <div
              key={i}
              className="brand-surface-soft rounded-xl p-4"
            >
              <h3 className="font-medium text-foreground">{trait.trait}</h3>
              <p className="mt-1 text-sm text-muted-foreground">{trait.description}</p>
              {trait.evidence && (
                <p className="mt-2 text-xs text-muted-foreground/60">
                  근거: {trait.evidence}
                </p>
              )}
            </div>
            ))}
          </div>
        </section>
      )}

      {analysis.recent_trends.length > 0 && (
        <section>
          <h2 className="mb-4 text-xl font-semibold text-foreground">
            최근 동향 ({analysis.recent_trends.length})
          </h2>
          <div className="space-y-3">
            {analysis.recent_trends.map((trend, i) => (
              <div
                key={i}
                className="rounded-xl border border-primary/25 bg-primary/10 p-4"
              >
                <h3 className="font-medium text-foreground">{trend.title}</h3>
                <p className="mt-1 text-sm text-muted-foreground">{trend.summary}</p>
              </div>
            ))}
          </div>
        </section>
      )}

      {analysis.strategy_keywords.length > 0 && (
        <section>
          <h2 className="mb-4 text-xl font-semibold text-foreground">
            자소서 전략 키워드
          </h2>
          <div className="flex flex-wrap gap-2">
            {analysis.strategy_keywords.map((keyword, i) => (
              <span
                key={i}
                className="inline-flex items-center rounded-full border border-primary/30 bg-primary/12 px-3 py-1 text-sm font-medium text-primary"
              >
                {keyword}
              </span>
            ))}
          </div>
        </section>
      )}

      {analysis.avoid_expressions.length > 0 && (
        <section>
          <h2 className="mb-4 text-xl font-semibold text-foreground">
            피해야 할 표현
          </h2>
          <div className="flex flex-wrap gap-2">
            {analysis.avoid_expressions.map((expr, i) => (
              <span
                key={i}
                className="inline-flex items-center rounded-full border border-red-400/30 bg-red-500/10 px-3 py-1 text-sm font-medium text-red-300"
              >
                {expr}
              </span>
            ))}
          </div>
        </section>
      )}

      {matches.length > 0 && (
        <section>
          <h2 className="mb-4 text-xl font-semibold text-foreground">
            내 경험 적합도 분석
          </h2>
          <p className="mb-4 text-sm text-muted-foreground">
            키워드와 역량 무기 기반 자동 매칭 (상위 {Math.min(5, matches.length)}개)
          </p>
          <div className="space-y-4">
            {matches.slice(0, 5).map((match) => {
              const exp = experiences.find((e) => e.id === match.experience_id);
              if (!exp) return null;
              return (
                <div key={match.experience_id} className="brand-surface-soft rounded-2xl p-6">
                  <div className="mb-3 flex items-start justify-between">
                    <h3 className="flex-1 text-lg font-semibold">{exp.title}</h3>
                    <FitScoreBadge score={match.overall_fit} />
                  </div>
                  <div className="mb-3 grid grid-cols-3 gap-4 text-sm">
                    <div><span className="text-muted-foreground">역량</span><div className="mt-1 font-medium">{match.weapon_score}점</div></div>
                    <div><span className="text-muted-foreground">키워드</span><div className="mt-1 font-medium">{match.keyword_score}점</div></div>
                    <div><span className="text-muted-foreground">유형</span><div className="mt-1 font-medium">{match.category_score}점</div></div>
                  </div>
                  <p className="text-sm text-muted-foreground">{match.reasoning}</p>
                </div>
              );
            })}
          </div>
        </section>
      )}

    </div>
  );
}

function getSourceLabel(source: string): string {
  switch (source) {
    case "talent_profiles":
      return "검증된 인재상 DB";
    case "cache":
      return "캐시 (최근 분석)";
    case "ai_generated":
      return "AI 실시간 분석";
    default:
      return source;
  }
}

function getStepMessage(step: AnalysisStep): string {
  switch (step) {
    case "crawling":
      return "채용공고를 분석하고 있습니다...";
    case "analyzing":
      return "AI가 기업을 분석하고 있습니다...";
    case "loading_experiences":
      return "내 경험을 불러오고 있습니다...";
    case "matching":
      return "경험 매칭 중...";
    case "completed":
      return "완료!";
  }
}

function StepIndicator({ active, completed }: { active: boolean; completed: boolean }) {
  if (completed) {
    return (
      <div className="flex h-5 w-5 items-center justify-center rounded-full bg-primary/20 text-primary">
        <svg className="h-3 w-3" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={3} d="M5 13l4 4L19 7" />
        </svg>
      </div>
    );
  }
  if (active) {
    return <div className="h-5 w-5 rounded-full border-2 border-primary bg-primary/20 animate-pulse" />;
  }
  return <div className="h-5 w-5 rounded-full border-2 border-muted-foreground/20" />;
}

function formatCacheAge(cachedAt: string): string {
  const cached = new Date(cachedAt);
  const now = new Date();
  const diffMs = now.getTime() - cached.getTime();
  const diffMins = Math.floor(diffMs / 60000);
  const diffHours = Math.floor(diffMs / 3600000);
  const diffDays = Math.floor(diffMs / 86400000);

  if (diffMins < 1) return "방금 전";
  if (diffMins < 60) return `${diffMins}분 전`;
  if (diffHours < 24) return `${diffHours}시간 전`;
  return `${diffDays}일 전`;
}

function FitScoreBadge({ score }: { score: number }) {
  const label = score >= 80 ? "최적" : score >= 60 ? "적합" : score >= 40 ? "보통" : "낮음";
  const colorClass = score >= 80 ? "bg-green-500/20 text-green-300 border-green-500/30" : score >= 60 ? "bg-blue-500/20 text-blue-300 border-blue-500/30" : score >= 40 ? "bg-amber-500/20 text-amber-300 border-amber-500/30" : "bg-gray-500/20 text-gray-300 border-gray-500/30";
  return (
    <span className={`inline-flex items-center rounded-full border px-3 py-1 text-sm font-medium ${colorClass}`}>
      {label} {score}점
    </span>
  );
}
