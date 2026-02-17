"use client";

import { StreamingProgress } from "./streaming-progress";
import { StreamingEditor } from "./streaming-editor";
import { AdvicePanel } from "./advice-panel";
import type { AdviceItem } from "@/lib/api/coaching";

interface DraftStreamingProps {
  content: string;
  isStreaming: boolean;
  charLimit: number;
  advice?: AdviceItem[];
  onComplete: () => void;
}

export function DraftStreaming({
  content,
  isStreaming,
  charLimit,
  advice = [],
  onComplete,
}: DraftStreamingProps) {
  // Count actual characters (not bytes) for Korean text
  const charCount = [...content].length;

  return (
    <div className="space-y-4">
      {/* Progress */}
      <StreamingProgress
        charCount={charCount}
        charLimit={charLimit}
        isStreaming={isStreaming}
      />

      {/* Editor */}
      <StreamingEditor content={content} isStreaming={isStreaming} />

      {/* Advice panel (shown after streaming completes) */}
      {!isStreaming && advice.length > 0 && <AdvicePanel advice={advice} />}

      {/* Actions */}
      {!isStreaming && content && (
        <div className="flex justify-end gap-3">
          <button
            onClick={onComplete}
            className="rounded-lg bg-primary px-4 py-2 text-sm font-medium text-primary-foreground hover:bg-primary/90"
          >
            ✏️ 에디터에서 편집하기
          </button>
        </div>
      )}
    </div>
  );
}
