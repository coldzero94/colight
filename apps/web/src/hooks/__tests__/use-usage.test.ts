import { describe, it, expect, vi, beforeEach } from "vitest";
import { renderHook, waitFor } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { createElement } from "react";
import { useUsage } from "../use-usage";

vi.mock("@/lib/api/usage", () => ({
  getUsage: vi.fn().mockResolvedValue({
    tokensUsed: 1000,
    tokensLimit: 5000,
    requestsUsed: 50,
    requestsLimit: 100,
  }),
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

describe("useUsage", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("fetches usage data successfully", async () => {
    const { result } = renderHook(() => useUsage(), {
      wrapper: createWrapper(),
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toEqual({
      tokensUsed: 1000,
      tokensLimit: 5000,
      requestsUsed: 50,
      requestsLimit: 100,
    });
  });

  it("calls getUsage API on mount", async () => {
    const { getUsage } = await import("@/lib/api/usage");
    const { result } = renderHook(() => useUsage(), {
      wrapper: createWrapper(),
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(getUsage).toHaveBeenCalled();
  });
});
