"use client";

import { useCallback, useEffect, useState } from "react";
import { toast } from "sonner";
import { apiClient } from "@/lib/api-client";

interface Feedback {
  id: string;
  user_id: string;
  category: string;
  content: string;
  page_url?: string;
  admin_status?: string;
  admin_note?: string;
  created_at: string;
}

const CATEGORY_LABELS: Record<string, string> = {
  bug: "버그",
  feature: "기능 요청",
  ux: "UX 개선",
  content: "콘텐츠",
  other: "기타",
};

const STATUS_LABELS: Record<string, string> = {
  pending: "대기",
  reviewed: "검토",
  resolved: "해결",
  dismissed: "기각",
};

const STATUS_STYLES: Record<string, string> = {
  pending: "bg-yellow-500/10 text-yellow-400",
  reviewed: "bg-blue-500/10 text-blue-400",
  resolved: "bg-green-500/10 text-green-400",
  dismissed: "bg-white/[0.06] text-foreground/50",
};

const PAGE_SIZE = 20;

export default function AdminFeedbacksPage() {
  const [feedbacks, setFeedbacks] = useState<Feedback[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(0);
  const [categoryFilter, setCategoryFilter] = useState("");
  const [loading, setLoading] = useState(true);
  const [editingId, setEditingId] = useState<string | null>(null);
  const [editStatus, setEditStatus] = useState("");
  const [editNote, setEditNote] = useState("");

  const fetchFeedbacks = useCallback(async () => {
    setLoading(true);
    try {
      const params = new URLSearchParams({
        limit: String(PAGE_SIZE),
        offset: String(page * PAGE_SIZE),
      });
      if (categoryFilter) params.set("category", categoryFilter);

      const { data } = await apiClient.get(`/v1/admin/feedbacks?${params}`);
      setFeedbacks(data.data ?? []);
      setTotal(data.total ?? 0);
    } catch {
      toast.error("피드백을 불러올 수 없습니다.");
    } finally {
      setLoading(false);
    }
  }, [page, categoryFilter]);

  useEffect(() => {
    fetchFeedbacks();
  }, [fetchFeedbacks]);

  const handleEdit = (fb: Feedback) => {
    setEditingId(fb.id);
    setEditStatus(fb.admin_status ?? "pending");
    setEditNote(fb.admin_note ?? "");
  };

  const handleSave = async () => {
    if (!editingId) return;
    try {
      await apiClient.put(`/v1/admin/feedbacks/${editingId}`, {
        admin_status: editStatus,
        admin_note: editNote,
      });
      toast.success("피드백 상태가 변경되었습니다.");
      setEditingId(null);
      fetchFeedbacks();
    } catch {
      toast.error("피드백 상태 변경에 실패했습니다.");
    }
  };

  const totalPages = Math.ceil(total / PAGE_SIZE);

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold font-display text-foreground">
          피드백 관리
        </h1>
        <p className="text-sm text-muted-foreground">총 {total}건</p>
      </div>

      <div className="flex gap-3">
        <select
          value={categoryFilter}
          onChange={(e) => {
            setCategoryFilter(e.target.value);
            setPage(0);
          }}
          className="px-3 py-2 border border-border rounded-lg text-sm bg-card text-foreground"
        >
          <option value="">전체 카테고리</option>
          {Object.entries(CATEGORY_LABELS).map(([key, label]) => (
            <option key={key} value={key}>
              {label}
            </option>
          ))}
        </select>
      </div>

      <div className="space-y-3">
        {loading ? (
          <div className="flex items-center justify-center py-12">
            <div className="animate-spin h-8 w-8 border-4 border-border border-t-foreground rounded-full" />
          </div>
        ) : feedbacks.length === 0 ? (
          <p className="text-center py-12 text-muted-foreground/60">
            피드백이 없습니다.
          </p>
        ) : (
          feedbacks.map((fb) => (
            <div
              key={fb.id}
              className="bg-card border border-border rounded-xl p-4 space-y-3"
            >
              <div className="flex items-start justify-between gap-4">
                <div className="flex items-center gap-2">
                  <span
                    className={`inline-block px-2 py-0.5 rounded text-xs font-medium ${
                      fb.category === "bug"
                        ? "bg-red-500/10 text-red-400"
                        : fb.category === "feature"
                          ? "bg-blue-500/10 text-blue-400"
                          : "bg-white/[0.06] text-foreground/80"
                    }`}
                  >
                    {CATEGORY_LABELS[fb.category] ?? fb.category}
                  </span>
                  <span
                    className={`inline-block px-2 py-0.5 rounded text-xs font-medium ${
                      STATUS_STYLES[fb.admin_status ?? "pending"] ??
                      STATUS_STYLES.pending
                    }`}
                  >
                    {STATUS_LABELS[fb.admin_status ?? "pending"] ?? "대기"}
                  </span>
                </div>
                <span className="text-xs text-muted-foreground whitespace-nowrap">
                  {new Date(fb.created_at).toLocaleString("ko-KR")}
                </span>
              </div>

              <p className="text-sm text-foreground whitespace-pre-wrap">
                {fb.content}
              </p>

              {fb.page_url && (
                <p className="text-xs text-muted-foreground">
                  페이지: {fb.page_url}
                </p>
              )}

              {fb.admin_note && editingId !== fb.id && (
                <div className="bg-white/[0.02] rounded-lg p-3 border border-border/50">
                  <p className="text-xs text-muted-foreground mb-1">
                    관리자 메모
                  </p>
                  <p className="text-sm text-foreground">{fb.admin_note}</p>
                </div>
              )}

              {editingId === fb.id ? (
                <div className="space-y-3 border-t border-border pt-3">
                  <div className="flex gap-3">
                    <select
                      value={editStatus}
                      onChange={(e) => setEditStatus(e.target.value)}
                      className="px-3 py-1.5 border border-border rounded-lg text-sm bg-card text-foreground"
                    >
                      {Object.entries(STATUS_LABELS).map(([key, label]) => (
                        <option key={key} value={key}>
                          {label}
                        </option>
                      ))}
                    </select>
                  </div>
                  <textarea
                    value={editNote}
                    onChange={(e) => setEditNote(e.target.value)}
                    placeholder="관리자 메모 입력..."
                    rows={2}
                    className="w-full px-3 py-2 border border-border rounded-lg text-sm bg-transparent text-foreground focus:outline-none focus:ring-2 focus:ring-primary/50 resize-none"
                  />
                  <div className="flex gap-2">
                    <button
                      onClick={handleSave}
                      className="px-3 py-1.5 text-sm bg-primary text-primary-foreground rounded-lg hover:bg-primary/90"
                    >
                      저장
                    </button>
                    <button
                      onClick={() => setEditingId(null)}
                      className="px-3 py-1.5 text-sm border border-border rounded-lg text-foreground hover:bg-white/[0.04]"
                    >
                      취소
                    </button>
                  </div>
                </div>
              ) : (
                <button
                  onClick={() => handleEdit(fb)}
                  className="text-xs text-primary hover:text-primary/80"
                >
                  상태 변경
                </button>
              )}
            </div>
          ))
        )}
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
