"use client";

import { useCallback, useRef, useState } from "react";
import { useAuthStore } from "@/stores/auth-store";
import type { GenerateDraftRequest } from "@/lib/api/coaching";

interface SSETextEvent {
  type: "text";
  content: string;
}

interface SSEDoneEvent {
  type: "done";
  cover_letter_id: string;
  session_id: string;
}

interface SSEErrorEvent {
  type: "error";
  message: string;
}

type SSEEvent = SSETextEvent | SSEDoneEvent | SSEErrorEvent;

interface DraftStreamingState {
  content: string;
  isStreaming: boolean;
  error: string | null;
  coverLetterId: string | null;
  sessionId: string | null;
}

export function useDraftStreaming() {
  const [state, setState] = useState<DraftStreamingState>({
    content: "",
    isStreaming: false,
    error: null,
    coverLetterId: null,
    sessionId: null,
  });
  const abortRef = useRef<AbortController | null>(null);

  const abort = useCallback(() => {
    abortRef.current?.abort();
    abortRef.current = null;
    setState((prev) => ({ ...prev, isStreaming: false }));
  }, []);

  const streamDraft = useCallback(
    async (req: GenerateDraftRequest) => {
      // Abort any previous stream
      abortRef.current?.abort();

      const controller = new AbortController();
      abortRef.current = controller;

      setState({
        content: "",
        isStreaming: true,
        error: null,
        coverLetterId: null,
        sessionId: null,
      });

      const baseURL =
        process.env.NEXT_PUBLIC_API_URL || "http://localhost:9000";
      const accessToken = useAuthStore.getState().accessToken;

      try {
        const response = await fetch(`${baseURL}/v1/coaching/draft`, {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
            ...(accessToken ? { Authorization: `Bearer ${accessToken}` } : {}),
          },
          body: JSON.stringify(req),
          signal: controller.signal,
        });

        if (!response.ok) {
          const body = await response.text();
          let message = "요청에 실패했습니다";
          try {
            const parsed = JSON.parse(body);
            message = parsed.error || message;
          } catch {
            // use default message
          }
          setState((prev) => ({
            ...prev,
            isStreaming: false,
            error: message,
          }));
          return;
        }

        const reader = response.body?.getReader();
        if (!reader) {
          setState((prev) => ({
            ...prev,
            isStreaming: false,
            error: "스트리밍을 시작할 수 없습니다",
          }));
          return;
        }

        const decoder = new TextDecoder();
        let buffer = "";

        while (true) {
          const { done, value } = await reader.read();
          if (done) break;

          buffer += decoder.decode(value, { stream: true });

          // Parse SSE events from buffer (events separated by double newline)
          const parts = buffer.split("\n\n");
          buffer = parts.pop() || "";

          for (const part of parts) {
            const trimmed = part.trim();
            if (!trimmed) continue;

            let dataStr = "";
            for (const line of trimmed.split("\n")) {
              if (line.startsWith("data: ")) {
                dataStr = line.slice(6);
              }
            }

            if (!dataStr) continue;

            try {
              const event = JSON.parse(dataStr) as SSEEvent;

              if (event.type === "text") {
                setState((prev) => ({
                  ...prev,
                  content: prev.content + event.content,
                }));
              } else if (event.type === "done") {
                setState((prev) => ({
                  ...prev,
                  isStreaming: false,
                  coverLetterId: event.cover_letter_id,
                  sessionId: event.session_id,
                }));
              } else if (event.type === "error") {
                setState((prev) => ({
                  ...prev,
                  isStreaming: false,
                  error: event.message,
                }));
              }
            } catch {
              // Skip malformed JSON
            }
          }
        }

        // If stream ended without done/error event
        setState((prev) => {
          if (prev.isStreaming) {
            return { ...prev, isStreaming: false };
          }
          return prev;
        });
      } catch (err) {
        if (err instanceof DOMException && err.name === "AbortError") {
          return; // User aborted, state already updated
        }
        setState((prev) => ({
          ...prev,
          isStreaming: false,
          error:
            err instanceof Error
              ? err.message
              : "알 수 없는 오류가 발생했습니다",
        }));
      } finally {
        if (abortRef.current === controller) {
          abortRef.current = null;
        }
      }
    },
    []
  );

  return {
    ...state,
    streamDraft,
    abort,
  };
}
