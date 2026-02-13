"use client";

import { useParams, useRouter } from "next/navigation";
import { toast } from "sonner";
import { LoadingSpinner } from "@/components/common/loading-spinner";
import { ExperienceForm } from "@/components/experiences/experience-form";
import { useExperience, useUpdateExperience } from "@/hooks/use-experiences";
import type { ExperienceFormValues } from "@/lib/validations/experience";

export default function EditExperiencePage() {
  const params = useParams();
  const router = useRouter();
  const id = params.id as string;

  const { data: experience, isLoading } = useExperience(id);
  const updateMutation = useUpdateExperience();

  const handleSubmit = async (data: ExperienceFormValues) => {
    try {
      await updateMutation.mutateAsync({
        id,
        data: {
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
        },
      });
      toast.success("경험이 수정되었습니다.");
      router.push(`/experiences/${id}`);
    } catch {
      toast.error("수정에 실패했습니다.");
    }
  };

  if (isLoading) {
    return (
      <div className="flex justify-center py-20">
        <LoadingSpinner />
      </div>
    );
  }

  if (!experience) {
    return (
      <div className="py-20 text-center text-gray-500">
        경험을 찾을 수 없습니다.
      </div>
    );
  }

  return (
    <div className="mx-auto max-w-2xl">
      <div className="mb-6">
        <button
          onClick={() => router.push(`/experiences/${id}`)}
          className="text-sm text-gray-500 hover:text-gray-900 transition-colors"
        >
          ← 경험 상세
        </button>
        <h1 className="mt-2 text-2xl font-bold text-gray-900">경험 수정</h1>
      </div>
      <ExperienceForm
        mode="edit"
        defaultValues={{
          title: experience.title,
          category: experience.category || "",
          period_start: experience.period_start || "",
          period_end: experience.period_end || "",
          role: experience.role || "",
          star_situation: experience.star_situation || "",
          star_task: experience.star_task || "",
          star_action: experience.star_action || "",
          star_result: experience.star_result || "",
          content: experience.content || "",
        }}
        onSubmit={handleSubmit}
        onCancel={() => router.push(`/experiences/${id}`)}
        isSubmitting={updateMutation.isPending}
      />
    </div>
  );
}
