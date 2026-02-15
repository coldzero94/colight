"use client";

import { OverallScore } from "./overall-score";
import { ScoreRadarChart } from "./score-radar-chart";
import { ScoreComparison } from "./score-comparison";
import { DimensionCard } from "./dimension-card";
import { SuggestionList } from "./suggestion-list";
import type {
  ReviewResult as ReviewResultType,
  ReviewScores,
  SpecificSuggestion,
} from "@/lib/api/coaching";

interface ReviewResultProps {
  review: ReviewResultType;
  previousScores?: ReviewScores;
  onApplySuggestion: (suggestion: SpecificSuggestion) => void;
}

export function ReviewResult({
  review,
  previousScores,
  onApplySuggestion,
}: ReviewResultProps) {
  return (
    <div className="space-y-6">
      <OverallScore
        score={review.overall}
        previousScore={previousScores ? computePreviousOverall(previousScores) : undefined}
      />

      <ScoreRadarChart
        scores={review.scores}
        previousScores={previousScores}
      />

      {previousScores && (
        <ScoreComparison current={review.scores} previous={previousScores} />
      )}

      <div className="space-y-2">
        <h3 className="text-sm font-medium text-gray-900">차원별 평가</h3>
        {review.per_dimension_feedback.map((fb) => (
          <DimensionCard key={fb.dimension} feedback={fb} />
        ))}
      </div>

      <SuggestionList
        suggestions={review.specific_suggestions}
        onApply={onApplySuggestion}
      />
    </div>
  );
}

function computePreviousOverall(scores: ReviewScores): number {
  return Math.round(
    (scores.specificity + scores.job_fit + scores.company_fit + scores.authenticity) / 4
  );
}
