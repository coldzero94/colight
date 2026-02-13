"use client";

import { EmptyState } from "@/components/common/empty-state";

export default function ExperiencesPage() {
  return (
    <EmptyState
      icon={<span>📋</span>}
      title="아직 등록된 경험이 없습니다"
      description="첫 번째 경험을 등록해보세요!"
      action={{
        label: "경험 등록",
        onClick: () => {
          // Phase 2 에서 구현
        },
      }}
    />
  );
}
