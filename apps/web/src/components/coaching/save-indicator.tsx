"use client";

import { Check, Loader2, AlertCircle } from "lucide-react";

interface SaveIndicatorProps {
  status: "saved" | "saving" | "unsaved";
}

export function SaveIndicator({ status }: SaveIndicatorProps) {
  return (
    <div data-testid="save-indicator" className="flex items-center gap-2 text-sm">
      {status === "saved" && (
        <>
          <Check className="h-4 w-4 text-green-600" />
          <span className="text-green-600">저장됨</span>
        </>
      )}
      {status === "saving" && (
        <>
          <Loader2 className="h-4 w-4 animate-spin text-blue-600" />
          <span className="text-blue-600">저장 중...</span>
        </>
      )}
      {status === "unsaved" && (
        <>
          <AlertCircle className="h-4 w-4 text-amber-600" />
          <span className="text-amber-600">저장 안 됨</span>
        </>
      )}
    </div>
  );
}
