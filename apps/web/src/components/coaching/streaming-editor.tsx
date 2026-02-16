"use client";

import { StarHighlighter } from "./star-highlighter";

interface StreamingEditorProps {
  content: string;
  isStreaming: boolean;
}

export function StreamingEditor({
  content,
  isStreaming,
}: StreamingEditorProps) {
  return (
    <div
      className={`min-h-[400px] rounded-lg border border-border bg-card p-6 ${
        isStreaming ? "animate-pulse" : ""
      }`}
    >
      {content ? (
        <StarHighlighter text={content} />
      ) : (
        <p className="text-sm text-muted-foreground/60">
          초안이 생성되면 여기에 표시됩니다...
        </p>
      )}
    </div>
  );
}
