"use client";

import { ArrowDown, ArrowUp, Check } from "lucide-react";
import type {
  CharCoachingResult as CharCoachingResultType,
  CharCoachingSuggestion,
} from "@/lib/api/coaching";

interface CharCoachingResultProps {
  result: CharCoachingResultType;
  onApplySuggestion: (suggestion: CharCoachingSuggestion) => void;
}

const STATUS_CONFIG = {
  over: {
    label: "초과",
    color: "bg-red-500/10 text-red-400",
    icon: ArrowUp,
  },
  under: {
    label: "부족",
    color: "bg-yellow-500/10 text-yellow-400",
    icon: ArrowDown,
  },
  good: {
    label: "적정",
    color: "bg-primary/10 text-primary",
    icon: Check,
  },
} as const;

export function CharCoachingResult({
  result,
  onApplySuggestion,
}: CharCoachingResultProps) {
  const config = STATUS_CONFIG[result.status];
  const StatusIcon = config.icon;

  return (
    <div className="space-y-4">
      {/* Status badge + counts */}
      <div className="flex items-center justify-between">
        <span
          className={`inline-flex items-center gap-1 rounded-full px-3 py-1 text-sm font-medium ${config.color}`}
        >
          <StatusIcon className="h-3.5 w-3.5" />
          {config.label}
        </span>
        <div className="text-sm text-muted-foreground">
          <span className="font-medium">{result.current_count}</span>
          <span className="text-muted-foreground/60"> / </span>
          <span>{result.char_limit}자</span>
          {result.diff !== 0 && (
            <span
              className={
                result.diff > 0 ? "ml-1 text-red-400" : "ml-1 text-yellow-400"
              }
            >
              ({result.diff > 0 ? "+" : ""}
              {result.diff})
            </span>
          )}
        </div>
      </div>

      {/* Summary */}
      <p className="rounded-lg bg-white/[0.02] p-3 text-sm text-foreground/80">
        {result.summary}
      </p>

      {/* Suggestions */}
      {result.suggestions.length > 0 && (
        <div className="space-y-3">
          <h4 className="text-sm font-medium text-foreground">
            수정 제안 ({result.suggestions.length}건)
          </h4>
          {result.suggestions.map((suggestion, idx) => (
            <div
              key={idx}
              className="rounded-lg border border-border p-3 text-sm"
            >
              <div className="mb-2 flex items-center justify-between">
                <span className="text-xs text-muted-foreground">
                  {suggestion.section}
                </span>
                <span
                  className={`text-xs font-medium ${
                    suggestion.type === "trim"
                      ? "text-red-400"
                      : "text-blue-400"
                  }`}
                >
                  {suggestion.type === "trim" ? "축약" : "보강"}{" "}
                  {suggestion.char_diff > 0 ? "+" : ""}
                  {suggestion.char_diff}자
                </span>
              </div>

              <div className="mb-2 space-y-1">
                <div className="rounded bg-red-500/10 px-2 py-1 text-foreground/80 line-through">
                  {suggestion.original}
                </div>
                <div className="rounded bg-primary/10 px-2 py-1 text-foreground">
                  {suggestion.suggested}
                </div>
              </div>

              <p className="mb-2 text-xs text-muted-foreground">{suggestion.reason}</p>

              <button
                onClick={() => onApplySuggestion(suggestion)}
                className="rounded border border-border px-3 py-1 text-xs font-medium text-foreground/80 hover:bg-white/[0.04]"
              >
                적용
              </button>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
