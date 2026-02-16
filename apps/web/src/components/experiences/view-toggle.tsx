"use client";

import { LayoutGrid, Rows3 } from "lucide-react";

interface ViewToggleProps {
  mode: "grid" | "list";
  onChange: (mode: "grid" | "list") => void;
}

export function ViewToggle({ mode, onChange }: ViewToggleProps) {
  return (
    <div
      className="inline-flex rounded-xl border border-white/14 bg-white/[0.03] p-1"
      role="group"
      aria-label="뷰 전환"
    >
      <button
        type="button"
        onClick={() => onChange("grid")}
        className={`inline-flex items-center gap-1.5 rounded-lg px-3 py-1.5 text-sm font-medium transition-all ${
          mode === "grid"
            ? "bg-primary text-primary-foreground shadow-[0_6px_18px_rgba(72,132,255,0.24)]"
            : "text-muted-foreground hover:bg-white/[0.06] hover:text-foreground"
        }`}
        aria-pressed={mode === "grid"}
        aria-label="그리드 뷰"
      >
        <LayoutGrid className="h-3.5 w-3.5" aria-hidden="true" />
        그리드
      </button>
      <button
        type="button"
        onClick={() => onChange("list")}
        className={`inline-flex items-center gap-1.5 rounded-lg px-3 py-1.5 text-sm font-medium transition-all ${
          mode === "list"
            ? "bg-primary text-primary-foreground shadow-[0_6px_18px_rgba(72,132,255,0.24)]"
            : "text-muted-foreground hover:bg-white/[0.06] hover:text-foreground"
        }`}
        aria-pressed={mode === "list"}
        aria-label="리스트 뷰"
      >
        <Rows3 className="h-3.5 w-3.5" aria-hidden="true" />
        리스트
      </button>
    </div>
  );
}
