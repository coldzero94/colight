"use client";

import { useState } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import Link from "next/link";
import { Loader2 } from "lucide-react";
import { CompanySelect } from "./company-select";
import { CharLimitInput } from "./char-limit-input";
import {
  questionAnalysisSchema,
  type QuestionAnalysisInput,
} from "@/lib/validations/coaching";

interface Company {
  id: string;
  company_name: string;
  position: string;
  created_at: string;
}

interface QuestionInputFormProps {
  companies: Company[];
  onSubmit: (data: QuestionAnalysisInput) => Promise<void> | void;
}

export function QuestionInputForm({
  companies,
  onSubmit,
}: QuestionInputFormProps) {
  const [isSubmitting, setIsSubmitting] = useState(false);

  const {
    register,
    handleSubmit,
    setValue,
    watch,
    formState: { errors },
  } = useForm<QuestionAnalysisInput>({
    resolver: zodResolver(questionAnalysisSchema),
    defaultValues: {
      application_id: "",
      question_text: "",
      char_limit: 800,
    },
  });

  const selectedCompanyId = watch("application_id");
  const questionText = watch("question_text");

  const handleFormSubmit = async (data: QuestionAnalysisInput) => {
    setIsSubmitting(true);
    try {
      await onSubmit(data);
    } finally {
      setIsSubmitting(false);
    }
  };

  // 빈 상태: 기업 분석 이력이 없을 때
  if (companies.length === 0) {
    return (
      <div className="rounded-lg border border-border bg-muted p-8 text-center">
        <p className="mb-4 text-muted-foreground">
          먼저 기업 분석을 진행해주세요
        </p>
        <Link
          href="/analysis"
          className="inline-block rounded-lg bg-primary px-4 py-2 text-sm text-primary-foreground hover:bg-primary/90"
        >
          기업 분석 시작하기
        </Link>
      </div>
    );
  }

  return (
    <form onSubmit={handleSubmit(handleFormSubmit)} className="space-y-6">
      {/* 기업 선택 */}
      <CompanySelect
        companies={companies}
        value={selectedCompanyId}
        onChange={(value) => setValue("application_id", value)}
        error={errors.application_id?.message}
      />

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
          className="w-full rounded-lg border border-border bg-input px-3 py-2 text-sm text-foreground focus:outline-none focus:ring-2 focus:ring-ring"
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

      {/* 제출 버튼 */}
      <button
        type="submit"
        disabled={isSubmitting}
        className="flex w-full items-center justify-center gap-2 rounded-lg bg-primary px-4 py-3 text-sm font-medium text-primary-foreground hover:bg-primary/90 disabled:opacity-50"
      >
        {isSubmitting ? (
          <>
            <Loader2 className="h-4 w-4 animate-spin" data-testid="loading-spinner" />
            분석 중...
          </>
        ) : (
          "🔍 분석 시작"
        )}
      </button>
    </form>
  );
}
