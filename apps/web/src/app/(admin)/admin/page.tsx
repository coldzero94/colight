"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { toast } from "sonner";
import { Users, Zap, AlertTriangle, MessageSquare, ArrowUpRight, ArrowDownRight } from "lucide-react";
import { apiClient } from "@/lib/api-client";

interface DashboardData {
  user_stats: {
    total_users: number;
    new_users_today: number;
    active_users_today: number;
  };
  ai_metrics_today: {
    calls: number;
    calls_delta_pct: number;
    error_rate: number;
    error_rate_delta: number;
    cost_krw: number;
  };
  daily_metrics: {
    date: string;
    calls: number;
    error_count: number;
    cost_krw: number;
  }[];
  quota_alerts: {
    model: string;
    provider: string;
    used_pct: number;
    remaining: number;
  }[];
  recent_errors: {
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
  }[];
  recent_feedbacks: {
    id: string;
    category: string;
    content_preview: string;
    created_at: string;
  }[];
  pending_feedback_count: number;
}

function DeltaBadge({ value, suffix = "%", invert = false }: { value: number; suffix?: string; invert?: boolean }) {
  const positive = invert ? value <= 0 : value >= 0;
  return (
    <span className={`inline-flex items-center gap-0.5 text-xs font-medium ${positive ? "text-green-400" : "text-red-400"}`}>
      {value >= 0 ? <ArrowUpRight className="h-3 w-3" /> : <ArrowDownRight className="h-3 w-3" />}
      {Math.abs(value).toFixed(1)}{suffix}
    </span>
  );
}

