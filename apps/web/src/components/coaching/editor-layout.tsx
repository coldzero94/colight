"use client";

import { useState, useCallback } from "react";
import Link from "next/link";
import dynamic from "next/dynamic";
import { ArrowLeft, Loader2, MessageSquareText, PanelRight } from "lucide-react";
import * as Tabs from "@radix-ui/react-tabs";
import { Sheet, SheetContent, SheetTitle } from "@/components/ui/sheet";
import { SaveIndicator } from "./save-indicator";
import { AnalysisSidebar } from "./analysis-sidebar";
import { VersionPreview } from "./version-preview";
import { ReviewResult as ReviewResultPanel } from "./review/review-result";
import { ReviewTimeline } from "./review/review-timeline";
import type { ReviewEntry } from "./review/review-timeline";
import { useAutoSave } from "@/hooks/use-auto-save";
import type {
  CoverLetterVersion,
  ReviewResult,
  ReviewScores,
  SpecificSuggestion,
} from "@/lib/api/coaching";

const TiptapEditor = dynamic(
  () => import("./tiptap-editor").then((m) => m.TiptapEditor),
  {
    ssr: false,
    loading: () => (
      <div className="space-y-3 p-4" data-testid="editor-skeleton">
        <div className="h-8 w-48 animate-pulse rounded bg-gray-200" />
        <div className="h-4 w-full animate-pulse rounded bg-gray-100" />
        <div className="h-4 w-3/4 animate-pulse rounded bg-gray-100" />
        <div className="h-4 w-5/6 animate-pulse rounded bg-gray-100" />
        <div className="h-4 w-2/3 animate-pulse rounded bg-gray-100" />
      </div>
    ),
  },
);

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
  reviewResult?: ReviewResult | null;
  previousScores?: ReviewScores;
  reviewHistory?: ReviewEntry[];
  isReviewing?: boolean;
  onRequestReview?: () => void;
  selectedVersion?: CoverLetterVersion | null;
  onClosePreview?: () => void;
  onRestore?: () => void;
}

