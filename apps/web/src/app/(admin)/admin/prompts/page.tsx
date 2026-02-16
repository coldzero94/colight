"use client";

import { useEffect, useState } from "react";
import { toast } from "sonner";
import { apiClient } from "@/lib/api-client";

interface PromptTemplate {
  id: string;
  category: string;
  sub_category: string;
  name: string;
  system_prompt: string;
  user_prompt_template: string;
  model_name: string;
  temperature: number;
  max_tokens: number;
  version: number;
  is_active: boolean;
  usage_count: number;
}

export default function AdminPromptsPage() {
  const [prompts, setPrompts] = useState<PromptTemplate[]>([]);
  const [categoryFilter, setCategoryFilter] = useState("");
  const [selected, setSelected] = useState<PromptTemplate | null>(null);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);

  const fetchPrompts = async () => {
    setLoading(true);
    try {
      const params = categoryFilter ? `?category=${categoryFilter}` : "";
      const { data } = await apiClient.get(`/v1/admin/prompts${params}`);
      setPrompts(data.data);
    } catch {
      toast.error("프롬프트 목록을 불러올 수 없습니다.");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchPrompts();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [categoryFilter]);

  const handleSave = async () => {
    if (!selected) return;
    setSaving(true);
    try {
      await apiClient.put(`/v1/admin/prompts/${selected.id}`, {
        system_prompt: selected.system_prompt,
        user_prompt_template: selected.user_prompt_template,
        model_name: selected.model_name,
        temperature: selected.temperature,
        max_tokens: selected.max_tokens,
        is_active: selected.is_active,
      });
      toast.success("프롬프트가 저장되었습니다.");
      setSelected(null);
      fetchPrompts();
    } catch {
      toast.error("프롬프트 저장에 실패했습니다.");
    } finally {
      setSaving(false);
    }
  };

  const categories = [...new Set(prompts.map((p) => p.category))];

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold text-foreground">프롬프트 관리</h1>
        <select
          value={categoryFilter}
          onChange={(e) => setCategoryFilter(e.target.value)}
          className="px-3 py-2 border border-border rounded-lg text-sm bg-card text-foreground focus:outline-none focus:ring-2 focus:ring-primary/50"
        >
          <option value="">전체 카테고리</option>
          {categories.map((cat) => (
            <option key={cat} value={cat}>
              {cat}
            </option>
          ))}
        </select>
      </div>

      <div className="bg-card border border-border rounded-xl overflow-hidden">
        <table className="w-full text-sm">
          <thead className="bg-white/[0.02] border-b border-border">
            <tr>
              <th className="text-left px-4 py-3 font-medium text-muted-foreground">
                카테고리
              </th>
              <th className="text-left px-4 py-3 font-medium text-muted-foreground">
                서브카테고리
              </th>
              <th className="text-left px-4 py-3 font-medium text-muted-foreground">
                이름
              </th>
              <th className="text-left px-4 py-3 font-medium text-muted-foreground">
                모델
              </th>
              <th className="text-left px-4 py-3 font-medium text-muted-foreground">
                상태
              </th>
              <th className="text-left px-4 py-3 font-medium text-muted-foreground">
                사용횟수
              </th>
            </tr>
          </thead>
          <tbody>
            {loading ? (
              <tr>
                <td colSpan={6} className="text-center py-8 text-muted-foreground/60">
                  로딩 중...
                </td>
              </tr>
            ) : prompts.length === 0 ? (
              <tr>
                <td colSpan={6} className="text-center py-8 text-muted-foreground/60">
                  프롬프트가 없습니다.
                </td>
              </tr>
            ) : (
              prompts.map((p) => (
                <tr
                  key={p.id}
                  onClick={() => setSelected({ ...p })}
                  className="border-b border-border hover:bg-white/[0.04] cursor-pointer"
                >
                  <td className="px-4 py-3">{p.category}</td>
                  <td className="px-4 py-3">{p.sub_category}</td>
                  <td className="px-4 py-3 font-medium">{p.name}</td>
                  <td className="px-4 py-3 text-muted-foreground">{p.model_name}</td>
                  <td className="px-4 py-3">
                    <span
                      className={`inline-block px-2 py-0.5 rounded text-xs font-medium ${
                        p.is_active
                          ? "bg-primary/10 text-primary"
                          : "bg-white/[0.06] text-muted-foreground"
                      }`}
                    >
                      {p.is_active ? "활성" : "비활성"}
                    </span>
                  </td>
                  <td className="px-4 py-3 text-muted-foreground">
                    {p.usage_count.toLocaleString()}
                  </td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>

      {/* Edit panel */}
      {selected && (
        <div className="fixed inset-0 bg-black/30 flex items-center justify-center z-50">
          <div className="bg-card rounded-2xl shadow-xl border border-border w-full max-w-2xl max-h-[90vh] overflow-auto p-6 space-y-4">
            <h2 className="text-lg font-bold text-foreground">
              프롬프트 편집: {selected.name}
            </h2>

            <div>
              <label className="block text-sm font-medium text-foreground/80 mb-1">
                System Prompt
              </label>
              <textarea
                value={selected.system_prompt}
                onChange={(e) =>
                  setSelected({ ...selected, system_prompt: e.target.value })
                }
                rows={6}
                className="w-full px-3 py-2 border border-border rounded-lg text-sm font-mono bg-transparent text-foreground focus:outline-none focus:ring-2 focus:ring-primary/50"
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-foreground/80 mb-1">
                User Prompt Template
              </label>
              <textarea
                value={selected.user_prompt_template}
                onChange={(e) =>
                  setSelected({
                    ...selected,
                    user_prompt_template: e.target.value,
                  })
                }
                rows={4}
                className="w-full px-3 py-2 border border-border rounded-lg text-sm font-mono bg-transparent text-foreground focus:outline-none focus:ring-2 focus:ring-primary/50"
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-foreground/80 mb-1">
                Model
              </label>
              <select
                value={selected.model_name}
                onChange={(e) =>
                  setSelected({ ...selected, model_name: e.target.value })
                }
                className="w-full px-3 py-2 border border-border rounded-lg text-sm bg-card text-foreground focus:outline-none focus:ring-2 focus:ring-primary/50"
              >
                <optgroup label="Heavy Models">
                  <option value="claude-sonnet-4-5">Claude Sonnet 4.5</option>
                  <option value="claude-opus-4">Claude Opus 4</option>
                </optgroup>
                <optgroup label="Light Models">
                  <option value="gemini-2.0-flash">Gemini 2.0 Flash</option>
                  <option value="groq/llama-3.3-70b">Groq Llama 3.3 70B</option>
                </optgroup>
              </select>
            </div>

            <div className="grid grid-cols-3 gap-4">
              <div>
                <label className="block text-sm font-medium text-foreground/80 mb-1">
                  Temperature
                </label>
                <input
                  type="number"
                  step="0.1"
                  min="0"
                  max="2"
                  value={selected.temperature}
                  onChange={(e) =>
                    setSelected({
                      ...selected,
                      temperature: parseFloat(e.target.value) || 0,
                    })
                  }
                  className="w-full px-3 py-2 border border-border rounded-lg text-sm bg-transparent text-foreground focus:outline-none focus:ring-2 focus:ring-primary/50"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-foreground/80 mb-1">
                  Max Tokens
                </label>
                <input
                  type="number"
                  value={selected.max_tokens}
                  onChange={(e) =>
                    setSelected({
                      ...selected,
                      max_tokens: parseInt(e.target.value) || 0,
                    })
                  }
                  className="w-full px-3 py-2 border border-border rounded-lg text-sm bg-transparent text-foreground focus:outline-none focus:ring-2 focus:ring-primary/50"
                />
              </div>
              <div className="flex items-end">
                <label className="flex items-center gap-2 text-sm">
                  <input
                    type="checkbox"
                    checked={selected.is_active}
                    onChange={(e) =>
                      setSelected({ ...selected, is_active: e.target.checked })
                    }
                    className="rounded"
                  />
                  Active
                </label>
              </div>
            </div>

            <div className="flex justify-end gap-3 pt-2">
              <button
                onClick={() => setSelected(null)}
                className="px-4 py-2 text-sm border border-border rounded-lg text-foreground/80 hover:bg-white/[0.04]"
              >
                취소
              </button>
              <button
                onClick={handleSave}
                disabled={saving}
                className="px-4 py-2 text-sm bg-primary text-primary-foreground rounded-lg hover:bg-primary/90 disabled:opacity-50"
              >
                {saving ? "저장 중..." : "저장"}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
