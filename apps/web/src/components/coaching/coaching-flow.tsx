"use client";

import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { Loader2 } from "lucide-react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { QuestionInputForm } from "./question-input-form";
import { AnalysisResult } from "./analysis-result";
import { ExperienceSelector } from "./experience-selector";
import { DraftStreaming } from "./draft-streaming";
import { useQuestionAnalysis } from "@/hooks/use-question-analysis";
import { useDraftCoaching } from "@/hooks/use-draft-coaching";
import {
  getApplications,
  type QuestionAnalysisResult,
  type QuestionAnalysisRequest,
} from "@/lib/api/coaching";

type FlowStep = "input" | "analysis" | "select" | "draft";

export function CoachingFlow() {
  const router = useRouter();
  const [step, setStep] = useState<FlowStep>("input");
  const [analysisResult, setAnalysisResult] =
    useState<QuestionAnalysisResult | null>(null);
  const [formData, setFormData] = useState<QuestionAnalysisRequest | null>(
    null
  );

  // Fetch user's applications
  const { data: appsData, isLoading: appsLoading } = useQuery({
    queryKey: ["applications"],
    queryFn: getApplications,
  });

  const questionAnalysis = useQuestionAnalysis();
  const draftCoaching = useDraftCoaching();

  // Handle question analysis submit
  const handleAnalyze = async (data: QuestionAnalysisRequest) => {
    setFormData(data);
    const result = await questionAnalysis.mutateAsync(data);
    setAnalysisResult(result);
    setStep("analysis");
  };

  // Handle experience selection confirm
  const handleExperienceConfirm = async (selectedIds: string[]) => {
    if (!formData) return;
    draftCoaching.mutate({
      application_id: formData.application_id,
      experience_ids: selectedIds,
      question_text: formData.question_text,
      char_limit: formData.char_limit,
      analysis_result: analysisResult ?? undefined,
    });
    setStep("draft");
  };

  // Handle draft complete - navigate to editor
  const handleDraftComplete = () => {
    // TODO: Navigate to editor with actual cover letter ID from draft response
    router.push("/coaching");
  };

  // Loading state
  if (appsLoading) {
    return (
      <div className="flex items-center justify-center py-20">
        <Loader2 className="h-6 w-6 animate-spin text-gray-400" />
        <span className="ml-2 text-gray-500">불러오는 중...</span>
      </div>
    );
  }

  const applications = appsData?.applications ?? [];

  // Empty state
  if (applications.length === 0) {
    return (
      <div className="rounded-lg border border-gray-200 bg-gray-50 p-8 text-center">
        <p className="mb-4 text-gray-600">먼저 기업 분석을 진행해주세요</p>
        <Link
          href="/analysis"
          className="inline-block rounded-lg bg-gray-900 px-4 py-2 text-sm text-white hover:bg-gray-800"
        >
          기업 분석 시작하기
        </Link>
      </div>
    );
  }

  // Map to QuestionInputForm's company format
  const companies = applications.map((app) => ({
    id: app.id,
    company_name: app.company_name,
    position: app.position,
    created_at: app.created_at,
  }));

  return (
    <div className="space-y-8">
      {/* Step 1: Question Input */}
      {step === "input" && (
        <QuestionInputForm companies={companies} onSubmit={handleAnalyze} />
      )}

      {/* Step 2: Analysis Result */}
      {step === "analysis" && analysisResult && (
        <div className="space-y-6">
          <AnalysisResult analysis={analysisResult} />
          <div className="flex gap-3">
            <button
              onClick={() => setStep("input")}
              className="rounded-lg border border-gray-200 px-4 py-2 text-sm text-gray-600 hover:bg-gray-50"
            >
              다시 분석하기
            </button>
            <button
              onClick={() => setStep("select")}
              className="rounded-lg bg-gray-900 px-4 py-2 text-sm font-medium text-white hover:bg-gray-800"
            >
              경험 선택하러 가기
            </button>
          </div>
        </div>
      )}

      {/* Step 3: Experience Selection */}
      {step === "select" && analysisResult && (
        <ExperienceSelector
          experiences={[]}
          maxSelect={3}
          requiredWeapons={analysisResult.required_weapons}
          onConfirm={handleExperienceConfirm}
        />
      )}

      {/* Step 4: Draft Generation */}
      {step === "draft" && (
        <DraftStreaming
          content={draftCoaching.data?.draft ?? ""}
          isStreaming={draftCoaching.isPending}
          charLimit={formData?.char_limit ?? 800}
          onComplete={handleDraftComplete}
        />
      )}

      {/* Error display */}
      {questionAnalysis.isError && (
        <div className="rounded-lg border border-red-200 bg-red-50 p-4 text-sm text-red-600">
          분석 중 오류가 발생했습니다:{" "}
          {questionAnalysis.error?.message ?? "알 수 없는 오류"}
        </div>
      )}

      {draftCoaching.isError && (
        <div className="rounded-lg border border-red-200 bg-red-50 p-4 text-sm text-red-600">
          초안 생성 중 오류가 발생했습니다:{" "}
          {draftCoaching.error?.message ?? "알 수 없는 오류"}
        </div>
      )}
    </div>
  );
}
