"use client";

import { SuggestionItem } from "./suggestion-item";
import type { SpecificSuggestion } from "@/lib/api/coaching";

interface SuggestionListProps {
  suggestions: SpecificSuggestion[];
  onApply: (suggestion: SpecificSuggestion) => void;
}

export function SuggestionList({ suggestions, onApply }: SuggestionListProps) {
  if (suggestions.length === 0) {
    return (
      <p className="py-4 text-center text-sm text-muted-foreground/60">
        수정 제안이 없습니다
      </p>
    );
  }

  return (
    <div className="space-y-3">
      <h3 className="text-sm font-medium text-foreground">
        수정 제안 ({suggestions.length})
      </h3>
      {suggestions.map((suggestion, i) => (
        <SuggestionItem key={i} suggestion={suggestion} onApply={onApply} />
      ))}
    </div>
  );
}
