"use client";

import { useCallback, useEffect, useState } from "react";
import { toast } from "sonner";
import { apiClient } from "@/lib/api-client";

// ─── Types ───

type TabId = "overview" | "errors" | "models";

interface DashboardData {
  user_stats: { total_users: number; new_users_today: number; active_users_today: number };
  ai_metrics_today: {
    calls: number;
    calls_delta_pct: number;
    error_rate: number;
    error_rate_delta: number;
    cost_krw: number;
  };
  daily_metrics: { date: string; calls: number; error_count: number; cost_krw: number }[];
  quota_alerts: { model: string; provider: string; used_pct: number; remaining: number }[];
  recent_errors: AIErrorItem[];
  recent_feedbacks: { id: string; category: string; content_preview: string; created_at: string }[];
  pending_feedback_count: number;
}

interface AIErrorItem {
  id: string;
  error_type: string;
  error_message: string;
  model?: string;
  provider?: string;
  feature: string;
  user_id: string;
  input_tokens: number;
  output_tokens: number;
  created_at: string;
}

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
  rpm: number;
  tpm: number;
  rpd: number;
  context_size: number;
  note: string;
}

interface ModelQuota {
  model: string;
  provider: string;
  limit: number;
  used: number;
  remaining: number;
  percentage: number;
  rpm: number;
  tpm: number;
}

interface QuotaHit {
  model: string;
  provider: string;
  hit_count: number;
  last_hit_at: string;
  last_error: string;
}

// ─── Tab Components ───

