"use client";

import { EmptyState } from "@/components/common/empty-state";

export default function AnalysisPage() {
  return (
    <EmptyState
      icon={<span>🏢</span>}
      title="분석된 기업이 없습니다"
      description="채용공고 URL을 입력해보세요!"
      action={{
        label: "기업 분석",
        onClick: () => {
          // Phase 3 에서 구현
        },
      }}
    />
  );
}
