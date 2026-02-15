import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { ExperienceSelector } from "../experience-selector";

const mockExperiences = [
  {
    id: "1",
    title: "캡스톤 프로젝트",
    category: "프로젝트",
    period_start: "2023-03",
    period_end: "2023-12",
    star_situation: "팀 프로젝트에서 기술 격차 발생",
    star_task: "코드 리뷰 시스템 도입 필요",
    star_action: "주간 세미나와 페어 프로그래밍 진행",
    star_result: "팀 생산성 30% 향상",
    weapons: [{ code: "W01", name: "문제해결" }, { code: "W02", name: "협업" }],
    matchScore: 92,
  },
  {
    id: "2",
    title: "해커톤 경험",
    category: "프로젝트",
    star_situation: "36시간 해커톤 참가",
    star_task: "AI 기반 추천 시스템 구축",
    star_action: "머신러닝 모델 설계 및 구현",
    star_result: "우승 및 기업 협업 제안",
    weapons: [{ code: "W01", name: "문제해결" }, { code: "W05", name: "창의성" }],
    matchScore: 78,
  },
  {
    id: "3",
    title: "인턴 경험",
    category: "인턴",
    star_situation: "스타트업 백엔드 인턴",
    star_task: "API 성능 개선",
    star_action: "쿼리 최적화 및 캐싱 도입",
    star_result: "응답 시간 50% 단축",
    weapons: [{ code: "W03", name: "실행력" }],
    matchScore: 65,
  },
];

const mockRequiredWeapons = {
  primary: { weapon_id: "W01", weapon_name: "문제해결", reason: "핵심" },
  secondary: [{ weapon_id: "W02", weapon_name: "협업", reason: "보조" }],
};

describe("ExperienceSelector", () => {
  it("toggles experience selection on card click", async () => {
    const user = userEvent.setup();
    const onConfirm = vi.fn();

    render(
      <ExperienceSelector
        experiences={mockExperiences}
        maxSelect={3}
        requiredWeapons={mockRequiredWeapons}
        onConfirm={onConfirm}
      />
    );

    const checkboxes = screen.getAllByRole("checkbox");

    // Initially unchecked
    expect(checkboxes[0]).not.toBeChecked();

    // Click to select
    await user.click(checkboxes[0]);
    expect(checkboxes[0]).toBeChecked();

    // Click again to deselect
    await user.click(checkboxes[0]);
    expect(checkboxes[0]).not.toBeChecked();
  });

  it("blocks selection beyond 3 experiences with message", async () => {
    const user = userEvent.setup();
    const onConfirm = vi.fn();

    render(
      <ExperienceSelector
        experiences={mockExperiences}
        maxSelect={3}
        requiredWeapons={mockRequiredWeapons}
        onConfirm={onConfirm}
      />
    );

    const checkboxes = screen.getAllByRole("checkbox");

    // Select first 3
    await user.click(checkboxes[0]);
    await user.click(checkboxes[1]);
    await user.click(checkboxes[2]);

    // All checkboxes should be checked or disabled
    expect(checkboxes[0]).toBeChecked();
    expect(checkboxes[1]).toBeChecked();
    expect(checkboxes[2]).toBeChecked();

    // Message about max selection
    expect(screen.getByText(/3개 선택/)).toBeInTheDocument();
  });

  it("shows validation error when submitting with no selection", async () => {
    const user = userEvent.setup();
    const onConfirm = vi.fn();

    render(
      <ExperienceSelector
        experiences={mockExperiences}
        maxSelect={3}
        requiredWeapons={mockRequiredWeapons}
        onConfirm={onConfirm}
      />
    );

    const confirmButton = screen.getByRole("button", { name: /초안 작성/ });
    expect(confirmButton).toBeDisabled();

    // Try to click disabled button
    await user.click(confirmButton);

    // Should not call onConfirm
    expect(onConfirm).not.toHaveBeenCalled();
  });

  it("expands card to show full STAR content on click", async () => {
    const user = userEvent.setup();
    const onConfirm = vi.fn();

    render(
      <ExperienceSelector
        experiences={mockExperiences}
        maxSelect={3}
        requiredWeapons={mockRequiredWeapons}
        onConfirm={onConfirm}
      />
    );

    // Initially, full STAR content is not visible
    expect(screen.queryByText("주간 세미나와 페어 프로그래밍 진행")).not.toBeInTheDocument();

    // Click expand button or card
    const expandButtons = screen.getAllByTestId("expand-button");
    await user.click(expandButtons[0]);

    // Full STAR content should be visible
    expect(screen.getByText(/주간 세미나와 페어 프로그래밍 진행/)).toBeInTheDocument();
    expect(screen.getByText(/팀 생산성 30% 향상/)).toBeInTheDocument();
  });

  it("updates weapon coverage display on selection change", async () => {
    const user = userEvent.setup();
    const onConfirm = vi.fn();

    render(
      <ExperienceSelector
        experiences={mockExperiences}
        maxSelect={3}
        requiredWeapons={mockRequiredWeapons}
        onConfirm={onConfirm}
      />
    );

    // Initially no weapons covered
    expect(screen.queryByTestId("weapon-coverage")).toBeInTheDocument();

    // Select first experience (has W01, W02)
    const checkboxes = screen.getAllByRole("checkbox");
    await user.click(checkboxes[0]);

    // Weapon coverage should update
    const coverage = screen.getByTestId("weapon-coverage");
    expect(coverage).toHaveTextContent("문제해결");
    expect(coverage).toHaveTextContent("협업");
  });

  it("sorts used experiences to bottom", () => {
    const onConfirm = vi.fn();

    const experiencesWithUsage = [
      {
        ...mockExperiences[0],
        id: "used-1",
        title: "사용된 경험",
        matchScore: 90,
        isUsed: true,
      },
      {
        ...mockExperiences[1],
        id: "fresh-1",
        title: "새로운 경험",
        matchScore: 70,
        isUsed: false,
      },
    ];

    render(
      <ExperienceSelector
        experiences={experiencesWithUsage}
        maxSelect={3}
        requiredWeapons={mockRequiredWeapons}
        onConfirm={onConfirm}
      />
    );

    // Cards should render unused first, then used (regardless of score)
    const headings = screen.getAllByRole("heading", { level: 3 });
    expect(headings[0]).toHaveTextContent("새로운 경험");
    expect(headings[1]).toHaveTextContent("사용된 경험");
  });
});
