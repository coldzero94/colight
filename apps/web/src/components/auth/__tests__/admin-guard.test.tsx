import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen } from "@testing-library/react";

const mockReplace = vi.fn();
vi.mock("next/navigation", () => ({
  useRouter: () => ({ replace: mockReplace }),
}));

const mockUseAuthStore = vi.fn();
vi.mock("@/stores/auth-store", () => ({
  useAuthStore: (...args: unknown[]) => mockUseAuthStore(...args),
}));

import { AdminGuard } from "../admin-guard";

describe("AdminGuard", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("redirects non-admin users to /experiences", () => {
    mockUseAuthStore.mockReturnValue({
      user: { id: "1", role: "user" },
    });

    render(
      <AdminGuard>
        <div>Admin Content</div>
      </AdminGuard>
    );

    expect(mockReplace).toHaveBeenCalledWith("/experiences");
    expect(screen.queryByText("Admin Content")).not.toBeInTheDocument();
  });

  it("shows loading spinner when no user yet", () => {
    mockUseAuthStore.mockReturnValue({ user: null });

    const { container } = render(
      <AdminGuard>
        <div>Admin Content</div>
      </AdminGuard>
    );

    expect(container.querySelector(".animate-spin")).toBeInTheDocument();
  });

  it("renders children for admin users", () => {
    mockUseAuthStore.mockReturnValue({
      user: { id: "1", role: "admin" },
    });

    render(
      <AdminGuard>
        <div>Admin Content</div>
      </AdminGuard>
    );

    expect(screen.getByText("Admin Content")).toBeInTheDocument();
    expect(mockReplace).not.toHaveBeenCalled();
  });
});
