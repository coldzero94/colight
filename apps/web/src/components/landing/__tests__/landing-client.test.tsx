import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen } from "@testing-library/react";

const mockReplace = vi.fn();
vi.mock("next/navigation", () => ({
  useRouter: () => ({ replace: mockReplace }),
}));

const mockUseAuthStore = vi.fn();
vi.mock("@/stores/auth-store", () => ({
  useAuthStore: (selector?: (state: { accessToken: string | null }) => unknown) => {
    const state = mockUseAuthStore();
    return selector ? selector(state) : state;
  },
}));

import { LandingClient } from "../landing-client";

describe("LandingClient", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("redirects to /experiences when user has access token", () => {
    mockUseAuthStore.mockReturnValue({ accessToken: "token-123" });

    render(<LandingClient />);

    expect(mockReplace).toHaveBeenCalledWith("/experiences");
    expect(screen.queryByText(/AI 코치/)).not.toBeInTheDocument();
  });

  it("renders landing sections when no access token", () => {
    mockUseAuthStore.mockReturnValue({ accessToken: null });

    render(<LandingClient />);

    expect(mockReplace).not.toHaveBeenCalled();
    expect(screen.getByText(/AI 코치/)).toBeInTheDocument();
    expect(screen.getByText("기업 분석")).toBeInTheDocument();
    expect(screen.getByText("경험 매칭")).toBeInTheDocument();
  });
});
