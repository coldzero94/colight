import { describe, it, expect, vi, beforeEach } from "vitest";
import { renderHook, waitFor, act } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { createElement } from "react";
import { useWeaponTagging } from "../use-weapon-tagging";

const mockTagExperience = vi.fn();

vi.mock("@/lib/api/experiences", () => ({
  tagExperience: (...args: unknown[]) => mockTagExperience(...args),
}));

function createWrapper() {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: { retry: false },
      mutations: { retry: false },
    },
  });
  return function Wrapper({ children }: { children: React.ReactNode }) {
    return createElement(
      QueryClientProvider,
      { client: queryClient },
      children
    );
  };
}

describe("useWeaponTagging", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("transitions through idle → tagging → success states", async () => {
    mockTagExperience.mockResolvedValue({
      primary_weapon: { code: "W01", confidence: 0.9, reasoning: "" },
      secondary_weapons: [],
    });

    const { result } = renderHook(() => useWeaponTagging(), {
      wrapper: createWrapper(),
    });

    expect(result.current.state).toBe("idle");

    act(() => {
      result.current.triggerTagging("exp-1");
    });

    expect(result.current.state).toBe("tagging");

    await waitFor(() => expect(result.current.state).toBe("success"));
  });

  it("handles tagging error and allows retry", async () => {
    mockTagExperience.mockRejectedValueOnce(new Error("API error"));
    mockTagExperience.mockResolvedValueOnce({
      primary_weapon: { code: "W01", confidence: 0.9, reasoning: "" },
      secondary_weapons: [],
    });

    const { result } = renderHook(() => useWeaponTagging(), {
      wrapper: createWrapper(),
    });

    act(() => {
      result.current.triggerTagging("exp-1");
    });

    await waitFor(() => expect(result.current.state).toBe("error"));

    // Retry
    act(() => {
      result.current.triggerTagging("exp-1");
    });

    await waitFor(() => expect(result.current.state).toBe("success"));
  });
});
