"use client";

import { useState, useCallback } from "react";
import { useQuery } from "@tanstack/react-query";
import { Loader2 } from "lucide-react";
import { EditorLayout } from "./editor-layout";
import { VersionHistory } from "./version-history";
import { getCoverLetter, getVersions } from "@/lib/api/coaching";
import { useReview } from "@/hooks/use-review";
import type { ReviewResult, ReviewScores } from "@/lib/api/coaching";

interface EditorPageClientProps {
  coverLetterId: string;
}

export function EditorPageClient({ coverLetterId }: EditorPageClientProps) {
  const [previousScores, setPreviousScores] = useState<ReviewScores | undefined>();
  const [reviewResult, setReviewResult] = useState<ReviewResult | null>(null);

  const {
    data: coverLetter,
    isLoading,
    error,
  } = useQuery({
    queryKey: ["cover-letter", coverLetterId],
    queryFn: () => getCoverLetter(coverLetterId),
  });

  const { data: versionsData } = useQuery({
    queryKey: ["cover-letter-versions", coverLetterId],
    queryFn: () => getVersions(coverLetterId),
  });

  const reviewMutation = useReview();

  const handleRequestReview = useCallback(() => {
    if (!coverLetter) return;

    // Store previous scores for comparison on re-review
    if (reviewResult) {
      setPreviousScores(reviewResult.scores);
    }

    reviewMutation.mutate(
      {
        cover_letter_id: coverLetterId,
        content: coverLetter.current_content ?? "",
      },
      {
        onSuccess: (data) => {
          setReviewResult(data);
        },
      },
    );
  }, [coverLetter, coverLetterId, reviewMutation, reviewResult]);

  if (isLoading) {
    return (
      <div className="flex items-center justify-center py-20">
        <Loader2 className="h-6 w-6 animate-spin text-gray-400" />
        <span className="ml-2 text-gray-500">에디터를 불러오는 중...</span>
      </div>
    );
  }

  if (error || !coverLetter) {
    return (
      <div className="flex flex-col items-center justify-center gap-4 py-20">
        <h2 className="text-lg font-semibold text-gray-900">
          자소서를 불러올 수 없습니다
        </h2>
        <p className="text-sm text-gray-500">
          {error?.message ?? "데이터를 찾을 수 없습니다"}
        </p>
      </div>
    );
  }

  // Build EditorLayout props from API data
  const editorCoverLetter = {
    id: coverLetter.id,
    question_text: coverLetter.question_text ?? "",
    char_limit: coverLetter.char_limit ?? 800,
    current_content: coverLetter.current_content ?? "",
    company_name: "",
    position: "",
  };

  // Placeholder analysis - will be populated from coaching session data
  const analysis = {
    required_weapons: {
      primary: { weapon_id: "", weapon_name: "", reason: "" },
      secondary: [] as Array<{
        weapon_id: string;
        weapon_name: string;
        reason: string;
      }>,
    },
    writing_structure: {
      total_chars: coverLetter.char_limit ?? 800,
      sections: [] as Array<{
        name: string;
        char_ratio: number;
        char_count: number;
        guide: string;
      }>,
    },
    key_keywords: [] as string[],
  };

  const versions = versionsData?.versions ?? [];

  return (
    <div className="flex h-screen flex-col">
      <EditorLayout
        coverLetter={editorCoverLetter}
        analysis={analysis}
        experiences={[]}
        onSave={() => {}}
        reviewResult={reviewResult}
        previousScores={previousScores}
        isReviewing={reviewMutation.isPending}
        onRequestReview={handleRequestReview}
      />
      {versions.length > 0 && (
        <VersionHistory
          versions={versions}
          coverLetterId={coverLetterId}
        />
      )}
    </div>
  );
}
