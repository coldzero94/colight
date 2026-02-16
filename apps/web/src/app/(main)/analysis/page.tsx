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
    <div className="mx-auto max-w-4xl px-4 py-8">
      <div className="mb-8">
        <h1 className="text-3xl font-bold font-display text-foreground">기업 분석</h1>
        <p className="mt-2 text-muted-foreground">
          채용공고 URL을 입력하면 AI가 기업을 자동 분석합니다.
        </p>
      </div>

      {/* URL Input */}
      <div className="rounded-lg border border-border bg-card p-6 shadow-sm">
        <h2 className="mb-4 text-lg font-semibold text-foreground">
          채용공고 URL 입력
        </h2>
        <JobUrlInput onSubmit={handleAnalyze} isLoading={isAnalyzing} />
      </div>

      {/* Placeholder for future analysis history */}
      <div className="mt-8">
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
