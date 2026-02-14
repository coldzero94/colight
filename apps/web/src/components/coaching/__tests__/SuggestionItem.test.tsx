import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { SuggestionItem } from "../review/suggestion-item";

const mockSuggestion = {
  original: "팀 프로젝트를 했습니다",
  suggested: "5명의 팀에서 백엔드 리드를 맡아 API 설계를 주도했습니다",
  reason: "구체적 역할과 규모를 명시하면 신뢰도가 높아집니다",
};

describe("SuggestionItem", () => {
  it("renders original and suggested text", () => {
    render(<SuggestionItem suggestion={mockSuggestion} onApply={vi.fn()} />);
    expect(screen.getByText(mockSuggestion.original)).toBeInTheDocument();
    expect(screen.getByText(mockSuggestion.suggested)).toBeInTheDocument();
  });

  it("renders reason", () => {
    render(<SuggestionItem suggestion={mockSuggestion} onApply={vi.fn()} />);
    expect(screen.getByText(mockSuggestion.reason)).toBeInTheDocument();
  });

  it("calls onApply when apply button is clicked", async () => {
    const user = userEvent.setup();
    const onApply = vi.fn();
    render(<SuggestionItem suggestion={mockSuggestion} onApply={onApply} />);

    await user.click(screen.getByText("적용하기"));

    expect(onApply).toHaveBeenCalledWith(mockSuggestion);
  });
});
