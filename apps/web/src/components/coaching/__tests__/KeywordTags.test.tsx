import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import { KeywordTags } from "../keyword-tags";

describe("KeywordTags", () => {
  const mockKeywords = ["데이터 기반", "개선율", "주도적", "소통", "성과"];
  const mockAvoidList = ["열심히", "최선을 다했습니다", "노력했습니다"];

  it("renders keyword tags and avoid-list in warning style", () => {
    render(
      <KeywordTags keywords={mockKeywords} avoidList={mockAvoidList} />
    );

    // Verify all keywords are rendered
    expect(screen.getByText("데이터 기반")).toBeInTheDocument();
    expect(screen.getByText("개선율")).toBeInTheDocument();
    expect(screen.getByText("주도적")).toBeInTheDocument();

    // Verify avoid list heading
    expect(screen.getByText(/피해야 할 표현/)).toBeInTheDocument();

    // Verify avoid expressions
    expect(screen.getByText(/열심히/)).toBeInTheDocument();
    expect(screen.getByText(/최선을 다했습니다/)).toBeInTheDocument();
  });

  it("displays avoid expressions with red/orange styling", () => {
    render(
      <KeywordTags keywords={mockKeywords} avoidList={mockAvoidList} />
    );

    // Find avoid list container and check for warning styling
    const avoidSection = screen.getByText(/피해야 할 표현/).closest("div");
    expect(avoidSection).toBeInTheDocument();

    // Avoid expressions should have warning styling (red/orange text)
    const avoidItems = screen.getAllByTestId("avoid-item");
    expect(avoidItems.length).toBe(3);

    // Check if avoid items have warning class (text-red or text-orange)
    avoidItems.forEach((item) => {
      expect(item.className).toMatch(/text-(red|orange)/);
    });
  });

  it("renders empty state when no keywords", () => {
    render(<KeywordTags keywords={[]} avoidList={[]} />);

    // Should still render the section but with empty state or placeholder
    const keywordSection = screen.queryByText(/핵심 키워드/);
    expect(keywordSection).toBeInTheDocument();
  });
});
