import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import { AnalysisResult } from "../analysis-result";

const mockAnalysis = {
  surface_question: "팀 프로젝트에서 어려움을 극복한 경험",
  real_intents: [
    { intent: "갈등 해결 능력", why: "팀워크 중시" },
    { intent: "주도적 문제 인식", why: "능동성 평가" },
    { intent: "성과 측정 역량", why: "정량적 사고" },
  ],
  required_weapons: {
    primary: {
      weapon_id: "W01",
      weapon_name: "문제해결",
      reason: "핵심 역량",
    },
    secondary: [
      {
        weapon_id: "W02",
        weapon_name: "협업",
        reason: "팀 경험",
      },
    ],
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
  key_keywords: ["데이터 기반", "개선율", "주도적", "소통", "성과"],
  avoid_list: ["열심히", "최선", "노력"],
  good_structure_example: "상황: 팀 프로젝트에서...\n과제: 해결해야 할...",
};

describe("AnalysisResult", () => {
  it("renders all analysis fields from API response", () => {
    render(<AnalysisResult analysis={mockAnalysis} />);

    // Surface question
    expect(screen.getByText(mockAnalysis.surface_question)).toBeInTheDocument();

    // Real intents
    expect(screen.getByText(/갈등 해결 능력/)).toBeInTheDocument();
    expect(screen.getByText(/주도적 문제 인식/)).toBeInTheDocument();
    expect(screen.getByText(/성과 측정 역량/)).toBeInTheDocument();

    // Weapons
    expect(screen.getByText("문제해결")).toBeInTheDocument();
    expect(screen.getByText("협업")).toBeInTheDocument();

    // Keywords
    expect(screen.getByText(/데이터 기반/)).toBeInTheDocument();
    expect(screen.getByText(/개선율/)).toBeInTheDocument();
  });

  it("highlights primary weapon badge visually", () => {
    render(<AnalysisResult analysis={mockAnalysis} />);

    // Primary weapon should have special styling or indicator
    const primaryWeapon = screen.getByText("문제해결");
    expect(primaryWeapon).toBeInTheDocument();
    // Check if parent has special class or data-attribute for primary
    const badge = primaryWeapon.closest("[data-primary='true']");
    expect(badge).toBeInTheDocument();
  });

  it("displays structure chart with correct ratios summing to 100%", () => {
    render(<AnalysisResult analysis={mockAnalysis} />);

    // Verify all percentages are displayed (use getAllByText for duplicates)
    const percentages = [
      screen.getByText(/20%/),
      screen.getByText(/15%/),
      screen.getByText(/40%/),
      screen.getByText(/25%/),
    ];

    percentages.forEach((p) => {
      expect(p).toBeInTheDocument();
    });

    // Verify structure heading
    expect(screen.getByText(/추천 작성 구조/)).toBeInTheDocument();
    expect(screen.getByText(/800자 기준/)).toBeInTheDocument();
  });
});
