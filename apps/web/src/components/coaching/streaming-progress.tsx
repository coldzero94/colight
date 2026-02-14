"use client";

import { Loader2 } from "lucide-react";

interface StreamingProgressProps {
  charCount: number;
  charLimit: number;
  isStreaming: boolean;
}

export function StreamingProgress({
  charCount,
  charLimit,
  isStreaming,
}: StreamingProgressProps) {
  const percentage = Math.min((charCount / charLimit) * 100, 100);
  const isOverLimit = charCount > charLimit;

  return (
    <div className="rounded-lg border border-gray-200 bg-white p-4">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          {isStreaming && (
            <Loader2 className="h-4 w-4 animate-spin text-blue-600" data-testid="streaming-indicator" />
          )}
          <span className="text-sm font-medium text-gray-700">
            {isStreaming ? "초안 생성 중..." : "작성 완료"}
          </span>
        </div>

        <div
          className={`text-sm font-semibold ${
            isOverLimit ? "text-red-600" : "text-gray-900"
          }`}
        >
          {charCount} / {charLimit}자
          {isOverLimit && " (초과)"}
        </div>
      </div>

      {/* Progress bar */}
      <div className="mt-2 h-2 w-full overflow-hidden rounded-full bg-gray-100">
        <div
          className={`h-full transition-all ${
            isOverLimit ? "bg-red-500" : "bg-blue-500"
          }`}
          style={{ width: `${Math.min(percentage, 100)}%` }}
        />
      </div>
    </div>
  );
}
