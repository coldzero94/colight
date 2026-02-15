import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import { LandingHeader } from "../landing-header";

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

describe("LandingHeader", () => {
  it("renders logo", () => {
    render(<LandingHeader />);
    const logo = screen.getByRole("link", { name: "Colight" });
    expect(logo).toBeInTheDocument();
    expect(logo).toHaveAttribute("href", "/");
  });

  it("renders login button", () => {
    render(<LandingHeader />);
    const loginButton = screen.getByRole("link", { name: "로그인" });
    expect(loginButton).toBeInTheDocument();
    expect(loginButton).toHaveAttribute("href", "/login");
  });

  it("renders signup button", () => {
    render(<LandingHeader />);
    const signupButton = screen.getByRole("link", { name: "시작하기" });
    expect(signupButton).toBeInTheDocument();
    expect(signupButton).toHaveAttribute("href", "/signup");
  });

  it("is wrapped in header element", () => {
    const { container } = render(<LandingHeader />);
    const header = container.querySelector("header");
    expect(header).toBeInTheDocument();
  });

  it("has sticky positioning", () => {
    const { container } = render(<LandingHeader />);
    const header = container.querySelector("header");
    expect(header).toHaveClass("sticky");
  });
});