export default function AdminDashboardPage() {
  const [data, setData] = useState<DashboardData | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    apiClient
      .get("/v1/admin/dashboard")
      .then(({ data: res }) => setData(res.data))
      .catch(() => toast.error("대시보드를 불러올 수 없습니다."))
      .finally(() => setLoading(false));
  }, []);

  if (loading) {
    return (
      <div className="flex items-center justify-center h-full">
        <div className="animate-spin h-8 w-8 border-4 border-border border-t-foreground rounded-full" />
      </div>
    );
  }

  if (!data) {
    return <p className="text-muted-foreground p-8">대시보드를 불러올 수 없습니다.</p>;
  }

  const { user_stats, ai_metrics_today: ai, daily_metrics, quota_alerts, recent_errors, recent_feedbacks, pending_feedback_count } = data;
  const maxCalls = Math.max(...daily_metrics.map((d) => d.calls), 1);

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-bold font-display text-foreground">대시보드</h1>

      {/* KPI Row */}
      <div className="grid grid-cols-2 lg:grid-cols-4 gap-4">
        {/* Users */}
        <div className="bg-card border border-border rounded-xl p-5">
          <div className="flex items-center gap-2 mb-3">
            <span className="inline-flex h-8 w-8 items-center justify-center rounded-lg bg-blue-500/10">
              <Users className="h-4 w-4 text-blue-400" />
            </span>
            <p className="text-xs text-muted-foreground uppercase tracking-wider">사용자</p>
          </div>
          <p className="text-3xl font-bold text-foreground">{user_stats.total_users.toLocaleString()}</p>
          <p className="text-xs text-muted-foreground mt-1">
            오늘 활성 {user_stats.active_users_today} &middot; 신규 {user_stats.new_users_today}
          </p>
        </div>

        {/* AI Calls */}
        <div className="bg-card border border-border rounded-xl p-5">
          <div className="flex items-center gap-2 mb-3">
            <span className="inline-flex h-8 w-8 items-center justify-center rounded-lg bg-primary/10">
              <Zap className="h-4 w-4 text-primary" />
            </span>
            <p className="text-xs text-muted-foreground uppercase tracking-wider">AI 호출</p>
          </div>
          <div className="flex items-baseline gap-2">
            <p className="text-3xl font-bold text-foreground">{ai.calls.toLocaleString()}</p>
            <DeltaBadge value={ai.calls_delta_pct} />
          </div>
          <p className="text-xs text-muted-foreground mt-1">오늘 &middot; 전일 대비</p>
        </div>

        {/* Error Rate */}
        <div className="bg-card border border-border rounded-xl p-5">
          <div className="flex items-center gap-2 mb-3">
            <span className="inline-flex h-8 w-8 items-center justify-center rounded-lg bg-red-500/10">
              <AlertTriangle className="h-4 w-4 text-red-400" />
            </span>
            <p className="text-xs text-muted-foreground uppercase tracking-wider">에러율</p>
          </div>
          <div className="flex items-baseline gap-2">
            <p className={`text-3xl font-bold ${ai.error_rate < 5 ? "text-green-400" : ai.error_rate < 15 ? "text-yellow-400" : "text-red-400"}`}>
              {ai.error_rate.toFixed(1)}%
            </p>
            <DeltaBadge value={ai.error_rate_delta} suffix="p" invert />
          </div>
          <p className="text-xs text-muted-foreground mt-1">오늘 &middot; 전일 대비</p>
        </div>

        {/* Cost */}
        <div className="bg-card border border-border rounded-xl p-5">
          <div className="flex items-center gap-2 mb-3">
            <span className="inline-flex h-8 w-8 items-center justify-center rounded-lg bg-emerald-500/10">
              <span className="text-emerald-400 text-sm font-bold">₩</span>
            </span>
            <p className="text-xs text-muted-foreground uppercase tracking-wider">오늘 비용</p>
          </div>
          <p className="text-3xl font-bold text-foreground">
            ₩{ai.cost_krw.toLocaleString(undefined, { maximumFractionDigits: 0 })}
          </p>
          <p className="text-xs text-muted-foreground mt-1">예상 비용 (KRW)</p>
        </div>
      </div>

      {/* Middle Row: Chart + Quota */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-4">
        {/* Daily Chart */}
        <div className="lg:col-span-2 bg-card border border-border rounded-xl p-6">
          <div className="flex items-center justify-between mb-4">
            <h2 className="text-sm font-semibold text-muted-foreground uppercase tracking-wider">일별 호출 추이</h2>
            <Link href="/admin/usage" className="text-xs text-primary hover:underline">자세히 보기</Link>
          </div>
          {daily_metrics.length > 0 ? (
            <div className="flex items-end gap-1.5 h-36">
              {daily_metrics.map((d) => {
                const height = (d.calls / maxCalls) * 100;
                return (
                  <div key={d.date} className="flex-1 flex flex-col items-center gap-1">
                    <span className="text-[10px] text-muted-foreground tabular-nums">{d.calls}</span>
                    <div className="w-full relative" style={{ height: `${height}%`, minHeight: "2px" }}>
                      <div className="absolute inset-0 bg-primary/60 rounded-t" />
                      {d.error_count > 0 && (
                        <div
                          className="absolute bottom-0 inset-x-0 bg-red-500/70 rounded-t"
                          style={{ height: `${(d.error_count / d.calls) * 100}%` }}
                        />
                      )}
                    </div>
                    <span className="text-[9px] text-muted-foreground/60">{d.date.slice(5)}</span>
                  </div>
                );
              })}
            </div>
          ) : (
            <p className="text-sm text-muted-foreground py-8 text-center">데이터가 없습니다.</p>
          )}
        </div>

        {/* Quota Alerts */}
        <div className="bg-card border border-border rounded-xl p-6">
          <h2 className="text-sm font-semibold text-muted-foreground uppercase tracking-wider mb-4">Quota 현황</h2>
          {quota_alerts.length === 0 ? (
            <p className="text-sm text-green-400 py-4 text-center">정상 범위 내</p>
          ) : (
            <div className="space-y-3">
              {quota_alerts.map((a) => (
                <div key={a.model} className="space-y-1">
                  <div className="flex justify-between text-xs">
                    <span className="text-muted-foreground truncate mr-2">{a.model}</span>
                    <span className={`font-medium ${a.used_pct > 80 ? "text-red-400" : a.used_pct > 50 ? "text-yellow-400" : "text-green-400"}`}>
                      {a.used_pct.toFixed(0)}%
                    </span>
                  </div>
                  <div className="h-1.5 bg-border rounded-full overflow-hidden">
                    <div
                      className={`h-full transition-all ${
                        a.used_pct > 80 ? "bg-red-500" : a.used_pct > 50 ? "bg-yellow-500" : "bg-green-500"
                      }`}
                      style={{ width: `${Math.min(a.used_pct, 100)}%` }}
                    />
                  </div>
                </div>
              ))}
            </div>
          )}
          <Link href="/admin/usage" className="block mt-4 text-xs text-primary hover:underline text-center">
            모든 모델 보기
          </Link>
        </div>
      </div>

      {/* Bottom Row: Errors + Feedbacks */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
        {/* Recent Errors */}
        <div className="bg-card border border-border rounded-xl p-6">
          <div className="flex items-center justify-between mb-4">
            <h2 className="text-sm font-semibold text-muted-foreground uppercase tracking-wider">최근 에러</h2>
            <Link href="/admin/usage" className="text-xs text-primary hover:underline">전체 보기</Link>
          </div>
          {recent_errors.length === 0 ? (
            <p className="text-sm text-green-400 py-4 text-center">최근 에러 없음</p>
          ) : (
            <div className="space-y-2">
              {recent_errors.slice(0, 4).map((e) => (
                <div key={e.id} className="flex items-start gap-2 p-2.5 bg-background rounded-lg">
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
                      {e.feature} &middot; {new Date(e.created_at).toLocaleString("ko-KR")}
                    </p>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>

        {/* Recent Feedbacks */}
        <div className="bg-card border border-border rounded-xl p-6">
          <div className="flex items-center justify-between mb-4">
            <div className="flex items-center gap-2">
              <h2 className="text-sm font-semibold text-muted-foreground uppercase tracking-wider">최근 피드백</h2>
              {pending_feedback_count > 0 && (
                <span className="text-[10px] bg-primary/20 text-primary px-1.5 py-0.5 rounded-full font-medium">
                  {pending_feedback_count} 대기
                </span>
              )}
            </div>
            <Link href="/admin/feedbacks" className="text-xs text-primary hover:underline">전체 보기</Link>
          </div>
          {recent_feedbacks.length === 0 ? (
            <p className="text-sm text-muted-foreground py-4 text-center">최근 피드백 없음</p>
          ) : (
            <div className="space-y-2">
              {recent_feedbacks.slice(0, 4).map((f) => (
                <div key={f.id} className="flex items-start gap-2 p-2.5 bg-background rounded-lg">
                  <span className="mt-0.5 shrink-0">
                    <MessageSquare className={`h-3.5 w-3.5 ${
                      f.category === "bug" ? "text-red-400"
                        : f.category === "improvement" ? "text-blue-400"
                        : "text-muted-foreground"
                    }`} />
                  </span>
                  <div className="flex-1 min-w-0">
                    <p className="text-xs text-foreground truncate">{f.content_preview}</p>
                    <p className="text-[10px] text-muted-foreground mt-0.5">
                      {f.category} &middot; {new Date(f.created_at).toLocaleString("ko-KR")}
                    </p>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      </div>

      {/* Quick Links */}
      <div className="flex gap-3">
        <Link
          href="/admin/users"
          className="bg-card border border-border rounded-lg px-4 py-2 text-sm font-medium text-foreground/80 hover:bg-white/[0.04] transition-colors"
        >
          사용자 관리
        </Link>
        <Link
          href="/admin/prompts"
          className="bg-card border border-border rounded-lg px-4 py-2 text-sm font-medium text-foreground/80 hover:bg-white/[0.04] transition-colors"
        >
          프롬프트 관리
        </Link>
        <Link
          href="/admin/usage"
          className="bg-card border border-border rounded-lg px-4 py-2 text-sm font-medium text-foreground/80 hover:bg-white/[0.04] transition-colors"
        >
          사용량 모니터링
        </Link>
      </div>
    </div>
  );
}
