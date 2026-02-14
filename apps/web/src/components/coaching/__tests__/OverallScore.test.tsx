import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import { OverallScore } from "../review/overall-score";

describe("OverallScore", () => {
  it("renders score and grade S for 90+", () => {
    render(<OverallScore score={92} />);
    expect(screen.getByText("92")).toBeInTheDocument();
    expect(screen.getByText("S")).toBeInTheDocument();
  });

  it("renders grade A for 80-89", () => {
    render(<OverallScore score={85} />);
    expect(screen.getByText("A")).toBeInTheDocument();
  });

  it("renders grade B for 70-79", () => {
    render(<OverallScore score={76} />);
    expect(screen.getByText("B")).toBeInTheDocument();
  });

  it("renders grade C for 60-69", () => {
    render(<OverallScore score={63} />);
    expect(screen.getByText("C")).toBeInTheDocument();
  });

  it("renders grade D for below 60", () => {
    render(<OverallScore score={45} />);
    expect(screen.getByText("D")).toBeInTheDocument();
  });

  it("shows positive diff when previousScore is lower", () => {
    render(<OverallScore score={80} previousScore={70} />);
    expect(screen.getByText("+10")).toBeInTheDocument();
  });

  it("shows negative diff when previousScore is higher", () => {
    render(<OverallScore score={65} previousScore={75} />);
    expect(screen.getByText("-10")).toBeInTheDocument();
  });

  it("does not show diff when no previousScore", () => {
    render(<OverallScore score={80} />);
    expect(screen.queryByText(/[+\-]/)).not.toBeInTheDocument();
  });
});
