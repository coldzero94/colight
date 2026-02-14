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
  if (score >= 80) return "text-blue-600 bg-blue-50";
  if (score >= 60) return "text-green-600 bg-green-50";
  if (score >= 40) return "text-yellow-600 bg-yellow-50";
  return "text-red-600 bg-red-50";
}

export function DimensionCard({
  feedback,
  defaultExpanded = false,
}: DimensionCardProps) {
  const [expanded, setExpanded] = useState(defaultExpanded);
  const label = DIMENSION_NAMES[feedback.dimension] ?? feedback.dimension;
  const scoreColor = getScoreColor(feedback.score);

  return (
    <div className="rounded-lg border border-gray-200 bg-white">
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
          <span className="text-sm font-medium text-gray-900">{label}</span>
        </div>
        {expanded ? (
          <ChevronUp className="h-4 w-4 text-gray-400" />
        ) : (
          <ChevronDown className="h-4 w-4 text-gray-400" />
        )}
      </button>

      {expanded && (
        <div className="border-t border-gray-100 px-4 pb-4 pt-3">
          {feedback.good.length > 0 && (
            <div className="mb-3">
              <p className="mb-1 text-xs font-medium text-green-700">
                잘한 점
              </p>
              <ul className="space-y-1">
                {feedback.good.map((item, i) => (
                  <li
                    key={i}
                    className="text-sm text-gray-600 before:mr-1 before:content-['·']"
                  >
                    {item}
                  </li>
                ))}
              </ul>
            </div>
          )}

          {feedback.improve.length > 0 && (
            <div>
              <p className="mb-1 text-xs font-medium text-amber-700">
                개선할 점
              </p>
              <ul className="space-y-1">
                {feedback.improve.map((item, i) => (
                  <li
                    key={i}
                    className="text-sm text-gray-600 before:mr-1 before:content-['·']"
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
