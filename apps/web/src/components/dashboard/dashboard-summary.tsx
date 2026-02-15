"use client";

import type { ApplicationStats } from "@/lib/api/applications";

interface DashboardSummaryProps {
  stats: ApplicationStats;
}

interface StatCardProps {
  testId: string;
  label: string;
  value: number;
  icon: string;
  color: string;
}

function StatCard({ testId, label, value, icon, color }: StatCardProps) {
  return (
    <div className="rounded-lg border bg-white p-4 shadow-sm">
      <div className="flex items-center gap-3">
        <span className="text-2xl">{icon}</span>
        <div>
          <p className="text-xs text-gray-500">{label}</p>
          <p
            data-testid={testId}
            className={`text-2xl font-bold ${color}`}
          >
            {value}
          </p>
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
    <div className="grid grid-cols-2 gap-3 lg:grid-cols-4">
      <StatCard
        testId="stat-total"
        label="전체 지원"
        value={stats.total}
        icon="📊"
        color="text-gray-900"
      />
      <StatCard
        testId="stat-active"
        label="진행중"
        value={active}
        icon="🔄"
        color="text-blue-600"
      />
      <StatCard
        testId="stat-deadlines"
        label="마감 임박"
        value={deadlineCount}
        icon="⏰"
        color={deadlineCount > 0 ? "text-red-600" : "text-gray-900"}
      />
      <StatCard
        testId="stat-completed"
        label="합격"
        value={completed}
        icon="🎉"
        color="text-green-600"
      />
    </div>
  );
}
