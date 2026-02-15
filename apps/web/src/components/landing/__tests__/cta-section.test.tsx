import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import { CTASection } from "../cta-section";

// Mock next/link
vi.mock("next/link", () => ({
  default: ({
    children,
    href,
    ...props
  }: {
    children: React.ReactNode;
    href: string;
    [key: string]: unknown;
  }) => (
    <a href={href} {...props}>
      {children}
    </a>
  ),
}));

describe("CTASection", () => {
  it("renders heading", () => {
    render(<CTASection />);
    expect(screen.getByText("지금 무료로 시작하세요")).toBeInTheDocument();
  });

  it("renders description text", () => {
    render(<CTASection />);
    expect(
      screen.getByText("경험 3건, 일 1회 분석·코칭 무료 제공")
    ).toBeInTheDocument();
  });

  it("renders signup button link", () => {
    render(<CTASection />);
    const signupLink = screen.getByRole("link", { name: /무료 회원가입/ });
    expect(signupLink).toBeInTheDocument();
    expect(signupLink).toHaveAttribute("href", "/signup");
  });

  it("renders button text with arrow icon", () => {
    render(<CTASection />);
    expect(screen.getByText("무료 회원가입")).toBeInTheDocument();
  });

  it("has dark background section", () => {
    const { container } = render(<CTASection />);
    const section = container.querySelector("section");
    expect(section).toHaveClass("bg-gray-900");
  });
});
