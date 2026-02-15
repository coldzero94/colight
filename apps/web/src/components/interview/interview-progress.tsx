"use client";

import { Check } from "lucide-react";
import { STAGES, STAGE_LABELS, type InterviewStage } from "@/lib/api/interview";

interface InterviewProgressProps {
  currentStage: InterviewStage;
  isComplete: boolean;
}

export function InterviewProgress({
  currentStage,
  isComplete,
}: InterviewProgressProps) {
  const currentIdx = STAGES.indexOf(currentStage);

  return (
    <div className="flex items-center gap-1">
      {STAGES.map((stage, idx) => {
        const isDone = isComplete || idx < currentIdx;
        const isCurrent = !isComplete && idx === currentIdx;

        return (
          <div key={stage} className="flex items-center">
            {idx > 0 && (
              <div
                className={`w-6 h-0.5 ${isDone ? "bg-gray-900" : "bg-gray-200"}`}
              />
            )}
            <div className="flex flex-col items-center gap-1">
              <div
                className={`w-7 h-7 rounded-full flex items-center justify-center text-xs font-medium ${
                  isDone
                    ? "bg-gray-900 text-white"
                    : isCurrent
                      ? "border-2 border-gray-900 text-gray-900"
                      : "border border-gray-300 text-gray-400"
                }`}
              >
                {isDone ? <Check className="h-3.5 w-3.5" /> : idx + 1}
              </div>
              <span
                className={`text-[10px] whitespace-nowrap ${
                  isDone || isCurrent ? "text-gray-900 font-medium" : "text-gray-400"
                }`}
              >
                {STAGE_LABELS[stage]}
              </span>
            </div>
          </div>
        );
      })}
    </div>
  );
}
