import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import GlobalError from "../global-error";

describe("GlobalError", () => {
  it("renders error message", () => {
    render(<GlobalError error={new Error("test")} reset={vi.fn()} />);
    expect(
      screen.getByText("서버에 문제가 발생했습니다"),
    ).toBeInTheDocument();
  });

  it("calls reset on retry button click", async () => {
    const user = userEvent.setup();
    const reset = vi.fn();
    render(<GlobalError error={new Error("test")} reset={reset} />);

    await user.click(screen.getByText("다시 시도"));
    expect(reset).toHaveBeenCalledOnce();
  });

  it("has a link to home", () => {
    render(<GlobalError error={new Error("test")} reset={vi.fn()} />);
    const link = screen.getByText("홈으로");
    expect(link).toHaveAttribute("href", "/");
  });
});
