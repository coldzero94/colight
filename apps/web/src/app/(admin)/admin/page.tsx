"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { apiClient } from "@/lib/api-client";

interface SystemStats {
  total_users: number;
  active_users_today: number;
  total_experiences: number;
  email_auth_users: number;
  naver_auth_users: number;
}

export default function AdminDashboardPage() {
  const [stats, setStats] = useState<SystemStats | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    apiClient
      .get("/v1/admin/stats")
      .then(({ data }) => setStats(data))
      .catch(() => {})
      .finally(() => setLoading(false));
  }, []);

  if (loading) {
    return (
      <div className="flex items-center justify-center h-full">
        <div className="animate-spin h-8 w-8 border-4 border-gray-300 border-t-gray-900 rounded-full" />
      </div>
    );
  }

  if (!stats) {
    return <p className="text-gray-500">통계를 불러올 수 없습니다.</p>;
  }

  const statCards = [
    { label: "전체 사용자", value: stats.total_users },
    { label: "오늘 활성 사용자", value: stats.active_users_today },
    { label: "전체 경험", value: stats.total_experiences },
    { label: "Naver 인증", value: stats.naver_auth_users },
    { label: "Email 인증", value: stats.email_auth_users },
  ];

  return (
    <div className="space-y-8">
      <h1 className="text-2xl font-bold text-gray-900">대시보드</h1>

      <div className="grid grid-cols-2 lg:grid-cols-5 gap-4">
        {statCards.map((card) => (
          <div
            key={card.label}
            className="bg-white rounded-xl border border-gray-200 p-4"
          >
            <p className="text-sm text-gray-500">{card.label}</p>
            <p className="text-2xl font-bold text-gray-900 mt-1">
              {card.value.toLocaleString()}
            </p>
          </div>
        ))}
      </div>

      <div className="flex gap-3">
        <Link
          href="/admin/users"
          className="bg-white border border-gray-200 rounded-lg px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50 transition-colors"
        >
          사용자 관리
        </Link>
        <Link
          href="/admin/prompts"
          className="bg-white border border-gray-200 rounded-lg px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50 transition-colors"
        >
          프롬프트 관리
        </Link>
      </div>
    </div>
  );
}
