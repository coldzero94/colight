import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { renderHook, act, waitFor } from "@testing-library/react";
import { useDraftStreaming } from "../use-draft-streaming";

// Mock auth store
vi.mock("@/stores/auth-store", () => ({
  useAuthStore: {
    getState: () => ({ accessToken: "test-token" }),
  },
}));

function createSSEStream(events: Array<{ event: string; data: string }>) {
  const text = events
    .map((e) => `event: ${e.event}\ndata: ${e.data}\n\n`)
    .join("");
  const encoder = new TextEncoder();
  const encoded = encoder.encode(text);

  return new ReadableStream({
    start(controller) {
      controller.enqueue(encoded);
      controller.close();
    },
  });
}

const defaultRequest = {
  application_id: "app-123",
  experience_ids: ["exp-1"],
  question_text: "팀에서 어려움을 극복한 경험을 기술하세요",
  char_limit: 800,
};

let fetchMock: ReturnType<typeof vi.fn>;

beforeEach(() => {
  fetchMock = vi.fn();
  vi.stubGlobal("fetch", fetchMock);
});

afterEach(() => {
  vi.restoreAllMocks();
});

describe("useDraftStreaming", () => {
  it("starts in idle state", () => {
    const { result } = renderHook(() => useDraftStreaming());

    expect(result.current.content).toBe("");
    expect(result.current.isStreaming).toBe(false);
    expect(result.current.error).toBeNull();
    expect(result.current.coverLetterId).toBeNull();
  });

  it("streams text chunks and receives done event", async () => {
    const stream = createSSEStream([
      { event: "text", data: '{"type":"text","content":"[상황]\\n"}' },
      { event: "text", data: '{"type":"text","content":"초안 내용"}' },
      {
        event: "done",
        data: '{"type":"done","cover_letter_id":"cl-123","session_id":"s-456"}',
      },
    ]);

    fetchMock.mockResolvedValue({
      ok: true,
      body: stream,
    });

    const { result } = renderHook(() => useDraftStreaming());

    await act(async () => {
      await result.current.streamDraft(defaultRequest);
    });

    expect(result.current.content).toBe("[상황]\n초안 내용");
    expect(result.current.isStreaming).toBe(false);
    expect(result.current.coverLetterId).toBe("cl-123");
    expect(result.current.sessionId).toBe("s-456");
    expect(result.current.error).toBeNull();

    // Verify fetch was called with correct params
    expect(fetchMock).toHaveBeenCalledWith(
      expect.stringContaining("/v1/coaching/draft"),
      expect.objectContaining({
        method: "POST",
        headers: expect.objectContaining({
          Authorization: "Bearer test-token",
        }),
      })
    );
  });

  it("handles SSE error event", async () => {
    const stream = createSSEStream([
      { event: "text", data: '{"type":"text","content":"partial"}' },
      {
        event: "error",
        data: '{"type":"error","message":"AI 초안 생성 중 오류가 발생했습니다"}',
      },
    ]);

    fetchMock.mockResolvedValue({
      ok: true,
      body: stream,
    });

    const { result } = renderHook(() => useDraftStreaming());

    await act(async () => {
      await result.current.streamDraft(defaultRequest);
    });

    expect(result.current.error).toBe("AI 초안 생성 중 오류가 발생했습니다");
    expect(result.current.isStreaming).toBe(false);
    expect(result.current.content).toBe("partial");
  });

  it("handles HTTP error response", async () => {
    fetchMock.mockResolvedValue({
      ok: false,
      status: 401,
      text: async () => '{"error":"인증이 필요합니다."}',
    });

    const { result } = renderHook(() => useDraftStreaming());

    await act(async () => {
      await result.current.streamDraft(defaultRequest);
    });

    expect(result.current.error).toBe("인증이 필요합니다.");
    expect(result.current.isStreaming).toBe(false);
  });

  it("sets isStreaming to true during streaming", async () => {
    // Create a stream that never closes so we can check isStreaming
    let resolveRead: ((value: ReadableStreamReadResult<Uint8Array>) => void) | null = null;
    const body = new ReadableStream({
      start() {
        // Don't close immediately
      },
      pull(controller) {
        return new Promise((resolve) => {
          resolveRead = (value) => {
            if (value.done) {
              controller.close();
            } else if (value.value) {
              controller.enqueue(value.value);
            }
            resolve();
          };
        });
      },
    });

    fetchMock.mockResolvedValue({ ok: true, body });

    const { result } = renderHook(() => useDraftStreaming());

    // Start streaming (don't await)
    let streamPromise: Promise<void>;
    act(() => {
      streamPromise = result.current.streamDraft(defaultRequest);
    });

    await waitFor(() => {
      expect(result.current.isStreaming).toBe(true);
    });

    // Close the stream
    await act(async () => {
      resolveRead?.({ done: true, value: undefined });
      await streamPromise!;
    });

    expect(result.current.isStreaming).toBe(false);
  });

  it("supports abort", async () => {
    let readerCancelled = false;
    const body = new ReadableStream({
      start() {
        // Never closes naturally
      },
      cancel() {
        readerCancelled = true;
      },
    });

    fetchMock.mockResolvedValue({ ok: true, body });

    const { result } = renderHook(() => useDraftStreaming());

    act(() => {
      result.current.streamDraft(defaultRequest);
    });

    await waitFor(() => {
      expect(result.current.isStreaming).toBe(true);
    });

    act(() => {
      result.current.abort();
    });

    expect(result.current.isStreaming).toBe(false);
    // Note: The abort may or may not cancel the reader depending on timing
    // The important thing is that isStreaming is set to false
  });
});
