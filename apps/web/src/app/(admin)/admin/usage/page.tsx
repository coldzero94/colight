"use client";

import { useCallback, useEffect, useState } from "react";
import { toast } from "sonner";
import { apiClient } from "@/lib/api-client";

interface UsageSummary {
  total_calls: number;
  total_tokens: number;
  total_cost_krw: number;
  error_rate: number;
  error_count: number;
  days: number;
}

interface ProviderCost {
  provider: string;
  total_tokens: number;
  total_cost_krw: number;
  call_count: number;
}

interface TopUser {
  user_id: string;
  total_tokens: number;
  total_cost_krw: number;
  call_count: number;
}

interface DailyUsage {
  date: string;
  total_calls: number;
  total_tokens: number;
  error_count: number;
}

export default function AdminUsagePage() {
  const [days, setDays] = useState(30);
  const [summary, setSummary] = useState<UsageSummary | null>(null);
  const [costs, setCosts] = useState<ProviderCost[]>([]);
  const [topUsers, setTopUsers] = useState<TopUser[]>([]);
  const [daily, setDaily] = useState<DailyUsage[]>([]);
  const [loading, setLoading] = useState(true);

  const fetchData = useCallback(async () => {
    setLoading(true);
    try {
      const [summaryRes, costsRes, topRes, dailyRes] = await Promise.all([
        apiClient.get(`/v1/admin/usage/summary?days=${days}`),
        apiClient.get(`/v1/admin/usage/costs?days=${days}`),
        apiClient.get(`/v1/admin/usage/top-users?days=${days}&limit=10`),
        apiClient.get(`/v1/admin/usage/daily?days=${days}`),
      ]);
      setSummary(summaryRes.data);
      setCosts(costsRes.data.data ?? []);
      setTopUsers(topRes.data.data ?? []);
      setDaily(
        (dailyRes.data.data ?? []).sort(
          (a: DailyUsage, b: DailyUsage) => a.date.localeCompare(b.date)
        )
      );
    } catch {
      toast.error("사용량 데이터를 불러올 수 없습니다.");
    } finally {
      setLoading(false);
    }
  }, [days]);

  useEffect(() => {
    fetchData();
  }, [fetchData]);

  if (loading) {
    return (
      <div className="flex items-center justify-center h-full">
        <div className="animate-spin h-8 w-8 border-4 border-border border-t-foreground rounded-full" />
      </div>
    );
  }

  return (
    <div className="space-y-8">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold font-display text-foreground">
          사용량 모니터링
        </h1>
        <select
          value={days}
          onChange={(e) => setDays(Number(e.target.value))}
          className="px-3 py-2 border border-border rounded-lg text-sm bg-card text-foreground"
        >
          <option value={7}>최근 7일</option>
          <option value={14}>최근 14일</option>
          <option value={30}>최근 30일</option>
          <option value={90}>최근 90일</option>
        </select>
      </div>

      {/* Summary Cards */}
      {summary && (
        <div className="grid grid-cols-2 lg:grid-cols-5 gap-4">
          <div className="bg-card rounded-xl border border-border p-4">
            <p className="text-sm text-muted-foreground">총 호출 수</p>
            <p className="text-2xl font-bold text-foreground mt-1">
              {summary.total_calls.toLocaleString()}
            </p>
          </div>
          <div className="bg-card rounded-xl border border-border p-4">
            <p className="text-sm text-muted-foreground">총 토큰</p>
            <p className="text-2xl font-bold text-foreground mt-1">
              {summary.total_tokens.toLocaleString()}
            </p>
          </div>
          <div className="bg-card rounded-xl border border-border p-4">
            <p className="text-sm text-muted-foreground">총 비용</p>
            <p className="text-2xl font-bold text-foreground mt-1">
              {summary.total_cost_krw.toLocaleString(undefined, {
                maximumFractionDigits: 0,
              })}
              원
            </p>
          </div>
          <div className="bg-card rounded-xl border border-border p-4">
            <p className="text-sm text-muted-foreground">에러율</p>
            <p className="text-2xl font-bold text-foreground mt-1">
              {summary.error_rate.toFixed(1)}%
            </p>
          </div>
          <div className="bg-card rounded-xl border border-border p-4">
            <p className="text-sm text-muted-foreground">에러 수</p>
            <p className="text-2xl font-bold text-foreground mt-1">
              {summary.error_count.toLocaleString()}
            </p>
          </div>
        </div>
      )}

      <div className="grid lg:grid-cols-2 gap-6">
        {/* Provider Cost Breakdown */}
        <div className="bg-card border border-border rounded-xl p-6">
          <h2 className="text-lg font-semibold text-foreground mb-4">
            프로바이더별 비용
          </h2>
          {costs.length === 0 ? (
            <p className="text-sm text-muted-foreground">데이터가 없습니다.</p>
          ) : (
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b border-border">
                  <th className="text-left py-2 text-muted-foreground font-medium">
                    프로바이더
                  </th>
                  <th className="text-right py-2 text-muted-foreground font-medium">
                    호출 수
                  </th>
                  <th className="text-right py-2 text-muted-foreground font-medium">
                    토큰
                  </th>
                  <th className="text-right py-2 text-muted-foreground font-medium">
                    비용 (원)
                  </th>
                </tr>
              </thead>
              <tbody>
                {costs.map((c) => (
                  <tr key={c.provider} className="border-b border-border/50">
                    <td className="py-2 font-medium text-foreground">
                      {c.provider}
                    </td>
                    <td className="py-2 text-right text-muted-foreground">
                      {c.call_count.toLocaleString()}
                    </td>
                    <td className="py-2 text-right text-muted-foreground">
                      {c.total_tokens.toLocaleString()}
                    </td>
                    <td className="py-2 text-right text-foreground">
                      {c.total_cost_krw.toLocaleString(undefined, {
                        maximumFractionDigits: 0,
                      })}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          )}
        </div>

        {/* Top Users */}
        <div className="bg-card border border-border rounded-xl p-6">
          <h2 className="text-lg font-semibold text-foreground mb-4">
            상위 사용자 (토큰 기준)
          </h2>
          {topUsers.length === 0 ? (
            <p className="text-sm text-muted-foreground">데이터가 없습니다.</p>
          ) : (
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b border-border">
                  <th className="text-left py-2 text-muted-foreground font-medium">
                    #
                  </th>
                  <th className="text-left py-2 text-muted-foreground font-medium">
                    사용자 ID
                  </th>
                  <th className="text-right py-2 text-muted-foreground font-medium">
                    호출
                  </th>
                  <th className="text-right py-2 text-muted-foreground font-medium">
                    토큰
                  </th>
                  <th className="text-right py-2 text-muted-foreground font-medium">
                    비용
                  </th>
                </tr>
              </thead>
              <tbody>
                {topUsers.map((u, i) => (
                  <tr
                    key={u.user_id}
                    className="border-b border-border/50"
                  >
                    <td className="py-2 text-muted-foreground">{i + 1}</td>
                    <td className="py-2 font-mono text-xs text-foreground">
                      {u.user_id.slice(0, 8)}...
                    </td>
                    <td className="py-2 text-right text-muted-foreground">
                      {u.call_count.toLocaleString()}
                    </td>
                    <td className="py-2 text-right text-muted-foreground">
                      {u.total_tokens.toLocaleString()}
                    </td>
                    <td className="py-2 text-right text-foreground">
                      {u.total_cost_krw.toLocaleString(undefined, {
                        maximumFractionDigits: 0,
                      })}
                      원
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          )}
        </div>
      </div>

      {/* Daily Usage Table */}
      <div className="bg-card border border-border rounded-xl p-6">
        <h2 className="text-lg font-semibold text-foreground mb-4">
          일별 사용량
        </h2>
        {daily.length === 0 ? (
          <p className="text-sm text-muted-foreground">데이터가 없습니다.</p>
        ) : (
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-border">
                <th className="text-left py-2 text-muted-foreground font-medium">
                  날짜
                </th>
                <th className="text-right py-2 text-muted-foreground font-medium">
                  호출 수
                </th>
                <th className="text-right py-2 text-muted-foreground font-medium">
                  토큰
                </th>
                <th className="text-right py-2 text-muted-foreground font-medium">
                  에러
                </th>
              </tr>
            </thead>
            <tbody>
              {daily.map((d) => (
                <tr key={d.date} className="border-b border-border/50">
                  <td className="py-2 text-foreground">{d.date}</td>
                  <td className="py-2 text-right text-muted-foreground">
                    {d.total_calls.toLocaleString()}
                  </td>
                  <td className="py-2 text-right text-muted-foreground">
                    {d.total_tokens.toLocaleString()}
                  </td>
                  <td className="py-2 text-right text-muted-foreground">
                    {d.error_count}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>
    </div>
  );
}
