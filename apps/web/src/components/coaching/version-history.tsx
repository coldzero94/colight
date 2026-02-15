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
    <div className="border-t border-gray-200 bg-white">
      <button
        onClick={() => setIsOpen(!isOpen)}
        className="flex w-full items-center justify-between px-6 py-3 text-sm text-gray-600 hover:bg-gray-50"
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
        <div className="max-h-48 overflow-y-auto border-t border-gray-100 px-6 py-2">
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
                    ? "border-blue-200 bg-blue-50"
                    : "border-transparent hover:bg-gray-50"
                } ${isLatest && !isSelected ? "border-gray-200 bg-gray-50" : ""}`}
              >
                <div>
                  <span className="text-sm font-medium text-gray-900">
                    v{version.version_number}
                  </span>
                  {isLatest && (
                    <span className="ml-1.5 text-xs text-blue-600">(현재)</span>
                  )}
                  <span className="ml-2 text-xs text-gray-500">
                    {version.char_count}자
                  </span>
                  {version.change_summary && (
                    <span className="ml-2 text-xs text-gray-400">
                      {version.change_summary}
                    </span>
                  )}
                </div>
                <span className="text-xs text-gray-400">
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
