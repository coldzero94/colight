"use client";

import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { toast } from "sonner";
import { X, Clock, BarChart3, Zap, Hash, ChevronRight } from "lucide-react";
import { apiClient } from "@/lib/api-client";

// ---------------------------------------------------------------------------
// Types
// ---------------------------------------------------------------------------

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
  avg_latency_ms: number;
  avg_quality_score: number;
  output_schema: Record<string, unknown> | null;
  created_at: string;
  updated_at: string;
}

// ---------------------------------------------------------------------------
// Constants
// ---------------------------------------------------------------------------

const CATEGORY_LABELS: Record<string, string> = {
  experience_classify: "경험 분류",
  coaching_draft: "코칭 초안",
  coaching: "코칭 평가",
  company_analysis: "기업 분석",
  crawling: "크롤링",
  interview: "인터뷰",
  matching: "매칭",
  star_generation: "STAR 생성",
};

const SUB_CATEGORY_LABELS: Record<string, string> = {
  weapon_tagging: "무기 자동 분류",
  interview: "대화형 인터뷰",
  question_analysis: "문항 분석",
  review: "자소서 첨삭",
  weapon_enhance: "무기별 강화",
  analyze: "AI 분석",
  extract_markdown: "마크다운 추출",
  extract_html: "HTML 추출",
  normalize: "정규화",
  generate_question: "질문 생성",
  extract_star: "STAR 추출",
  match_experience: "경험 매칭",
  generate: "자동 생성",
};

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

function relativeTime(dateStr: string): string {
  if (!dateStr) return "-";
  const diff = Date.now() - new Date(dateStr).getTime();
  const mins = Math.floor(diff / 60000);
  if (mins < 1) return "방금 전";
  if (mins < 60) return `${mins}분 전`;
  const hours = Math.floor(mins / 60);
  if (hours < 24) return `${hours}시간 전`;
  const days = Math.floor(hours / 24);
  if (days < 30) return `${days}일 전`;
  return new Date(dateStr).toLocaleDateString("ko-KR");
}

function extractPlaceholders(template: string): string[] {
  const matches = template.match(/\{\{(\w+)\}\}/g);
  if (!matches) return [];
  return [...new Set(matches.map((m) => m.slice(2, -2)))];
}

function hasChanges(a: PromptTemplate, b: PromptTemplate): boolean {
  return (
    a.system_prompt !== b.system_prompt ||
    a.user_prompt_template !== b.user_prompt_template ||
    a.model_name !== b.model_name ||
    a.temperature !== b.temperature ||
    a.max_tokens !== b.max_tokens ||
    a.is_active !== b.is_active
  );
}

// ---------------------------------------------------------------------------
// Component
// ---------------------------------------------------------------------------

