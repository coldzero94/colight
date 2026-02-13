"use client";

import { useRouter } from "next/navigation";

interface ExperienceActionsProps {
  experienceId: string;
  onDeleteClick: () => void;
}

export function ExperienceActions({
  experienceId,
  onDeleteClick,
}: ExperienceActionsProps) {
  const router = useRouter();

  return (
    <div className="flex gap-2">
      <button
        type="button"
        onClick={() => router.push(`/experiences/${experienceId}/edit`)}
        className="rounded-lg border border-gray-200 px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50 transition-colors"
      >
        수정
      </button>
      <button
        type="button"
        onClick={onDeleteClick}
        className="rounded-lg border border-red-200 px-4 py-2 text-sm font-medium text-red-600 hover:bg-red-50 transition-colors"
      >
        삭제
      </button>
    </div>
  );
}
