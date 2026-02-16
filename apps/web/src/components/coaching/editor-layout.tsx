"use client";

import { useState, useCallback } from "react";
import Link from "next/link";
import dynamic from "next/dynamic";
import { ArrowLeft, Loader2, MessageSquareText, Ruler, PanelRight } from "lucide-react";
import * as Tabs from "@radix-ui/react-tabs";
import { Sheet, SheetContent, SheetTitle } from "@/components/ui/sheet";
import { SaveIndicator } from "./save-indicator";
import { AnalysisSidebar } from "./analysis-sidebar";
import { VersionPreview } from "./version-preview";
import { ReviewResult as ReviewResultPanel } from "./review/review-result";
import { ReviewTimeline } from "./review/review-timeline";
import { CharCoachingResult as CharCoachingResultPanel } from "./char-coaching-result";
import type { ReviewEntry } from "./review/review-timeline";
import { useAutoSave } from "@/hooks/use-auto-save";
import type {
  CharCoachingResult,
  CharCoachingSuggestion,
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
        <div className="h-8 w-48 animate-pulse rounded bg-white/[0.08]" />
        <div className="h-4 w-full animate-pulse rounded bg-white/[0.06]" />
        <div className="h-4 w-3/4 animate-pulse rounded bg-white/[0.06]" />
        <div className="h-4 w-5/6 animate-pulse rounded bg-white/[0.06]" />
        <div className="h-4 w-2/3 animate-pulse rounded bg-white/[0.06]" />
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
  charCoachingResult?: CharCoachingResult | null;
  isCharCoaching?: boolean;
  onRequestCharCoaching?: () => void;
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
  charCoachingResult,
  isCharCoaching,
  onRequestCharCoaching,
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

  const handleApplyCharSuggestion = useCallback(
    (suggestion: CharCoachingSuggestion) => {
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
      <Tabs.List className="flex shrink-0 border-b border-border">
        <Tabs.Trigger
          value="analysis"
          className="flex-1 px-4 py-3 text-sm font-medium text-muted-foreground data-[state=active]:border-b-2 data-[state=active]:border-primary data-[state=active]:text-foreground"
        >
          분석
        </Tabs.Trigger>
        <Tabs.Trigger
          value="review"
          className="flex-1 px-4 py-3 text-sm font-medium text-muted-foreground data-[state=active]:border-b-2 data-[state=active]:border-primary data-[state=active]:text-foreground"
        >
          첨삭 결과
        </Tabs.Trigger>
        <Tabs.Trigger
          value="char-coaching"
          className="flex-1 px-4 py-3 text-sm font-medium text-muted-foreground data-[state=active]:border-b-2 data-[state=active]:border-primary data-[state=active]:text-foreground"
        >
          글자수
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
            <MessageSquareText className="mb-3 h-8 w-8 text-muted-foreground/40" />
            <p className="text-sm text-muted-foreground">
              첨삭 결과가 없습니다
            </p>
            <p className="mt-1 text-xs text-muted-foreground/60">
              &quot;첨삭 요청&quot; 버튼을 눌러 AI 첨삭을 받아보세요
            </p>
          </div>
        )}
      </Tabs.Content>

      <Tabs.Content value="char-coaching" className="flex-1 overflow-y-auto p-6">
        {charCoachingResult ? (
          <CharCoachingResultPanel
            result={charCoachingResult}
            onApplySuggestion={handleApplyCharSuggestion}
          />
        ) : (
          <div className="flex flex-col items-center justify-center py-12 text-center">
            <Ruler className="mb-3 h-8 w-8 text-muted-foreground/40" />
            <p className="text-sm text-muted-foreground">
              글자수 코칭 결과가 없습니다
            </p>
            <p className="mt-1 text-xs text-muted-foreground/60">
              &quot;글자수 코칭&quot; 버튼을 눌러 AI 코칭을 받아보세요
            </p>
          </div>
        )}
      </Tabs.Content>
    </Tabs.Root>
  );

  return (
    <div className="flex h-screen flex-col">
      {/* Header */}
      <div className="border-b border-border bg-card px-4 py-4 lg:px-6">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-4 min-w-0">
            <Link
              href="/coaching"
              className="flex shrink-0 items-center gap-2 text-sm text-muted-foreground hover:text-foreground"
            >
              <ArrowLeft className="h-4 w-4" />
              <span className="hidden sm:inline">코칭 목록</span>
            </Link>
            <div className="h-4 w-px bg-border hidden sm:block" />
            <div className="min-w-0">
              <h1 className="truncate text-lg font-semibold text-foreground">
                {coverLetter.company_name} - {coverLetter.position}
              </h1>
              <p className="truncate text-sm text-muted-foreground">{coverLetter.question_text}</p>
            </div>
          </div>
          <div className="flex items-center gap-2">
            <SaveIndicator status={status} />
            <button
              onClick={() => setSheetOpen(true)}
              className="rounded-lg p-2 text-muted-foreground hover:bg-white/[0.06] lg:hidden"
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
                  className="rounded-lg bg-primary px-4 py-2 text-sm font-medium text-primary-foreground hover:bg-primary/90 disabled:opacity-50"
                >
                  버전 저장
                </button>
                {onRequestReview && (
                  <button
                    onClick={onRequestReview}
                    disabled={isReviewing}
                    className="flex items-center gap-2 rounded-lg border border-primary px-4 py-2 text-sm font-medium text-primary hover:bg-primary/10 disabled:opacity-50"
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
                {onRequestCharCoaching && (
                  <button
                    onClick={onRequestCharCoaching}
                    disabled={isCharCoaching}
                    className="flex items-center gap-2 rounded-lg border border-border px-4 py-2 text-sm font-medium text-foreground/80 hover:bg-white/[0.04] disabled:opacity-50"
                  >
                    {isCharCoaching ? (
                      <>
                        <Loader2 className="h-4 w-4 animate-spin" />
                        분석 중...
                      </>
                    ) : (
                      <>
                        <Ruler className="h-4 w-4" />
                        글자수 코칭
                      </>
                    )}
                  </button>
                )}
              </div>
            </div>
          )}
        </div>

        {/* Desktop sidebar */}
        <div className="hidden lg:flex w-80 flex-col border-l border-border bg-card">
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
