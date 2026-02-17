"use client";

import { useRecentAnalyses } from "@/hooks/use-company-analysis";
import { BarChart3, ExternalLink } from "lucide-react";
import Link from "next/link";
import { EmptyState } from "@/components/common/empty-state";

function formatTimeAgo(dateString: string): string {
  const date = new Date(dateString);
  const now = new Date();
  const diffMs = now.getTime() - date.getTime();
  const diffMins = Math.floor(diffMs / 60000);
  const diffHours = Math.floor(diffMs / 3600000);
  const diffDays = Math.floor(diffMs / 86400000);

  if (diffMins < 1) return "방금 전";
  if (diffMins < 60) return `${diffMins}분 전`;
  if (diffHours < 24) return `${diffHours}시간 전`;
  return `${diffDays}일 전`;
}

export function RecentAnalysesList() {
  const { data, isLoading, error } = useRecentAnalyses(10);

  if (isLoading) {
    return (
      <div className="space-y-3">
        {[...Array(3)].map((_, i) => (
          <div key={i} className="h-20 animate-pulse rounded-lg bg-accent" />
        ))}
      </div>
    );
  }

  if (error) {
    return (
      <div className="rounded-lg border border-destructive/50 bg-destructive/10 p-4 text-sm text-destructive">
        분석 기록을 불러올 수 없습니다.
      </div>
    );
  }

  if (!data?.analyses || data.analyses.length === 0) {
    return (
      <EmptyState
        icon={<BarChart3 className="h-7 w-7" />}
        title="분석 기록이 없습니다"
        description="채용공고 URL을 입력하여 첫 기업 분석을 시작하세요."
      />
    );
  }

  return (
    <div className="space-y-3">
      {data.analyses.map((analysis) => (
        <Link
          key={analysis.id}
          href={`/analysis/result?company=${encodeURIComponent(analysis.company_name)}`}
          className="group block rounded-lg border border-border bg-card p-4 transition-colors hover:border-primary/30 hover:bg-accent"
        >
          <div className="flex items-start justify-between gap-4">
            <div className="flex-1 min-w-0">
              <h3 className="font-semibold text-foreground group-hover:text-primary transition-colors truncate">
                {analysis.company_name}
              </h3>
              <p className="mt-1 text-xs text-muted-foreground">
                {formatTimeAgo(analysis.created_at)}
              </p>
            </div>
            <ExternalLink className="h-4 w-4 flex-shrink-0 text-muted-foreground group-hover:text-primary transition-colors" />
          </div>
        </Link>
      ))}
    </div>
  );
}
