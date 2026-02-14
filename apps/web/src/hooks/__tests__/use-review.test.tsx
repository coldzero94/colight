import { describe, it, expect, vi, beforeEach } from "vitest";
import { renderHook, waitFor } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { useReview } from "../use-review";
import * as coachingApi from "@/lib/api/coaching";

vi.mock("@/lib/api/coaching", () => ({
  requestReview: vi.fn(),
}));

const mockRequestReview = vi.mocked(coachingApi.requestReview);

function createWrapper() {
  const queryClient = new QueryClient({
    defaultOptions: { mutations: { retry: false } },
  });
  return function Wrapper({ children }: { children: React.ReactNode }) {
    return (
      <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
    );
  };
}

beforeEach(() => {
  vi.clearAllMocks();
});

describe("useReview", () => {
  it("starts in idle state", () => {
    const { result } = renderHook(() => useReview(), {
      wrapper: createWrapper(),
    });
    expect(result.current.isPending).toBe(false);
    expect(result.current.data).toBeUndefined();
  });

  it("calls requestReview and returns result", async () => {
    const mockResult: coachingApi.ReviewResult = {
      scores: { specificity: 75, job_fit: 80, company_fit: 65, authenticity: 85 },
      overall: 76,
      per_dimension_feedback: [],
      specific_suggestions: [],
    };

    mockRequestReview.mockResolvedValue(mockResult);

    const { result } = renderHook(() => useReview(), {
      wrapper: createWrapper(),
    });

    result.current.mutate({
      cover_letter_id: "cl-123",
      content: "자소서 내용",
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(result.current.data).toEqual(mockResult);
    expect(mockRequestReview.mock.calls[0][0]).toEqual({
      cover_letter_id: "cl-123",
      content: "자소서 내용",
    });
  });

  it("handles error", async () => {
    mockRequestReview.mockRejectedValue(new Error("review failed"));

    const { result } = renderHook(() => useReview(), {
      wrapper: createWrapper(),
    });

    result.current.mutate({
      cover_letter_id: "cl-123",
      content: "자소서 내용",
    });

    await waitFor(() => expect(result.current.isError).toBe(true));
    expect(result.current.error?.message).toBe("review failed");
  });
});
