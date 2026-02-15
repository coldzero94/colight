import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import { StructureChart } from "../structure-chart";

describe("StructureChart", () => {
  const mockSections = [
    { name: "도입", char_ratio: 0.2, char_count: 200, guide: "도입 가이드" },
    { name: "본문", char_ratio: 0.6, char_count: 600, guide: "본문 가이드" },
    { name: "결론", char_ratio: 0.2, char_count: 200, guide: "결론 가이드" },
  ];

  it("renders section names and percentages", () => {
    render(<StructureChart sections={mockSections} totalChars={1000} />);
    expect(screen.getByText("도입")).toBeInTheDocument();
    expect(screen.getByText("본문")).toBeInTheDocument();
    expect(screen.getByText("결론")).toBeInTheDocument();
    // 도입 and 결론 both have 20% (200자), so there should be 2 matches
    expect(screen.getAllByText(/20%.*200자/).length).toBe(2);
    expect(screen.getByText(/60%.*600자/)).toBeInTheDocument();
  });

  it("renders progress bars", () => {
    const { container } = render(
      <StructureChart sections={mockSections} totalChars={1000} />
    );
    // Progress bars use bg-gradient-to-r class
    const progressBars = container.querySelectorAll(".bg-gradient-to-r");
    expect(progressBars.length).toBe(3);
  });

  it("handles empty sections array", () => {
    const { container } = render(
      <StructureChart sections={[]} totalChars={0} />
    );
    const progressBars = container.querySelectorAll(".bg-gradient-to-r");
    expect(progressBars.length).toBe(0);
  });

  it("displays guide text for each section", () => {
    render(<StructureChart sections={mockSections} totalChars={1000} />);
    expect(screen.getByText("도입 가이드")).toBeInTheDocument();
    expect(screen.getByText("본문 가이드")).toBeInTheDocument();
    expect(screen.getByText("결론 가이드")).toBeInTheDocument();
  });

  it("displays total character count in heading", () => {
    render(<StructureChart sections={mockSections} totalChars={1000} />);
    expect(screen.getByText("추천 작성 구조 (1000자 기준)")).toBeInTheDocument();
  });
});
