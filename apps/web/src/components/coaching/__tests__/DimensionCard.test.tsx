import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { DimensionCard } from "../review/dimension-card";

const mockFeedback = {
  dimension: "specificity",
  score: 75,
  good: ["구체적인 수치 활용이 좋습니다"],
  improve: ["행동의 구체적 과정을 더 서술하세요"],
};

describe("DimensionCard", () => {
  it("renders dimension name and score", () => {
    render(<DimensionCard feedback={mockFeedback} />);
    expect(screen.getByText("구체성")).toBeInTheDocument();
    expect(screen.getByText("75")).toBeInTheDocument();
  });

  it("is collapsed by default", () => {
    render(<DimensionCard feedback={mockFeedback} />);
    expect(
      screen.queryByText("구체적인 수치 활용이 좋습니다")
    ).not.toBeInTheDocument();
  });

  it("expands on click to show feedback", async () => {
    const user = userEvent.setup();
    render(<DimensionCard feedback={mockFeedback} />);

    await user.click(screen.getByText("구체성"));

    expect(
      screen.getByText("구체적인 수치 활용이 좋습니다")
    ).toBeInTheDocument();
    expect(
      screen.getByText("행동의 구체적 과정을 더 서술하세요")
    ).toBeInTheDocument();
  });

  it("shows expanded by default when defaultExpanded is true", () => {
    render(<DimensionCard feedback={mockFeedback} defaultExpanded />);
    expect(
      screen.getByText("구체적인 수치 활용이 좋습니다")
    ).toBeInTheDocument();
  });
});
