import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

vi.mock("next/link", () => ({
  default: ({
    children,
    href,
    onClick,
    ...props
  }: {
    children: React.ReactNode;
    href: string;
    onClick?: () => void;
    className?: string;
  }) => (
    <a href={href} onClick={onClick} {...props}>
      {children}
    </a>
  ),
}));

vi.mock("next/navigation", () => ({
  usePathname: () => "/experiences",
}));

const mockUseAuthStore = vi.fn();
vi.mock("@/stores/auth-store", () => ({
  useAuthStore: (...args: unknown[]) => mockUseAuthStore(...args),
}));

import { MobileNav } from "../mobile-nav";

describe("MobileNav", () => {
  it("renders hamburger menu button", () => {
    mockUseAuthStore.mockReturnValue({ user: { role: "user" } });

    render(<MobileNav />);

    expect(screen.getByLabelText("메뉴 열기")).toBeInTheDocument();
  });

  it("opens menu on hamburger click", async () => {
    mockUseAuthStore.mockReturnValue({ user: { role: "user" } });
    const user = userEvent.setup();

    render(<MobileNav />);
    await user.click(screen.getByLabelText("메뉴 열기"));

    expect(screen.getByText("경험 관리")).toBeInTheDocument();
    expect(screen.getByText("기업 분석")).toBeInTheDocument();
    expect(screen.getByText("자소서 코칭")).toBeInTheDocument();
    expect(screen.getByText("대시보드")).toBeInTheDocument();
  });

  it("shows admin link for admin users", async () => {
    mockUseAuthStore.mockReturnValue({ user: { role: "admin" } });
    const user = userEvent.setup();

    render(<MobileNav />);
    await user.click(screen.getByLabelText("메뉴 열기"));

    expect(screen.getByText("어드민")).toBeInTheDocument();
  });

  it("hides admin link for regular users", async () => {
    mockUseAuthStore.mockReturnValue({ user: { role: "user" } });
    const user = userEvent.setup();

    render(<MobileNav />);
    await user.click(screen.getByLabelText("메뉴 열기"));

    expect(screen.queryByText("어드민")).not.toBeInTheDocument();
  });
});
