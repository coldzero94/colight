import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen } from "@testing-library/react";

let mockPathname = "/admin";
vi.mock("next/navigation", () => ({
  usePathname: () => mockPathname,
}));

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

import type { UserRole } from "@/lib/auth-utils";

const mockAuthStore = vi.hoisted(() => ({
  user: { id: "1", role: "super_admin" as UserRole },
}));

vi.mock("@/stores/auth-store", () => ({
  useAuthStore: () => mockAuthStore,
}));

import { AdminSidebar } from "../admin-sidebar";

describe("AdminSidebar", () => {
  beforeEach(() => {
    mockAuthStore.user = { id: "1", role: "super_admin" as UserRole };
  });

  it("renders all admin menu items for super_admin", () => {
    render(<AdminSidebar />);

    expect(screen.getByText("대시보드")).toBeInTheDocument();
    expect(screen.getByText("사용자 관리")).toBeInTheDocument();
    expect(screen.getByText("프롬프트 관리")).toBeInTheDocument();
    expect(screen.getByText("시스템 설정")).toBeInTheDocument();
    expect(screen.getByText("사용량 모니터링")).toBeInTheDocument();
    expect(screen.getByText("감사 로그")).toBeInTheDocument();
  });

  it("hides super_admin-only items for admin role", () => {
    mockAuthStore.user = { id: "1", role: "admin" as UserRole };
    render(<AdminSidebar />);

    expect(screen.getByText("대시보드")).toBeInTheDocument();
    expect(screen.getByText("사용자 관리")).toBeInTheDocument();
    expect(screen.getByText("프롬프트 관리")).toBeInTheDocument();
    expect(screen.queryByText("시스템 설정")).not.toBeInTheDocument();
  });

  it("renders back to main app link", () => {
    render(<AdminSidebar />);

    const backLink = screen.getByText(/메인 앱으로/);
    expect(backLink.closest("a")).toHaveAttribute("href", "/experiences");
  });

  it("highlights /admin exactly for dashboard", () => {
    mockPathname = "/admin";
    render(<AdminSidebar />);

    const dashboardLink = screen.getByText("대시보드").closest("a");
    expect(dashboardLink?.className).toContain("bg-primary/10");
  });

  it("does not highlight /admin for /admin/users", () => {
    mockPathname = "/admin/users";
    render(<AdminSidebar />);

    const dashboardLink = screen.getByText("대시보드").closest("a");
    expect(dashboardLink?.className).not.toContain("bg-primary/10");

    const usersLink = screen.getByText("사용자 관리").closest("a");
    expect(usersLink?.className).toContain("bg-primary/10");
  });

  it("highlights /admin/prompts for prompts page", () => {
    mockPathname = "/admin/prompts";
    render(<AdminSidebar />);

    const promptsLink = screen.getByText("프롬프트 관리").closest("a");
    expect(promptsLink?.className).toContain("bg-primary/10");
  });
});
