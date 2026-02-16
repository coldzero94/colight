"use client";

import { useState } from "react";
import { ArrowLeft, GitCompare, RotateCcw } from "lucide-react";
import { Button } from "@/components/ui/button";
import type { CoverLetterVersion } from "@/lib/api/coaching";
import { computeDiff } from "@/lib/diff-utils";

interface VersionPreviewProps {
  version: CoverLetterVersion;
  currentContent: string;
  onRestore: () => void;
  onClose: () => void;
}

export function VersionPreview({
  version,
  currentContent,
  onRestore,
  onClose,
}: VersionPreviewProps) {
  const [showDiff, setShowDiff] = useState(false);

  const diff = showDiff
    ? computeDiff(version.content, currentContent)
    : null;

  return (
    <div className="flex h-full flex-col">
      {/* Header */}
      <div className="flex items-center justify-between border-b border-border px-6 py-3">
        <div className="flex items-center gap-3">
          <Button variant="ghost" size="sm" onClick={onClose}>
            <ArrowLeft className="h-4 w-4" />
            돌아가기
          </Button>
          <div className="text-sm text-muted-foreground">
            <span className="font-medium text-foreground">
              v{version.version_number}
            </span>
            <span className="ml-2">{version.char_count}자</span>
            <span className="ml-2 text-muted-foreground/60">
              {new Date(version.created_at).toLocaleString("ko-KR")}
            </span>
            {version.change_summary && (
              <span className="ml-2 text-muted-foreground/60">
                — {version.change_summary}
              </span>
            )}
          </div>
        </div>
        <div className="flex items-center gap-2">
          <Button
            variant={showDiff ? "secondary" : "outline"}
            size="sm"
            onClick={() => setShowDiff(!showDiff)}
          >
            <GitCompare className="h-4 w-4" />
            비교 보기
          </Button>
          <Button variant="default" size="sm" onClick={onRestore}>
            <RotateCcw className="h-4 w-4" />
            이 버전으로 복원
          </Button>
        </div>
      </div>

      {/* Diff stats */}
      {diff && (
        <div className="flex gap-3 border-b border-border bg-card px-6 py-2 text-xs">
          <span className="text-green-400">+{diff.stats.added}자 추가</span>
          <span className="text-red-400">-{diff.stats.removed}자 삭제</span>
          <span className="text-muted-foreground">
            {diff.stats.unchanged}자 동일
          </span>
        </div>
      )}

      {/* Content */}
      <div className="flex-1 overflow-y-auto px-6 py-4">
        {showDiff && diff ? (
          <div className="whitespace-pre-wrap text-sm leading-relaxed">
            {diff.parts.map((part, i) => {
              if (part.type === "added") {
                return (
                  <span
                    key={i}
                    className="bg-green-500/20 text-green-400"
                  >
                    {part.value}
                  </span>
                );
              }
              if (part.type === "removed") {
                return (
                  <span
                    key={i}
                    className="bg-red-500/20 text-red-400 line-through"
                  >
                    {part.value}
                  </span>
                );
              }
              return <span key={i}>{part.value}</span>;
            })}
          </div>
        ) : (
          <div className="whitespace-pre-wrap text-sm leading-relaxed text-foreground/90">
            {version.content}
          </div>
        )}
      </div>
    </div>
  );
}
