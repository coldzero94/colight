import { describe, it, expect } from "vitest";
import { renderHook } from "@testing-library/react";
import { useAutoSave } from "../use-auto-save";

describe("useAutoSave", () => {
  it("initializes with saved status", () => {
    const { result } = renderHook(() => useAutoSave("test-id"));
    expect(result.current.status).toBe("saved");
  });

  it("provides debouncedSave function", () => {
    const { result } = renderHook(() => useAutoSave("test-id"));
    expect(typeof result.current.debouncedSave).toBe("function");
  });

  it("provides saveVersion function", () => {
    const { result } = renderHook(() => useAutoSave("test-id"));
    expect(typeof result.current.saveVersion).toBe("function");
  });

  it("has status property", () => {
    const { result } = renderHook(() => useAutoSave("test-id"));
    expect(result.current.status).toBeDefined();
    expect(["saved", "saving", "unsaved"]).toContain(result.current.status);
  });

  // TODO: E2E tests for debounce timing, network calls, beforeunload
  // These require full DOM environment and async handling
});
