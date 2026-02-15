"use client";

import { useState, useCallback, useRef, useEffect } from "react";
import {
  generateQuestion,
  STAGES,
  type InterviewStage,
  type ChatMessage,
} from "@/lib/api/interview";

interface UseInterviewChatReturn {
  messages: ChatMessage[];
  stage: InterviewStage;
  isLoading: boolean;
  isComplete: boolean;
  elapsedSeconds: number;
  error: string | null;
  startInterview: () => Promise<void>;
  sendMessage: (content: string) => Promise<boolean>;
  reset: () => void;
}

export function useInterviewChat(): UseInterviewChatReturn {
  const [messages, setMessages] = useState<ChatMessage[]>([]);
  const [stage, setStage] = useState<InterviewStage>("warmup");
  const [isLoading, setIsLoading] = useState(false);
  const [isComplete, setIsComplete] = useState(false);
  const [elapsedSeconds, setElapsedSeconds] = useState(0);
  const [error, setError] = useState<string | null>(null);
  const timerRef = useRef<ReturnType<typeof setInterval> | null>(null);

  // Timer management
  const startTimer = useCallback(() => {
    if (timerRef.current) return;
    timerRef.current = setInterval(() => {
      setElapsedSeconds((prev) => prev + 1);
    }, 1000);
  }, []);

  const stopTimer = useCallback(() => {
    if (timerRef.current) {
      clearInterval(timerRef.current);
      timerRef.current = null;
    }
  }, []);

  // Cleanup timer on unmount
  useEffect(() => {
    return () => stopTimer();
  }, [stopTimer]);

  const startInterview = useCallback(async () => {
    setIsLoading(true);
    setError(null);
    try {
      const resp = await generateQuestion({ stage: "warmup", messages: [] });
      setMessages([{ role: "assistant", content: resp.question }]);
      setStage(resp.stage);
      startTimer();
    } catch (e) {
      setError(e instanceof Error ? e.message : "인터뷰를 시작할 수 없습니다");
    } finally {
      setIsLoading(false);
    }
  }, [startTimer]);

  const sendMessage = useCallback(
    async (content: string): Promise<boolean> => {
      const userMsg: ChatMessage = { role: "user", content };
      const updated = [...messages, userMsg];
      setMessages(updated);
      setIsLoading(true);
      setError(null);

      try {
        const resp = await generateQuestion({ stage, messages: updated });
        setMessages((prev) => [
          ...prev,
          { role: "assistant", content: resp.question },
        ]);
        setStage(resp.stage);
        if (resp.is_complete) {
          setIsComplete(true);
          stopTimer();
          return true;
        }
        if (resp.next_stage && resp.next_stage !== resp.stage) {
          setStage(resp.next_stage as InterviewStage);
        }
        return false;
      } catch (e) {
        setError(e instanceof Error ? e.message : "질문 생성에 실패했습니다");
        return false;
      } finally {
        setIsLoading(false);
      }
    },
    [messages, stage, stopTimer]
  );

  const reset = useCallback(() => {
    stopTimer();
    setMessages([]);
    setStage("warmup");
    setIsLoading(false);
    setIsComplete(false);
    setElapsedSeconds(0);
    setError(null);
  }, [stopTimer]);

  return {
    messages,
    stage,
    isLoading,
    isComplete,
    elapsedSeconds,
    error,
    startInterview,
    sendMessage,
    reset,
  };
}
