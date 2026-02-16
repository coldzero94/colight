"use client";

import { useState } from "react";
import { ChevronDown, ChevronUp } from "lucide-react";
import type { DimensionFeedback } from "@/lib/api/coaching";

const DIMENSION_NAMES: Record<string, string> = {
  specificity: "구체성",
  job_fit: "직무적합성",
  company_fit: "기업적합성",
  authenticity: "진정성",
};

interface DimensionCardProps {
  feedback: DimensionFeedback;
  defaultExpanded?: boolean;
}

function getScoreColor(score: number): string {
  if (score >= 80) return "text-blue-400 bg-blue-500/10";
  if (score >= 60) return "text-green-400 bg-green-500/10";
  if (score >= 40) return "text-yellow-400 bg-yellow-500/10";
  return "text-red-400 bg-red-500/10";
}

export function DimensionCard({
  feedback,
  defaultExpanded = false,
}: DimensionCardProps) {
  const [expanded, setExpanded] = useState(defaultExpanded);
  const label = DIMENSION_NAMES[feedback.dimension] ?? feedback.dimension;
  const scoreColor = getScoreColor(feedback.score);

  return (
    <div className="rounded-lg border border-border bg-card">
      <button
        onClick={() => setExpanded(!expanded)}
        className="flex w-full items-center justify-between p-4"
      >
        <div className="flex items-center gap-3">
          <span
            className={`rounded-md px-2 py-1 text-sm font-bold ${scoreColor}`}
          >
            {feedback.score}
          </span>
          <span className="text-sm font-medium text-foreground">{label}</span>
        </div>
        {expanded ? (
          <ChevronUp className="h-4 w-4 text-muted-foreground/60" />
        ) : (
          <ChevronDown className="h-4 w-4 text-muted-foreground/60" />
        )}
      </button>

      {expanded && (
        <div className="border-t border-border px-4 pb-4 pt-3">
          {feedback.good.length > 0 && (
            <div className="mb-3">
              <p className="mb-1 text-xs font-medium text-green-400">
                잘한 점
              </p>
              <ul className="space-y-1">
                {feedback.good.map((item, i) => (
                  <li
                    key={i}
                    className="text-sm text-muted-foreground before:mr-1 before:content-['·']"
                  >
                    {item}
                  </li>
                ))}
              </ul>
            </div>
          )}

          {feedback.improve.length > 0 && (
            <div>
              <p className="mb-1 text-xs font-medium text-yellow-400">
                개선할 점
              </p>
              <ul className="space-y-1">
                {feedback.improve.map((item, i) => (
                  <li
                    key={i}
                    className="text-sm text-muted-foreground before:mr-1 before:content-['·']"
                  >
                    {item}
                  </li>
                ))}
              </ul>
            </div>
          )}
        </div>
      )}
    </div>
  );
}
