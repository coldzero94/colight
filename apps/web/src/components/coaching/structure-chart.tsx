"use client";

import { Lightbulb } from "lucide-react";

interface Section {
  name: string;
  char_ratio: number;
  char_count: number;
  guide: string;
}

interface StructureChartProps {
  sections: Section[];
  totalChars: number;
}

export function StructureChart({ sections, totalChars }: StructureChartProps) {
  return (
    <div className="rounded-lg border border-border bg-card p-6">
      <h3 className="mb-4 text-lg font-semibold text-foreground">
        추천 작성 구조 ({totalChars}자 기준)
      </h3>

      <div className="space-y-3">
        {sections.map((section, index) => {
          const percentage = Math.round(section.char_ratio * 100);

          return (
            <div key={index} className="space-y-1">
              <div className="flex items-center justify-between text-sm">
                <span className="font-medium text-foreground/80">
                  {section.name}
                </span>
                <span className="text-muted-foreground">
                  {percentage}% ({section.char_count}자)
                </span>
              </div>

              {/* Progress bar */}
              <div className="h-3 w-full overflow-hidden rounded-full bg-white/[0.06]">
                <div
                  className="h-full bg-gradient-to-r from-blue-500 to-blue-600"
                  style={{ width: `${percentage}%` }}
                />
              </div>

              <p className="text-xs text-muted-foreground">{section.guide}</p>
            </div>
          );
        })}
      </div>

      <div className="mt-4 flex items-center gap-2 rounded-lg bg-blue-500/10 p-3 text-sm text-blue-300">
        <Lightbulb className="h-4 w-4" aria-hidden="true" />
        <span>실행 과정에 가장 많은 비중을 할애하세요</span>
      </div>
    </div>
  );
}
