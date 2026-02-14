"use client";

interface KeywordTagsProps {
  keywords: string[];
  avoidList: string[];
}

export function KeywordTags({ keywords, avoidList }: KeywordTagsProps) {
  return (
    <div className="rounded-lg border border-gray-200 bg-white p-6">
      {/* 핵심 키워드 */}
      <div className="mb-6">
        <h3 className="mb-3 text-lg font-semibold text-gray-900">
          핵심 키워드
        </h3>
        {keywords.length > 0 ? (
          <div className="flex flex-wrap gap-2">
            {keywords.map((keyword, index) => (
              <span
                key={index}
                className="rounded-full bg-green-100 px-3 py-1 text-sm font-medium text-green-800"
              >
                {keyword}
              </span>
            ))}
          </div>
        ) : (
          <p className="text-sm text-gray-500">키워드가 없습니다</p>
        )}
      </div>

      {/* 피해야 할 표현 */}
      <div>
        <h3 className="mb-3 text-lg font-semibold text-gray-900">
          ⚠️ 피해야 할 표현
        </h3>
        {avoidList.length > 0 ? (
          <ul className="space-y-2">
            {avoidList.map((avoid, index) => (
              <li
                key={index}
                data-testid="avoid-item"
                className="text-sm text-red-600"
              >
                • {avoid}
              </li>
            ))}
          </ul>
        ) : (
          <p className="text-sm text-gray-500">피해야 할 표현이 없습니다</p>
        )}
      </div>
    </div>
  );
}
