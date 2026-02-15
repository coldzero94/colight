import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { CharCoachingResult } from "../char-coaching-result";
import type { CharCoachingResult as CharCoachingResultType } from "@/lib/api/coaching";

const overResult: CharCoachingResultType = {
  status: "over",
  current_count: 850,
  char_limit: 800,
  diff: 50,
  suggestions: [
    {
      type: "trim",
      section: "도입부",
      original: "저는 대학교 시절부터 다양한 프로젝트를 수행하면서",
      suggested: "대학 시절 다양한 프로젝트를 통해",
      reason: "불필요한 수식어 제거",
      char_diff: -12,
    },
  ],
  summary: "50자 초과입니다. 도입부를 간결하게 정리해보세요.",
};

const underResult: CharCoachingResultType = {
  status: "under",
  current_count: 500,
  char_limit: 800,
  diff: -300,
  suggestions: [
    {
      type: "expand",
      section: "결과 부분",
      original: "프로젝트가 성공했습니다.",
      suggested:
        "프로젝트 완료 후 사용자 만족도가 30% 향상되었습니다.",
      reason: "구체적 수치 추가",
      char_diff: 18,
    },
  ],
  summary: "300자 부족합니다.",
};

const goodResult: CharCoachingResultType = {
  status: "good",
  current_count: 790,
  char_limit: 800,
  diff: -10,
  suggestions: [],
  summary: "적정 범위입니다.",
};

describe("CharCoachingResult", () => {
  it("renders over status badge and diff", () => {
    render(
      <CharCoachingResult result={overResult} onApplySuggestion={vi.fn()} />
    );
    expect(screen.getByText("초과")).toBeInTheDocument();
    expect(screen.getByText("850")).toBeInTheDocument();
    expect(screen.getByText("800자")).toBeInTheDocument();
    expect(screen.getByText("(+50)")).toBeInTheDocument();
  });

  it("renders under status badge and diff", () => {
    render(
      <CharCoachingResult result={underResult} onApplySuggestion={vi.fn()} />
    );
    expect(screen.getByText("부족")).toBeInTheDocument();
    expect(screen.getByText("(-300)")).toBeInTheDocument();
  });

  it("renders good status badge", () => {
    render(
      <CharCoachingResult result={goodResult} onApplySuggestion={vi.fn()} />
    );
    expect(screen.getByText("적정")).toBeInTheDocument();
  });

  it("renders summary text", () => {
    render(
      <CharCoachingResult result={overResult} onApplySuggestion={vi.fn()} />
    );
    expect(
      screen.getByText("50자 초과입니다. 도입부를 간결하게 정리해보세요.")
    ).toBeInTheDocument();
  });

  it("renders suggestion with original and suggested text", () => {
    render(
      <CharCoachingResult result={overResult} onApplySuggestion={vi.fn()} />
    );
    expect(screen.getByText("수정 제안 (1건)")).toBeInTheDocument();
    expect(screen.getByText("도입부")).toBeInTheDocument();
    expect(
      screen.getByText("저는 대학교 시절부터 다양한 프로젝트를 수행하면서")
    ).toBeInTheDocument();
    expect(
      screen.getByText("대학 시절 다양한 프로젝트를 통해")
    ).toBeInTheDocument();
    expect(screen.getByText("축약 -12자")).toBeInTheDocument();
  });

  it("calls onApplySuggestion when clicking apply button", async () => {
    const user = userEvent.setup();
    const handler = vi.fn();
    render(
      <CharCoachingResult result={overResult} onApplySuggestion={handler} />
    );

    await user.click(screen.getByText("적용"));
    expect(handler).toHaveBeenCalledWith(overResult.suggestions[0]);
  });

  it("shows no suggestions section when suggestions are empty", () => {
    render(
      <CharCoachingResult result={goodResult} onApplySuggestion={vi.fn()} />
    );
    expect(screen.queryByText(/수정 제안/)).not.toBeInTheDocument();
  });

  it("renders expand suggestion type correctly", () => {
    render(
      <CharCoachingResult result={underResult} onApplySuggestion={vi.fn()} />
    );
    expect(screen.getByText("보강 +18자")).toBeInTheDocument();
  });
});
