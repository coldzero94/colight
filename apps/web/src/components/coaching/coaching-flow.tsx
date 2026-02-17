"use client";

import { useState } from "react";
import { useMutation } from "@tanstack/react-query";
import { Building2, Loader2, PenLine } from "lucide-react";
import { useRouter } from "next/navigation";
import { QuestionInputForm } from "./question-input-form";
import { StandaloneQuestionForm } from "./standalone-question-form";
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
type CoachingMode = "application" | "standalone";

export function CoachingFlow() {
  const router = useRouter();
  const [step, setStep] = useState<FlowStep>("input");
  const [mode, setMode] = useState<CoachingMode>("application");
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
      applicationId?: string;
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
      company_name: formData.company_name,
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
      <div className="brand-surface-soft flex items-center justify-center rounded-2xl py-20">
        <Loader2 className="h-6 w-6 animate-spin text-muted-foreground/60" />
        <span className="ml-2 text-muted-foreground">불러오는 중...</span>
      </div>
    );
  }

  const appList = applications ?? [];
  const hasApplications = appList.length > 0;

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
      {/* Mode Toggle (only on input step) */}
      {step === "input" && (
        <div className="flex gap-2">
          {hasApplications && (
            <button
              onClick={() => setMode("application")}
              className={`flex items-center gap-2 rounded-xl px-4 py-2 text-sm font-medium transition-all ${
                mode === "application"
                  ? "bg-primary text-primary-foreground shadow-[0_4px_12px_rgba(72,132,255,0.2)]"
                  : "brand-outline-btn text-muted-foreground hover:text-foreground"
              }`}
            >
              <Building2 className="h-4 w-4" />
              기업 분석 기반
            </button>
          )}
          <button
            onClick={() => setMode("standalone")}
            className={`flex items-center gap-2 rounded-xl px-4 py-2 text-sm font-medium transition-all ${
              mode === "standalone" || !hasApplications
                ? "bg-primary text-primary-foreground shadow-[0_4px_12px_rgba(72,132,255,0.2)]"
                : "brand-outline-btn text-muted-foreground hover:text-foreground"
            }`}
          >
            <PenLine className="h-4 w-4" />
            자유 코칭
          </button>
        </div>
      )}

      {/* Step 1: Question Input */}
      {step === "input" &&
        (mode === "application" && hasApplications ? (
          <QuestionInputForm companies={companies} onSubmit={handleAnalyze} />
        ) : (
          <StandaloneQuestionForm onSubmit={handleAnalyze} />
        ))}

      {/* Step 2: Analysis Result */}
      {step === "analysis" && analysisResult && (
        <div className="space-y-6">
          <AnalysisResult analysis={analysisResult} />
          <div className="flex gap-3">
            <button
              onClick={() => setStep("input")}
              className="brand-outline-btn rounded-xl px-4 py-2 text-sm text-muted-foreground hover:text-foreground"
            >
              다시 분석하기
            </button>
            <button
              onClick={handleGoToSelect}
              disabled={experienceRecommend.isPending}
              className="flex items-center gap-2 rounded-xl bg-primary px-4 py-2 text-sm font-medium text-primary-foreground shadow-[0_10px_24px_rgba(72,132,255,0.25)] transition-all hover:-translate-y-0.5 hover:bg-primary/90 disabled:opacity-50"
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
          advice={draftStreaming.advice}
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
