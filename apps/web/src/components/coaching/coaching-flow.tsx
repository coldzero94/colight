"use client";

import { useState } from "react";
import { useMutation } from "@tanstack/react-query";
import { Loader2 } from "lucide-react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { QuestionInputForm } from "./question-input-form";
import { AnalysisResult } from "./analysis-result";
import { ExperienceSelector } from "./experience-selector";
import { DraftStreaming } from "./draft-streaming";
import { useQuestionAnalysis } from "@/hooks/use-question-analysis";
import { useDraftStreaming } from "@/hooks/use-draft-streaming";
import { useApplications } from "@/hooks/use-applications";
import {
  recommendExperiences,
  type QuestionAnalysisResult,
  type QuestionAnalysisRequest,
  type ExperienceRecommendation,
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
  const [recommendations, setRecommendations] = useState<
    ExperienceRecommendation[]
  >([]);

  // Fetch user's applications (shared query key with dashboard)
  const { data: applications, isLoading: appsLoading } = useApplications();

  const questionAnalysis = useQuestionAnalysis();
  const draftStreaming = useDraftStreaming();

  const experienceRecommend = useMutation({
    mutationFn: ({
      requiredWeapons,
      keyKeywords,
      applicationId,
    }: {
      requiredWeapons: QuestionAnalysisResult["required_weapons"];
      keyKeywords: string[];
      applicationId: string;
    }) => recommendExperiences(requiredWeapons, keyKeywords, applicationId),
  });

  // Handle question analysis submit
  const handleAnalyze = async (data: QuestionAnalysisRequest) => {
    setFormData(data);
    const result = await questionAnalysis.mutateAsync(data);
    setAnalysisResult(result);
    setStep("analysis");
  };

  // Handle move to experience selection - fetch recommendations
  const handleGoToSelect = async () => {
    if (!analysisResult || !formData) return;
    const result = await experienceRecommend.mutateAsync({
      requiredWeapons: analysisResult.required_weapons,
      keyKeywords: analysisResult.key_keywords,
      applicationId: formData.application_id,
    });
    setRecommendations(result.recommendations);
    setStep("select");
  };

  // Handle experience selection confirm
  const handleExperienceConfirm = async (selectedIds: string[]) => {
    if (!formData) return;
    setStep("draft");
    draftStreaming.streamDraft({
      application_id: formData.application_id,
      experience_ids: selectedIds,
      question_text: formData.question_text,
      char_limit: formData.char_limit,
      analysis_result: analysisResult ?? undefined,
    });
  };

  // Handle draft complete - navigate to editor
  const handleDraftComplete = () => {
    if (draftStreaming.coverLetterId) {
      router.push(`/coaching/${draftStreaming.coverLetterId}/edit`);
    } else {
      router.push("/coaching");
    }
  };

  // Loading state
  if (appsLoading) {
    return (
      <div className="flex items-center justify-center py-20">
        <Loader2 className="h-6 w-6 animate-spin text-muted-foreground/60" />
        <span className="ml-2 text-muted-foreground">불러오는 중...</span>
      </div>
    );
  }

  const appList = applications ?? [];

  // Empty state
  if (appList.length === 0) {
    return (
      <div className="rounded-lg border border-border bg-muted p-8 text-center">
        <p className="mb-4 text-muted-foreground">먼저 기업 분석을 진행해주세요</p>
        <Link
          href="/analysis"
          className="inline-block rounded-lg bg-primary px-4 py-2 text-sm text-primary-foreground hover:bg-primary/90"
        >
          기업 분석 시작하기
        </Link>
      </div>
    );
  }

  // Map to QuestionInputForm's company format
  const companies = appList.map((app) => ({
    id: app.id,
    company_name: app.company_name,
    position: app.position,
    created_at: app.created_at,
  }));

  // Map recommendations to ExperienceSelector format
  const selectorExperiences = recommendations.map((rec) => ({
    id: rec.id,
    title: rec.title,
    category: rec.category,
    period_start: rec.period_start,
    period_end: rec.period_end,
    star_situation: rec.star_situation,
    star_task: "",
    star_action: "",
    star_result: "",
    weapons: rec.weapons.map((w) => ({ code: w, name: w })),
    matchScore: rec.match_score,
    matchReasons: rec.match_reasons,
    isUsed: rec.is_used,
    keywordMatches: rec.keyword_matches,
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
              className="rounded-lg border border-border px-4 py-2 text-sm text-muted-foreground hover:bg-white/[0.04]"
            >
              다시 분석하기
            </button>
            <button
              onClick={handleGoToSelect}
              disabled={experienceRecommend.isPending}
              className="flex items-center gap-2 rounded-lg bg-primary px-4 py-2 text-sm font-medium text-primary-foreground hover:bg-primary/90 disabled:opacity-50"
            >
              {experienceRecommend.isPending ? (
                <>
                  <Loader2 className="h-4 w-4 animate-spin" />
                  경험 추천 중...
                </>
              ) : (
                "경험 선택하러 가기"
              )}
            </button>
          </div>
        </div>
      )}

      {/* Step 3: Experience Selection */}
      {step === "select" && analysisResult && (
        <ExperienceSelector
          experiences={selectorExperiences}
          maxSelect={3}
          requiredWeapons={analysisResult.required_weapons}
          onConfirm={handleExperienceConfirm}
        />
      )}

      {/* Step 4: Draft Generation */}
      {step === "draft" && (
        <DraftStreaming
          content={draftStreaming.content}
          isStreaming={draftStreaming.isStreaming}
          charLimit={formData?.char_limit ?? 800}
          onComplete={handleDraftComplete}
        />
      )}

      {/* Error display */}
      {questionAnalysis.isError && (
        <div className="rounded-lg border border-red-500/20 bg-red-500/10 p-4 text-sm text-red-400">
          분석 중 오류가 발생했습니다:{" "}
          {questionAnalysis.error?.message ?? "알 수 없는 오류"}
        </div>
      )}

      {experienceRecommend.isError && (
        <div className="rounded-lg border border-red-500/20 bg-red-500/10 p-4 text-sm text-red-400">
          경험 추천 중 오류가 발생했습니다:{" "}
          {experienceRecommend.error?.message ?? "알 수 없는 오류"}
        </div>
      )}

      {draftStreaming.error && (
        <div className="rounded-lg border border-red-500/20 bg-red-500/10 p-4 text-sm text-red-400">
          초안 생성 중 오류가 발생했습니다: {draftStreaming.error}
        </div>
      )}
    </div>
  );
}
