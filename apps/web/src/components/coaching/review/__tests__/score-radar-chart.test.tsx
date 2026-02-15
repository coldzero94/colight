import { describe, it, expect, vi } from "vitest";
import { render } from "@testing-library/react";
import { ScoreRadarChart } from "../score-radar-chart";
import type { ReviewScores } from "@/lib/api/coaching";

// Mock next/dynamic to return a simple component
vi.mock("next/dynamic", () => ({
  default: (fn: () => Promise<unknown>) => {
    const Component = () => null;
    Component.displayName = "DynamicComponent";
    return Component;
  },
}));

const mockScores: ReviewScores = {
  specificity: 80,
  job_fit: 85,
  company_fit: 90,
  authenticity: 85,
};

const mockPreviousScores: ReviewScores = {
  specificity: 75,
  job_fit: 80,
  company_fit: 85,
  authenticity: 80,
};

describe("ScoreRadarChart", () => {
  it("renders without errors (smoke test)", () => {
    const { container } = render(<ScoreRadarChart scores={mockScores} />);
    expect(container).toBeInTheDocument();
  });

  it("renders with previous scores without errors", () => {
    const { container } = render(
      <ScoreRadarChart scores={mockScores} previousScores={mockPreviousScores} />
    );
    expect(container).toBeInTheDocument();
  });

  it("renders container with correct height class", () => {
    const { container } = render(<ScoreRadarChart scores={mockScores} />);
    const chartContainer = container.querySelector(".h-64");
    expect(chartContainer).toBeInTheDocument();
  });

  it("renders container with full width", () => {
    const { container } = render(<ScoreRadarChart scores={mockScores} />);
    const chartContainer = container.querySelector(".w-full");
    expect(chartContainer).toBeInTheDocument();
  });

  it("accepts all valid score properties", () => {
    const scores: ReviewScores = {
      specificity: 100,
      job_fit: 100,
      company_fit: 100,
      authenticity: 100,
    };

    const { container } = render(<ScoreRadarChart scores={scores} />);
    expect(container).toBeInTheDocument();
  });

  it("handles zero scores without errors", () => {
    const zeroScores: ReviewScores = {
      specificity: 0,
      job_fit: 0,
      company_fit: 0,
      authenticity: 0,
    };

    const { container } = render(<ScoreRadarChart scores={zeroScores} />);
    expect(container).toBeInTheDocument();
  });
});
