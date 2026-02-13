import { describe, it, expect, vi } from "vitest";
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

import { AdminSidebar } from "../admin-sidebar";

describe("AdminSidebar", () => {
  it("renders all admin menu items", () => {
    render(<AdminSidebar />);

    expect(screen.getByText("대시보드")).toBeInTheDocument();
    expect(screen.getByText("사용자 관리")).toBeInTheDocument();
    expect(screen.getByText("프롬프트 관리")).toBeInTheDocument();
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
    expect(dashboardLink?.className).toContain("bg-gray-700");
  });

  it("does not highlight /admin for /admin/users", () => {
    mockPathname = "/admin/users";
    render(<AdminSidebar />);

    const dashboardLink = screen.getByText("대시보드").closest("a");
    expect(dashboardLink?.className).not.toContain("bg-gray-700");

    const usersLink = screen.getByText("사용자 관리").closest("a");
    expect(usersLink?.className).toContain("bg-gray-700");
  });

  it("highlights /admin/prompts for prompts page", () => {
    mockPathname = "/admin/prompts";
    render(<AdminSidebar />);

    const promptsLink = screen.getByText("프롬프트 관리").closest("a");
    expect(promptsLink?.className).toContain("bg-gray-700");
  });
});
