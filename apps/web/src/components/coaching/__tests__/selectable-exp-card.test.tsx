import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import { SelectableExpCard } from "../selectable-exp-card";

const baseExperience = {
  id: "1",
  title: "캡스톤 프로젝트",
  category: "프로젝트",
  period_start: "2023-03",
  period_end: "2023-12",
  star_situation: "팀 프로젝트에서 기술 격차 발생",
  star_task: "코드 리뷰 시스템 도입 필요",
  star_action: "주간 세미나와 페어 프로그래밍 진행",
  star_result: "팀 생산성 30% 향상",
  weapons: [{ code: "W01", name: "문제해결" }],
  matchScore: 85,
};

describe("SelectableExpCard", () => {
  it("shows match reasons when provided", () => {
    render(
      <SelectableExpCard
        experience={{
          ...baseExperience,
          matchReasons: ["주 무기 '문제해결' 일치 (신뢰도 90%)", "키워드 2개 매칭: 데이터, 분석"],
        }}
        selected={false}
        disabled={false}
        expanded={false}
        onToggle={vi.fn()}
        onExpand={vi.fn()}
      />,
    );

    const reasons = screen.getByTestId("match-reasons");
    expect(reasons).toHaveTextContent("주 무기 '문제해결' 일치");
    expect(reasons).toHaveTextContent("키워드 2개 매칭");
  });

  it("does not show match reasons when empty", () => {
    render(
      <SelectableExpCard
        experience={{ ...baseExperience, matchReasons: [] }}
        selected={false}
        disabled={false}
        expanded={false}
        onToggle={vi.fn()}
        onExpand={vi.fn()}
      />,
    );

    expect(screen.queryByTestId("match-reasons")).not.toBeInTheDocument();
  });

  it("shows 'already used' badge when isUsed is true", () => {
    render(
      <SelectableExpCard
        experience={{ ...baseExperience, isUsed: true }}
        selected={false}
        disabled={false}
        expanded={false}
        onToggle={vi.fn()}
        onExpand={vi.fn()}
      />,
    );

    expect(screen.getByTestId("used-badge")).toHaveTextContent("이미 사용됨");
  });

  it("does not show 'already used' badge when isUsed is false", () => {
    render(
      <SelectableExpCard
        experience={{ ...baseExperience, isUsed: false }}
        selected={false}
        disabled={false}
        expanded={false}
        onToggle={vi.fn()}
        onExpand={vi.fn()}
      />,
    );

    expect(screen.queryByTestId("used-badge")).not.toBeInTheDocument();
  });

  it("uses green color for high scores (>= 70)", () => {
    render(
      <SelectableExpCard
        experience={{ ...baseExperience, matchScore: 85 }}
        selected={false}
        disabled={false}
        expanded={false}
        onToggle={vi.fn()}
        onExpand={vi.fn()}
      />,
    );

    const score = screen.getByTestId("match-score");
    expect(score).toHaveClass("text-green-400");
  });

  it("uses blue color for medium scores (40-69)", () => {
    render(
      <SelectableExpCard
        experience={{ ...baseExperience, matchScore: 55 }}
        selected={false}
        disabled={false}
        expanded={false}
        onToggle={vi.fn()}
        onExpand={vi.fn()}
      />,
    );

    const score = screen.getByTestId("match-score");
    expect(score).toHaveClass("text-blue-400");
  });

  it("uses gray color for low scores (< 40)", () => {
    render(
      <SelectableExpCard
        experience={{ ...baseExperience, matchScore: 25 }}
        selected={false}
        disabled={false}
        expanded={false}
        onToggle={vi.fn()}
        onExpand={vi.fn()}
      />,
    );

    const score = screen.getByTestId("match-score");
    expect(score).toHaveClass("text-muted-foreground");
  });
});
