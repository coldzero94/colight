"use client";

import { useRouter } from "next/navigation";
import Link from "next/link";
import { toast } from "sonner";
import { ExperienceForm } from "@/components/experiences/experience-form";
import { useCreateExperience } from "@/hooks/use-experiences";
import { useWeaponTagging } from "@/hooks/use-weapon-tagging";
import type { ExperienceFormValues } from "@/lib/validations/experience";

export default function NewExperiencePage() {
  const router = useRouter();
  const createMutation = useCreateExperience();
  const { triggerTagging } = useWeaponTagging();

  const handleSubmit = async (data: ExperienceFormValues) => {
    try {
      const result = await createMutation.mutateAsync({
        title: data.title,
        category: data.category || undefined,
        period_start: data.period_start || undefined,
        period_end: data.period_end || undefined,
        role: data.role || undefined,
        content: data.content || undefined,
        star_situation: data.star_situation || undefined,
        star_task: data.star_task || undefined,
        star_action: data.star_action || undefined,
        star_result: data.star_result || undefined,
      });
      toast.success("경험이 등록되었습니다.");

      // Trigger AI weapon tagging
      triggerTagging(result.id);

      router.push(`/experiences/${result.id}`);
    } catch {
      toast.error("경험 등록에 실패했습니다.");
    }
  };

  return (
    <div className="mx-auto max-w-2xl space-y-5 animate-fade-in">
      <div className="mb-6">
        <Link
          href="/experiences"
          className="text-sm text-muted-foreground transition-colors hover:text-foreground"
          aria-label="경험 목록으로 돌아가기"
        >
          ← 경험 목록
        </Link>
        <div className="brand-surface mt-3 rounded-2xl px-6 py-5">
          <span className="brand-kicker inline-flex rounded-full px-3 py-1 text-[11px] font-semibold tracking-[0.18em] uppercase">
            New Experience
          </span>
          <h1 className="mt-3 text-2xl font-bold font-display text-foreground">
            경험 등록
          </h1>
          <p className="mt-1 text-sm text-muted-foreground">
            STAR 구조와 자유 서술을 함께 기록하면 AI 활용도가 높아집니다.
          </p>
        </div>
      </div>
      <ExperienceForm
        mode="create"
        onSubmit={handleSubmit}
        onCancel={() => router.push("/experiences")}
        isSubmitting={createMutation.isPending}
      />
    </div>
  );
}
