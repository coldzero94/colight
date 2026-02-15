import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import { InterviewProgress } from "../interview-progress";

describe("InterviewProgress", () => {
  it("renders all 5 stage labels", () => {
    render(<InterviewProgress currentStage="warmup" isComplete={false} />);

    expect(screen.getByText("가볍게")).toBeInTheDocument();
    expect(screen.getByText("기억에 남는 순간")).toBeInTheDocument();
    expect(screen.getByText("어려웠던 점")).toBeInTheDocument();
    expect(screen.getByText("해결법")).toBeInTheDocument();
    expect(screen.getByText("결과/배운 점")).toBeInTheDocument();
  });

  it("highlights current stage", () => {
    render(<InterviewProgress currentStage="challenge" isComplete={false} />);

    // Current stage (3rd = index 2) should have border style
    const currentLabel = screen.getByText("어려웠던 점");
    expect(currentLabel).toHaveClass("font-medium");
  });

  it("marks previous stages as done", () => {
    render(<InterviewProgress currentStage="challenge" isComplete={false} />);

    // First two stages should be completed (font-medium for labels)
    expect(screen.getByText("가볍게")).toHaveClass("font-medium");
    expect(screen.getByText("기억에 남는 순간")).toHaveClass("font-medium");
  });

  it("marks all stages as done when complete", () => {
    render(<InterviewProgress currentStage="outcome" isComplete={true} />);

    // All labels should have font-medium
    expect(screen.getByText("가볍게")).toHaveClass("font-medium");
    expect(screen.getByText("기억에 남는 순간")).toHaveClass("font-medium");
    expect(screen.getByText("어려웠던 점")).toHaveClass("font-medium");
    expect(screen.getByText("해결법")).toHaveClass("font-medium");
    expect(screen.getByText("결과/배운 점")).toHaveClass("font-medium");
  });
});
