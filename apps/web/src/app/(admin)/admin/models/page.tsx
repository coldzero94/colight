"use client";

import { useCallback, useEffect, useState } from "react";
import { toast } from "sonner";
import { apiClient } from "@/lib/api-client";

interface ModelStat {
  model: string;
  provider: string;
  call_count: number;
  success_count: number;
  error_count: number;
  error_rate: number;
  total_tokens: number;
  input_tokens: number;
  output_tokens: number;
  total_cost_krw: number;
  avg_cost_krw: number;
  avg_latency_ms: number;
  last_used: string;
}

interface GeminiQuota {
  model: string;
  limit: number;
  used: number;
  remaining: number;
  percentage: number;
}

export default function AdminModelsPage() {
  const [days, setDays] = useState(30);
  const [stats, setStats] = useState<ModelStat[]>([]);
  const [featureModels, setFeatureModels] = useState<Record<string, string>>({});
  const [geminiQuotas, setGeminiQuotas] = useState<GeminiQuota[]>([]);
  const [loading, setLoading] = useState(true);

  const fetchData = useCallback(async () => {
    setLoading(true);
    try {
      const res = await apiClient.get(`/v1/admin/models/stats?days=${days}`);
      setStats(res.data.data ?? []);
      setFeatureModels(res.data.feature_models ?? {});
      setGeminiQuotas(res.data.gemini_quotas ?? []);
    } catch {
      toast.error("모델 통계를 불러올 수 없습니다.");
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

  const sortedStats = [...stats].sort((a, b) => b.call_count - a.call_count);

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold font-display text-foreground">
            AI 모델 사용 통계
          </h1>
          <p className="mt-1 text-sm text-muted-foreground">
            모델별 호출 횟수, 비용, 성공률, Quota 모니터링
          </p>
        </div>
        <select
          value={days}
          onChange={(e) => setDays(Number(e.target.value))}
          className="px-3 py-2 border border-border rounded-lg text-sm bg-card text-foreground"
        >
          <option value={7}>최근 7일</option>
          <option value={30}>최근 30일</option>
          <option value={90}>최근 90일</option>
        </select>
      </div>

      {/* Gemini Quota Status */}
      {geminiQuotas.length > 0 && (
        <div className="bg-card border border-border rounded-xl p-6">
          <h2 className="text-sm font-semibold text-muted-foreground uppercase tracking-wider mb-4">
            Gemini Quota 상태 (오늘)
          </h2>
          <div className="grid grid-cols-1 md:grid-cols-3 lg:grid-cols-5 gap-4">
            {geminiQuotas.map((q) => (
              <div key={q.model} className="p-4 bg-background rounded-lg">
                <p className="text-xs text-muted-foreground mb-2 truncate">{q.model}</p>
                <p className="text-2xl font-bold text-foreground">
                  {q.remaining.toLocaleString()}
                </p>
                <p className="text-xs text-muted-foreground mb-2">
                  / {q.limit.toLocaleString()}
                </p>
                <div className="h-2 bg-border rounded-full overflow-hidden">
                  <div
                    className={`h-full transition-all ${
                      q.percentage < 50
                        ? "bg-green-500"
                        : q.percentage < 80
                        ? "bg-yellow-500"
                        : "bg-red-500"
                    }`}
                    style={{ width: `${Math.min(q.percentage, 100)}%` }}
                  />
                </div>
                <p className="mt-1 text-xs text-muted-foreground">
                  {q.percentage.toFixed(1)}% 사용
                </p>
              </div>
            ))}
          </div>
          <p className="mt-4 text-xs text-muted-foreground/60">
            ⏰ Quota는 자정(PT)에 리셋됩니다. Groq는 response headers 기반 실시간 추적.
          </p>
        </div>
      )}

      {/* Feature Configuration */}
      <div className="bg-card border border-border rounded-xl p-6">
        <h2 className="text-sm font-semibold text-muted-foreground uppercase tracking-wider mb-4">
          기능별 설정 모델
        </h2>
        <div className="grid grid-cols-2 lg:grid-cols-3 gap-3 text-sm">
          {Object.entries(featureModels).map(([feature, model]) => (
            <div
              key={feature}
              className="flex justify-between p-3 bg-background rounded-lg"
            >
              <span className="text-muted-foreground truncate mr-2">{feature}</span>
              <span className="font-mono text-xs text-primary">{model}</span>
            </div>
          ))}
        </div>
      </div>

      {/* Model Statistics Table */}
      <div className="bg-card border border-border rounded-xl overflow-hidden">
        <div className="p-6 pb-0">
          <h2 className="text-sm font-semibold text-muted-foreground uppercase tracking-wider mb-4">
            모델별 사용 통계
          </h2>
        </div>
        {sortedStats.length === 0 ? (
          <div className="p-8 text-center text-muted-foreground">
            선택한 기간에 데이터가 없습니다.
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead className="bg-white/[0.02] border-b border-border">
                <tr>
                  <th className="text-left px-4 py-3 text-muted-foreground font-medium">
                    모델
                  </th>
                  <th className="text-left px-4 py-3 text-muted-foreground font-medium">
                    프로바이더
                  </th>
                  <th className="text-right px-4 py-3 text-muted-foreground font-medium">
                    호출 수
                  </th>
                  <th className="text-right px-4 py-3 text-muted-foreground font-medium">
                    성공률
                  </th>
                  <th className="text-right px-4 py-3 text-muted-foreground font-medium">
                    토큰
                  </th>
                  <th className="text-right px-4 py-3 text-muted-foreground font-medium">
                    평균 비용
                  </th>
                  <th className="text-right px-4 py-3 text-muted-foreground font-medium">
                    지연시간
                  </th>
                  <th className="text-right px-4 py-3 text-muted-foreground font-medium">
                    최근 사용
                  </th>
                </tr>
              </thead>
              <tbody>
                {sortedStats.map((s) => (
                  <tr
                    key={s.model}
                    className="border-b border-border/50 hover:bg-white/[0.02]"
                  >
                    <td className="px-4 py-3 font-mono text-xs text-foreground">
                      {s.model}
                    </td>
                    <td className="px-4 py-3 text-muted-foreground capitalize">
                      {s.provider}
                    </td>
                    <td className="px-4 py-3 text-right text-muted-foreground">
                      {s.call_count.toLocaleString()}
                    </td>
                    <td className="px-4 py-3 text-right">
                      <span
                        className={
                          s.error_rate < 5
                            ? "text-green-400"
                            : s.error_rate < 20
                            ? "text-yellow-400"
                            : "text-red-400"
                        }
                      >
                        {((s.success_count / s.call_count) * 100).toFixed(1)}%
                      </span>
                    </td>
                    <td className="px-4 py-3 text-right text-muted-foreground">
                      {s.total_tokens.toLocaleString()}
                    </td>
                    <td className="px-4 py-3 text-right text-foreground">
                      ₩{s.avg_cost_krw.toFixed(0)}
                    </td>
                    <td className="px-4 py-3 text-right text-muted-foreground">
                      {s.avg_latency_ms.toFixed(0)}ms
                    </td>
                    <td className="px-4 py-3 text-right text-muted-foreground text-xs">
                      {new Date(s.last_used).toLocaleDateString("ko-KR")}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>
    </div>
  );
}
