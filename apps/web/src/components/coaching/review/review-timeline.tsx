"use client";

import type { ReviewScores } from "@/lib/api/coaching";

interface ReviewEntry {
  reviewNumber: number;
  overall: number;
  scores: ReviewScores;
  timestamp?: string;
}

interface ReviewTimelineProps {
  entries: ReviewEntry[];
}

function getGradeColor(score: number): string {
  if (score >= 90) return "bg-purple-500";
  if (score >= 80) return "bg-blue-500";
  if (score >= 70) return "bg-green-500";
  if (score >= 60) return "bg-yellow-500";
  return "bg-red-500";
}

export type { ReviewEntry };

export function ReviewTimeline({ entries }: ReviewTimelineProps) {
  if (entries.length === 0) return null;

  return (
    <div className="space-y-2">
      <h3 className="text-sm font-medium text-foreground">첨삭 이력</h3>
      <div className="space-y-2">
        {entries.map((entry, i) => {
          const prev = i > 0 ? entries[i - 1] : null;
          const diff = prev ? entry.overall - prev.overall : null;

          return (
            <div
              key={entry.reviewNumber}
              className="flex items-center gap-3 rounded-lg border border-border bg-card px-4 py-3"
            >
              <div
                className={`flex h-8 w-8 shrink-0 items-center justify-center rounded-full text-xs font-bold text-white ${getGradeColor(entry.overall)}`}
              >
                {entry.overall}
              </div>
              <div className="flex-1">
                <span className="text-sm font-medium text-foreground">
                  {entry.reviewNumber}차 첨삭
                </span>
              </div>
              {diff != null && (
                <span
                  className={`text-xs font-medium ${
                    diff > 0
                      ? "text-green-400"
                      : diff < 0
                        ? "text-red-400"
                        : "text-muted-foreground/60"
                  }`}
                >
                  {diff > 0 ? `+${diff}` : diff === 0 ? "-" : diff}
                </span>
              )}
            </div>
          );
        })}
      </div>
    </div>
  );
}
