"use client";

import type { ReviewScores } from "@/lib/api/coaching";

interface ScoreComparisonProps {
  current: ReviewScores;
  previous: ReviewScores;
}

const DIMENSIONS: { key: keyof ReviewScores; label: string }[] = [
  { key: "specificity", label: "구체성" },
  { key: "job_fit", label: "직무 적합성" },
  { key: "company_fit", label: "기업 적합성" },
  { key: "authenticity", label: "진정성" },
];

function DiffBadge({ diff }: { diff: number }) {
  if (diff > 0) {
    return (
      <span className="rounded-full bg-green-100 px-2 py-0.5 text-xs font-medium text-green-700">
        +{diff}
      </span>
    );
  }
  if (diff < 0) {
    return (
      <span className="rounded-full bg-red-100 px-2 py-0.5 text-xs font-medium text-red-700">
        {diff}
      </span>
    );
  }
  return (
    <span className="rounded-full bg-gray-100 px-2 py-0.5 text-xs font-medium text-gray-500">
      -
    </span>
  );
}

export function ScoreComparison({ current, previous }: ScoreComparisonProps) {
  return (
    <div className="space-y-2">
      <h3 className="text-sm font-medium text-gray-900">점수 변화</h3>
      <div className="rounded-lg border border-gray-200 bg-white">
        {DIMENSIONS.map(({ key, label }, i) => {
          const diff = current[key] - previous[key];
          return (
            <div
              key={key}
              className={`flex items-center justify-between px-4 py-3 ${
                i < DIMENSIONS.length - 1 ? "border-b border-gray-100" : ""
              }`}
            >
              <span className="text-sm text-gray-700">{label}</span>
              <div className="flex items-center gap-3">
                <span className="text-xs text-gray-400">{previous[key]}</span>
                <span className="text-xs text-gray-400">&rarr;</span>
                <span className="text-sm font-medium text-gray-900">
                  {current[key]}
                </span>
                <DiffBadge diff={diff} />
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
}
