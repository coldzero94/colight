import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import { FeatureSection } from "../feature-section";

describe("FeatureSection", () => {
  it("renders 3 feature cards", () => {
    render(<FeatureSection />);
    expect(screen.getByText("기업 분석")).toBeInTheDocument();
    expect(screen.getByText("경험 매칭")).toBeInTheDocument();
    expect(screen.getByText("AI 코칭")).toBeInTheDocument();
  });

  it("renders feature descriptions", () => {
    render(<FeatureSection />);
    expect(screen.getByText(/채용공고 URL만 입력/)).toBeInTheDocument();
    expect(screen.getByText(/AI가 매칭/)).toBeInTheDocument();
    expect(screen.getByText(/STAR 구조/)).toBeInTheDocument();
  });
});
