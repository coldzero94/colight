"use client";

import { useEffect, useState } from "react";
import { toast } from "sonner";
import { apiClient } from "@/lib/api-client";
import { AdminGuard } from "@/components/auth/admin-guard";

interface SystemConfig {
  id: string;
  config_key: string;
  config_value: string;
  category: string;
  description?: string;
  is_secret: boolean;
  updated_by?: string;
}

const CATEGORY_LABELS: Record<string, string> = {
  api_key: "API 키",
  model: "모델 설정",
  limit: "제한 설정",
  cost: "비용 설정",
};

export default function AdminSettingsPage() {
  const [configs, setConfigs] = useState<SystemConfig[]>([]);
  const [categoryFilter, setCategoryFilter] = useState("");
  const [loading, setLoading] = useState(true);
  const [editKey, setEditKey] = useState<string | null>(null);
  const [editValue, setEditValue] = useState("");
  const [saving, setSaving] = useState(false);

  const fetchConfigs = async () => {
    setLoading(true);
    try {
      const params = categoryFilter ? `?category=${categoryFilter}` : "";
      const { data } = await apiClient.get(`/v1/admin/configs${params}`);
      setConfigs(data.data);
    } catch {
      toast.error("설정 목록을 불러올 수 없습니다.");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchConfigs();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [categoryFilter]);

  const handleEdit = (config: SystemConfig) => {
    setEditKey(config.config_key);
    setEditValue(config.is_secret ? "" : config.config_value);
  };

  const handleSave = async () => {
    if (!editKey || !editValue) return;
    setSaving(true);
    try {
      await apiClient.put(`/v1/admin/configs/${editKey}`, {
        value: editValue,
      });
      toast.success("설정이 저장되었습니다.");
      setEditKey(null);
      setEditValue("");
      fetchConfigs();
    } catch {
      toast.error("설정 저장에 실패했습니다.");
    } finally {
      setSaving(false);
    }
  };

  const categories = [...new Set(configs.map((c) => c.category))];

  return (
    <AdminGuard requiredRole="super_admin">
      <div className="space-y-6">
        <div className="flex items-center justify-between">
          <h1 className="text-2xl font-bold font-display text-foreground">시스템 설정</h1>
          <select
            value={categoryFilter}
            onChange={(e) => setCategoryFilter(e.target.value)}
            className="px-3 py-2 border border-border rounded-lg text-sm bg-card text-foreground focus:outline-none focus:ring-2 focus:ring-primary/50"
          >
            <option value="">전체 카테고리</option>
            {categories.map((cat) => (
              <option key={cat} value={cat}>
                {CATEGORY_LABELS[cat] || cat}
              </option>
            ))}
          </select>
        </div>

        <div className="bg-card border border-border rounded-xl overflow-hidden">
          <table className="w-full text-sm">
            <thead className="bg-white/[0.02] border-b border-border">
              <tr>
                <th className="text-left px-4 py-3 font-medium text-muted-foreground">
                  설정 키
                </th>
                <th className="text-left px-4 py-3 font-medium text-muted-foreground">
                  카테고리
                </th>
                <th className="text-left px-4 py-3 font-medium text-muted-foreground">
                  값
                </th>
                <th className="text-left px-4 py-3 font-medium text-muted-foreground">
                  설명
                </th>
                <th className="text-left px-4 py-3 font-medium text-muted-foreground">
                  작업
                </th>
              </tr>
            </thead>
            <tbody>
              {loading ? (
                <tr>
                  <td colSpan={5} className="text-center py-8 text-muted-foreground/60">
                    로딩 중...
                  </td>
                </tr>
              ) : configs.length === 0 ? (
                <tr>
                  <td colSpan={5} className="text-center py-8 text-muted-foreground/60">
                    설정이 없습니다.
                  </td>
                </tr>
              ) : (
                configs.map((cfg) => (
                  <tr
                    key={cfg.id}
                    className="border-b border-border hover:bg-white/[0.04]"
                  >
                    <td className="px-4 py-3 font-mono text-xs">
                      {cfg.config_key}
                    </td>
                    <td className="px-4 py-3">
                      <span className="inline-block px-2 py-0.5 rounded text-xs font-medium bg-blue-500/10 text-blue-400">
                        {CATEGORY_LABELS[cfg.category] || cfg.category}
                      </span>
                    </td>
                    <td className="px-4 py-3 font-mono text-xs text-muted-foreground">
                      {cfg.is_secret ? (
                        <span className="text-amber-600">
                          {cfg.config_value}
                        </span>
                      ) : (
                        cfg.config_value
                      )}
                    </td>
                    <td className="px-4 py-3 text-muted-foreground text-xs">
                      {cfg.description || "-"}
                    </td>
                    <td className="px-4 py-3">
                      <button
                        onClick={() => handleEdit(cfg)}
                        className="text-xs text-blue-400 hover:text-blue-300 font-medium"
                      >
                        수정
                      </button>
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>

        {/* Edit modal */}
        {editKey && (
          <div className="fixed inset-0 bg-black/30 flex items-center justify-center z-50">
            <div className="bg-card rounded-2xl shadow-xl border border-border w-full max-w-md p-6 space-y-4">
              <h2 className="text-lg font-bold text-foreground">
                설정 수정: {editKey}
              </h2>

              <div>
                <label className="block text-sm font-medium text-foreground/80 mb-1">
                  새 값
                </label>
                <input
                  type="text"
                  value={editValue}
                  onChange={(e) => setEditValue(e.target.value)}
                  placeholder="새 값을 입력하세요"
                  className="w-full px-3 py-2 border border-border rounded-lg text-sm font-mono bg-transparent text-foreground focus:outline-none focus:ring-2 focus:ring-primary/50"
                />
              </div>

              <div className="flex justify-end gap-3 pt-2">
                <button
                  onClick={() => {
                    setEditKey(null);
                    setEditValue("");
                  }}
                  className="px-4 py-2 text-sm border border-border rounded-lg text-foreground/80 hover:bg-white/[0.04]"
                >
                  취소
                </button>
                <button
                  onClick={handleSave}
                  disabled={saving || !editValue}
                  className="px-4 py-2 text-sm bg-primary text-primary-foreground rounded-lg hover:bg-primary/90 disabled:opacity-50"
                >
                  {saving ? "저장 중..." : "저장"}
                </button>
              </div>
            </div>
          </div>
        )}
      </div>
    </AdminGuard>
  );
}
