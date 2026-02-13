"use client";

import { EmptyState } from "@/components/common/empty-state";

export default function DashboardPage() {
  return (
    <EmptyState
      icon={<span>📊</span>}
      title="지원 현황이 없습니다"
      description="첫 지원을 등록해보세요!"
    />
  );
}
