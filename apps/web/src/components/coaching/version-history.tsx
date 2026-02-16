"use client";

import { useState } from "react";
import { Clock, ChevronDown, ChevronUp } from "lucide-react";
import type { CoverLetterVersion } from "@/lib/api/coaching";

interface VersionHistoryProps {
  versions: CoverLetterVersion[];
  coverLetterId: string;
  selectedVersionId?: string;
  onSelectVersion?: (version: CoverLetterVersion) => void;
}

export function VersionHistory({
  versions,
  selectedVersionId,
  onSelectVersion,
}: VersionHistoryProps) {
  const [isOpen, setIsOpen] = useState(false);

  if (versions.length === 0) return null;

  const latestVersionNumber = versions[0]?.version_number;

  return (
    <div className="border-t border-border bg-card">
      <button
        onClick={() => setIsOpen(!isOpen)}
        className="flex w-full items-center justify-between px-6 py-3 text-sm text-muted-foreground hover:bg-white/[0.04]"
      >
        <div className="flex items-center gap-2">
          <Clock className="h-4 w-4" />
          <span>버전 이력 ({versions.length}개)</span>
        </div>
        {isOpen ? (
          <ChevronDown className="h-4 w-4" />
        ) : (
          <ChevronUp className="h-4 w-4" />
        )}
      </button>

      {isOpen && (
        <div className="max-h-48 overflow-y-auto border-t border-border px-6 py-2">
          {versions.map((version) => {
            const isLatest = version.version_number === latestVersionNumber;
            const isSelected = version.id === selectedVersionId;

            return (
              <button
                key={version.id}
                type="button"
                onClick={() => onSelectVersion?.(version)}
                className={`flex w-full items-center justify-between rounded-md border px-3 py-2 text-left transition-colors ${
                  isSelected
                    ? "border-blue-500/30 bg-blue-500/10"
                    : "border-transparent hover:bg-white/[0.04]"
                } ${isLatest && !isSelected ? "border-border bg-white/[0.02]" : ""}`}
              >
                <div>
                  <span className="text-sm font-medium text-foreground">
                    v{version.version_number}
                  </span>
                  {isLatest && (
                    <span className="ml-1.5 text-xs text-blue-400">(현재)</span>
                  )}
                  <span className="ml-2 text-xs text-muted-foreground">
                    {version.char_count}자
                  </span>
                  {version.change_summary && (
                    <span className="ml-2 text-xs text-muted-foreground/60">
                      {version.change_summary}
                    </span>
                  )}
                </div>
                <span className="text-xs text-muted-foreground/60">
                  {new Date(version.created_at).toLocaleString("ko-KR")}
                </span>
              </button>
            );
          })}
        </div>
      )}
    </div>
  );
}