export function EditorLayout({
  coverLetter,
  analysis,
  experiences,
  onSave,
  reviewResult,
  previousScores,
  reviewHistory = [],
  isReviewing,
  onRequestReview,
  selectedVersion,
  onClosePreview,
  onRestore,
}: EditorLayoutProps) {
  const [content, setContent] = useState(coverLetter.current_content);
  const [editorKey, setEditorKey] = useState(0);
  const { status, debouncedSave, saveVersion } = useAutoSave(coverLetter.id);

  const handleChange = (newContent: string) => {
    setContent(newContent);
    debouncedSave(newContent);
  };

  const handleSave = () => {
    saveVersion(content);
    onSave(content);
  };

  const handleApplySuggestion = useCallback(
    (suggestion: SpecificSuggestion) => {
      const idx = content.indexOf(suggestion.original);
      if (idx === -1) return;

      const newContent =
        content.slice(0, idx) +
        suggestion.suggested +
        content.slice(idx + suggestion.original.length);
      setContent(newContent);
      setEditorKey((k) => k + 1);
      debouncedSave(newContent);
    },
    [content, debouncedSave],
  );

  const [sheetOpen, setSheetOpen] = useState(false);

  const sidebarContent = (
    <Tabs.Root defaultValue="analysis" className="flex h-full flex-col">
      <Tabs.List className="flex shrink-0 border-b border-gray-200">
        <Tabs.Trigger
          value="analysis"
          className="flex-1 px-4 py-3 text-sm font-medium text-gray-500 data-[state=active]:border-b-2 data-[state=active]:border-gray-900 data-[state=active]:text-gray-900"
        >
          분석
        </Tabs.Trigger>
        <Tabs.Trigger
          value="review"
          className="flex-1 px-4 py-3 text-sm font-medium text-gray-500 data-[state=active]:border-b-2 data-[state=active]:border-gray-900 data-[state=active]:text-gray-900"
        >
          첨삭 결과
        </Tabs.Trigger>
      </Tabs.List>

      <Tabs.Content value="analysis" className="flex-1 overflow-y-auto p-6">
        <AnalysisSidebar analysis={analysis} experiences={experiences} />
      </Tabs.Content>

      <Tabs.Content value="review" className="flex-1 overflow-y-auto p-6">
        {reviewResult ? (
          <div className="space-y-6">
            <ReviewResultPanel
              review={reviewResult}
              previousScores={previousScores}
              onApplySuggestion={handleApplySuggestion}
            />
            {reviewHistory.length > 1 && (
              <ReviewTimeline entries={reviewHistory} />
            )}
          </div>
        ) : (
          <div className="flex flex-col items-center justify-center py-12 text-center">
            <MessageSquareText className="mb-3 h-8 w-8 text-gray-300" />
            <p className="text-sm text-gray-500">
              첨삭 결과가 없습니다
            </p>
            <p className="mt-1 text-xs text-gray-400">
              &quot;첨삭 요청&quot; 버튼을 눌러 AI 첨삭을 받아보세요
            </p>
          </div>
        )}
      </Tabs.Content>
    </Tabs.Root>
  );

  return (
    <div className="flex h-screen flex-col">
      {/* Header */}
      <div className="border-b border-gray-200 bg-white px-4 py-4 lg:px-6">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-4 min-w-0">
            <Link
              href="/coaching"
              className="flex shrink-0 items-center gap-2 text-sm text-gray-600 hover:text-gray-900"
            >
              <ArrowLeft className="h-4 w-4" />
              <span className="hidden sm:inline">코칭 목록</span>
            </Link>
            <div className="h-4 w-px bg-gray-300 hidden sm:block" />
            <div className="min-w-0">
              <h1 className="truncate text-lg font-semibold text-gray-900">
                {coverLetter.company_name} - {coverLetter.position}
              </h1>
              <p className="truncate text-sm text-gray-600">{coverLetter.question_text}</p>
            </div>
          </div>
          <div className="flex items-center gap-2">
            <SaveIndicator status={status} />
            <button
              onClick={() => setSheetOpen(true)}
              className="rounded-lg p-2 text-gray-500 hover:bg-gray-100 lg:hidden"
              aria-label="사이드바 열기"
            >
              <PanelRight className="h-5 w-5" />
            </button>
          </div>
        </div>
      </div>

      {/* Main content */}
      <div className="flex flex-1 overflow-hidden">
        {/* Editor area */}
        <div className="flex-1 overflow-hidden">
          {selectedVersion && onClosePreview && onRestore ? (
            <VersionPreview
              version={selectedVersion}
              currentContent={content}
              onRestore={onRestore}
              onClose={onClosePreview}
            />
          ) : (
            <div className="h-full overflow-y-auto p-4 lg:p-6">
              <TiptapEditor
                key={editorKey}
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
                  버전 저장
                </button>
                {onRequestReview && (
                  <button
                    onClick={onRequestReview}
                    disabled={isReviewing}
                    className="flex items-center gap-2 rounded-lg border border-gray-900 px-4 py-2 text-sm font-medium text-gray-900 hover:bg-gray-100 disabled:opacity-50"
                  >
                    {isReviewing ? (
                      <>
                        <Loader2 className="h-4 w-4 animate-spin" />
                        첨삭 중...
                      </>
                    ) : (
                      <>
                        <MessageSquareText className="h-4 w-4" />
                        첨삭 요청
                      </>
                    )}
                  </button>
                )}
              </div>
            </div>
          )}
        </div>

        {/* Desktop sidebar */}
        <div className="hidden lg:flex w-80 flex-col border-l border-gray-200 bg-gray-50">
          {sidebarContent}
        </div>

        {/* Mobile sidebar (Sheet) */}
        <Sheet open={sheetOpen} onOpenChange={setSheetOpen}>
          <SheetContent side="right" className="w-80 p-0 sm:max-w-80">
            <SheetTitle className="sr-only">분석 및 첨삭</SheetTitle>
            {sidebarContent}
          </SheetContent>
        </Sheet>
      </div>
    </div>
  );
}
