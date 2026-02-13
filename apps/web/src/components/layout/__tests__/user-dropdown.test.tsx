import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

const mockReplace = vi.fn();
vi.mock("next/navigation", () => ({
  useRouter: () => ({ replace: mockReplace }),
}));

vi.mock("next/link", () => ({
  default: ({
    children,
    href,
    ...props
  }: {
    children: React.ReactNode;
    href: string;
    onClick?: () => void;
    className?: string;
  }) => (
    <a href={href} {...props}>
      {children}
    </a>
  ),
}));

const mockLogout = vi.fn();
const mockUseAuthStore = vi.fn();
vi.mock("@/stores/auth-store", () => ({
  useAuthStore: (...args: unknown[]) => mockUseAuthStore(...args),
}));

import { UserDropdown } from "../user-dropdown";

describe("UserDropdown", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockUseAuthStore.mockReturnValue({
      user: { nickname: "홍길동", email: "hong@test.com", role: "user" },
      logout: mockLogout,
    });
  });

  it("renders user avatar initial and nickname", () => {
    render(<UserDropdown />);

    expect(screen.getByText("홍")).toBeInTheDocument();
    expect(screen.getByText("홍길동")).toBeInTheDocument();
  });

  it("opens dropdown on click", async () => {
    const user = userEvent.setup();
    render(<UserDropdown />);

    expect(screen.queryByText("로그아웃")).not.toBeInTheDocument();

    await user.click(screen.getByText("홍길동"));

    expect(screen.getByText("로그아웃")).toBeInTheDocument();
    expect(screen.getByText("hong@test.com")).toBeInTheDocument();
  });

  it("calls logout and redirects on logout click", async () => {
    const user = userEvent.setup();
    render(<UserDropdown />);

    await user.click(screen.getByText("홍길동"));
    await user.click(screen.getByText("로그아웃"));

    expect(mockLogout).toHaveBeenCalled();
    expect(mockReplace).toHaveBeenCalledWith("/login");
  });

  it("shows admin link for admin users", async () => {
    mockUseAuthStore.mockReturnValue({
      user: { nickname: "관리자", role: "admin" },
      logout: mockLogout,
    });
    const user = userEvent.setup();

    render(<UserDropdown />);
    await user.click(screen.getByText("관리자"));

    expect(screen.getByText("어드민")).toBeInTheDocument();
  });

  it("hides admin link for regular users", async () => {
    const user = userEvent.setup();
    render(<UserDropdown />);

    await user.click(screen.getByText("홍길동"));

    expect(screen.queryByText("어드민")).not.toBeInTheDocument();
  });
});
