"use client";

import { useState } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { BarChart3, Search, Sparkles, Target, Zap } from "lucide-react";
import { toast } from "sonner";
import { apiClient } from "@/lib/api-client";
import {
  experienceSchema,
  type ExperienceFormValues,
} from "@/lib/validations/experience";
import { StarField } from "./star-field";
import { CategorySelect } from "./category-select";
import { StarDiffDialog } from "./star-diff-dialog";

interface ExperienceFormProps {
  mode: "create" | "edit";
  defaultValues?: Partial<ExperienceFormValues>;
  onSubmit: (data: ExperienceFormValues) => void;
  onCancel: () => void;
  isSubmitting?: boolean;
}

interface StarFields {
  star_situation: string;
  star_task: string;
  star_action: string;
  star_result: string;
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
    setValue,
    getValues,
    formState: { errors },
  } = useForm<ExperienceFormValues>({
    resolver: zodResolver(experienceSchema),
    defaultValues: mergedDefaults,
  });

  const [isGenerating, setIsGenerating] = useState(false);
  const [diffOpen, setDiffOpen] = useState(false);
  const [diffBefore, setDiffBefore] = useState<StarFields | null>(null);
  const [diffAfter, setDiffAfter] = useState<StarFields | null>(null);

  const handleGenerateSTAR = async () => {
    const content = getValues("content");
    const title = getValues("title");

    if (!content || content.length < 30) {
      toast.error("자유 입력란에 최소 30자 이상 작성해주세요.");
      return;
    }

    setIsGenerating(true);
    try {
      const { data } = await apiClient.post("/v1/experiences/generate-star", {
        content,
        title: title || "",
      });

      const before: StarFields = {
        star_situation: getValues("star_situation") || "",
        star_task: getValues("star_task") || "",
        star_action: getValues("star_action") || "",
        star_result: getValues("star_result") || "",
      };

      const after: StarFields = {
        star_situation: data.star_situation || "",
        star_task: data.star_task || "",
        star_action: data.star_action || "",
        star_result: data.star_result || "",
      };

      const hasExisting = Object.values(before).some((v) => v.trim() !== "");

      if (hasExisting) {
        setDiffBefore(before);
        setDiffAfter(after);
        setDiffOpen(true);
      } else {
        applyStarFields(after);
        toast.success("AI가 STAR 구조를 생성했습니다.");
      }
    } catch {
      toast.error("AI STAR 생성에 실패했습니다. 다시 시도해주세요.");
    } finally {
      setIsGenerating(false);
    }
  };

  const applyStarFields = (fields: StarFields) => {
    setValue("star_situation", fields.star_situation, { shouldDirty: true });
    setValue("star_task", fields.star_task, { shouldDirty: true });
    setValue("star_action", fields.star_action, { shouldDirty: true });
    setValue("star_result", fields.star_result, { shouldDirty: true });
  };

  const handleDiffAccept = () => {
    if (diffAfter) {
      applyStarFields(diffAfter);
      toast.success("AI가 생성한 STAR 구조를 적용했습니다.");
    }
    setDiffOpen(false);
  };

  return (
    <>
      <form onSubmit={handleSubmit(onSubmit)} className="space-y-8">
        <section className="brand-surface rounded-2xl p-5 space-y-4">
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
              className="brand-input w-full rounded-xl px-3 py-2 text-sm"
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
                className="brand-input w-full rounded-xl px-3 py-2 text-sm"
              />
            </div>
            <div className="space-y-1">
              <label htmlFor="period_end" className="text-sm font-medium text-foreground/80">종료일</label>
              <input
                id="period_end"
                {...register("period_end")}
                type="date"
                className="brand-input w-full rounded-xl px-3 py-2 text-sm"
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
              className="brand-input w-full rounded-xl px-3 py-2 text-sm"
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

        <section className="brand-surface rounded-2xl p-5 space-y-4">
          <h2 className="text-lg font-semibold text-foreground">자유 입력</h2>
          <p className="text-xs text-muted-foreground/60">
            경험을 자유롭게 작성한 후 버튼을 눌러 AI가 STAR 구조로 정리해드려요
          </p>
          <textarea
            id="content"
            {...register("content")}
            maxLength={5000}
            placeholder="경험을 자유롭게 작성해주세요..."
            rows={6}
            className="brand-input w-full resize-none rounded-xl px-3 py-2 text-sm"
          />
          {errors.content && (
            <p role="alert" className="text-xs text-red-500">{errors.content.message}</p>
          )}

          <button
            type="button"
            onClick={handleGenerateSTAR}
            disabled={isGenerating}
            className="flex items-center gap-2 rounded-xl bg-violet-600 px-4 py-2 text-sm font-medium text-white shadow-md shadow-violet-600/20 transition-all hover:-translate-y-0.5 hover:bg-violet-500 disabled:opacity-50 disabled:hover:translate-y-0"
          >
            <Sparkles className="h-4 w-4" />
            {isGenerating ? "AI 생성 중..." : "AI로 STAR 생성"}
          </button>
        </section>

        <section className="brand-surface rounded-2xl p-5 space-y-4">
          <h2 className="text-lg font-semibold text-foreground">STAR 구조</h2>

          <StarField
            label="Situation"
            icon={<Search className="h-3.5 w-3.5 text-blue-300" />}
            placeholder="어떤 상황이었나요? (배경, 맥락)"
            guide="언제, 어디서, 어떤 조직에서, 규모는?"
            maxLength={1000}
            defaultValue={mergedDefaults.star_situation}
            registration={register("star_situation")}
            error={errors.star_situation?.message}
          />

          <StarField
            label="Task"
            icon={<Target className="h-3.5 w-3.5 text-amber-300" />}
            placeholder="무엇을 해야 했나요? (과제, 목표)"
            guide="구체적 목표, 기대 성과, 주어진 제약 조건은?"
            maxLength={1000}
            defaultValue={mergedDefaults.star_task}
            registration={register("star_task")}
            error={errors.star_task?.message}
          />

          <StarField
            label="Action"
            icon={<Zap className="h-3.5 w-3.5 text-emerald-300" />}
            placeholder="어떻게 행동했나요? (구체적 행동)"
            guide="어떤 전략을 세웠고, 구체적으로 무엇을 했나요? 수치를 포함하면 좋아요."
            maxLength={2000}
            defaultValue={mergedDefaults.star_action}
            registration={register("star_action")}
            error={errors.star_action?.message}
          />

          <StarField
            label="Result"
            icon={<BarChart3 className="h-3.5 w-3.5 text-violet-300" />}
            placeholder="결과는 어땠나요? (성과, 교훈)"
            guide="정량적 성과 (%, 건수, 금액), 질적 변화, 배운 점은?"
            maxLength={1000}
            defaultValue={mergedDefaults.star_result}
            registration={register("star_result")}
            error={errors.star_result?.message}
          />
        </section>

        <div className="flex justify-end gap-3 border-t border-white/10 pt-6">
          <button
            type="button"
            onClick={onCancel}
            className="brand-outline-btn rounded-xl px-6 py-2 text-sm font-medium text-foreground/80"
            disabled={isSubmitting}
          >
            취소
          </button>
          <button
            type="submit"
            className="rounded-xl bg-primary px-6 py-2 text-sm font-medium text-primary-foreground shadow-[0_12px_26px_rgba(72,132,255,0.25)] transition-all hover:-translate-y-0.5 hover:bg-primary/90 disabled:opacity-50"
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

      <StarDiffDialog
        open={diffOpen}
        onClose={() => setDiffOpen(false)}
        onAccept={handleDiffAccept}
        before={diffBefore}
        after={diffAfter}
      />
    </>
  );
}
