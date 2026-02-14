"use client";

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
    <div className="rounded-lg border border-gray-200 bg-white p-6">
      <h3 className="mb-4 text-lg font-semibold text-gray-900">
        📊 문항 분석 결과
      </h3>

      <div className="grid grid-cols-1 gap-6 md:grid-cols-2">
        {/* 표면적 질문 */}
        <div className="space-y-2">
          <h4 className="text-sm font-medium text-gray-700">표면적 질문</h4>
          <div className="rounded-lg bg-gray-50 p-4">
            <p className="text-sm text-gray-900">{surface}</p>
          </div>
        </div>

        {/* 진짜 의도 */}
        <div className="space-y-2">
          <h4 className="text-sm font-medium text-gray-700">진짜 의도</h4>
          <div className="space-y-3">
            {intents.map((intent, index) => (
              <div
                key={index}
                data-testid="intent-item"
                className="rounded-lg bg-blue-50 p-3"
              >
                <div className="flex items-start gap-2">
                  <span className="text-blue-600">🎯</span>
                  <div className="flex-1">
                    <p className="text-sm font-medium text-gray-900">
                      {intent.intent}
                    </p>
                    <p className="mt-1 text-xs text-gray-600">
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
