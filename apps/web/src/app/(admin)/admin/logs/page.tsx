"use client";

import { useCallback, useEffect, useState } from "react";
import { toast } from "sonner";
import { apiClient } from "@/lib/api-client";

interface AuditLog {
  id: string;
  admin_id: string;
  action: string;
  target_type: string;
  target_id?: string;
  old_value?: string;
  new_value?: string;
  created_at: string;
}

const ACTION_LABELS: Record<string, string> = {
  role_change: "역할 변경",
  config_update: "설정 변경",
  suspend: "계정 정지",
  unsuspend: "정지 해제",
  force_logout: "강제 로그아웃",
  plan_change: "플랜 변경",
  deletion_request: "삭제 요청",
};

const ACTION_BADGE_STYLES: Record<string, string> = {
  role_change: "bg-blue-500/10 text-blue-400",
  config_update: "bg-purple-500/10 text-purple-400",
  suspend: "bg-red-500/10 text-red-400",
  unsuspend: "bg-green-500/10 text-green-400",
  force_logout: "bg-orange-500/10 text-orange-400",
  plan_change: "bg-yellow-500/10 text-yellow-400",
  deletion_request: "bg-red-500/10 text-red-400",
};

const PAGE_SIZE = 50;

export default function AdminLogsPage() {
  const [logs, setLogs] = useState<AuditLog[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(0);
  const [actionFilter, setActionFilter] = useState("");
  const [loading, setLoading] = useState(true);

  const fetchLogs = useCallback(async () => {
    setLoading(true);
    try {
      const params = new URLSearchParams({
        limit: String(PAGE_SIZE),
        offset: String(page * PAGE_SIZE),
      });
      if (actionFilter) params.set("action", actionFilter);

      const { data } = await apiClient.get(`/v1/admin/audit-logs?${params}`);
      setLogs(data.data ?? []);
      setTotal(data.total ?? 0);
    } catch {
      toast.error("감사 로그를 불러올 수 없습니다.");
    } finally {
      setLoading(false);
    }
  }, [page, actionFilter]);

  useEffect(() => {
    fetchLogs();
  }, [fetchLogs]);

  const totalPages = Math.ceil(total / PAGE_SIZE);

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold font-display text-foreground">
          감사 로그
        </h1>
        <p className="text-sm text-muted-foreground">총 {total}건</p>
      </div>

      <div className="flex gap-3">
        <select
          value={actionFilter}
          onChange={(e) => {
            setActionFilter(e.target.value);
            setPage(0);
          }}
          className="px-3 py-2 border border-border rounded-lg text-sm bg-card text-foreground"
        >
          <option value="">전체 액션</option>
          {Object.entries(ACTION_LABELS).map(([key, label]) => (
            <option key={key} value={key}>
              {label}
            </option>
          ))}
        </select>
      </div>

      <div className="bg-card border border-border rounded-xl overflow-hidden">
        <table className="w-full text-sm">
          <thead className="bg-white/[0.02] border-b border-border">
            <tr>
              <th className="text-left px-4 py-3 font-medium text-muted-foreground">
                시각
              </th>
              <th className="text-left px-4 py-3 font-medium text-muted-foreground">
                액션
              </th>
              <th className="text-left px-4 py-3 font-medium text-muted-foreground">
                대상
              </th>
              <th className="text-left px-4 py-3 font-medium text-muted-foreground">
                관리자 ID
              </th>
              <th className="text-left px-4 py-3 font-medium text-muted-foreground">
                변경 내용
              </th>
            </tr>
          </thead>
          <tbody>
            {loading ? (
              <tr>
                <td
                  colSpan={5}
                  className="text-center py-8 text-muted-foreground/60"
                >
                  로딩 중...
                </td>
              </tr>
            ) : logs.length === 0 ? (
              <tr>
                <td
                  colSpan={5}
                  className="text-center py-8 text-muted-foreground/60"
                >
                  로그가 없습니다.
                </td>
              </tr>
            ) : (
              logs.map((log) => (
                <tr
                  key={log.id}
                  className="border-b border-border hover:bg-white/[0.04]"
                >
                  <td className="px-4 py-3 text-muted-foreground whitespace-nowrap">
                    {new Date(log.created_at).toLocaleString("ko-KR")}
                  </td>
                  <td className="px-4 py-3">
                    <span
                      className={`inline-block px-2 py-0.5 rounded text-xs font-medium ${
                        ACTION_BADGE_STYLES[log.action] ??
                        "bg-white/[0.06] text-foreground/80"
                      }`}
                    >
                      {ACTION_LABELS[log.action] ?? log.action}
                    </span>
                  </td>
                  <td className="px-4 py-3">
                    <span className="text-muted-foreground">
                      {log.target_type}
                    </span>
                    {log.target_id && (
                      <span className="ml-1 font-mono text-xs text-foreground/60">
                        {log.target_id.length > 8
                          ? log.target_id.slice(0, 8) + "..."
                          : log.target_id}
                      </span>
                    )}
                  </td>
                  <td className="px-4 py-3 font-mono text-xs text-foreground/60">
                    {log.admin_id.slice(0, 8)}...
                  </td>
                  <td className="px-4 py-3 text-muted-foreground">
                    {log.old_value && log.new_value ? (
                      <span>
                        <span className="line-through text-red-400/70">
                          {log.old_value}
                        </span>
                        {" → "}
                        <span className="text-green-400">
                          {log.new_value}
                        </span>
                      </span>
                    ) : log.new_value ? (
                      <span className="text-green-400">{log.new_value}</span>
                    ) : (
                      "-"
                    )}
                  </td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>

      {totalPages > 1 && (
        <div className="flex items-center justify-center gap-2">
          <button
            onClick={() => setPage((p) => Math.max(0, p - 1))}
            disabled={page === 0}
            className="px-3 py-1 text-sm border border-border rounded text-foreground disabled:opacity-50"
          >
            이전
          </button>
          <span className="text-sm text-muted-foreground">
            {page + 1} / {totalPages}
          </span>
          <button
            onClick={() => setPage((p) => Math.min(totalPages - 1, p + 1))}
            disabled={page >= totalPages - 1}
            className="px-3 py-1 text-sm border border-border rounded text-foreground disabled:opacity-50"
          >
            다음
          </button>
        </div>
      )}
    </div>
  );
}
