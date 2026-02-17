"use client";

import { useMemo, useState } from "react";
import { useRouter } from "next/navigation";
import { BarChart3, Plus } from "lucide-react";
import { useApplications, useApplicationStats, useUpdateApplicationStatus } from "@/hooks/use-applications";
import { DashboardSummary } from "@/components/dashboard/dashboard-summary";
import { KanbanBoard } from "@/components/dashboard/kanban-board";
import { KanbanFilterBar, filterApplications, type KanbanFilters } from "@/components/dashboard/kanban-filter-bar";
import { EmptyState } from "@/components/common/empty-state";
import { ApplicationCreateModal } from "@/components/dashboard/application-create-modal";
import { ApplicationEditModal } from "@/components/dashboard/application-edit-modal";
import { ApplicationDeleteDialog } from "@/components/dashboard/application-delete-dialog";
import { Button } from "@/components/ui/button";
import type { ApplicationDetail, ApplicationStatus } from "@/lib/api/applications";

export default function DashboardPage() {
  const router = useRouter();
  const { data: applications, isLoading: appsLoading } = useApplications();
  const { data: stats, isLoading: statsLoading } = useApplicationStats();
  const updateStatus = useUpdateApplicationStatus();

  const [createOpen, setCreateOpen] = useState(false);
  const [editApp, setEditApp] = useState<ApplicationDetail | null>(null);
  const [deleteApp, setDeleteApp] = useState<ApplicationDetail | null>(null);
  const [filters, setFilters] = useState<KanbanFilters>({ query: "", tags: [] });

  const filteredApplications = useMemo(
    () => (applications ? filterApplications(applications, filters) : []),
    [applications, filters],
  );

  const handleStatusChange = (id: string, status: ApplicationStatus) => {
    updateStatus.mutate({ id, status });
  };

  if (appsLoading || statsLoading) {
    return (
      <div className="space-y-5 animate-fade-in">
        <div className="brand-surface-soft grid grid-cols-2 gap-3 rounded-2xl p-4 lg:grid-cols-4">
          {Array.from({ length: 4 }).map((_, i) => (
            <div
              key={i}
              className="h-20 animate-pulse rounded-xl bg-white/[0.07]"
            />
          ))}
        </div>
        <div className="brand-surface-soft flex gap-3 rounded-2xl p-4">
          {Array.from({ length: 6 }).map((_, i) => (
            <div
              key={i}
              className="h-64 w-[280px] shrink-0 animate-pulse rounded-xl bg-white/[0.07] lg:flex-1"
            />
          ))}
        </div>
      </div>
    );
  }

  if (!applications || applications.length === 0) {
    return (
      <>
        <EmptyState
          icon={<BarChart3 className="h-8 w-8" />}
          title="지원 현황이 없습니다"
          description="직접 추가하거나 기업 분석을 시작하면 지원 현황이 여기에 표시됩니다."
          action={{ label: "지원 현황 추가", onClick: () => setCreateOpen(true) }}
        />
        <ApplicationCreateModal open={createOpen} onOpenChange={setCreateOpen} />
      </>
    );
  }

  return (
    <div className="space-y-6 animate-fade-in">
      <section className="brand-surface relative overflow-hidden rounded-2xl px-6 py-5">
        <div className="pointer-events-none absolute inset-y-0 right-[-8%] w-1/2 bg-[radial-gradient(circle,_rgba(61,169,255,0.22)_0%,_transparent_70%)]" />
        <div className="relative flex items-start justify-between gap-4">
          <div>
            <span className="brand-kicker inline-flex items-center rounded-full px-3 py-1 text-[11px] font-semibold tracking-[0.18em] uppercase">
              Pipeline Command
            </span>
            <h1 className="mt-3 text-2xl font-bold font-display text-foreground">
              대시보드
            </h1>
            <p className="mt-1 text-sm text-muted-foreground">
              지원 진행 흐름과 마감 상태를 한눈에 정리하세요.
            </p>
          </div>
          <Button size="sm" onClick={() => setCreateOpen(true)} className="shrink-0">
            <Plus className="mr-1.5 h-4 w-4" />
            추가
          </Button>
        </div>
      </section>

      {stats && <DashboardSummary stats={stats} />}

      <KanbanFilterBar
        applications={applications}
        filters={filters}
        onFiltersChange={setFilters}
      />

      <KanbanBoard
        applications={filteredApplications}
        onStatusChange={handleStatusChange}
        onEdit={setEditApp}
        onDelete={setDeleteApp}
      />

      <ApplicationCreateModal open={createOpen} onOpenChange={setCreateOpen} />
      <ApplicationEditModal
        open={editApp !== null}
        onOpenChange={(v) => { if (!v) setEditApp(null); }}
        application={editApp}
      />
      <ApplicationDeleteDialog
        open={deleteApp !== null}
        onOpenChange={(v) => { if (!v) setDeleteApp(null); }}
        application={deleteApp}
      />
    </div>
  );
}
