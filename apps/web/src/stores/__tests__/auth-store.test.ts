import { describe, it, expect, beforeEach, vi } from "vitest";

vi.mock("@/lib/api-client", () => ({
  apiClient: {
    get: vi.fn(),
    post: vi.fn().mockResolvedValue({ data: {} }),
  },
}));

import { useAuthStore } from "../auth-store";

describe("useAuthStore", () => {
  beforeEach(() => {
    // Reset store state before each test
    useAuthStore.setState({
      accessToken: null,
      refreshToken: null,
      user: null,
      isLoading: false,
    });
  });

  it("has correct initial state", () => {
    const state = useAuthStore.getState();
    expect(state.accessToken).toBeNull();
    expect(state.refreshToken).toBeNull();
    expect(state.user).toBeNull();
    expect(state.isLoading).toBe(false);
  });

  it("setTokens updates access and refresh tokens", () => {
    useAuthStore.getState().setTokens("access-123", "refresh-456");

    const state = useAuthStore.getState();
    expect(state.accessToken).toBe("access-123");
    expect(state.refreshToken).toBe("refresh-456");
  });

  it("logout clears all auth state", () => {
    // Set some state first
    useAuthStore.setState({
      accessToken: "access-123",
      refreshToken: "refresh-456",
      user: {
        id: "user-1",
        role: "user",
        auth_provider: "email",
        onboarding_completed: false,
      },
    });

    useAuthStore.getState().logout();

    const state = useAuthStore.getState();
    expect(state.accessToken).toBeNull();
    expect(state.refreshToken).toBeNull();
    expect(state.user).toBeNull();
  });
});
