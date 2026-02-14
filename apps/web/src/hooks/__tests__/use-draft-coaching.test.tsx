import { describe, it, expect, vi, beforeEach } from "vitest";
import { renderHook, waitFor } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { useDraftCoaching } from "../use-draft-coaching";
import * as coachingApi from "@/lib/api/coaching";

vi.mock("@/lib/api/coaching", () => ({
  generateDraft: vi.fn(),
}));

const mockGenerateDraft = vi.mocked(coachingApi.generateDraft);

function createWrapper() {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false }, mutations: { retry: false } },
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

describe("useDraftCoaching", () => {
  it("starts in idle state", () => {
    const { result } = renderHook(() => useDraftCoaching(), {
      wrapper: createWrapper(),
    });

    expect(result.current.isPending).toBe(false);
    expect(result.current.data).toBeUndefined();
  });

  it("calls generateDraft and returns result on success", async () => {
    const mockResult = { draft: "[상황]\n프로젝트에서..." };
    mockGenerateDraft.mockResolvedValue(mockResult);

    const { result } = renderHook(() => useDraftCoaching(), {
      wrapper: createWrapper(),
    });

    result.current.mutate({
      application_id: "app-123",
      experience_ids: ["exp-1"],
      question_text: "팀에서 어려움을 극복한 경험을 기술하세요",
      char_limit: 800,
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(mockGenerateDraft).toHaveBeenCalledWith({
      application_id: "app-123",
      experience_ids: ["exp-1"],
      question_text: "팀에서 어려움을 극복한 경험을 기술하세요",
      char_limit: 800,
    });
    expect(result.current.data).toEqual(mockResult);
  });

  it("handles error state", async () => {
    mockGenerateDraft.mockRejectedValue(new Error("AI generation failed"));

    const { result } = renderHook(() => useDraftCoaching(), {
      wrapper: createWrapper(),
    });

    result.current.mutate({
      application_id: "app-123",
      experience_ids: ["exp-1"],
      question_text: "팀에서 어려움을 극복한 경험을 기술하세요",
      char_limit: 800,
    });

    await waitFor(() => expect(result.current.isError).toBe(true));
    expect(result.current.error?.message).toBe("AI generation failed");
  });
});
