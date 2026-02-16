"use client";

import type { LucideIcon } from "lucide-react";
import { AlarmClockCheck, BarChart3, CircleCheckBig, Gauge } from "lucide-react";
import type { ApplicationStats } from "@/lib/api/applications";
import { cn } from "@/lib/utils";

interface DashboardSummaryProps {
  stats: ApplicationStats;
}

interface StatCardProps {
  testId: string;
  label: string;
  hint: string;
  value: number;
  icon: LucideIcon;
  color: string;
  iconTone: string;
  iconBgTone: string;
}

function StatCard({
  testId,
  label,
  hint,
  value,
  icon: Icon,
  color,
  iconTone,
  iconBgTone,
}: StatCardProps) {
  return (
    <div className="brand-surface-soft group relative overflow-hidden rounded-2xl p-4 shadow-[0_16px_32px_rgba(0,0,0,0.18)] transition-all duration-300 hover:-translate-y-1 hover:border-primary/35 hover:shadow-[0_20px_38px_rgba(0,0,0,0.26)]">
      <div className="pointer-events-none absolute inset-x-0 top-0 h-18 bg-gradient-to-b from-white/[0.08] to-transparent opacity-0 transition-opacity duration-300 group-hover:opacity-100" />
      <div className="relative flex items-center gap-3">
        <span
          className={cn(
            "inline-flex h-11 w-11 items-center justify-center rounded-xl border border-white/15 shadow-[inset_0_1px_0_rgba(255,255,255,0.15)]",
            iconBgTone
          )}
        >
          <Icon className={cn("h-5 w-5", iconTone)} aria-hidden="true" />
        </span>
        <div className="min-w-0">
          <p className="text-xs font-medium text-muted-foreground">{label}</p>
          <p
            data-testid={testId}
            className={`text-2xl font-bold tracking-tight ${color}`}
          >
            {value}
          </p>
          <p className="truncate text-[11px] text-muted-foreground/80">{hint}</p>
        </div>
      </div>
    </div>
  );
}

export function DashboardSummary({ stats }: DashboardSummaryProps) {
  const active =
    (stats.by_status["preparing"] ?? 0) +
    (stats.by_status["submitted"] ?? 0) +
    (stats.by_status["in_review"] ?? 0) +
    (stats.by_status["interview"] ?? 0);

  const completed = stats.by_status["accepted"] ?? 0;
  const deadlineCount = stats.upcoming_deadlines?.length ?? 0;

  return (
    <div className="grid grid-cols-2 gap-3 lg:grid-cols-4 animate-slide-up">
      <StatCard
        testId="stat-total"
        label="전체 지원"
        hint="지원 파이프라인 총량"
        value={stats.total}
        icon={BarChart3}
        color="text-foreground"
        iconTone="text-slate-100"
        iconBgTone="bg-slate-500/20"
      />
      <StatCard
        testId="stat-active"
        label="진행중"
        hint={`${Math.round((active / Math.max(stats.total, 1)) * 100)}% 진행 중`}
        value={active}
        icon={Gauge}
        color="text-blue-400"
        iconTone="text-blue-300"
        iconBgTone="bg-blue-500/20"
      />
      <StatCard
        testId="stat-deadlines"
        label="마감 임박"
        hint={deadlineCount > 0 ? "빠른 정리 필요" : "여유 있음"}
        value={deadlineCount}
        icon={AlarmClockCheck}
        color={deadlineCount > 0 ? "text-red-400" : "text-foreground"}
        iconTone={deadlineCount > 0 ? "text-red-300" : "text-slate-100"}
        iconBgTone={deadlineCount > 0 ? "bg-red-500/20" : "bg-slate-500/20"}
      />
      <StatCard
        testId="stat-completed"
        label="합격"
        hint={completed > 0 ? "성과 축적 중" : "첫 합격을 노려보세요"}
        value={completed}
        icon={CircleCheckBig}
        color="text-primary"
        iconTone="text-primary"
        iconBgTone="bg-primary/20"
      />
    </div>
  );
}
