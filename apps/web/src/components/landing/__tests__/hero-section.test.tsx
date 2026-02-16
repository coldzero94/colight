import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import { HeroSection } from "../hero-section";

describe("HeroSection", () => {
  it("renders main headline", () => {
    render(<HeroSection />);
    expect(screen.getByText(/AI 코치/)).toBeInTheDocument();
  });

  it("renders CTA button linking to /signup", () => {
    render(<HeroSection />);
    const cta = screen.getByRole("link", { name: /무료로 시작하기/ });
    expect(cta).toHaveAttribute("href", "/signup");
  });

  it("renders login link", () => {
    render(<HeroSection />);
    const login = screen.getByRole("link", { name: "로그인" });
    expect(login).toHaveAttribute("href", "/login");
  });

  it("renders badge with platform description", () => {
    render(<HeroSection />);
    expect(screen.getByText("AI 기반 취업 코칭 플랫폼")).toBeInTheDocument();
  });

  it("renders product preview mockup", () => {
    const { container } = render(<HeroSection />);
    // Product mockup has window-style dots
    const dots = container.querySelectorAll(".rounded-full.bg-white\\/10");
    expect(dots.length).toBeGreaterThanOrEqual(3);
  });
});
