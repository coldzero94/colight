import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { TaggingStatus } from "../tagging-status";

describe("TaggingStatus", () => {
  const onRetry = vi.fn();

  it("renders nothing for idle state", () => {
    const { container } = render(
      <TaggingStatus status="idle" onRetry={onRetry} />
    );
    expect(container.firstChild).toBeNull();
  });

  it("renders nothing for success state", () => {
    const { container } = render(
      <TaggingStatus status="success" onRetry={onRetry} />
    );
    expect(container.firstChild).toBeNull();
  });

  it("shows loading skeleton during tagging", () => {
    render(<TaggingStatus status="tagging" onRetry={onRetry} />);

    expect(
      screen.getByText(/AI가 경험을 분석하고 있어요/)
    ).toBeInTheDocument();
  });

  it("shows error message with retry button on failure", async () => {
    const user = userEvent.setup();
    render(<TaggingStatus status="error" onRetry={onRetry} />);

    expect(screen.getByText("무기 분석에 실패했습니다.")).toBeInTheDocument();

    const retryButton = screen.getByText("다시 시도");
    await user.click(retryButton);

    expect(onRetry).toHaveBeenCalledOnce();
  });
});
