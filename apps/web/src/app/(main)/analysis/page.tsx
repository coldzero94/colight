"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { toast } from "sonner";
import { JobUrlInput } from "@/components/analysis/job-url-input";
import { EmptyState } from "@/components/common/empty-state";
import { LoadingSpinner } from "@/components/common/loading-spinner";

export default function AnalysisPage() {
  const router = useRouter();
  const [isAnalyzing, setIsAnalyzing] = useState(false);

  const handleAnalyze = async (url: string) => {
    setIsAnalyzing(true);
    try {
      // Extract company name from job posting first
      toast.success("분석을 시작합니다.");

      // Navigate to result page with URL
      setTimeout(() => {
        router.push("/analysis/result?url=" + encodeURIComponent(url));
        setIsAnalyzing(false);
      }, 500);
    } catch (error) {
      toast.error("분석에 실패했습니다.");
      setIsAnalyzing(false);
    }
  };

  return (
    <div className="mx-auto max-w-4xl px-4 py-8">
      <div className="mb-8">
        <h1 className="text-3xl font-bold text-gray-900">기업 분석</h1>
        <p className="mt-2 text-gray-600">
          채용공고 URL을 입력하면 AI가 기업을 자동 분석합니다.
        </p>
      </div>

      {/* URL Input */}
      <div className="rounded-lg border border-gray-200 bg-white p-6 shadow-sm">
        <h2 className="mb-4 text-lg font-semibold text-gray-900">
          채용공고 URL 입력
        </h2>
        <JobUrlInput onSubmit={handleAnalyze} isLoading={isAnalyzing} />
      </div>

      {/* Recent Analyses */}
      <div className="mt-8">
        <h2 className="mb-4 text-lg font-semibold text-gray-900">
          최근 분석 기록
        </h2>
        <RecentAnalysisList />
      </div>
    </div>
  );
}

function RecentAnalysisList() {
  // TODO: Fetch from API in Phase 3.3.4
  const analyses: any[] = [];
  const isLoading = false;

  if (isLoading) {
    return (
      <div className="flex justify-center py-8">
        <LoadingSpinner />
      </div>
    );
  }

  if (analyses.length === 0) {
    return (
      <EmptyState
        icon="📊"
        title="분석 기록이 없습니다"
        description="채용공고 URL을 입력하여 첫 기업 분석을 시작하세요."
      />
    );
  }

  return (
    <div className="space-y-3">
      {analyses.map((analysis: any) => (
        <div
          key={analysis.id}
          className="rounded-lg border border-gray-200 bg-white p-4 hover:border-gray-300 transition-colors cursor-pointer"
        >
          <h3 className="font-medium text-gray-900">{analysis.company_name}</h3>
          <p className="text-sm text-gray-500">{analysis.created_at}</p>
        </div>
      ))}
    </div>
  );
}
