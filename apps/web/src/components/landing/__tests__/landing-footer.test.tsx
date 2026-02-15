import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import { LandingFooter } from "../landing-footer";

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

describe("LandingFooter", () => {
  it("renders brand name", () => {
    render(<LandingFooter />);
    expect(screen.getAllByText("Colight")[0]).toBeInTheDocument();
  });

  it("renders privacy policy link", () => {
    render(<LandingFooter />);
    const privacyLink = screen.getByRole("link", { name: "개인정보처리방침" });
    expect(privacyLink).toBeInTheDocument();
    expect(privacyLink).toHaveAttribute("href", "/privacy");
  });

  it("renders terms of service link", () => {
    render(<LandingFooter />);
    const termsLink = screen.getByRole("link", { name: "이용약관" });
    expect(termsLink).toBeInTheDocument();
    expect(termsLink).toHaveAttribute("href", "/terms");
  });

  it("renders copyright text", () => {
    render(<LandingFooter />);
    expect(screen.getByText("© 2026 Colight")).toBeInTheDocument();
  });

  it("is wrapped in footer element", () => {
    const { container } = render(<LandingFooter />);
    const footer = container.querySelector("footer");
    expect(footer).toBeInTheDocument();
  });
});
