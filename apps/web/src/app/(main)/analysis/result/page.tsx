"use client";

import { useRouter, useSearchParams } from "next/navigation";
import { useEffect, useState } from "react";
import { toast } from "sonner";
import { LoadingSpinner } from "@/components/common/loading-spinner";
import { MatchingResults } from "@/components/analysis/matching-results";
import { useAnalyzeCompany } from "@/hooks/use-company-analysis";
import { useAutoMatching } from "@/hooks/use-auto-matching";
import { crawlJobPosting } from "@/lib/api/crawl";
import { fetchExperiences, type Experience } from "@/lib/api/experiences";
import type { CompanyAnalysis } from "@/lib/api/analysis";
import { isUsageLimitError, getUsageLimitInfo } from "@/lib/api/errors";

export default function AnalysisResultPage() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const url = searchParams.get("url");
  const [companyName, setCompanyName] = useState<string | null>(null);
  const [analysis, setAnalysis] = useState<CompanyAnalysis | null>(null);
  const [experiences, setExperiences] = useState<Experience[]>([]);
  const [crawlError, setCrawlError] = useState<string | null>(null);

  const analyzeMutation = useAnalyzeCompany();

  const { isMatching, matchingResult, triggerMatching } = useAutoMatching({
    companyName,
    hasMatching: false,
    experienceCount: experiences.length,
    enabled: !!analysis,
  });

  // Step 1: Crawl job posting to extract company name
  useEffect(() => {
    if (!url || companyName) return;

    (async () => {
      try {
        const jobPosting = await crawlJobPosting(url);
        setCompanyName(jobPosting.company_name);
      } catch {
        setCrawlError("채용공고에서 회사명을 추출하지 못했습니다.");
        toast.error("크롤링에 실패했습니다. URL을 확인해주세요.");
      }
    })();
  }, [url, companyName]);

  // Step 2: Analyze company once we have the name
  useEffect(() => {
    if (!companyName || analysis || analyzeMutation.isPending) return;

    analyzeMutation.mutate(companyName, {
      onSuccess: (data) => {
        setAnalysis(data);
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

  // Step 3: Load experiences for matching
  useEffect(() => {
    if (!analysis) return;

    (async () => {
      try {
        const data = await fetchExperiences();
        setExperiences(data);
      } catch {
        // Non-critical: matching will show empty
      }
    })();
  }, [analysis]);

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
          <p className="mt-4 text-muted-foreground">
            {!companyName
              ? "채용공고를 분석하고 있습니다..."
              : "AI가 기업을 분석하고 있습니다..."}
          </p>
          <p className="mt-2 text-sm text-muted-foreground/60">약 10-15초 소요됩니다</p>
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
            <p className="mt-1 text-sm text-muted-foreground">
              분석 출처: {getSourceLabel(analysis.source)}
            </p>
          </div>
          <button
            onClick={() => window.history.back()}
            className="brand-outline-btn rounded-xl px-3 py-1.5 text-sm text-muted-foreground hover:text-foreground"
          >
            ← 돌아가기
          </button>
        </div>
      </div>

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

      <section>
        <div className="mb-4 flex items-center justify-between">
          <h2 className="text-xl font-semibold text-foreground">
            내 경험 적합도 매칭
          </h2>
          {!isMatching && matchingResult && (
            <button
              className="text-sm text-primary hover:text-primary/80"
              onClick={triggerMatching}
            >
              다시 매칭 →
            </button>
          )}
        </div>

        {isMatching ? (
          <div className="flex flex-col items-center py-8">
            <LoadingSpinner />
            <p className="mt-4 text-sm text-muted-foreground">
              경험 매칭 중입니다...
            </p>
          </div>
        ) : matchingResult ? (
          <MatchingResults
            matches={matchingResult.matches}
            experiences={experiences}
          />
        ) : experiences.length === 0 ? (
          <p className="text-sm text-muted-foreground py-4">
            등록된 경험이 없습니다. 경험을 먼저 추가해주세요.
          </p>
        ) : null}
      </section>
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
