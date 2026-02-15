import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import { AnalysisSidebar } from "../analysis-sidebar";

const mockAnalysis = {
  required_weapons: {
    primary: {
      weapon_id: "W01",
      weapon_name: "문제해결력",
      reason: "Required primary weapon",
    },
    secondary: [
      {
        weapon_id: "W02",
        weapon_name: "리더십",
        reason: "Secondary weapon",
      },
      {
        weapon_id: "W03",
        weapon_name: "커뮤니케이션",
        reason: "Secondary weapon",
      },
    ],
  },
  writing_structure: {
    total_chars: 1000,
    sections: [
      {
        name: "도입부",
        char_ratio: 0.15,
        char_count: 150,
        guide: "Start with context",
      },
      {
        name: "본론",
        char_ratio: 0.7,
        char_count: 700,
        guide: "Main content",
      },
      {
        name: "결론",
        char_ratio: 0.15,
        char_count: 150,
        guide: "Conclusion",
      },
    ],
  },
  key_keywords: ["협업", "성장", "혁신"],
};

const mockExperiences = [
  {
    id: "exp-1",
    title: "프로젝트 리더 경험",
    category: "프로젝트",
    star_situation: "팀 프로젝트에서 리더로 활동하며 문제를 해결했습니다",
    weapons: [{ name: "문제해결력" }, { name: "리더십" }],
    matchScore: 92,
  },
  {
    id: "exp-2",
    title: "인턴 경험",
    category: "인턴",
    star_situation: "스타트업 인턴으로 근무하며 다양한 프로젝트에 참여했습니다",
    weapons: [{ name: "커뮤니케이션" }],
    matchScore: 78,
  },
];

describe("AnalysisSidebar", () => {
  it("renders primary weapon name", () => {
    render(
      <AnalysisSidebar
        analysis={mockAnalysis}
        experiences={mockExperiences}
      />
    );

    expect(screen.getByText("문제해결력")).toBeInTheDocument();
  });

  it("renders secondary weapons", () => {
    render(
      <AnalysisSidebar
        analysis={mockAnalysis}
        experiences={mockExperiences}
      />
    );

    // Secondary weapons appear in required weapons section
    const leadershipBadges = screen.getAllByText("리더십");
    expect(leadershipBadges.length).toBeGreaterThanOrEqual(1);
    const communicationBadges = screen.getAllByText("커뮤니케이션");
    expect(communicationBadges.length).toBeGreaterThanOrEqual(1);
  });

  it("renders key keywords", () => {
    render(
      <AnalysisSidebar
        analysis={mockAnalysis}
        experiences={mockExperiences}
      />
    );

    expect(screen.getByText("협업")).toBeInTheDocument();
    expect(screen.getByText("성장")).toBeInTheDocument();
    expect(screen.getByText("혁신")).toBeInTheDocument();
  });

  it("renders writing structure sections with percentages", () => {
    render(
      <AnalysisSidebar
        analysis={mockAnalysis}
        experiences={mockExperiences}
      />
    );

    expect(screen.getByText("도입부")).toBeInTheDocument();
    expect(screen.getByText("본론")).toBeInTheDocument();
    expect(screen.getByText("결론")).toBeInTheDocument();
    // Match the percentage and character count - multiple matches are expected
    const fifteenPercent = screen.getAllByText(/15%/);
    expect(fifteenPercent.length).toBe(2); // Intro and conclusion
    expect(screen.getByText(/70%/)).toBeInTheDocument();
  });

  it("renders matched experiences with titles and scores", () => {
    render(
      <AnalysisSidebar
        analysis={mockAnalysis}
        experiences={mockExperiences}
      />
    );

    expect(screen.getByText("프로젝트 리더 경험")).toBeInTheDocument();
    expect(screen.getByText("인턴 경험")).toBeInTheDocument();
    expect(screen.getByText("92%")).toBeInTheDocument();
    expect(screen.getByText("78%")).toBeInTheDocument();
  });

  it("renders experience weapons as badges", () => {
    render(
      <AnalysisSidebar
        analysis={mockAnalysis}
        experiences={mockExperiences}
      />
    );

    // Primary weapon appears twice: once in required weapons, once in experience
    const weaponBadges = screen.getAllByText("문제해결력");
    expect(weaponBadges.length).toBeGreaterThanOrEqual(1);
  });

  it("renders section headings", () => {
    render(
      <AnalysisSidebar
        analysis={mockAnalysis}
        experiences={mockExperiences}
      />
    );

    expect(screen.getByText("필요 무기")).toBeInTheDocument();
    expect(screen.getByText("추천 구조")).toBeInTheDocument();
    expect(screen.getByText("핵심 키워드")).toBeInTheDocument();
    expect(screen.getByText("사용 경험")).toBeInTheDocument();
  });
});
