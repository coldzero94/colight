"use client";

import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import {
  experienceSchema,
  type ExperienceFormValues,
} from "@/lib/validations/experience";
import { StarField } from "./star-field";
import { CategorySelect } from "./category-select";

interface ExperienceFormProps {
  mode: "create" | "edit";
  defaultValues?: Partial<ExperienceFormValues>;
  onSubmit: (data: ExperienceFormValues) => void;
  onCancel: () => void;
  isSubmitting?: boolean;
}

export function ExperienceForm({
  mode,
  defaultValues,
  onSubmit,
  onCancel,
  isSubmitting = false,
}: ExperienceFormProps) {
  const mergedDefaults = {
    title: "",
    category: "",
    period_start: "",
    period_end: "",
    role: "",
    star_situation: "",
    star_task: "",
    star_action: "",
    star_result: "",
    content: "",
    ...defaultValues,
  };

  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<ExperienceFormValues>({
    resolver: zodResolver(experienceSchema),
    defaultValues: mergedDefaults,
  });

  return (
    <form onSubmit={handleSubmit(onSubmit)} className="space-y-8">
      {/* Basic info */}
      <section className="space-y-4">
        <h2 className="text-lg font-semibold text-foreground">기본 정보</h2>

        <div className="space-y-1">
          <label htmlFor="title" className="text-sm font-medium text-foreground/80">
            제목 <span className="text-red-500">*</span>
          </label>
          <input
            id="title"
            {...register("title")}
            maxLength={100}
            placeholder="동아리 축제 부스 운영"
            className="w-full rounded-lg border border-border bg-input px-3 py-2 text-sm text-foreground placeholder:text-muted-foreground/50 focus:outline-none focus:ring-2 focus:ring-ring"
          />
          {errors.title && (
            <p role="alert" className="text-xs text-red-500">{errors.title.message}</p>
          )}
        </div>

        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div className="space-y-1">
            <label htmlFor="period_start" className="text-sm font-medium text-foreground/80">시작일</label>
            <input
              id="period_start"
              {...register("period_start")}
              type="date"
              className="w-full rounded-lg border border-border bg-input px-3 py-2 text-sm text-foreground focus:outline-none focus:ring-2 focus:ring-ring"
            />
          </div>
          <div className="space-y-1">
            <label htmlFor="period_end" className="text-sm font-medium text-foreground/80">종료일</label>
            <input
              id="period_end"
              {...register("period_end")}
              type="date"
              className="w-full rounded-lg border border-border bg-input px-3 py-2 text-sm text-foreground focus:outline-none focus:ring-2 focus:ring-ring"
            />
            {errors.period_end && (
              <p role="alert" className="text-xs text-red-500">
                {errors.period_end.message}
              </p>
            )}
          </div>
        </div>

        <div className="space-y-1">
          <label htmlFor="role" className="text-sm font-medium text-foreground/80">역할</label>
          <input
            id="role"
            {...register("role")}
            maxLength={50}
            placeholder="부스 운영 총괄"
            className="w-full rounded-lg border border-border bg-input px-3 py-2 text-sm text-foreground placeholder:text-muted-foreground/50 focus:outline-none focus:ring-2 focus:ring-ring"
          />
          {errors.role && (
            <p role="alert" className="text-xs text-red-500">{errors.role.message}</p>
          )}
        </div>

        <CategorySelect
          registration={register("category")}
          error={errors.category?.message}
        />
      </section>

      {/* STAR section */}
      <section className="space-y-4">
        <h2 className="text-lg font-semibold text-foreground">STAR 구조</h2>

        <StarField
          label="Situation"
          icon="🔍"
          placeholder="어떤 상황이었나요? (배경, 맥락)"
          guide="언제, 어디서, 어떤 조직에서, 규모는?"
          maxLength={1000}
          defaultValue={mergedDefaults.star_situation}
          registration={register("star_situation")}
          error={errors.star_situation?.message}
        />

        <StarField
          label="Task"
          icon="🎯"
          placeholder="무엇을 해야 했나요? (과제, 목표)"
          guide="구체적 목표, 기대 성과, 주어진 제약 조건은?"
          maxLength={1000}
          defaultValue={mergedDefaults.star_task}
          registration={register("star_task")}
          error={errors.star_task?.message}
        />

        <StarField
          label="Action"
          icon="⚡"
          placeholder="어떻게 행동했나요? (구체적 행동)"
          guide="어떤 전략을 세웠고, 구체적으로 무엇을 했나요? 수치를 포함하면 좋아요."
          maxLength={2000}
          defaultValue={mergedDefaults.star_action}
          registration={register("star_action")}
          error={errors.star_action?.message}
        />

        <StarField
          label="Result"
          icon="📊"
          placeholder="결과는 어땠나요? (성과, 교훈)"
          guide="정량적 성과 (%, 건수, 금액), 질적 변화, 배운 점은?"
          maxLength={1000}
          defaultValue={mergedDefaults.star_result}
          registration={register("star_result")}
          error={errors.star_result?.message}
        />
      </section>

      {/* Free text section */}
      <section className="space-y-4">
        <h2 className="text-lg font-semibold text-foreground">자유 입력</h2>
        <p className="text-xs text-muted-foreground/60">
          STAR 구조가 어렵다면 자유롭게 작성해도 AI가 정리해드려요
        </p>
        <textarea
          id="content"
          {...register("content")}
          maxLength={5000}
          placeholder="경험을 자유롭게 작성해주세요..."
          rows={6}
          className="w-full rounded-lg border border-border bg-input px-3 py-2 text-sm text-foreground placeholder:text-muted-foreground/50 focus:outline-none focus:ring-2 focus:ring-ring resize-none"
        />
        {errors.content && (
          <p role="alert" className="text-xs text-red-500">{errors.content.message}</p>
        )}
      </section>

      {/* Form actions */}
      <div className="flex gap-3 justify-end border-t border-border pt-6">
        <button
          type="button"
          onClick={onCancel}
          className="rounded-lg px-6 py-2 text-sm font-medium text-foreground/80 hover:bg-white/[0.06] transition-colors"
          disabled={isSubmitting}
        >
          취소
        </button>
        <button
          type="submit"
          className="rounded-lg bg-primary px-6 py-2 text-sm font-medium text-primary-foreground hover:bg-primary/90 transition-colors disabled:opacity-50"
          disabled={isSubmitting}
        >
          {isSubmitting
            ? "저장 중..."
            : mode === "create"
              ? "등록"
              : "수정"}
        </button>
      </div>
    </form>
  );
}
