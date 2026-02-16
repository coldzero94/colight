"use client";

import { useState } from "react";
import { Loader2, Save, RefreshCw } from "lucide-react";
import { Button } from "@/components/ui/button";
import type { ExtractSTARResult } from "@/lib/api/interview";

const CATEGORY_OPTIONS = [
  { value: "project", label: "프로젝트" },
  { value: "work", label: "업무경험" },
  { value: "activity", label: "대외활동" },
  { value: "competition", label: "대회/공모전" },
  { value: "education", label: "교육/학습" },
  { value: "volunteer", label: "봉사활동" },
  { value: "other", label: "기타" },
];

interface STARPreviewProps {
  data: ExtractSTARResult;
  isSaving: boolean;
  onSave: (data: ExtractSTARResult) => void;
  onReExtract: () => void;
}

export function STARPreview({
  data,
  isSaving,
  onSave,
  onReExtract,
}: STARPreviewProps) {
  const [form, setForm] = useState<ExtractSTARResult>(data);

  const updateField = (field: keyof ExtractSTARResult, value: string) => {
    setForm((prev) => ({ ...prev, [field]: value }));
  };

  return (
    <div className="space-y-6">
      <div className="rounded-lg border border-border bg-card p-6 space-y-4">
        <h3 className="text-lg font-semibold text-foreground">
          STAR 경험 카드 미리보기
        </h3>

        {/* Title */}
        <div className="space-y-1.5">
          <label className="text-sm font-medium text-foreground/80">제목</label>
          <input
            type="text"
            value={form.title}
            onChange={(e) => updateField("title", e.target.value)}
            className="w-full rounded-lg border border-border bg-background px-3 py-2 text-sm text-foreground focus:outline-none focus:ring-2 focus:ring-primary/40"
          />
        </div>

        {/* Category */}
        <div className="space-y-1.5">
          <label className="text-sm font-medium text-foreground/80">카테고리</label>
          <select
            value={form.category}
            onChange={(e) => updateField("category", e.target.value)}
            className="w-full rounded-lg border border-border bg-background px-3 py-2 text-sm text-foreground focus:outline-none focus:ring-2 focus:ring-primary/40"
          >
            {CATEGORY_OPTIONS.map((opt) => (
              <option key={opt.value} value={opt.value}>
                {opt.label}
              </option>
            ))}
          </select>
        </div>

        {/* STAR fields */}
        {(
          [
            ["star_situation", "상황 (Situation)"],
            ["star_task", "과제 (Task)"],
            ["star_action", "행동 (Action)"],
            ["star_result", "결과 (Result)"],
          ] as const
        ).map(([field, label]) => (
          <div key={field} className="space-y-1.5">
            <label className="text-sm font-medium text-foreground/80">
              {label}
            </label>
            <textarea
              value={form[field]}
              onChange={(e) => updateField(field, e.target.value)}
              rows={3}
              className="w-full rounded-lg border border-border bg-background px-3 py-2 text-sm text-foreground focus:outline-none focus:ring-2 focus:ring-primary/40 resize-none"
            />
          </div>
        ))}

        {/* Keywords */}
        <div className="space-y-1.5">
          <label className="text-sm font-medium text-foreground/80">키워드</label>
          <div className="flex flex-wrap gap-2">
            {form.keywords.map((kw, i) => (
              <span
                key={i}
                className="rounded-full bg-white/[0.06] px-3 py-1 text-xs text-foreground/80"
              >
                {kw}
              </span>
            ))}
          </div>
        </div>
      </div>

      {/* Actions */}
      <div className="flex gap-3">
        <Button
          variant="outline"
          onClick={onReExtract}
          disabled={isSaving}
          className="gap-2"
        >
          <RefreshCw className="h-4 w-4" />
          다시 추출
        </Button>
        <Button
          onClick={() => onSave(form)}
          disabled={isSaving}
          className="gap-2"
        >
          {isSaving ? (
            <>
              <Loader2 className="h-4 w-4 animate-spin" />
              저장 중...
            </>
          ) : (
            <>
              <Save className="h-4 w-4" />
              저장하기
            </>
          )}
        </Button>
      </div>
    </div>
  );
}
