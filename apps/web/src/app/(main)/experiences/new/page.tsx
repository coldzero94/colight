"use client";

import { useRouter } from "next/navigation";
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
    <div className="mx-auto max-w-2xl">
      <div className="mb-6">
        <button
          onClick={() => router.push("/experiences")}
          className="text-sm text-gray-500 hover:text-gray-900 transition-colors"
        >
          ← 경험 목록
        </button>
        <h1 className="mt-2 text-2xl font-bold text-gray-900">경험 등록</h1>
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
