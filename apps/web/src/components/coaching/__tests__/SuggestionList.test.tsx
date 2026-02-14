import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import { SuggestionList } from "../review/suggestion-list";

describe("SuggestionList", () => {
  it("renders empty state when no suggestions", () => {
    render(<SuggestionList suggestions={[]} onApply={vi.fn()} />);
    expect(screen.getByText("수정 제안이 없습니다")).toBeInTheDocument();
  });

  it("renders suggestion count and items", () => {
    const suggestions = [
      {
        original: "원문1",
        suggested: "수정1",
        reason: "이유1",
      },
      {
        original: "원문2",
        suggested: "수정2",
        reason: "이유2",
      },
    ];

    render(<SuggestionList suggestions={suggestions} onApply={vi.fn()} />);
    expect(screen.getByText("수정 제안 (2)")).toBeInTheDocument();
    expect(screen.getByText("원문1")).toBeInTheDocument();
    expect(screen.getByText("원문2")).toBeInTheDocument();
  });
});
