import { describe, it, expect, vi, beforeEach } from "vitest";
import { renderHook, waitFor } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { useQuestionAnalysis } from "../use-question-analysis";
import * as coachingApi from "@/lib/api/coaching";

vi.mock("@/lib/api/coaching", () => ({
  analyzeQuestion: vi.fn(),
}));

const mockAnalyzeQuestion = vi.mocked(coachingApi.analyzeQuestion);

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

describe("useQuestionAnalysis", () => {
  it("starts in idle state", () => {
    const { result } = renderHook(() => useQuestionAnalysis(), {
      wrapper: createWrapper(),
    });

    expect(result.current.isPending).toBe(false);
    expect(result.current.data).toBeUndefined();
  });

  it("calls analyzeQuestion and returns result on success", async () => {
    const mockResult = {
      surface_question: "팀워크 경험",
      real_intents: [{ intent: "협업", description: "팀 경험" }],
      required_weapons: {
        primary: { weapon_id: "W01", weapon_name: "문제해결", reason: "핵심" },
        secondary: [],
      },
      writing_structure: { total_chars: 800, sections: [] },
      key_keywords: ["데이터"],
    };
    mockAnalyzeQuestion.mockResolvedValue(mockResult);

    const { result } = renderHook(() => useQuestionAnalysis(), {
      wrapper: createWrapper(),
    });

    result.current.mutate({
      application_id: "app-123",
      question_text: "팀에서 어려움을 극복한 경험을 기술하세요",
      char_limit: 800,
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(mockAnalyzeQuestion).toHaveBeenCalledWith({
      application_id: "app-123",
      question_text: "팀에서 어려움을 극복한 경험을 기술하세요",
      char_limit: 800,
    });
    expect(result.current.data).toEqual(mockResult);
  });

  it("handles error state", async () => {
    mockAnalyzeQuestion.mockRejectedValue(new Error("API Error"));

    const { result } = renderHook(() => useQuestionAnalysis(), {
      wrapper: createWrapper(),
    });

    result.current.mutate({
      application_id: "app-123",
      question_text: "팀에서 어려움을 극복한 경험을 기술하세요",
      char_limit: 800,
    });

    await waitFor(() => expect(result.current.isError).toBe(true));
    expect(result.current.error?.message).toBe("API Error");
  });
});
