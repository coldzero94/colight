import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { OutdatedMatchingBanner } from "../outdated-matching-banner";

describe("OutdatedMatchingBanner", () => {
  it("displays outdated banner with reason text", () => {
    render(
      <OutdatedMatchingBanner
        reason="매칭 이후 경험이 변경되었습니다"
        onRefresh={vi.fn()}
      />
    );

    expect(
      screen.getByText("매칭 결과가 최신이 아닐 수 있습니다")
    ).toBeInTheDocument();
    expect(
      screen.getByText("매칭 이후 경험이 변경되었습니다")
    ).toBeInTheDocument();
    expect(screen.getByText("다시 매칭하기")).toBeInTheDocument();
  });

  it("calls onRefresh when re-match button clicked", async () => {
    const user = userEvent.setup();
    const onRefresh = vi.fn();

    render(
      <OutdatedMatchingBanner
        reason="매칭 결과가 7일 이상 지났습니다"
        onRefresh={onRefresh}
      />
    );

    const button = screen.getByText("다시 매칭하기");
    await user.click(button);

    expect(onRefresh).toHaveBeenCalledTimes(1);
  });

  it("hides banner when reason is empty", () => {
    const { container } = render(
      <OutdatedMatchingBanner reason="" onRefresh={vi.fn()} />
    );

    expect(container.firstChild).toBeNull();
  });

  it("shows refreshing state when isRefreshing is true", () => {
    render(
      <OutdatedMatchingBanner
        reason="경험이 변경되었습니다"
        onRefresh={vi.fn()}
        isRefreshing={true}
      />
    );

    const button = screen.getByText("매칭 중...");
    expect(button).toBeDisabled();
  });

  it("has alert role for accessibility", () => {
    render(
      <OutdatedMatchingBanner
        reason="Test reason"
        onRefresh={vi.fn()}
      />
    );

    expect(screen.getByRole("alert")).toBeInTheDocument();
  });
});
