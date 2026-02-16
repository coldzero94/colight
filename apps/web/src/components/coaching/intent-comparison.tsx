"use client";

import { BarChart3, Target } from "lucide-react";

interface RealIntent {
  intent: string;
  why: string;
}

interface IntentComparisonProps {
  surface: string;
  intents: RealIntent[];
}

export function IntentComparison({ surface, intents }: IntentComparisonProps) {
  return (
    <div className="rounded-lg border border-border bg-card p-6">
      <h3 className="mb-4 flex items-center gap-2 text-lg font-semibold text-foreground">
        <span className="inline-flex h-7 w-7 items-center justify-center rounded-lg bg-blue-500/15 text-blue-300">
          <BarChart3 className="h-4 w-4" aria-hidden="true" />
        </span>
        문항 분석 결과
      </h3>

      <div className="grid grid-cols-1 gap-6 md:grid-cols-2">
        {/* 표면적 질문 */}
        <div className="space-y-2">
          <h4 className="text-sm font-medium text-foreground/80">표면적 질문</h4>
          <div className="rounded-lg bg-white/[0.02] p-4">
            <p className="text-sm text-foreground">{surface}</p>
          </div>
        </div>

        {/* 진짜 의도 */}
        <div className="space-y-2">
          <h4 className="text-sm font-medium text-foreground/80">진짜 의도</h4>
          <div className="space-y-3">
            {intents.map((intent, index) => (
              <div
                key={index}
                data-testid="intent-item"
                className="rounded-lg bg-blue-500/10 p-3"
              >
                <div className="flex items-start gap-2">
                  <span className="inline-flex h-6 w-6 items-center justify-center rounded-md bg-blue-500/20 text-blue-300">
                    <Target className="h-3.5 w-3.5" aria-hidden="true" />
                  </span>
                  <div className="flex-1">
                    <p className="text-sm font-medium text-foreground">
                      {intent.intent}
                    </p>
                    <p className="mt-1 text-xs text-muted-foreground">
                      → {intent.why}
                    </p>
                  </div>
                </div>
              </div>
            ))}
          </div>
        </div>
      </div>
    </div>
  );
}
