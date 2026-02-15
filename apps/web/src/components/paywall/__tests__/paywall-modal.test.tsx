import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { PaywallModal } from "../paywall-modal";

describe("PaywallModal", () => {
  it("renders when open", () => {
    render(
      <PaywallModal open onOpenChange={vi.fn()} feature="draft" used={1} limit={1} />,
    );
    expect(screen.getByText("무료 사용 횟수 초과")).toBeInTheDocument();
    expect(screen.getByText("초안 생성")).toBeInTheDocument();
    expect(screen.getByText("1/1회 사용")).toBeInTheDocument();
  });

  it("shows plan options", () => {
    render(<PaywallModal open onOpenChange={vi.fn()} />);
    expect(screen.getByText("Starter")).toBeInTheDocument();
    expect(screen.getByText("Pro")).toBeInTheDocument();
  });

  it("has link to pricing page", () => {
    render(<PaywallModal open onOpenChange={vi.fn()} />);
    const link = screen.getByRole("link", { name: "플랜 보기" });
    expect(link).toHaveAttribute("href", "/pricing");
  });

  it("calls onOpenChange when dismiss clicked", async () => {
    const user = userEvent.setup();
    const onOpenChange = vi.fn();
    render(<PaywallModal open onOpenChange={onOpenChange} />);

    await user.click(screen.getByText("닫기"));
    expect(onOpenChange).toHaveBeenCalledWith(false);
  });
});
