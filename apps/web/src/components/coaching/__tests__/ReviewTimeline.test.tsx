import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import { ReviewTimeline } from "../review/review-timeline";
import type { ReviewEntry } from "../review/review-timeline";

const mockEntries: ReviewEntry[] = [
  {
    reviewNumber: 1,
    overall: 65,
    scores: { specificity: 60, job_fit: 70, company_fit: 65, authenticity: 65 },
  },
  {
    reviewNumber: 2,
    overall: 75,
    scores: { specificity: 75, job_fit: 75, company_fit: 70, authenticity: 80 },
  },
  {
    reviewNumber: 3,
    overall: 82,
    scores: { specificity: 85, job_fit: 80, company_fit: 78, authenticity: 85 },
  },
];

describe("ReviewTimeline", () => {
  it("renders nothing for empty entries", () => {
    const { container } = render(<ReviewTimeline entries={[]} />);
    expect(container.firstChild).toBeNull();
  });

  it("renders entries with review numbers", () => {
    render(<ReviewTimeline entries={mockEntries} />);

    expect(screen.getByText("1차 첨삭")).toBeInTheDocument();
    expect(screen.getByText("2차 첨삭")).toBeInTheDocument();
    expect(screen.getByText("3차 첨삭")).toBeInTheDocument();
  });

  it("shows score diffs from previous entry", () => {
    render(<ReviewTimeline entries={mockEntries} />);

    // 2nd entry: 75 - 65 = +10
    expect(screen.getByText("+10")).toBeInTheDocument();
    // 3rd entry: 82 - 75 = +7
    expect(screen.getByText("+7")).toBeInTheDocument();
  });

  it("shows heading", () => {
    render(<ReviewTimeline entries={mockEntries} />);
    expect(screen.getByText("첨삭 이력")).toBeInTheDocument();
  });
});
