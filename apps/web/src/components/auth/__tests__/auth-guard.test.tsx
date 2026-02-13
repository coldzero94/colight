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

import { AuthGuard } from "../auth-guard";

describe("AuthGuard", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("redirects to /login when no access token", () => {
    mockUseAuthStore.mockReturnValue({
      accessToken: null,
      user: null,
      isLoading: false,
      fetchUser: vi.fn(),
    });

    render(
      <AuthGuard>
        <div>Protected</div>
      </AuthGuard>
    );

    expect(mockReplace).toHaveBeenCalledWith("/login");
    expect(screen.queryByText("Protected")).not.toBeInTheDocument();
  });

  it("shows loading spinner when loading", () => {
    mockUseAuthStore.mockReturnValue({
      accessToken: "token",
      user: null,
      isLoading: true,
      fetchUser: vi.fn(),
    });

    const { container } = render(
      <AuthGuard>
        <div>Protected</div>
      </AuthGuard>
    );

    expect(container.querySelector(".animate-spin")).toBeInTheDocument();
    expect(screen.queryByText("Protected")).not.toBeInTheDocument();
  });

  it("renders children when authenticated", () => {
    mockUseAuthStore.mockReturnValue({
      accessToken: "token",
      user: { id: "1", role: "user" },
      isLoading: false,
      fetchUser: vi.fn(),
    });

    render(
      <AuthGuard>
        <div>Protected</div>
      </AuthGuard>
    );

    expect(screen.getByText("Protected")).toBeInTheDocument();
  });

  it("calls fetchUser when token exists but no user", () => {
    const fetchUser = vi.fn();
    mockUseAuthStore.mockReturnValue({
      accessToken: "token",
      user: null,
      isLoading: false,
      fetchUser,
    });

    render(
      <AuthGuard>
        <div>Protected</div>
      </AuthGuard>
    );

    expect(fetchUser).toHaveBeenCalled();
  });
});
