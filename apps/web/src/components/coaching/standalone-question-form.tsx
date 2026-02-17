"use client";

import { useState } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { Loader2, Search } from "lucide-react";
import { CharLimitInput } from "./char-limit-input";
import { ExperiencePicker } from "./experience-picker";
import {
  standaloneQuestionSchema,
  type StandaloneQuestionInput,
} from "@/lib/validations/coaching";
import type { QuestionAnalysisRequest } from "@/lib/api/coaching";

interface StandaloneQuestionFormProps {
  onSubmit: (data: QuestionAnalysisRequest) => Promise<void> | void;
}

export function StandaloneQuestionForm({
  onSubmit,
}: StandaloneQuestionFormProps) {
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [selectedExpIds, setSelectedExpIds] = useState<string[]>([]);

  const {
    register,
    handleSubmit,
    watch,
    formState: { errors },
  } = useForm<StandaloneQuestionInput>({
    resolver: zodResolver(standaloneQuestionSchema),
    defaultValues: {
      company_name: "",
      question_text: "",
      char_limit: 800,
    },
  });

  const questionText = watch("question_text");

  const handleFormSubmit = async (data: StandaloneQuestionInput) => {
    setIsSubmitting(true);
    try {
      await onSubmit({
        company_name: data.company_name || undefined,
        question_text: data.question_text,
        char_limit: data.char_limit,
        experience_ids: selectedExpIds.length > 0 ? selectedExpIds : undefined,
      });
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <form onSubmit={handleSubmit(handleFormSubmit)} className="space-y-6">
      {/* 회사명 (선택) */}
      <div className="space-y-1">
        <label
          htmlFor="company_name"
          className="text-sm font-medium text-foreground/80"
        >
          회사명 <span className="text-muted-foreground">(선택)</span>
        </label>
        <input
          id="company_name"
          type="text"
          {...register("company_name")}
          placeholder="예: 삼성전자, 네이버 (입력하면 맞춤 코칭)"
          className="brand-input w-full rounded-xl px-3 py-2 text-sm"
        />
      </div>

      {/* 자소서 문항 입력 */}
      <div className="space-y-1">
        <label
          htmlFor="question_text"
          className="text-sm font-medium text-foreground/80"
        >
          자소서 문항
        </label>
        <textarea
          id="question_text"
          {...register("question_text")}
          placeholder="자소서 문항을 입력하세요 (예: 본인이 팀 프로젝트에서 어려움을 극복한 경험을 구체적으로 기술하세요.)"
          rows={4}
          className="brand-input w-full rounded-xl px-3 py-2 text-sm"
        />
        <div className="flex items-center justify-between">
          <div>
            {errors.question_text && (
              <p role="alert" className="text-xs text-red-500">
                {errors.question_text.message}
              </p>
            )}
          </div>
          <p className="text-xs text-muted-foreground">
            {questionText.length}/500자
          </p>
        </div>
      </div>

      {/* 글자수 제한 */}
      <CharLimitInput
        registration={register("char_limit", {
          valueAsNumber: true,
        })}
        defaultValue={800}
        error={errors.char_limit?.message}
      />

      {/* 경험 선택 */}
      <ExperiencePicker
        selectedIds={selectedExpIds}
        onSelectionChange={setSelectedExpIds}
        maxSelect={3}
      />

      {/* 제출 버튼 */}
      <button
        type="submit"
        disabled={isSubmitting}
        className="group relative flex w-full items-center justify-center gap-2 overflow-hidden rounded-xl bg-primary px-4 py-3 text-sm font-medium text-primary-foreground transition-all hover:-translate-y-0.5 hover:bg-primary/90 hover:shadow-[0_12px_26px_rgba(72,132,255,0.26)] disabled:opacity-50"
      >
        <span className="pointer-events-none absolute inset-0 bg-gradient-to-r from-white/0 via-white/10 to-white/0 opacity-0 transition-opacity group-hover:opacity-100" />
        {isSubmitting ? (
          <>
            <Loader2 className="h-4 w-4 animate-spin" data-testid="loading-spinner" />
            분석 중...
          </>
        ) : (
          <>
            <Search className="h-4 w-4" aria-hidden="true" />
            분석 시작
          </>
        )}
      </button>
    </form>
  );
}
