import { describe, it, expect, vi } from "vitest";
import { render, screen, fireEvent } from "@testing-library/react";
import { EditorLayout } from "../editor-layout";

const mockCoverLetter = {
  id: "cl-id",
  question_text: "팀 프로젝트에서 어려움을 극복한 경험을 기술하세요",
  char_limit: 800,
  current_content: "[상황]\n초안 내용",
  company_name: "삼성전자",
  position: "소프트웨어 개발직",
};

const mockAnalysis = {
  required_weapons: {
    primary: { weapon_id: "W01", weapon_name: "문제해결", reason: "핵심" },
    secondary: [{ weapon_id: "W02", weapon_name: "협업", reason: "보조" }],
  },
  writing_structure: {
    total_chars: 800,
    sections: [
      { name: "상황", char_ratio: 0.2, char_count: 160, guide: "상황 설정" },
      { name: "과제", char_ratio: 0.15, char_count: 120, guide: "과제 정의" },
      { name: "행동", char_ratio: 0.4, char_count: 320, guide: "실행 과정" },
      { name: "결과", char_ratio: 0.25, char_count: 200, guide: "성과" },
    ],
  },
  key_keywords: ["데이터 기반", "개선율"],
};

const mockExperiences = [
  {
    id: "exp-1",
    title: "캡스톤 프로젝트",
    category: "프로젝트",
    star_situation: "팀 프로젝트 상황",
    weapons: [{ name: "문제해결" }, { name: "협업" }],
    matchScore: 92,
  },
];

describe("EditorLayout", () => {
  it("loads latest version content into editor", () => {
    const onSave = vi.fn();

    render(
      <EditorLayout
        coverLetter={mockCoverLetter}
        analysis={mockAnalysis}
        experiences={mockExperiences}
        onSave={onSave}
      />
    );

    expect(screen.getByText(/초안 내용/)).toBeInTheDocument();
  });

  it("displays analysis sidebar with weapon badges and structure", () => {
    const onSave = vi.fn();

    render(
      <EditorLayout
        coverLetter={mockCoverLetter}
        analysis={mockAnalysis}
        experiences={mockExperiences}
        onSave={onSave}
      />
    );

    // Weapon badges
    const problemSolving = screen.getAllByText("문제해결");
    expect(problemSolving.length).toBeGreaterThan(0);

    const teamwork = screen.getAllByText("협업");
    expect(teamwork.length).toBeGreaterThan(0);

    // Writing structure
    expect(screen.getByText("상황")).toBeInTheDocument();
    expect(screen.getByText(/20%.*160/)).toBeInTheDocument();
    expect(screen.getByText("과제")).toBeInTheDocument();
    expect(screen.getByText(/15%.*120/)).toBeInTheDocument();
  });

  it("shows company name and question in header", () => {
    const onSave = vi.fn();

    render(
      <EditorLayout
        coverLetter={mockCoverLetter}
        analysis={mockAnalysis}
        experiences={mockExperiences}
        onSave={onSave}
      />
    );

    const heading = screen.getByRole("heading", { level: 1 });
    expect(heading).toHaveTextContent("삼성전자");
    expect(heading).toHaveTextContent("소프트웨어 개발직");
    expect(screen.getByText(/팀 프로젝트에서 어려움을 극복한 경험/)).toBeInTheDocument();
  });

  it("displays save status indicator", () => {
    const onSave = vi.fn();

    render(
      <EditorLayout
        coverLetter={mockCoverLetter}
        analysis={mockAnalysis}
        experiences={mockExperiences}
        onSave={onSave}
      />
    );

    expect(screen.getByTestId("save-indicator")).toBeInTheDocument();
  });

  it("shows tabbed sidebar with analysis and review tabs", () => {
    render(
      <EditorLayout
        coverLetter={mockCoverLetter}
        analysis={mockAnalysis}
        experiences={mockExperiences}
        onSave={vi.fn()}
      />
    );

    expect(screen.getByRole("tab", { name: "분석" })).toBeInTheDocument();
    expect(screen.getByRole("tab", { name: "첨삭 결과" })).toBeInTheDocument();
  });

  it("shows review request button when onRequestReview is provided", () => {
    const onRequestReview = vi.fn();

    render(
      <EditorLayout
        coverLetter={mockCoverLetter}
        analysis={mockAnalysis}
        experiences={mockExperiences}
        onSave={vi.fn()}
        onRequestReview={onRequestReview}
      />
    );

    const reviewButton = screen.getByRole("button", { name: /첨삭 요청/ });
    expect(reviewButton).toBeInTheDocument();

    fireEvent.click(reviewButton);
    expect(onRequestReview).toHaveBeenCalledOnce();
  });

  it("shows loading state on review button when isReviewing", () => {
    render(
      <EditorLayout
        coverLetter={mockCoverLetter}
        analysis={mockAnalysis}
        experiences={mockExperiences}
        onSave={vi.fn()}
        onRequestReview={vi.fn()}
        isReviewing={true}
      />
    );

    expect(screen.getByText("첨삭 중...")).toBeInTheDocument();
  });

  it("does not show review button when onRequestReview is not provided", () => {
    render(
      <EditorLayout
        coverLetter={mockCoverLetter}
        analysis={mockAnalysis}
        experiences={mockExperiences}
        onSave={vi.fn()}
      />
    );

    expect(screen.queryByRole("button", { name: /첨삭 요청/ })).not.toBeInTheDocument();
  });
});
