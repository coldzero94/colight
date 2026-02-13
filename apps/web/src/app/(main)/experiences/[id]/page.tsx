"use client";

import { useParams, useRouter } from "next/navigation";
import Link from "next/link";
import { toast } from "sonner";
import { LoadingSpinner } from "@/components/common/loading-spinner";
import { StarDisplay } from "@/components/experiences/star-display";
import { WeaponBadges } from "@/components/experiences/weapon-badges";
import { ExperienceActions } from "@/components/experiences/experience-actions";
import {
  DeleteDialog,
  useDeleteDialog,
} from "@/components/experiences/delete-dialog";
import { CategoryIcon } from "@/components/experiences/category-icon";
import { useExperience, useDeleteExperience } from "@/hooks/use-experiences";

export default function ExperienceDetailPage() {
  const params = useParams();
  const router = useRouter();
  const id = params.id as string;

  const { data: experience, isLoading, error } = useExperience(id);
  const deleteMutation = useDeleteExperience();
  const deleteDialog = useDeleteDialog();

  const handleDelete = async () => {
    try {
      await deleteMutation.mutateAsync(id);
      toast.success("경험이 삭제되었습니다.");
      router.push("/experiences");
    } catch {
      toast.error("삭제에 실패했습니다.");
    }
  };

  if (isLoading) {
    return (
      <div className="flex justify-center py-20">
        <LoadingSpinner />
      </div>
    );
  }

  if (error || !experience) {
    return (
      <div className="py-20 text-center text-gray-500">
        경험을 찾을 수 없습니다.
      </div>
    );
  }

  return (
    <div className="mx-auto max-w-2xl space-y-6">
      <Link
        href="/experiences"
        className="text-sm text-gray-500 hover:text-gray-900 transition-colors"
      >
        ← 경험 목록
      </Link>

      <div className="flex items-start justify-between">
        <div>
          <div className="flex items-center gap-2 mb-2">
            {experience.category && (
              <CategoryIcon category={experience.category} showLabel />
            )}
            {experience.period_start && (
              <span className="text-xs text-gray-400">
                {experience.period_start.slice(0, 7).replace("-", ".")}
                {experience.period_end &&
                  ` ~ ${experience.period_end.slice(0, 7).replace("-", ".")}`}
              </span>
            )}
          </div>
          <h1 className="text-2xl font-bold text-gray-900">
            {experience.title}
          </h1>
          {experience.role && (
            <p className="mt-1 text-sm text-gray-500">{experience.role}</p>
          )}
        </div>
        <ExperienceActions
          experienceId={experience.id}
          onDeleteClick={deleteDialog.open}
        />
      </div>

      {experience.weapons && experience.weapons.length > 0 ? (
        <WeaponBadges weapons={experience.weapons} />
      ) : (
        <p className="text-sm text-gray-400">AI 분석 대기 중...</p>
      )}

      <StarDisplay
        situation={experience.star_situation}
        task={experience.star_task}
        action={experience.star_action}
        result={experience.star_result}
      />

      {experience.content && (
        <div className="rounded-lg border border-gray-100 bg-white p-4">
          <h4 className="mb-2 text-sm font-semibold text-gray-900">
            자유 입력
          </h4>
          <p className="text-sm leading-relaxed text-gray-700 whitespace-pre-wrap">
            {experience.content}
          </p>
        </div>
      )}

      <div className="text-xs text-gray-400 space-y-1 border-t border-gray-100 pt-4">
        <p>
          생성일:{" "}
          {new Date(experience.created_at).toLocaleDateString("ko-KR")}
        </p>
        <p>
          최종 수정일:{" "}
          {new Date(experience.updated_at).toLocaleDateString("ko-KR")}
        </p>
      </div>

      {deleteDialog.isOpen && (
        <DeleteDialog
          title={experience.title}
          onConfirm={handleDelete}
          onCancel={deleteDialog.close}
          isDeleting={deleteMutation.isPending}
        />
      )}
    </div>
  );
}
