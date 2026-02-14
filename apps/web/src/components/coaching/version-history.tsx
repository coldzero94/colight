"use client";

import { useState } from "react";
import { Clock, ChevronDown, ChevronUp } from "lucide-react";
import type { CoverLetterVersion } from "@/lib/api/coaching";

interface VersionHistoryProps {
  versions: CoverLetterVersion[];
  coverLetterId: string;
}

export function VersionHistory({ versions }: VersionHistoryProps) {
  const [isOpen, setIsOpen] = useState(false);

  if (versions.length === 0) return null;

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
          {versions.map((version) => (
            <div
              key={version.id}
              className="flex items-center justify-between border-b border-gray-50 py-2 last:border-0"
            >
              <div>
                <span className="text-sm font-medium text-gray-900">
                  v{version.version_number}
                </span>
                <span className="ml-2 text-xs text-gray-500">
                  {version.char_count}자
                </span>
              </div>
              <span className="text-xs text-gray-400">
                {new Date(version.created_at).toLocaleString("ko-KR")}
              </span>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
