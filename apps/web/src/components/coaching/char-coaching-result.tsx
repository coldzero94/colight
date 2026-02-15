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
    color: "bg-red-100 text-red-700",
    icon: ArrowUp,
  },
  under: {
    label: "부족",
    color: "bg-amber-100 text-amber-700",
    icon: ArrowDown,
  },
  good: {
    label: "적정",
    color: "bg-green-100 text-green-700",
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
        <div className="text-sm text-gray-600">
          <span className="font-medium">{result.current_count}</span>
          <span className="text-gray-400"> / </span>
          <span>{result.char_limit}자</span>
          {result.diff !== 0 && (
            <span
              className={
                result.diff > 0 ? "ml-1 text-red-600" : "ml-1 text-amber-600"
              }
            >
              ({result.diff > 0 ? "+" : ""}
              {result.diff})
            </span>
          )}
        </div>
      </div>

      {/* Summary */}
      <p className="rounded-lg bg-gray-50 p-3 text-sm text-gray-700">
        {result.summary}
      </p>

      {/* Suggestions */}
      {result.suggestions.length > 0 && (
        <div className="space-y-3">
          <h4 className="text-sm font-medium text-gray-900">
            수정 제안 ({result.suggestions.length}건)
          </h4>
          {result.suggestions.map((suggestion, idx) => (
            <div
              key={idx}
              className="rounded-lg border border-gray-200 p-3 text-sm"
            >
              <div className="mb-2 flex items-center justify-between">
                <span className="text-xs text-gray-500">
                  {suggestion.section}
                </span>
                <span
                  className={`text-xs font-medium ${
                    suggestion.type === "trim"
                      ? "text-red-600"
                      : "text-blue-600"
                  }`}
                >
                  {suggestion.type === "trim" ? "축약" : "보강"}{" "}
                  {suggestion.char_diff > 0 ? "+" : ""}
                  {suggestion.char_diff}자
                </span>
              </div>

              <div className="mb-2 space-y-1">
                <div className="rounded bg-red-50 px-2 py-1 text-gray-700 line-through">
                  {suggestion.original}
                </div>
                <div className="rounded bg-green-50 px-2 py-1 text-gray-900">
                  {suggestion.suggested}
                </div>
              </div>

              <p className="mb-2 text-xs text-gray-500">{suggestion.reason}</p>

              <button
                onClick={() => onApplySuggestion(suggestion)}
                className="rounded border border-gray-300 px-3 py-1 text-xs font-medium text-gray-700 hover:bg-gray-50"
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
