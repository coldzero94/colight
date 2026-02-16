"use client";

import { StreamingProgress } from "./streaming-progress";
import { StreamingEditor } from "./streaming-editor";

interface DraftStreamingProps {
  content: string;
  isStreaming: boolean;
  charLimit: number;
  onComplete: () => void;
}

export function DraftStreaming({
  content,
  isStreaming,
  charLimit,
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
