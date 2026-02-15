import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { renderHook, act } from "@testing-library/react";
import { useInterviewChat } from "../use-interview-chat";
import * as interviewApi from "@/lib/api/interview";

vi.mock("@/lib/api/interview", () => ({
  generateQuestion: vi.fn(),
  STAGES: ["warmup", "memory", "challenge", "solution", "outcome"],
  STAGE_LABELS: {
    warmup: "가볍게",
    memory: "기억에 남는 순간",
    challenge: "어려웠던 점",
    solution: "해결법",
    outcome: "결과/배운 점",
  },
}));

const mockGenerateQuestion = vi.mocked(interviewApi.generateQuestion);

beforeEach(() => {
  vi.clearAllMocks();
  vi.useFakeTimers();
});

afterEach(() => {
  vi.useRealTimers();
});

describe("useInterviewChat", () => {
  it("starts in initial state", () => {
    const { result } = renderHook(() => useInterviewChat());

    expect(result.current.messages).toEqual([]);
    expect(result.current.stage).toBe("warmup");
    expect(result.current.isLoading).toBe(false);
    expect(result.current.isComplete).toBe(false);
    expect(result.current.elapsedSeconds).toBe(0);
    expect(result.current.error).toBeNull();
  });

  it("starts interview and adds first AI message", async () => {
    mockGenerateQuestion.mockResolvedValue({
      question: "안녕하세요! 최근 어떤 활동을 하셨나요?",
      stage: "warmup",
      next_stage: "memory",
      is_complete: false,
    });

    const { result } = renderHook(() => useInterviewChat());

    await act(async () => {
      await result.current.startInterview();
    });

    expect(result.current.messages).toHaveLength(1);
    expect(result.current.messages[0]).toEqual({
      role: "assistant",
      content: "안녕하세요! 최근 어떤 활동을 하셨나요?",
    });
    expect(result.current.stage).toBe("warmup");
  });

  it("sends user message and receives AI response", async () => {
    mockGenerateQuestion
      .mockResolvedValueOnce({
        question: "첫 질문입니다",
        stage: "warmup",
        next_stage: "memory",
        is_complete: false,
      })
      .mockResolvedValueOnce({
        question: "더 자세히 말씀해주세요",
        stage: "warmup",
        next_stage: "memory",
        is_complete: false,
      });

    const { result } = renderHook(() => useInterviewChat());

    await act(async () => {
      await result.current.startInterview();
    });

    await act(async () => {
      await result.current.sendMessage("대학교에서 프로젝트를 했습니다");
    });

    expect(result.current.messages).toHaveLength(3);
    expect(result.current.messages[1]).toEqual({
      role: "user",
      content: "대학교에서 프로젝트를 했습니다",
    });
    expect(result.current.messages[2]).toEqual({
      role: "assistant",
      content: "더 자세히 말씀해주세요",
    });
  });

  it("transitions stage when next_stage differs", async () => {
    mockGenerateQuestion
      .mockResolvedValueOnce({
        question: "첫 질문",
        stage: "warmup",
        next_stage: "memory",
        is_complete: false,
      })
      .mockResolvedValueOnce({
        question: "기억에 남는 건?",
        stage: "warmup",
        next_stage: "memory",
        is_complete: false,
      });

    const { result } = renderHook(() => useInterviewChat());

    await act(async () => {
      await result.current.startInterview();
    });

    await act(async () => {
      await result.current.sendMessage("답변");
    });

    expect(result.current.stage).toBe("memory");
  });

  it("marks complete when is_complete is true", async () => {
    mockGenerateQuestion
      .mockResolvedValueOnce({
        question: "첫 질문",
        stage: "warmup",
        next_stage: "",
        is_complete: false,
      })
      .mockResolvedValueOnce({
        question: "마지막 질문 감사합니다",
        stage: "outcome",
        next_stage: "",
        is_complete: true,
      });

    const { result } = renderHook(() => useInterviewChat());

    await act(async () => {
      await result.current.startInterview();
    });

    await act(async () => {
      await result.current.sendMessage("답변");
    });

    expect(result.current.isComplete).toBe(true);
  });

  it("increments timer after start", async () => {
    mockGenerateQuestion.mockResolvedValue({
      question: "질문",
      stage: "warmup",
      next_stage: "memory",
      is_complete: false,
    });

    const { result } = renderHook(() => useInterviewChat());

    await act(async () => {
      await result.current.startInterview();
    });

    expect(result.current.elapsedSeconds).toBe(0);

    act(() => {
      vi.advanceTimersByTime(3000);
    });

    expect(result.current.elapsedSeconds).toBe(3);
  });

  it("resets all state", async () => {
    mockGenerateQuestion.mockResolvedValue({
      question: "질문",
      stage: "warmup",
      next_stage: "memory",
      is_complete: false,
    });

    const { result } = renderHook(() => useInterviewChat());

    await act(async () => {
      await result.current.startInterview();
    });

    act(() => {
      vi.advanceTimersByTime(5000);
    });

    act(() => {
      result.current.reset();
    });

    expect(result.current.messages).toEqual([]);
    expect(result.current.stage).toBe("warmup");
    expect(result.current.elapsedSeconds).toBe(0);
    expect(result.current.isComplete).toBe(false);
  });

  it("sets error on API failure", async () => {
    mockGenerateQuestion.mockRejectedValue(new Error("Network error"));

    const { result } = renderHook(() => useInterviewChat());

    await act(async () => {
      await result.current.startInterview();
    });

    expect(result.current.error).toBe("Network error");
    expect(result.current.messages).toEqual([]);
  });
});
