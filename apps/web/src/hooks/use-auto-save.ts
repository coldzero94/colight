"use client";

import { useState, useEffect, useRef, useCallback } from "react";
import { useDebouncedCallback } from "use-debounce";

type SaveStatus = "saved" | "saving" | "unsaved";

export function useAutoSave(coverLetterId: string) {
  const [status, setStatus] = useState<SaveStatus>("saved");
  const lastSavedContent = useRef<string>("");

  // 자동 저장 (2초 debounce)
  const debouncedSave = useDebouncedCallback(
    async (content: string) => {
      if (content === lastSavedContent.current) return;

      setStatus("saving");
      try {
        await fetch(`/v1/coaching/cover-letters/${coverLetterId}`, {
          method: "PATCH",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ content }),
        });
        lastSavedContent.current = content;
        setStatus("saved");
      } catch (error) {
        setStatus("unsaved");
      }
    },
    2000
  );

  // 명시적 버전 저장
  const saveVersion = useCallback(
    async (content: string) => {
      setStatus("saving");
      try {
        await fetch(`/v1/coaching/cover-letters/${coverLetterId}/versions`, {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ content }),
        });
        lastSavedContent.current = content;
        setStatus("saved");
      } catch (error) {
        setStatus("unsaved");
      }
    },
    [coverLetterId]
  );

  // 페이지 이탈 경고
  useEffect(() => {
    const handler = (e: BeforeUnloadEvent) => {
      if (status === "unsaved") {
        e.preventDefault();
      }
    };
    window.addEventListener("beforeunload", handler);
    return () => window.removeEventListener("beforeunload", handler);
  }, [status]);

  return { status, debouncedSave, saveVersion };
}
