"use client";

import { ArrowRight } from "lucide-react";
import type { SpecificSuggestion } from "@/lib/api/coaching";

interface SuggestionItemProps {
  suggestion: SpecificSuggestion;
  onApply: (suggestion: SpecificSuggestion) => void;
}

export function SuggestionItem({ suggestion, onApply }: SuggestionItemProps) {
  return (
    <div className="rounded-lg border border-border bg-card p-4">
      <div className="mb-2 flex items-start gap-2">
        <div className="flex-1">
          <p className="text-sm text-red-400 line-through">
            {suggestion.original}
          </p>
        </div>
        <ArrowRight className="mt-0.5 h-4 w-4 shrink-0 text-muted-foreground/60" />
        <div className="flex-1">
          <p className="text-sm font-medium text-blue-400">
            {suggestion.suggested}
          </p>
        </div>
      </div>
      <p className="mb-2 text-xs text-muted-foreground">{suggestion.reason}</p>
      <button
        onClick={() => onApply(suggestion)}
        className="rounded-md bg-white/[0.06] px-3 py-1 text-xs font-medium text-foreground/80 hover:bg-white/[0.08]"
      >
        적용하기
      </button>
    </div>
  );
}
