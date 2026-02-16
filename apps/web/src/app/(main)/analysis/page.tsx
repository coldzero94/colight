"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { BarChart3 } from "lucide-react";
import { toast } from "sonner";
import { JobUrlInput } from "@/components/analysis/job-url-input";
import { EmptyState } from "@/components/common/empty-state";

export default function AnalysisPage() {
  const router = useRouter();
  const [isAnalyzing, setIsAnalyzing] = useState(false);

  const handleAnalyze = async (url: string) => {
    setIsAnalyzing(true);
    try {
      toast.success("분석을 시작합니다.");
      router.push("/analysis/result?url=" + encodeURIComponent(url));
    } catch {
      toast.error("분석에 실패했습니다.");
    } finally {
      setIsAnalyzing(false);
    }
  };

  return (
    <div className="mx-auto max-w-4xl space-y-6 px-4 py-8 animate-fade-in">
      <div className="brand-surface relative overflow-hidden rounded-2xl px-6 py-5">
        <div className="pointer-events-none absolute -right-10 -top-8 h-48 w-48 rounded-full bg-cyan-400/16 blur-3xl" />
        <div className="relative">
          <span className="brand-kicker inline-flex rounded-full px-3 py-1 text-[11px] font-semibold tracking-[0.18em] uppercase">
            Insight Engine
          </span>
          <h1 className="mt-3 text-3xl font-bold font-display text-foreground">
            기업 분석
          </h1>
          <p className="mt-2 text-muted-foreground">
            채용공고 URL을 입력하면 AI가 기업의 핵심가치와 전략 키워드를 정리합니다.
          </p>
        </div>
      </div>

      <div className="brand-surface rounded-2xl p-6 shadow-[0_20px_36px_rgba(0,0,0,0.2)]">
        <h2 className="mb-4 text-lg font-semibold text-foreground">
          채용공고 URL 입력
        </h2>
        <JobUrlInput onSubmit={handleAnalyze} isLoading={isAnalyzing} />
      </div>

      <div>
        <h2 className="mb-4 text-lg font-semibold text-foreground">
          최근 분석 기록
        </h2>
        <EmptyState
          icon={<BarChart3 className="h-7 w-7" />}
          title="분석 기록이 없습니다"
          description="채용공고 URL을 입력하여 첫 기업 분석을 시작하세요."
        />
      </div>
    </div>
  );
}
