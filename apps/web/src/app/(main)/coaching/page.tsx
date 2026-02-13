"use client";

import { EmptyState } from "@/components/common/empty-state";

export default function CoachingPage() {
  return (
    <EmptyState
      icon={<span>✍️</span>}
      title="작성 중인 자소서가 없습니다"
      description="기업 분석 후 코칭을 시작하세요"
    />
  );
}
