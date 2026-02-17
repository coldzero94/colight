"use client";

import { X } from "lucide-react";

interface StarFields {
  star_situation: string;
  star_task: string;
  star_action: string;
  star_result: string;
}

const LABELS: Record<keyof StarFields, string> = {
  star_situation: "Situation",
  star_task: "Task",
  star_action: "Action",
  star_result: "Result",
};

interface StarDiffDialogProps {
  open: boolean;
  onClose: () => void;
  onAccept: () => void;
  before: StarFields | null;
  after: StarFields | null;
}

export function StarDiffDialog({
  open,
  onClose,
  onAccept,
  before,
  after,
}: StarDiffDialogProps) {
  if (!open || !before || !after) return null;

  const fields = Object.keys(LABELS) as (keyof StarFields)[];

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm">
      <div className="relative mx-4 max-h-[85vh] w-full max-w-2xl overflow-y-auto rounded-2xl border border-white/10 bg-[#0d1220] p-6 shadow-2xl">
        <button
          onClick={onClose}
          className="absolute right-4 top-4 text-muted-foreground hover:text-foreground"
        >
          <X className="h-5 w-5" />
        </button>

        <h3 className="mb-1 text-lg font-semibold text-foreground">
          AI STAR 생성 결과
        </h3>
        <p className="mb-5 text-sm text-muted-foreground">
          기존 내용과 AI 생성 결과를 비교하고 적용할 수 있습니다.
        </p>

        <div className="space-y-5">
          {fields.map((key) => {
            const hasChange = before[key] !== after[key];
            if (!hasChange && !before[key] && !after[key]) return null;

            return (
              <div key={key} className="space-y-2">
                <h4 className="text-sm font-medium text-foreground/90">
                  {LABELS[key]}
                  {hasChange && (
                    <span className="ml-2 text-xs text-amber-400">변경됨</span>
                  )}
                </h4>

                {before[key] && hasChange && (
                  <div className="rounded-lg border border-red-500/20 bg-red-500/5 p-3">
                    <p className="mb-1 text-xs font-medium text-red-400">기존</p>
                    <p className="whitespace-pre-wrap text-sm text-foreground/70">
                      {before[key]}
                    </p>
                  </div>
                )}

                <div
                  className={`rounded-lg border p-3 ${
                    hasChange
                      ? "border-emerald-500/20 bg-emerald-500/5"
                      : "border-white/10 bg-white/[0.02]"
                  }`}
                >
                  {hasChange && (
                    <p className="mb-1 text-xs font-medium text-emerald-400">AI 생성</p>
                  )}
                  <p className="whitespace-pre-wrap text-sm text-foreground/80">
                    {after[key] || "(비어있음)"}
                  </p>
                </div>
              </div>
            );
          })}
        </div>

        <div className="mt-6 flex justify-end gap-3">
          <button
            onClick={onClose}
            className="rounded-xl border border-white/10 px-5 py-2 text-sm font-medium text-foreground/70 transition-colors hover:bg-white/5"
          >
            취소
          </button>
          <button
            onClick={onAccept}
            className="rounded-xl bg-violet-600 px-5 py-2 text-sm font-medium text-white shadow-md shadow-violet-600/20 transition-all hover:-translate-y-0.5 hover:bg-violet-500"
          >
            적용하기
          </button>
        </div>
      </div>
    </div>
  );
}
