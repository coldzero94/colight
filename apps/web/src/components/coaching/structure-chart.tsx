"use client";

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
    <div className="rounded-lg border border-gray-200 bg-white p-6">
      <h3 className="mb-4 text-lg font-semibold text-gray-900">
        추천 작성 구조 ({totalChars}자 기준)
      </h3>

      <div className="space-y-3">
        {sections.map((section, index) => {
          const percentage = Math.round(section.char_ratio * 100);

          return (
            <div key={index} className="space-y-1">
              <div className="flex items-center justify-between text-sm">
                <span className="font-medium text-gray-700">
                  {section.name}
                </span>
                <span className="text-gray-500">
                  {percentage}% ({section.char_count}자)
                </span>
              </div>

              {/* Progress bar */}
              <div className="h-3 w-full overflow-hidden rounded-full bg-gray-100">
                <div
                  className="h-full bg-gradient-to-r from-blue-500 to-blue-600"
                  style={{ width: `${percentage}%` }}
                />
              </div>

              <p className="text-xs text-gray-500">{section.guide}</p>
            </div>
          );
        })}
      </div>

      <div className="mt-4 rounded-lg bg-blue-50 p-3 text-sm text-blue-800">
        💡 실행 과정에 가장 많은 비중을 할애하세요
      </div>
    </div>
  );
}
