"use client";

import { useRouter } from "next/navigation";
import { useApplications, useApplicationStats, useUpdateApplicationStatus } from "@/hooks/use-applications";
import { DashboardSummary } from "@/components/dashboard/dashboard-summary";
import { KanbanBoard } from "@/components/dashboard/kanban-board";
import { EmptyState } from "@/components/common/empty-state";
import type { ApplicationStatus } from "@/lib/api/applications";

export default function DashboardPage() {
  const router = useRouter();
  const { data: applications, isLoading: appsLoading } = useApplications();
  const { data: stats, isLoading: statsLoading } = useApplicationStats();
  const updateStatus = useUpdateApplicationStatus();

  const handleStatusChange = (id: string, status: ApplicationStatus) => {
    updateStatus.mutate({ id, status });
  };

  if (appsLoading || statsLoading) {
    return (
      <div className="space-y-4">
        <div className="grid grid-cols-2 gap-3 lg:grid-cols-4">
          {Array.from({ length: 4 }).map((_, i) => (
            <div key={i} className="h-20 animate-pulse rounded-lg bg-white/[0.06]" />
          ))}
        </div>
        <div className="flex gap-3">
          {Array.from({ length: 6 }).map((_, i) => (
            <div key={i} className="h-64 w-[280px] shrink-0 animate-pulse rounded-lg bg-white/[0.06] lg:flex-1" />
          ))}
        </div>
      </div>
    );
  }

  if (!applications || applications.length === 0) {
    return (
      <EmptyState
        icon={<span className="text-4xl">📊</span>}
        title="지원 현황이 없습니다"
        description="기업 분석을 시작하면 지원 현황이 여기에 표시됩니다."
        action={{ label: "기업 분석 시작", onClick: () => router.push("/analysis") }}
      />
    );
  }

  return (
    <div className="space-y-6">
      <h1 className="text-xl font-bold text-foreground">대시보드</h1>

      {stats && <DashboardSummary stats={stats} />}

      <KanbanBoard
        applications={applications}
        onStatusChange={handleStatusChange}
      />
    </div>
  );
}
