import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import { Header } from "../header";
import { usePathname } from "next/navigation";

// Mock next/navigation
vi.mock("next/navigation", () => ({
  usePathname: vi.fn(),
}));

// Mock child components
vi.mock("../mobile-nav", () => ({
  MobileNav: () => <div data-testid="mobile-nav">MobileNav</div>,
}));

vi.mock("../user-dropdown", () => ({
  UserDropdown: () => <div data-testid="user-dropdown">UserDropdown</div>,
}));

describe("Header", () => {
  it("renders page title based on pathname", () => {
    vi.mocked(usePathname).mockReturnValue("/experiences");

    render(<Header />);
    expect(screen.getByText("경험 관리")).toBeInTheDocument();
  });

  it("renders interview page title", () => {
    vi.mocked(usePathname).mockReturnValue("/interview");

    render(<Header />);
    expect(screen.getByText("경험 인터뷰")).toBeInTheDocument();
  });

  it("renders analysis page title", () => {
    vi.mocked(usePathname).mockReturnValue("/analysis");

    render(<Header />);
    expect(screen.getByText("기업 분석")).toBeInTheDocument();
  });

  it("renders coaching page title", () => {
    vi.mocked(usePathname).mockReturnValue("/coaching");

    render(<Header />);
    expect(screen.getByText("자소서 코칭")).toBeInTheDocument();
  });

  it("renders dashboard page title", () => {
    vi.mocked(usePathname).mockReturnValue("/dashboard");

    render(<Header />);
    expect(screen.getByText("대시보드")).toBeInTheDocument();
  });

  it("renders empty title for unknown routes", () => {
    vi.mocked(usePathname).mockReturnValue("/unknown-route");

    render(<Header />);
    const heading = screen.queryByRole("heading");
    expect(heading).toHaveTextContent("");
  });

  it("renders MobileNav component", () => {
    vi.mocked(usePathname).mockReturnValue("/experiences");

    render(<Header />);
    expect(screen.getByTestId("mobile-nav")).toBeInTheDocument();
  });

  it("renders UserDropdown component", () => {
    vi.mocked(usePathname).mockReturnValue("/experiences");

    render(<Header />);
    expect(screen.getByTestId("user-dropdown")).toBeInTheDocument();
  });

  it("is wrapped in header element", () => {
    vi.mocked(usePathname).mockReturnValue("/experiences");

    const { container } = render(<Header />);
    const header = container.querySelector("header");
    expect(header).toBeInTheDocument();
  });

  it("matches title for nested routes", () => {
    vi.mocked(usePathname).mockReturnValue("/experiences/new");

    render(<Header />);
    expect(screen.getByText("경험 관리")).toBeInTheDocument();
  });
});
