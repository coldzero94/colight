import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { EmptyState } from "../empty-state";

describe("EmptyState", () => {
  it("renders title", () => {
    render(<EmptyState title="데이터가 없습니다" />);
    expect(screen.getByText("데이터가 없습니다")).toBeInTheDocument();
  });

  it("renders description when provided", () => {
    render(
      <EmptyState
        title="경험 없음"
        description="경험을 추가해주세요."
      />
    );
    expect(screen.getByText("경험을 추가해주세요.")).toBeInTheDocument();
  });

  it("does not render description when not provided", () => {
    render(<EmptyState title="제목만" />);
    expect(screen.queryByText("경험을 추가해주세요.")).not.toBeInTheDocument();
  });

  it("renders icon when provided", () => {
    render(<EmptyState title="아이콘 테스트" icon={<span data-testid="icon">📋</span>} />);
    expect(screen.getByTestId("icon")).toBeInTheDocument();
  });

  it("renders action button and handles click", async () => {
    const onClick = vi.fn();
    render(
      <EmptyState
        title="액션 테스트"
        action={{ label: "추가하기", onClick }}
      />
    );

    const button = screen.getByRole("button", { name: "추가하기" });
    expect(button).toBeInTheDocument();

    await userEvent.click(button);
    expect(onClick).toHaveBeenCalledOnce();
  });

  it("does not render button when no action provided", () => {
    render(<EmptyState title="버튼 없음" />);
    expect(screen.queryByRole("button")).not.toBeInTheDocument();
  });
});
