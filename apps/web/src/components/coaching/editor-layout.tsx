"use client";

import { useState } from "react";
import Link from "next/link";
import { ArrowLeft } from "lucide-react";
import { TiptapEditor } from "./tiptap-editor";
import { SaveIndicator } from "./save-indicator";
import { AnalysisSidebar } from "./analysis-sidebar";
import { useAutoSave } from "@/hooks/use-auto-save";

interface CoverLetter {
  id: string;
  question_text: string;
  char_limit: number;
  current_content: string;
  company_name: string;
  position: string;
}

interface AnalysisResult {
  required_weapons: {
    primary: { weapon_id: string; weapon_name: string; reason: string };
    secondary: Array<{ weapon_id: string; weapon_name: string; reason: string }>;
  };
  writing_structure: {
    total_chars: number;
    sections: Array<{
      name: string;
      char_ratio: number;
      char_count: number;
      guide: string;
    }>;
  };
  key_keywords: string[];
}

interface Experience {
  id: string;
  title: string;
  category: string;
  star_situation: string;
  weapons: Array<{ name: string }>;
  matchScore: number;
}

interface EditorLayoutProps {
  coverLetter: CoverLetter;
  analysis: AnalysisResult;
  experiences: Experience[];
  onSave: (content: string) => void;
}

export function EditorLayout({
  coverLetter,
  analysis,
  experiences,
  onSave,
}: EditorLayoutProps) {
  const [content, setContent] = useState(coverLetter.current_content);
  const { status, debouncedSave, saveVersion } = useAutoSave(coverLetter.id);

  const handleChange = (newContent: string) => {
    setContent(newContent);
    debouncedSave(newContent); // Auto-save with 2s debounce
  };

  const handleSave = () => {
    saveVersion(content); // Manual version save
    onSave(content);
  };

  return (
    <div className="h-screen flex flex-col">
      {/* Header */}
      <div className="border-b border-gray-200 bg-white px-6 py-4">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-4">
            <Link
              href="/coaching"
              className="flex items-center gap-2 text-sm text-gray-600 hover:text-gray-900"
            >
              <ArrowLeft className="h-4 w-4" />
              코칭 목록
            </Link>
            <div className="h-4 w-px bg-gray-300" />
            <div>
              <h1 className="text-lg font-semibold text-gray-900">
                {coverLetter.company_name} - {coverLetter.position}
              </h1>
              <p className="text-sm text-gray-600">{coverLetter.question_text}</p>
            </div>
          </div>
          <SaveIndicator status={status} />
        </div>
      </div>

      {/* Main content */}
      <div className="flex flex-1 overflow-hidden">
        {/* Editor area */}
        <div className="flex-1 overflow-y-auto p-6">
          <TiptapEditor
            content={content}
            charLimit={coverLetter.char_limit}
            onChange={handleChange}
            editable={true}
          />

          {/* Actions */}
          <div className="mt-4 flex gap-3">
            <button
              onClick={handleSave}
              disabled={status === "saved"}
              className="rounded-lg bg-gray-900 px-4 py-2 text-sm font-medium text-white hover:bg-gray-800 disabled:opacity-50"
            >
              💾 버전 저장
            </button>
          </div>
        </div>

        {/* Sidebar */}
        <div className="w-80 overflow-y-auto border-l border-gray-200 bg-gray-50 p-6">
          <AnalysisSidebar analysis={analysis} experiences={experiences} />
        </div>
      </div>
    </div>
  );
}
