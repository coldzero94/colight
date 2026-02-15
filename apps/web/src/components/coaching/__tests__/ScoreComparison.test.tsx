import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import { ScoreComparison } from "../review/score-comparison";

describe("ScoreComparison", () => {
  const current = { specificity: 80, job_fit: 75, company_fit: 70, authenticity: 85 };
  const previous = { specificity: 70, job_fit: 80, company_fit: 70, authenticity: 75 };

  it("renders all four dimension labels", () => {
    render(<ScoreComparison current={current} previous={previous} />);

    expect(screen.getByText("구체성")).toBeInTheDocument();
    expect(screen.getByText("직무 적합성")).toBeInTheDocument();
    expect(screen.getByText("기업 적합성")).toBeInTheDocument();
    expect(screen.getByText("진정성")).toBeInTheDocument();
  });

  it("shows positive diff with +N format", () => {
    render(<ScoreComparison current={current} previous={previous} />);

    // specificity: 80 - 70 = +10, authenticity: 85 - 75 = +10
    const positives = screen.getAllByText("+10");
    expect(positives).toHaveLength(2);
  });

  it("shows negative diff", () => {
    render(<ScoreComparison current={current} previous={previous} />);

    // job_fit: 75 - 80 = -5
    expect(screen.getByText("-5")).toBeInTheDocument();
  });

  it("shows dash for no change", () => {
    render(<ScoreComparison current={current} previous={previous} />);

    // company_fit: 70 - 70 = 0
    expect(screen.getByText("-")).toBeInTheDocument();
  });
});
