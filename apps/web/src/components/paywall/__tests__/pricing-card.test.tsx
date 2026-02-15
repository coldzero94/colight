import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { PricingCard } from "../pricing-card";

describe("PricingCard", () => {
  const defaultProps = {
    name: "Pro",
    price: "₩19,900/월",
    description: "합격률을 높이고 싶은 분을 위해",
    features: ["모든 기능 무제한", "우선 AI 처리"],
    cta: "시작하기",
    onCtaClick: vi.fn(),
  };

  it("renders plan info", () => {
    render(<PricingCard {...defaultProps} />);
    expect(screen.getByText("Pro")).toBeInTheDocument();
    expect(screen.getByText("₩19,900/월")).toBeInTheDocument();
    expect(screen.getByText("모든 기능 무제한")).toBeInTheDocument();
    expect(screen.getByText("우선 AI 처리")).toBeInTheDocument();
  });

  it("calls onCtaClick when CTA clicked", async () => {
    const user = userEvent.setup();
    const onClick = vi.fn();
    render(<PricingCard {...defaultProps} onCtaClick={onClick} />);

    await user.click(screen.getByText("시작하기"));
    expect(onClick).toHaveBeenCalledOnce();
  });

  it("shows disabled current plan button", () => {
    render(<PricingCard {...defaultProps} current />);
    const btn = screen.getByText("현재 플랜");
    expect(btn).toBeDisabled();
  });

  it("applies highlighted border", () => {
    const { container } = render(
      <PricingCard {...defaultProps} highlighted />,
    );
    const card = container.firstChild as HTMLElement;
    expect(card.className).toContain("border-gray-900");
  });
});
