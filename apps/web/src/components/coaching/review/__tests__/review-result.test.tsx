import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import { ReviewResult } from "../review-result";
import type { ReviewResult as ReviewResultType } from "@/lib/api/coaching";

// Mock dynamic imports for radar chart
vi.mock("next/dynamic", () => ({
  default: (fn: () => Promise<unknown>) => {
    const Component = () => null;
    Component.displayName = "DynamicComponent";
    return Component;
  },
}));

const mockReview: ReviewResultType = {
  overall: 85,
  scores: {
    specificity: 80,
    job_fit: 85,
    company_fit: 90,
    authenticity: 85,
  },
  per_dimension_feedback: [
    {
      dimension: "specificity",
      score: 80,
      good: ["구체적인 수치 제시", "명확한 상황 설명"],
      improve: ["더 많은 세부사항 필요"],
    },
    {
      dimension: "job_fit",
      score: 85,
      good: ["직무 연관성 높음"],
      improve: ["기술 스택 언급 추가"],
    },
  ],
  specific_suggestions: [
    {
      original: "프로젝트를 진행했습니다",
      suggested: "5명 규모의 팀에서 프로젝트를 진행했습니다",
      reason: "팀 규모를 명시하여 구체성을 높입니다",
    },
    {
      original: "좋은 결과를 얻었습니다",
      suggested: "15% 성능 개선이라는 결과를 얻었습니다",
      reason: "정량적 지표로 성과를 표현합니다",
    },
  ],
};

const mockPreviousScores = {
  specificity: 75,
  job_fit: 80,
  company_fit: 85,
  authenticity: 80,
};

describe("ReviewResult", () => {
  it("renders overall score", () => {
    render(
      <ReviewResult
        review={mockReview}
        onApplySuggestion={vi.fn()}
      />
    );

    // Overall score 85 appears multiple times (overall + dimension scores)
    const scores = screen.getAllByText("85");
    expect(scores.length).toBeGreaterThanOrEqual(1);
  });

  it("renders dimension feedback cards", () => {
    render(
      <ReviewResult
        review={mockReview}
        onApplySuggestion={vi.fn()}
      />
    );

    expect(screen.getByText("구체성")).toBeInTheDocument();
    expect(screen.getByText("직무적합성")).toBeInTheDocument();
  });

  it("renders dimension scores", () => {
    render(
      <ReviewResult
        review={mockReview}
        onApplySuggestion={vi.fn()}
      />
    );

    expect(screen.getByText("80")).toBeInTheDocument();
    const scores85 = screen.getAllByText("85");
    expect(scores85.length).toBeGreaterThanOrEqual(1);
  });

  it("renders suggestion list with suggestions", () => {
    render(
      <ReviewResult
        review={mockReview}
        onApplySuggestion={vi.fn()}
      />
    );

    expect(screen.getByText("프로젝트를 진행했습니다")).toBeInTheDocument();
    expect(
      screen.getByText("5명 규모의 팀에서 프로젝트를 진행했습니다")
    ).toBeInTheDocument();
  });

  it("renders with previous scores for comparison", () => {
    render(
      <ReviewResult
        review={mockReview}
        previousScores={mockPreviousScores}
        onApplySuggestion={vi.fn()}
      />
    );

    const scores = screen.getAllByText("85");
    expect(scores.length).toBeGreaterThanOrEqual(1);
  });

  it("renders section heading for dimension feedback", () => {
    render(
      <ReviewResult
        review={mockReview}
        onApplySuggestion={vi.fn()}
      />
    );

    expect(screen.getByText("차원별 평가")).toBeInTheDocument();
  });

  it("renders all dimension feedback items", () => {
    render(
      <ReviewResult
        review={mockReview}
        onApplySuggestion={vi.fn()}
      />
    );

    expect(mockReview.per_dimension_feedback).toHaveLength(2);
    expect(screen.getByText("구체성")).toBeInTheDocument();
    expect(screen.getByText("직무적합성")).toBeInTheDocument();
  });
});
