"use client";

import { useState, useCallback } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { Loader2 } from "lucide-react";
import { EditorLayout } from "./editor-layout";
import { VersionHistory } from "./version-history";
import { RollbackDialog } from "./rollback-dialog";
import {
  getCoverLetter,
  getVersions,
  createVersion,
  updateCoverLetter,
} from "@/lib/api/coaching";
import { useReview } from "@/hooks/use-review";
import { useCharCoaching } from "@/hooks/use-char-coaching";
import type {
  CharCoachingResult,
  CoverLetterVersion,
  ReviewResult,
  ReviewScores,
} from "@/lib/api/coaching";
import type { ReviewEntry } from "./review/review-timeline";

interface EditorPageClientProps {
  coverLetterId: string;
}

export function EditorPageClient({ coverLetterId }: EditorPageClientProps) {
  const queryClient = useQueryClient();
  const [previousScores, setPreviousScores] = useState<
    ReviewScores | undefined
  >();
  const [reviewResult, setReviewResult] = useState<ReviewResult | null>(null);
  const [reviewHistory, setReviewHistory] = useState<ReviewEntry[]>([]);
  const [selectedVersion, setSelectedVersion] =
    useState<CoverLetterVersion | null>(null);
  const [rollbackDialogOpen, setRollbackDialogOpen] = useState(false);
  const [charCoachingResult, setCharCoachingResult] =
    useState<CharCoachingResult | null>(null);

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
  const charCoachingMutation = useCharCoaching();

  const rollbackMutation = useMutation({
    mutationFn: async (version: CoverLetterVersion) => {
      // Save current content as a new version first, then restore
      await createVersion(
        coverLetterId,
        version.content,
        `v${version.version_number}에서 복원`
      );
      await updateCoverLetter(coverLetterId, version.content);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: ["cover-letter", coverLetterId],
      });
      queryClient.invalidateQueries({
        queryKey: ["cover-letter-versions", coverLetterId],
      });
      setSelectedVersion(null);
      setRollbackDialogOpen(false);
    },
  });

  const handleRequestReview = useCallback(() => {
    if (!coverLetter) return;

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
          setReviewHistory((prev) => [
            ...prev,
            {
              reviewNumber: prev.length + 1,
              overall: data.overall,
              scores: data.scores,
            },
          ]);
        },
      }
    );
  }, [coverLetter, coverLetterId, reviewMutation, reviewResult]);

  const handleRequestCharCoaching = useCallback(() => {
    if (!coverLetter) return;

    charCoachingMutation.mutate(
      {
        cover_letter_id: coverLetterId,
        content: coverLetter.current_content ?? "",
      },
      {
        onSuccess: (data) => {
          setCharCoachingResult(data);
        },
      }
    );
  }, [coverLetter, coverLetterId, charCoachingMutation]);

  const handleSelectVersion = useCallback(
    (version: CoverLetterVersion) => {
      setSelectedVersion(version);
    },
    []
  );

  const handleClosePreview = useCallback(() => {
    setSelectedVersion(null);
  }, []);

  const handleRestore = useCallback(() => {
    setRollbackDialogOpen(true);
  }, []);

  const handleConfirmRollback = useCallback(() => {
    if (!selectedVersion) return;
    rollbackMutation.mutate(selectedVersion);
  }, [selectedVersion, rollbackMutation]);

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

  const editorCoverLetter = {
    id: coverLetter.id,
    question_text: coverLetter.question_text ?? "",
    char_limit: coverLetter.char_limit ?? 800,
    current_content: coverLetter.current_content ?? "",
    company_name: "",
    position: "",
  };

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
        reviewHistory={reviewHistory}
        isReviewing={reviewMutation.isPending}
        onRequestReview={handleRequestReview}
        charCoachingResult={charCoachingResult}
        isCharCoaching={charCoachingMutation.isPending}
        onRequestCharCoaching={handleRequestCharCoaching}
        selectedVersion={selectedVersion}
        onClosePreview={handleClosePreview}
        onRestore={handleRestore}
      />
      {versions.length > 0 && (
        <VersionHistory
          versions={versions}
          coverLetterId={coverLetterId}
          selectedVersionId={selectedVersion?.id}
          onSelectVersion={handleSelectVersion}
        />
      )}
      {selectedVersion && (
        <RollbackDialog
          open={rollbackDialogOpen}
          onOpenChange={setRollbackDialogOpen}
          versionNumber={selectedVersion.version_number}
          isLoading={rollbackMutation.isPending}
          onConfirm={handleConfirmRollback}
        />
      )}
    </div>
  );
}
