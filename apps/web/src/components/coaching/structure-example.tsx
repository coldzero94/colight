"use client";

import { useState } from "react";
import { ChevronDown, ChevronUp } from "lucide-react";

interface StructureExampleProps {
  example: string;
}

export function StructureExample({ example }: StructureExampleProps) {
  const [isExpanded, setIsExpanded] = useState(false);

  return (
    <div className="rounded-lg border border-border bg-card p-6">
      <button
        onClick={() => setIsExpanded(!isExpanded)}
        className="flex w-full items-center justify-between text-left"
      >
        <h3 className="text-lg font-semibold text-foreground">
          좋은 구조 예시
        </h3>
        {isExpanded ? (
          <ChevronUp className="h-5 w-5 text-muted-foreground" />
        ) : (
          <ChevronDown className="h-5 w-5 text-muted-foreground" />
        )}
      </button>

      {isExpanded && (
        <div className="mt-4 whitespace-pre-wrap rounded-lg bg-white/[0.02] p-4 text-sm text-foreground/80">
          {example}
        </div>
      )}
    </div>
  );
}
