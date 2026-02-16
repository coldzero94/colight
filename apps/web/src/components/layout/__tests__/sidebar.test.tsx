import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";

vi.mock("next/link", () => ({
  default: ({
    children,
    href,
    ...props
  }: {
    children: React.ReactNode;
    href: string;
    className?: string;
  }) => (
    <a href={href} {...props}>
      {children}
    </a>
  ),
}));

let mockPathname = "/experiences";
vi.mock("next/navigation", () => ({
  usePathname: () => mockPathname,
}));

const mockUseAuthStore = vi.fn();
vi.mock("@/stores/auth-store", () => ({
  useAuthStore: (...args: unknown[]) => mockUseAuthStore(...args),
}));

import { Sidebar } from "../sidebar";

describe("Sidebar", () => {
  it("renders all navigation items", () => {
    mockUseAuthStore.mockReturnValue({ user: { role: "user" } });

    render(<Sidebar />);

    expect(screen.getByText("경험 관리")).toBeInTheDocument();
    expect(screen.getByText("기업 분석")).toBeInTheDocument();
    expect(screen.getByText("자소서 코칭")).toBeInTheDocument();
    expect(screen.getByText("대시보드")).toBeInTheDocument();
  });

  it("shows admin link for admin users", () => {
    mockUseAuthStore.mockReturnValue({
      user: { role: "admin", nickname: "관리자" },
    });

    render(<Sidebar />);

    expect(screen.getByText("어드민")).toBeInTheDocument();
  });

  it("hides admin link for regular users", () => {
    mockUseAuthStore.mockReturnValue({
      user: { role: "user", nickname: "유저" },
    });

    render(<Sidebar />);

    expect(screen.queryByText("어드민")).not.toBeInTheDocument();
  });

  it("shows user nickname and avatar initial", () => {
    mockUseAuthStore.mockReturnValue({
      user: { role: "user", nickname: "홍길동" },
    });

    render(<Sidebar />);

    expect(screen.getByText("홍길동")).toBeInTheDocument();
    expect(screen.getByText("홍")).toBeInTheDocument();
  });

  it("shows default text when no user info", () => {
    mockUseAuthStore.mockReturnValue({ user: null });

    render(<Sidebar />);

    expect(screen.getByText("사용자")).toBeInTheDocument();
    expect(screen.getByText("U")).toBeInTheDocument();
  });

  it("highlights active navigation item", () => {
    mockPathname = "/experiences";
    mockUseAuthStore.mockReturnValue({ user: { role: "user" } });

    render(<Sidebar />);

    const experiencesLink = screen.getByText("경험 관리").closest("a");
    expect(experiencesLink?.className).toContain("bg-primary/14");
    expect(experiencesLink?.className).toContain("text-primary");
  });
});