function OverviewTab({ dashboard }: { dashboard: DashboardData | null }) {
  if (!dashboard) {
    return <div className="p-8 text-center text-muted-foreground">데이터를 불러오는 중...</div>;
  }

  const { ai_metrics_today: ai, daily_metrics, quota_alerts, recent_errors } = dashboard;

  return (
    <div className="space-y-6">
      {/* KPI Cards */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        <div className="bg-card border border-border rounded-xl p-5">
          <p className="text-xs text-muted-foreground uppercase tracking-wider mb-1">오늘 AI 호출</p>
          <div className="flex items-baseline gap-2">
            <p className="text-3xl font-bold text-foreground">{ai.calls.toLocaleString()}</p>
            <span className={`text-xs ${ai.calls_delta_pct >= 0 ? "text-green-400" : "text-red-400"}`}>
              {ai.calls_delta_pct >= 0 ? "+" : ""}{ai.calls_delta_pct.toFixed(1)}%
            </span>
          </div>
          <p className="text-[10px] text-muted-foreground/60 mt-1">전일 대비</p>
        </div>

        <div className="bg-card border border-border rounded-xl p-5">
          <p className="text-xs text-muted-foreground uppercase tracking-wider mb-1">오늘 에러율</p>
          <div className="flex items-baseline gap-2">
            <p className={`text-3xl font-bold ${ai.error_rate < 5 ? "text-green-400" : ai.error_rate < 15 ? "text-yellow-400" : "text-red-400"}`}>
              {ai.error_rate.toFixed(1)}%
            </p>
            <span className={`text-xs ${ai.error_rate_delta <= 0 ? "text-green-400" : "text-red-400"}`}>
              {ai.error_rate_delta >= 0 ? "+" : ""}{ai.error_rate_delta.toFixed(1)}p
            </span>
          </div>
          <p className="text-[10px] text-muted-foreground/60 mt-1">전일 대비</p>
        </div>

        <div className="bg-card border border-border rounded-xl p-5">
          <p className="text-xs text-muted-foreground uppercase tracking-wider mb-1">오늘 비용</p>
          <p className="text-3xl font-bold text-foreground">₩{ai.cost_krw.toLocaleString(undefined, { maximumFractionDigits: 0 })}</p>
          <p className="text-[10px] text-muted-foreground/60 mt-1">예상 비용 (KRW)</p>
        </div>
      </div>

      {/* Daily Metrics Chart (simple bar) */}
      {daily_metrics.length > 0 && (
        <div className="bg-card border border-border rounded-xl p-6">
          <h2 className="text-sm font-semibold text-muted-foreground uppercase tracking-wider mb-4">
            일별 호출 추이 (7일)
          </h2>
          <div className="flex items-end gap-1 h-32">
            {daily_metrics.map((d) => {
              const max = Math.max(...daily_metrics.map((m) => m.calls), 1);
              const height = (d.calls / max) * 100;
              return (
                <div key={d.date} className="flex-1 flex flex-col items-center gap-1">
                  <span className="text-[10px] text-muted-foreground">{d.calls}</span>
                  <div className="w-full relative" style={{ height: `${height}%`, minHeight: "2px" }}>
                    <div className="absolute inset-0 bg-primary/70 rounded-t" />
                    {d.error_count > 0 && (
                      <div
                        className="absolute bottom-0 inset-x-0 bg-red-500/80 rounded-t"
                        style={{ height: `${(d.error_count / d.calls) * 100}%` }}
                      />
                    )}
                  </div>
                  <span className="text-[9px] text-muted-foreground/60">{d.date.slice(5)}</span>
                </div>
              );
            })}
          </div>
          <div className="flex gap-4 mt-3 text-[10px] text-muted-foreground/60">
            <span className="flex items-center gap-1"><span className="w-2 h-2 bg-primary/70 rounded" /> 호출</span>
            <span className="flex items-center gap-1"><span className="w-2 h-2 bg-red-500/80 rounded" /> 에러</span>
          </div>
        </div>
      )}

      {/* Quota Alerts */}
      {quota_alerts.length > 0 && (
        <div className="bg-card border border-yellow-500/30 rounded-xl p-6">
          <h2 className="text-sm font-semibold text-yellow-400 uppercase tracking-wider mb-4">
            Quota 경고
          </h2>
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-3">
            {quota_alerts.map((a) => (
              <div key={a.model} className="p-3 bg-yellow-500/5 border border-yellow-500/20 rounded-lg">
                <p className="text-xs font-mono text-foreground truncate">{a.model}</p>
                <p className="text-[10px] text-muted-foreground capitalize">{a.provider}</p>
                <div className="flex items-baseline gap-1 mt-1">
                  <span className={`text-lg font-bold ${a.used_pct > 80 ? "text-red-400" : a.used_pct > 50 ? "text-yellow-400" : "text-green-400"}`}>
                    {a.used_pct.toFixed(0)}%
                  </span>
                  <span className="text-xs text-muted-foreground">사용 / 잔여 {a.remaining}</span>
                </div>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Recent Errors */}
      {recent_errors.length > 0 && (
        <div className="bg-card border border-border rounded-xl p-6">
          <h2 className="text-sm font-semibold text-muted-foreground uppercase tracking-wider mb-4">
            최근 에러 (5건)
          </h2>
          <div className="space-y-2">
            {recent_errors.slice(0, 5).map((e) => (
              <div key={e.id} className="flex items-start gap-3 p-3 bg-background rounded-lg">
                <span className={`mt-0.5 shrink-0 text-[10px] px-1.5 py-0.5 rounded font-medium ${
                  e.error_type === "rate_limit" ? "bg-yellow-500/20 text-yellow-400"
                    : e.error_type === "timeout" ? "bg-orange-500/20 text-orange-400"
                    : "bg-red-500/20 text-red-400"
                }`}>
                  {e.error_type}
                </span>
                <div className="flex-1 min-w-0">
                  <p className="text-xs text-foreground truncate">{e.error_message}</p>
                  <p className="text-[10px] text-muted-foreground mt-0.5">
                    {e.model ?? "N/A"} &middot; {e.feature} &middot; {new Date(e.created_at).toLocaleString("ko-KR")}
                  </p>
                </div>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}

const ERROR_TYPE_OPTIONS = [
  { value: "", label: "전체" },
  { value: "rate_limit", label: "Rate Limit" },
  { value: "timeout", label: "Timeout" },
  { value: "provider_error", label: "Provider Error" },
  { value: "invalid_request", label: "Invalid Request" },
  { value: "context_exceeded", label: "Context Exceeded" },
];

function ErrorsTab() {
  const [errors, setErrors] = useState<AIErrorItem[]>([]);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(true);
  const [days, setDays] = useState(7);
  const [errorType, setErrorType] = useState("");
  const [page, setPage] = useState(0);
  const limit = 20;

  const fetchErrors = useCallback(async () => {
    setLoading(true);
    try {
      const params = new URLSearchParams({
        days: String(days),
        limit: String(limit),
        offset: String(page * limit),
      });
      if (errorType) params.set("error_type", errorType);
      const res = await apiClient.get(`/v1/admin/usage/errors?${params}`);
      setErrors(res.data.data ?? []);
      setTotal(res.data.count ?? 0);
    } catch {
      toast.error("에러 목록을 불러올 수 없습니다.");
    } finally {
      setLoading(false);
    }
  }, [days, errorType, page]);

  useEffect(() => {
    fetchErrors();
  }, [fetchErrors]);

  const totalPages = Math.ceil(total / limit);

  return (
    <div className="space-y-4">
      {/* Filters */}
      <div className="flex flex-wrap gap-3">
        <select
          value={errorType}
          onChange={(e) => { setErrorType(e.target.value); setPage(0); }}
          className="px-3 py-2 border border-border rounded-lg text-sm bg-card text-foreground"
        >
          {ERROR_TYPE_OPTIONS.map((o) => (
            <option key={o.value} value={o.value}>{o.label}</option>
          ))}
        </select>
        <select
          value={days}
          onChange={(e) => { setDays(Number(e.target.value)); setPage(0); }}
          className="px-3 py-2 border border-border rounded-lg text-sm bg-card text-foreground"
        >
          <option value={1}>최근 1일</option>
          <option value={7}>최근 7일</option>
          <option value={30}>최근 30일</option>
        </select>
        <span className="flex items-center text-xs text-muted-foreground ml-auto">
          총 {total.toLocaleString()}건
        </span>
      </div>

      {/* Error Table */}
      {loading ? (
        <div className="flex justify-center py-12">
          <div className="animate-spin h-8 w-8 border-4 border-border border-t-foreground rounded-full" />
        </div>
      ) : errors.length === 0 ? (
        <div className="bg-card border border-border rounded-xl p-8 text-center text-muted-foreground">
          선택한 조건에 해당하는 에러가 없습니다.
        </div>
      ) : (
        <div className="bg-card border border-border rounded-xl overflow-hidden">
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead className="bg-white/[0.02] border-b border-border">
                <tr>
                  <th className="text-left px-4 py-3 text-muted-foreground font-medium">유형</th>
                  <th className="text-left px-4 py-3 text-muted-foreground font-medium">메시지</th>
                  <th className="text-left px-4 py-3 text-muted-foreground font-medium">모델</th>
                  <th className="text-left px-4 py-3 text-muted-foreground font-medium">기능</th>
                  <th className="text-right px-4 py-3 text-muted-foreground font-medium">토큰</th>
                  <th className="text-right px-4 py-3 text-muted-foreground font-medium">시간</th>
                </tr>
              </thead>
              <tbody>
                {errors.map((e) => (
                  <tr key={e.id} className="border-b border-border/50 hover:bg-white/[0.02]">
                    <td className="px-4 py-3">
                      <span className={`text-[10px] px-1.5 py-0.5 rounded font-medium ${
                        e.error_type === "rate_limit" ? "bg-yellow-500/20 text-yellow-400"
                          : e.error_type === "timeout" ? "bg-orange-500/20 text-orange-400"
                          : e.error_type === "context_exceeded" ? "bg-purple-500/20 text-purple-400"
                          : "bg-red-500/20 text-red-400"
                      }`}>
                        {e.error_type}
                      </span>
                    </td>
                    <td className="px-4 py-3 text-xs text-foreground max-w-xs truncate">{e.error_message}</td>
                    <td className="px-4 py-3 text-xs font-mono text-muted-foreground">{e.model ?? "-"}</td>
                    <td className="px-4 py-3 text-xs text-muted-foreground">{e.feature}</td>
                    <td className="px-4 py-3 text-right text-xs text-muted-foreground">
                      {(e.input_tokens + e.output_tokens).toLocaleString()}
                    </td>
                    <td className="px-4 py-3 text-right text-xs text-muted-foreground whitespace-nowrap">
                      {new Date(e.created_at).toLocaleString("ko-KR")}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {/* Pagination */}
      {totalPages > 1 && (
        <div className="flex items-center justify-center gap-2">
          <button
            onClick={() => setPage((p) => Math.max(0, p - 1))}
            disabled={page === 0}
            className="px-3 py-1.5 text-xs rounded-lg border border-border text-muted-foreground hover:bg-white/[0.04] disabled:opacity-40"
          >
            이전
          </button>
          <span className="text-xs text-muted-foreground">
            {page + 1} / {totalPages}
          </span>
          <button
            onClick={() => setPage((p) => Math.min(totalPages - 1, p + 1))}
            disabled={page >= totalPages - 1}
            className="px-3 py-1.5 text-xs rounded-lg border border-border text-muted-foreground hover:bg-white/[0.04] disabled:opacity-40"
          >
            다음
          </button>
        </div>
      )}
    </div>
  );
}

function ModelsTab() {
  const [days, setDays] = useState(30);
  const [stats, setStats] = useState<ModelStat[]>([]);
  const [featureModels, setFeatureModels] = useState<Record<string, string>>({});
  const [modelQuotas, setModelQuotas] = useState<ModelQuota[]>([]);
  const [quotaHits, setQuotaHits] = useState<QuotaHit[]>([]);
  const [loading, setLoading] = useState(true);

  const fetchData = useCallback(async () => {
    setLoading(true);
    try {
      const [statsRes, hitsRes] = await Promise.all([
        apiClient.get(`/v1/admin/models/stats?days=${days}`),
        apiClient.get(`/v1/admin/models/rate-limit-hits?hours=24`),
      ]);
      setStats(statsRes.data.data ?? []);
      setFeatureModels(statsRes.data.feature_models ?? {});
      setModelQuotas(statsRes.data.model_quotas ?? []);
      setQuotaHits(hitsRes.data.data ?? []);
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
      <div className="flex items-center justify-center py-12">
        <div className="animate-spin h-8 w-8 border-4 border-border border-t-foreground rounded-full" />
      </div>
    );
  }

  const sortedStats = [...stats].sort((a, b) => b.call_count - a.call_count);

  return (
    <div className="space-y-6">
      {/* Days selector */}
      <div className="flex justify-end">
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

      {/* Model Quota Status */}
      {modelQuotas.length > 0 && (
        <div className="bg-card border border-border rounded-xl p-6">
          <h2 className="text-sm font-semibold text-muted-foreground uppercase tracking-wider mb-4">
            모델 Quota 상태 (오늘)
          </h2>
          {["gemini", "groq"].map((provider) => {
            const providerQuotas = modelQuotas.filter((q) => q.provider === provider);
            if (providerQuotas.length === 0) return null;
            return (
              <div key={provider} className="mb-4">
                <h3 className="text-xs font-medium text-muted-foreground/70 uppercase mb-2">
                  {provider === "gemini" ? "Google Gemini" : "Groq"}
                </h3>
                <div className="grid grid-cols-1 md:grid-cols-3 lg:grid-cols-4 gap-4">
                  {providerQuotas.map((q) => (
                    <div key={q.model} className="p-4 bg-background rounded-lg">
                      <p className="text-xs text-muted-foreground mb-1 truncate">{q.model}</p>
                      <div className="flex items-baseline gap-1 mb-1">
                        <p className="text-2xl font-bold text-foreground">{q.remaining.toLocaleString()}</p>
                        <p className="text-xs text-muted-foreground">/ {q.limit.toLocaleString()} RPD</p>
                      </div>
                      <div className="flex gap-2 text-[10px] text-muted-foreground/70 mb-2">
                        <span>{q.rpm} RPM</span>
                        <span>{q.tpm === 0 ? "Unlimited" : `${q.tpm}K`} TPM</span>
                      </div>
                      <div className="h-2 bg-border rounded-full overflow-hidden">
                        <div
                          className={`h-full transition-all ${
                            q.percentage < 50 ? "bg-green-500"
                              : q.percentage < 80 ? "bg-yellow-500"
                              : "bg-red-500"
                          }`}
                          style={{ width: `${Math.min(q.percentage, 100)}%` }}
                        />
                      </div>
                      <p className="mt-1 text-xs text-muted-foreground">{q.percentage.toFixed(1)}% 사용</p>
                    </div>
                  ))}
                </div>
              </div>
            );
          })}
          <p className="mt-2 text-xs text-muted-foreground/60">
            Gemini: 자정(PT) 리셋 / Groq: 자정(UTC) 리셋
          </p>
        </div>
      )}

      {/* Rate Limit Hits (24h) */}
      {quotaHits.length > 0 && (
        <div className="bg-card border border-red-500/30 rounded-xl p-6">
          <h2 className="text-sm font-semibold text-red-400 uppercase tracking-wider mb-4">
            Rate Limit 발생 (최근 24시간)
          </h2>
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-3">
            {quotaHits.sort((a, b) => b.hit_count - a.hit_count).map((h) => (
              <div key={h.model} className="p-4 bg-red-500/5 border border-red-500/20 rounded-lg">
                <div className="flex items-center justify-between mb-1">
                  <p className="text-xs font-mono text-foreground truncate mr-2">{h.model}</p>
                  <span className="text-xs bg-red-500/20 text-red-400 px-2 py-0.5 rounded-full whitespace-nowrap">
                    {h.hit_count}회
                  </span>
                </div>
                <p className="text-[10px] text-muted-foreground capitalize">{h.provider}</p>
                <p className="text-[10px] text-red-400/70 mt-1 truncate">{h.last_error}</p>
                <p className="text-[10px] text-muted-foreground/60 mt-0.5">
                  마지막: {new Date(h.last_hit_at).toLocaleString("ko-KR")}
                </p>
              </div>
            ))}
          </div>
          <p className="mt-2 text-xs text-muted-foreground/60">
            Rate Limit 발생 시 자동으로 폴백 모델로 전환됩니다. 폴백 성공 시 해당 프롬프트의 모델이 자동으로 업데이트됩니다.
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
            <div key={feature} className="flex justify-between p-3 bg-background rounded-lg">
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
                  <th className="text-left px-4 py-3 text-muted-foreground font-medium">모델</th>
                  <th className="text-left px-4 py-3 text-muted-foreground font-medium">프로바이더</th>
                  <th className="text-right px-4 py-3 text-muted-foreground font-medium">호출 수</th>
                  <th className="text-right px-4 py-3 text-muted-foreground font-medium">성공률</th>
                  <th className="text-right px-4 py-3 text-muted-foreground font-medium">토큰</th>
                  <th className="text-right px-4 py-3 text-muted-foreground font-medium">평균 비용</th>
                  <th className="text-right px-4 py-3 text-muted-foreground font-medium">지연시간</th>
                  <th className="text-right px-4 py-3 text-muted-foreground font-medium">RPM</th>
                  <th className="text-right px-4 py-3 text-muted-foreground font-medium">TPM</th>
                  <th className="text-right px-4 py-3 text-muted-foreground font-medium">RPD</th>
                  <th className="text-right px-4 py-3 text-muted-foreground font-medium">최근 사용</th>
                </tr>
              </thead>
              <tbody>
                {sortedStats.map((s) => (
                  <tr key={s.model} className="border-b border-border/50 hover:bg-white/[0.02]">
                    <td className="px-4 py-3 font-mono text-xs text-foreground">{s.model}</td>
                    <td className="px-4 py-3 text-muted-foreground capitalize">{s.provider}</td>
                    <td className="px-4 py-3 text-right text-muted-foreground">{s.call_count.toLocaleString()}</td>
                    <td className="px-4 py-3 text-right">
                      <span className={
                        s.error_rate < 5 ? "text-green-400"
                          : s.error_rate < 20 ? "text-yellow-400"
                          : "text-red-400"
                      }>
                        {((s.success_count / s.call_count) * 100).toFixed(1)}%
                      </span>
                    </td>
                    <td className="px-4 py-3 text-right text-muted-foreground">{s.total_tokens.toLocaleString()}</td>
                    <td className="px-4 py-3 text-right text-foreground">₩{s.avg_cost_krw.toFixed(0)}</td>
                    <td className="px-4 py-3 text-right text-muted-foreground">{s.avg_latency_ms.toFixed(0)}ms</td>
                    <td className="px-4 py-3 text-right text-muted-foreground text-xs">{s.rpm || "-"}</td>
                    <td className="px-4 py-3 text-right text-muted-foreground text-xs">
                      {s.tpm === 0 ? "Unlimited" : s.tpm ? `${s.tpm}K` : "-"}
                    </td>
                    <td className="px-4 py-3 text-right text-muted-foreground text-xs">
                      {s.rpd ? s.rpd.toLocaleString() : "-"}
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

// ─── Main Page ───

const TABS: { id: TabId; label: string }[] = [
  { id: "overview", label: "개요" },
  { id: "errors", label: "에러 로그" },
  { id: "models", label: "모델 통계" },
];

export default function AdminUsagePage() {
  const [activeTab, setActiveTab] = useState<TabId>("overview");
  const [dashboard, setDashboard] = useState<DashboardData | null>(null);
  const [dashboardLoading, setDashboardLoading] = useState(true);

  useEffect(() => {
    (async () => {
      try {
        const res = await apiClient.get("/v1/admin/dashboard");
        setDashboard(res.data.data);
      } catch {
        toast.error("대시보드 데이터를 불러올 수 없습니다.");
      } finally {
        setDashboardLoading(false);
      }
    })();
  }, []);

  return (
    <div className="space-y-6">
      {/* Header */}
      <div>
        <h1 className="text-2xl font-bold font-display text-foreground">사용량 모니터링</h1>
        <p className="mt-1 text-sm text-muted-foreground">
          AI 호출 현황, 에러 추적, 모델별 통계를 한눈에 확인합니다.
        </p>
      </div>

      {/* Tabs */}
      <div className="flex gap-1 border-b border-border">
        {TABS.map((tab) => (
          <button
            key={tab.id}
            onClick={() => setActiveTab(tab.id)}
            className={`px-4 py-2.5 text-sm font-medium transition-colors relative ${
              activeTab === tab.id
                ? "text-foreground"
                : "text-muted-foreground hover:text-foreground/80"
            }`}
          >
            {tab.label}
            {activeTab === tab.id && (
              <span className="absolute bottom-0 left-0 right-0 h-0.5 bg-primary rounded-full" />
            )}
          </button>
        ))}
      </div>

      {/* Tab Content */}
      {dashboardLoading && activeTab === "overview" ? (
        <div className="flex items-center justify-center py-12">
          <div className="animate-spin h-8 w-8 border-4 border-border border-t-foreground rounded-full" />
        </div>
      ) : (
        <>
          {activeTab === "overview" && <OverviewTab dashboard={dashboard} />}
          {activeTab === "errors" && <ErrorsTab />}
          {activeTab === "models" && <ModelsTab />}
        </>
      )}
    </div>
  );
}
