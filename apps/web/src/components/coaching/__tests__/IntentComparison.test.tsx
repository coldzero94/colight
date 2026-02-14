import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import { IntentComparison } from "../intent-comparison";

describe("IntentComparison", () => {
  const mockIntents = [
    { intent: "갈등 해결 능력 확인", why: "팀워크 중시" },
    { intent: "주도적 문제 인식", why: "능동성 평가" },
    { intent: "성과 측정 역량", why: "정량적 사고" },
  ];

  it("shows surface question and three real intents", () => {
    render(
      <IntentComparison
        surface="팀 프로젝트에서 어려움을 극복한 경험"
        intents={mockIntents}
      />
    );

    // Surface question
    expect(
      screen.getByText(/팀 프로젝트에서 어려움을 극복한 경험/)
    ).toBeInTheDocument();

    // All three intents
    expect(screen.getByText(/갈등 해결 능력 확인/)).toBeInTheDocument();
    expect(screen.getByText(/팀워크 중시/)).toBeInTheDocument();

    expect(screen.getByText(/주도적 문제 인식/)).toBeInTheDocument();
    expect(screen.getByText(/능동성 평가/)).toBeInTheDocument();

    expect(screen.getByText(/성과 측정 역량/)).toBeInTheDocument();
    expect(screen.getByText(/정량적 사고/)).toBeInTheDocument();
  });

  it("displays exactly 3 intents", () => {
    render(
      <IntentComparison
        surface="표면 질문"
        intents={mockIntents}
      />
    );

    // Find all intent items (using data-testid or role)
    const intentItems = screen.getAllByTestId("intent-item");
    expect(intentItems).toHaveLength(3);
  });
});
