import { describe, it, expect, vi, beforeEach } from "vitest";
import { renderHook, waitFor } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { createElement } from "react";
import { useFeedback } from "../use-feedback";

vi.mock("@/lib/api/feedback", () => ({
  submitFeedback: vi.fn().mockResolvedValue({ id: "feedback-1", success: true }),
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

describe("useFeedback", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("submits feedback successfully", async () => {
    const { result } = renderHook(() => useFeedback(), {
      wrapper: createWrapper(),
    });

    result.current.mutate({ category: "bug", content: "Found a bug" });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toEqual({ id: "feedback-1", success: true });
  });

  it("handles error when submission fails", async () => {
    const { submitFeedback } = await import("@/lib/api/feedback");
    vi.mocked(submitFeedback).mockRejectedValue(new Error("Network error"));

    const { result } = renderHook(() => useFeedback(), {
      wrapper: createWrapper(),
    });

    result.current.mutate({ category: "bug", content: "Found a bug" });

    await waitFor(() => expect(result.current.isError).toBe(true));
    expect(result.current.error).toEqual(new Error("Network error"));
  });
});
