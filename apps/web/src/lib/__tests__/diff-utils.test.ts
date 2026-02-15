import { describe, it, expect } from "vitest";
import { computeDiff } from "../diff-utils";

describe("computeDiff", () => {
  it("returns unchanged for identical strings", () => {
    const result = computeDiff("hello", "hello");
    expect(result.parts).toEqual([{ value: "hello", type: "unchanged" }]);
    expect(result.stats).toEqual({ added: 0, removed: 0, unchanged: 5 });
  });

  it("detects added text", () => {
    const result = computeDiff("", "새로운 텍스트");
    expect(result.parts).toEqual([{ value: "새로운 텍스트", type: "added" }]);
    expect(result.stats.added).toBe(7);
    expect(result.stats.removed).toBe(0);
  });

  it("detects removed text", () => {
    const result = computeDiff("삭제될 내용", "");
    expect(result.parts).toEqual([{ value: "삭제될 내용", type: "removed" }]);
    expect(result.stats.removed).toBe(6);
    expect(result.stats.added).toBe(0);
  });

  it("handles mixed changes with Korean text", () => {
    const result = computeDiff("저는 개발자입니다", "저는 디자이너입니다");
    expect(result.parts.length).toBeGreaterThan(1);
    expect(result.stats.added).toBeGreaterThan(0);
    expect(result.stats.removed).toBeGreaterThan(0);
    expect(result.stats.unchanged).toBeGreaterThan(0);
  });

  it("handles both empty strings", () => {
    const result = computeDiff("", "");
    expect(result.parts).toEqual([]);
    expect(result.stats).toEqual({ added: 0, removed: 0, unchanged: 0 });
  });

  it("counts Korean characters correctly", () => {
    const result = computeDiff("가나다", "가라마");
    // "가" unchanged, "나다" removed, "라마" added
    expect(result.stats.unchanged).toBe(1);
    expect(result.stats.removed).toBe(2);
    expect(result.stats.added).toBe(2);
  });
});
