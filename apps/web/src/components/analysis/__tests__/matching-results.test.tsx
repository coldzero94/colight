import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import { MatchingResults } from "../matching-results";
import type { MatchResult } from "@/lib/api/matching";
import type { Experience } from "@/lib/api/experiences";

const mockExperience = (id: string, title: string): Experience => ({
  id,
  user_id: "user-1",
  title,
  content: "test content",
  category: "인턴",
  role: "developer",
  result: "good result",
  star_situation: "S",
  star_task: "T",
  star_action: "A",
  star_result: "R",
  keywords: [],
  source: "manual",
  is_archived: false,
  created_at: "2026-01-01",
  updated_at: "2026-01-01",
  weapons: [],
});

const mockMatch = (
  id: string,
  overallFit: number,
  jobRelevance = 80,
  talentFit = 70,
  uniqueness = 60
): MatchResult => ({
  experience_id: id,
  overall_fit: overallFit,
  job_relevance: jobRelevance,
  talent_fit: talentFit,
  uniqueness: uniqueness,
  reasoning: "Test reasoning for " + id,
  suggested_angle: "Test angle for " + id,
});

describe("MatchingResults", () => {
  it("renders experience title and overall fit score", () => {
    const experiences = [mockExperience("exp-1", "백엔드 개발 경험")];
    const matches = [mockMatch("exp-1", 85)];

    render(<MatchingResults matches={matches} experiences={experiences} />);

    expect(screen.getByText("백엔드 개발 경험")).toBeInTheDocument();
    expect(screen.getByText("85점")).toBeInTheDocument();
  });

  it("applies green color class for score >= 80", () => {
    const experiences = [mockExperience("exp-1", "High score")];
    const matches = [mockMatch("exp-1", 90)];

    const { container } = render(
      <MatchingResults matches={matches} experiences={experiences} />
    );

    // FitScoreBadge uses bg-primary/10 for score >= 80
    const badge = container.querySelector('[class*="bg-primary"]');
    expect(badge).toBeInTheDocument();
  });

  it("applies yellow color class for score 40-69", () => {
    const experiences = [mockExperience("exp-1", "Mid score")];
    const matches = [mockMatch("exp-1", 55)];

    const { container } = render(
      <MatchingResults matches={matches} experiences={experiences} />
    );

    const badge = container.querySelector('[class*="bg-amber"]');
    expect(badge).toBeInTheDocument();
  });

  it("applies red color class for score < 40", () => {
    const experiences = [mockExperience("exp-1", "Low score")];
    const matches = [mockMatch("exp-1", 25)];

    const { container } = render(
      <MatchingResults matches={matches} experiences={experiences} />
    );

    const badge = container.querySelector('[class*="bg-red"]');
    expect(badge).toBeInTheDocument();
  });

  it("sorts experiences by score descending", () => {
    const experiences = [
      mockExperience("exp-1", "Low"),
      mockExperience("exp-2", "High"),
      mockExperience("exp-3", "Mid"),
    ];
    const matches = [
      mockMatch("exp-1", 30),
      mockMatch("exp-2", 90),
      mockMatch("exp-3", 60),
    ];

    render(<MatchingResults matches={matches} experiences={experiences} />);

    const titles = screen.getAllByRole("heading", { level: 3 });
    expect(titles[0]).toHaveTextContent("High");
    expect(titles[1]).toHaveTextContent("Mid");
    expect(titles[2]).toHaveTextContent("Low");
  });

  it("shows empty state when no matches", () => {
    render(<MatchingResults matches={[]} experiences={[]} />);

    expect(screen.getByText(/매칭 결과가 없습니다/)).toBeInTheDocument();
  });

  it("displays reasoning text", () => {
    const experiences = [mockExperience("exp-1", "Test")];
    const matches = [mockMatch("exp-1", 75)];

    render(<MatchingResults matches={matches} experiences={experiences} />);

    expect(screen.getByText("Test reasoning for exp-1")).toBeInTheDocument();
  });

  it("displays suggested angle", () => {
    const experiences = [mockExperience("exp-1", "Test")];
    const matches = [mockMatch("exp-1", 75)];

    render(<MatchingResults matches={matches} experiences={experiences} />);

    expect(screen.getByText("Test angle for exp-1")).toBeInTheDocument();
  });

  it("renders score bars with correct width percentage", () => {
    const experiences = [mockExperience("exp-1", "Test")];
    const matches = [mockMatch("exp-1", 75, 80, 70, 60)];

    const { container } = render(
      <MatchingResults matches={matches} experiences={experiences} />
    );

    // Check for score bar widths
    const bars = container.querySelectorAll('[style*="width"]');
    const widths = Array.from(bars).map((bar) =>
      (bar as HTMLElement).style.width
    );
    expect(widths).toContain("80%");
    expect(widths).toContain("70%");
    expect(widths).toContain("60%");
  });
});
