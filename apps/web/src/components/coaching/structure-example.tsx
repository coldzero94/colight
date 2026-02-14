"use client";

import { useState } from "react";
import { ChevronDown, ChevronUp } from "lucide-react";

interface StructureExampleProps {
  example: string;
}

export function StructureExample({ example }: StructureExampleProps) {
  const [isExpanded, setIsExpanded] = useState(false);

  return (
    <div className="rounded-lg border border-gray-200 bg-white p-6">
      <button
        onClick={() => setIsExpanded(!isExpanded)}
        className="flex w-full items-center justify-between text-left"
      >
        <h3 className="text-lg font-semibold text-gray-900">
          좋은 구조 예시
        </h3>
        {isExpanded ? (
          <ChevronUp className="h-5 w-5 text-gray-500" />
        ) : (
          <ChevronDown className="h-5 w-5 text-gray-500" />
        )}
      </button>

      {isExpanded && (
        <div className="mt-4 whitespace-pre-wrap rounded-lg bg-gray-50 p-4 text-sm text-gray-700">
          {example}
        </div>
      )}
    </div>
  );
}