export default function AdminPromptsPage() {
  const [prompts, setPrompts] = useState<PromptTemplate[]>([]);
  const [categoryFilter, setCategoryFilter] = useState("");
  const [selected, setSelected] = useState<PromptTemplate | null>(null);
  const [original, setOriginal] = useState<PromptTemplate | null>(null);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const panelRef = useRef<HTMLDivElement>(null);

  // ── Fetch ────────────────────────────────────────────────────────────────
  const fetchPrompts = useCallback(async () => {
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
  }, [categoryFilter]);

  useEffect(() => {
    fetchPrompts();
  }, [fetchPrompts]);

  // ── Select / unsaved guard ───────────────────────────────────────────────
  const handleSelect = (p: PromptTemplate) => {
    if (selected && original && hasChanges(selected, original)) {
      if (!window.confirm("저장하지 않은 변경사항이 있습니다. 이동하시겠습니까?")) return;
    }
    const copy = { ...p };
    setSelected(copy);
    setOriginal({ ...p });
    // scroll panel to top
    panelRef.current?.scrollTo(0, 0);
  };

  const handleClose = () => {
    if (selected && original && hasChanges(selected, original)) {
      if (!window.confirm("저장하지 않은 변경사항이 있습니다. 닫으시겠습니까?")) return;
    }
    setSelected(null);
    setOriginal(null);
  };

  // ── Save ─────────────────────────────────────────────────────────────────
  const handleSave = async () => {
    if (!selected) return;
    setSaving(true);
    try {
      const { data } = await apiClient.put(`/v1/admin/prompts/${selected.id}`, {
        system_prompt: selected.system_prompt,
        user_prompt_template: selected.user_prompt_template,
        model_name: selected.model_name,
        temperature: selected.temperature,
        max_tokens: selected.max_tokens,
        is_active: selected.is_active,
      });
      toast.success("프롬프트가 저장되었습니다.");
      // refresh list + keep panel open with new data
      const updated: PromptTemplate = data;
      setSelected(updated);
      setOriginal({ ...updated });
      fetchPrompts();
    } catch {
      toast.error("프롬프트 저장에 실패했습니다.");
    } finally {
      setSaving(false);
    }
  };

  // ── Derived ──────────────────────────────────────────────────────────────
  const categories = useMemo(
    () => [...new Set(prompts.map((p) => p.category))],
    [prompts],
  );

  const grouped = useMemo(() => {
    const map = new Map<string, PromptTemplate[]>();
    for (const p of prompts) {
      const list = map.get(p.category) ?? [];
      list.push(p);
      map.set(p.category, list);
    }
    return map;
  }, [prompts]);

  const dirty = !!(selected && original && hasChanges(selected, original));

  // ── Render ───────────────────────────────────────────────────────────────
  return (
    <div className="flex h-[calc(100vh-3rem)] gap-0">
      {/* ── Left: List ──────────────────────────────────────────────── */}
      <div
        className={`flex-1 overflow-auto p-6 space-y-6 transition-all duration-200 ${
          selected ? "pr-0" : ""
        }`}
      >
        {/* Header */}
        <div className="flex items-center justify-between">
          <h1 className="text-2xl font-bold font-display text-foreground">
            프롬프트 관리
          </h1>
          <select
            value={categoryFilter}
            onChange={(e) => setCategoryFilter(e.target.value)}
            className="px-3 py-2 border border-border rounded-lg text-sm bg-card text-foreground focus:outline-none focus:ring-2 focus:ring-primary/50"
          >
            <option value="">전체 카테고리</option>
            {categories.map((cat) => (
              <option key={cat} value={cat}>
                {CATEGORY_LABELS[cat] ?? cat}
              </option>
            ))}
          </select>
        </div>

        {/* Loading / empty */}
        {loading ? (
          <div className="flex items-center justify-center py-20 text-muted-foreground/60">
            <div className="h-6 w-6 border-2 border-border border-t-primary rounded-full animate-spin" />
            <span className="ml-3 text-sm">로딩 중...</span>
          </div>
        ) : prompts.length === 0 ? (
          <div className="text-center py-20 text-muted-foreground/60 text-sm">
            프롬프트가 없습니다.
          </div>
        ) : (
          /* Category groups */
          [...grouped.entries()].map(([category, items]) => (
            <div
              key={category}
              className="bg-card border border-border rounded-xl overflow-hidden"
            >
              {/* Group header */}
              <div className="px-4 py-3 bg-white/[0.02] border-b border-border flex items-center gap-2">
                <span className="text-sm font-semibold text-foreground">
                  {CATEGORY_LABELS[category] ?? category}
                </span>
                <span className="text-xs text-muted-foreground/60">
                  {category}
                </span>
                <span className="ml-auto text-xs text-muted-foreground/50">
                  {items.length}개
                </span>
              </div>

              {/* Rows */}
              {items.map((p) => {
                const isSelected = selected?.id === p.id;
                return (
                  <div
                    key={p.id}
                    onClick={() => handleSelect(p)}
                    className={`flex items-center gap-3 px-4 py-3 border-b border-border last:border-b-0 cursor-pointer transition-colors ${
                      isSelected
                        ? "bg-primary/[0.06]"
                        : "hover:bg-white/[0.04]"
                    }`}
                  >
                    {/* Sub-category label */}
                    <div className="flex-1 min-w-0">
                      <p className="text-sm font-medium text-foreground truncate">
                        {SUB_CATEGORY_LABELS[p.sub_category] ?? p.sub_category}
                      </p>
                      <p className="text-xs text-muted-foreground/60 truncate">
                        {p.name}
                      </p>
                    </div>

                    {/* Model */}
                    <span className="hidden sm:inline-block text-xs text-muted-foreground font-mono bg-white/[0.04] px-2 py-0.5 rounded">
                      {p.model_name}
                    </span>

                    {/* Status */}
                    <span
                      className={`inline-block px-2 py-0.5 rounded text-xs font-medium ${
                        p.is_active
                          ? "bg-primary/10 text-primary"
                          : "bg-white/[0.06] text-muted-foreground"
                      }`}
                    >
                      {p.is_active ? "활성" : "비활성"}
                    </span>

                    {/* Usage */}
                    <span className="text-xs text-muted-foreground/60 w-12 text-right tabular-nums">
                      {p.usage_count.toLocaleString()}회
                    </span>

                    {/* Latency */}
                    <span className="hidden md:inline-block text-xs text-muted-foreground/50 w-14 text-right tabular-nums">
                      {p.avg_latency_ms > 0 ? `${p.avg_latency_ms}ms` : "-"}
                    </span>

                    <ChevronRight className="h-4 w-4 text-muted-foreground/30 flex-shrink-0" />
                  </div>
                );
              })}
            </div>
          ))
        )}
      </div>

      {/* ── Right: Edit Panel ───────────────────────────────────────── */}
      {selected && (
        <div
          ref={panelRef}
          className="w-[480px] flex-shrink-0 border-l border-border bg-card overflow-auto"
        >
          <div className="p-5 space-y-5">
            {/* Panel header */}
            <div className="flex items-start justify-between">
              <div>
                <h2 className="text-lg font-bold text-foreground leading-tight">
                  {selected.name}
                </h2>
                <p className="text-xs text-muted-foreground mt-1">
                  {selected.category} / {selected.sub_category}
                </p>
              </div>
              <button
                onClick={handleClose}
                className="p-1.5 rounded-lg hover:bg-white/[0.06] text-muted-foreground"
              >
                <X className="h-4 w-4" />
              </button>
            </div>

            {/* ── AI Settings ─────────────────────────────────────── */}
            <Section title="AI 설정">
              <div className="space-y-3">
                <Field label="모델">
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
                      <option value="llama-3.3-70b-versatile">Groq Llama 3.3 70B</option>
                      <option value="qwen/qwen3-32b">Groq Qwen3 32B</option>
                      <option value="openai/gpt-oss-120b">Groq GPT-OSS 120B</option>
                      <option value="meta-llama/llama-4-scout-17b-16e-instruct">Groq Llama 4 Scout</option>
                      <option value="moonshotai/kimi-k2-instruct">Groq Kimi K2</option>
                      <option value="groq/compound">Groq Compound</option>
                    </optgroup>
                  </select>
                </Field>

                <div className="grid grid-cols-2 gap-3">
                  <Field label="Temperature">
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
                      className="w-full px-3 py-2 border border-border rounded-lg text-sm bg-transparent text-foreground tabular-nums focus:outline-none focus:ring-2 focus:ring-primary/50"
                    />
                  </Field>
                  <Field label="Max Tokens">
                    <input
                      type="number"
                      step="100"
                      min="100"
                      max="16000"
                      value={selected.max_tokens}
                      onChange={(e) =>
                        setSelected({
                          ...selected,
                          max_tokens: parseInt(e.target.value) || 0,
                        })
                      }
                      className="w-full px-3 py-2 border border-border rounded-lg text-sm bg-transparent text-foreground tabular-nums focus:outline-none focus:ring-2 focus:ring-primary/50"
                    />
                  </Field>
                </div>

                <label className="flex items-center gap-2 text-sm text-foreground/80 cursor-pointer">
                  <input
                    type="checkbox"
                    checked={selected.is_active}
                    onChange={(e) =>
                      setSelected({ ...selected, is_active: e.target.checked })
                    }
                    className="rounded border-border"
                  />
                  활성 상태
                </label>
              </div>
            </Section>

            {/* ── System Prompt ────────────────────────────────────── */}
            <Section title="System Prompt">
              <textarea
                value={selected.system_prompt}
                onChange={(e) =>
                  setSelected({ ...selected, system_prompt: e.target.value })
                }
                rows={10}
                className="w-full px-3 py-2 border border-border rounded-lg text-xs font-mono leading-relaxed bg-transparent text-foreground resize-y focus:outline-none focus:ring-2 focus:ring-primary/50"
              />
            </Section>

            {/* ── User Prompt Template ─────────────────────────────── */}
            <Section title="User Prompt Template">
              <textarea
                value={selected.user_prompt_template}
                onChange={(e) =>
                  setSelected({
                    ...selected,
                    user_prompt_template: e.target.value,
                  })
                }
                rows={8}
                className="w-full px-3 py-2 border border-border rounded-lg text-xs font-mono leading-relaxed bg-transparent text-foreground resize-y focus:outline-none focus:ring-2 focus:ring-primary/50"
              />
              {/* Placeholder tags */}
              <div className="flex flex-wrap gap-1.5 mt-2">
                {extractPlaceholders(selected.user_prompt_template).map(
                  (ph) => (
                    <span
                      key={ph}
                      className="inline-block px-2 py-0.5 bg-primary/10 text-primary text-xs rounded font-mono"
                    >
                      {`{{${ph}}}`}
                    </span>
                  ),
                )}
              </div>
            </Section>

            {/* ── Stats ────────────────────────────────────────────── */}
            <Section title="통계">
              <div className="grid grid-cols-2 gap-3">
                <StatItem
                  icon={<Hash className="h-3.5 w-3.5" />}
                  label="사용 횟수"
                  value={`${selected.usage_count.toLocaleString()}회`}
                />
                <StatItem
                  icon={<Clock className="h-3.5 w-3.5" />}
                  label="평균 지연"
                  value={
                    selected.avg_latency_ms > 0
                      ? `${selected.avg_latency_ms}ms`
                      : "측정 전"
                  }
                />
                <StatItem
                  icon={<BarChart3 className="h-3.5 w-3.5" />}
                  label="평균 품질"
                  value={
                    selected.avg_quality_score > 0
                      ? selected.avg_quality_score.toFixed(1)
                      : "측정 전"
                  }
                />
                <StatItem
                  icon={<Zap className="h-3.5 w-3.5" />}
                  label="버전"
                  value={`v${selected.version}`}
                />
              </div>
              <p className="text-xs text-muted-foreground/50 mt-2">
                마지막 수정: {relativeTime(selected.updated_at)}
              </p>
            </Section>

            {/* ── Actions ──────────────────────────────────────────── */}
            <div className="flex justify-end gap-3 pt-2 border-t border-border">
              <button
                onClick={handleClose}
                className="px-4 py-2 text-sm border border-border rounded-lg text-foreground/80 hover:bg-white/[0.04]"
              >
                취소
              </button>
              <button
                onClick={handleSave}
                disabled={saving || !dirty}
                className="px-4 py-2 text-sm bg-primary text-primary-foreground rounded-lg hover:bg-primary/90 disabled:opacity-40 disabled:cursor-not-allowed"
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

// ---------------------------------------------------------------------------
// Sub-components
// ---------------------------------------------------------------------------

function Section({
  title,
  children,
}: {
  title: string;
  children: React.ReactNode;
}) {
  return (
    <div>
      <h3 className="text-xs font-semibold text-muted-foreground uppercase tracking-wider mb-2">
        {title}
      </h3>
      {children}
    </div>
  );
}

function Field({
  label,
  children,
}: {
  label: string;
  children: React.ReactNode;
}) {
  return (
    <div>
      <label className="block text-sm font-medium text-foreground/80 mb-1">
        {label}
      </label>
      {children}
    </div>
  );
}

function StatItem({
  icon,
  label,
  value,
}: {
  icon: React.ReactNode;
  label: string;
  value: string;
}) {
  return (
    <div className="flex items-center gap-2 bg-white/[0.02] rounded-lg px-3 py-2">
      <span className="text-muted-foreground/50">{icon}</span>
      <div>
        <p className="text-xs text-muted-foreground/60">{label}</p>
        <p className="text-sm font-medium text-foreground tabular-nums">
          {value}
        </p>
      </div>
    </div>
  );
}
