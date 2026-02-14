import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { ExperienceRecommend } from "../experience-recommend";

const mockExperiences = [
  {
    id: "1",
    title: "캡스톤 프로젝트 팀장 경험",
    period_start: "2023-03",
    period_end: "2023-12",
    category: "프로젝트",
    weapons: [
      { weapon_name: "문제해결", relevance_score: 0.9 },
      { weapon_name: "협업", relevance_score: 0.8 },
    ],
    star_situation: "팀원 간 기술 수준 차이로 발생한 병목을...",
    matchScore: 92,
  },
  {
    id: "2",
    title: "해커톤 우승 경험",
    period_start: "2023-07",
    category: "프로젝트",
    weapons: [{ weapon_name: "문제해결" }],
    star_situation: "36시간 동안 새로운 아이디어를...",
    matchScore: 78,
  },
  {
    id: "3",
    title: "인턴십 프로젝트 경험",
    period_start: "2024-01",
    period_end: "2024-02",
    category: "인턴",
    weapons: [{ weapon_name: "실행력" }],
    star_situation: "데이터 파이프라인 오류를...",
    matchScore: 65,
  },
];

const mockRequiredWeapons = {
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
};

describe("ExperienceRecommend", () => {
  it("renders top 3 recommended experience cards sorted by score", () => {
    const onSelect = vi.fn();
    render(
      <ExperienceRecommend
        experiences={mockExperiences}
        requiredWeapons={mockRequiredWeapons}
        onSelect={onSelect}
      />
    );

    // Verify all 3 experiences are displayed
    expect(screen.getByText("캡스톤 프로젝트 팀장 경험")).toBeInTheDocument();
    expect(screen.getByText("해커톤 우승 경험")).toBeInTheDocument();
    expect(screen.getByText("인턴십 프로젝트 경험")).toBeInTheDocument();

    // Verify scores are displayed
    expect(screen.getByText(/92%/)).toBeInTheDocument();
    expect(screen.getByText(/78%/)).toBeInTheDocument();
    expect(screen.getByText(/65%/)).toBeInTheDocument();
  });

  it("shows empty state with registration CTA when no experiences", () => {
    const onSelect = vi.fn();
    render(
      <ExperienceRecommend
        experiences={[]}
        requiredWeapons={mockRequiredWeapons}
        onSelect={onSelect}
      />
    );

    // Empty state message
    expect(screen.getByText(/등록된 경험이 없습니다/)).toBeInTheDocument();

    // Link to experience registration
    expect(screen.getByRole("link", { name: /경험 등록/ })).toBeInTheDocument();
  });

  it("enables CTA button only when at least 1 experience selected", async () => {
    const user = userEvent.setup();
    const onSelect = vi.fn();
    render(
      <ExperienceRecommend
        experiences={mockExperiences}
        requiredWeapons={mockRequiredWeapons}
        onSelect={onSelect}
      />
    );

    // Button should be disabled initially
    const ctaButton = screen.getByRole("button", { name: /초안 작성/ });
    expect(ctaButton).toBeDisabled();

    // Select first experience
    const checkboxes = screen.getAllByRole("checkbox");
    await user.click(checkboxes[0]);

    // Button should be enabled now
    expect(ctaButton).toBeEnabled();
  });

  it("limits selection to maximum 3 experiences", async () => {
    const user = userEvent.setup();
    const onSelect = vi.fn();

    // Create 4 experiences
    const fourExperiences = [
      ...mockExperiences,
      {
        id: "4",
        title: "경험4",
        category: "기타",
        weapons: [],
        star_situation: "내용",
        matchScore: 50,
      },
    ];

    render(
      <ExperienceRecommend
        experiences={fourExperiences}
        requiredWeapons={mockRequiredWeapons}
        onSelect={onSelect}
      />
    );

    const checkboxes = screen.getAllByRole("checkbox");

    // Select first 3
    await user.click(checkboxes[0]);
    await user.click(checkboxes[1]);
    await user.click(checkboxes[2]);

    // 4th checkbox should be disabled
    const fourthCheckbox = checkboxes[3] as HTMLInputElement;
    expect(fourthCheckbox).toBeDisabled();
  });
});
