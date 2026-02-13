import { describe, it, expect } from "vitest";
import { experienceSchema } from "../experience";

describe("experienceSchema", () => {
  it("validates valid experience data", () => {
    const result = experienceSchema.safeParse({
      title: "인턴 경험",
      category: "인턴",
      role: "백엔드 개발",
      star_situation: "스타트업에서 인턴",
      star_task: "API 개발",
      star_action: "Go로 REST API 구현",
      star_result: "3개월 내 5개 API 완성",
    });
    expect(result.success).toBe(true);
  });

  it("rejects empty title", () => {
    const result = experienceSchema.safeParse({
      title: "",
    });
    expect(result.success).toBe(false);
    if (!result.success) {
      const titleError = result.error.issues.find((i) =>
        i.path.includes("title")
      );
      expect(titleError).toBeDefined();
    }
  });

  it("rejects title exceeding 100 chars", () => {
    const result = experienceSchema.safeParse({
      title: "a".repeat(101),
    });
    expect(result.success).toBe(false);
    if (!result.success) {
      const titleError = result.error.issues.find((i) =>
        i.path.includes("title")
      );
      expect(titleError).toBeDefined();
    }
  });

  it("rejects action exceeding 2000 chars", () => {
    const result = experienceSchema.safeParse({
      title: "Valid Title",
      star_action: "a".repeat(2001),
    });
    expect(result.success).toBe(false);
    if (!result.success) {
      const actionError = result.error.issues.find((i) =>
        i.path.includes("star_action")
      );
      expect(actionError).toBeDefined();
    }
  });

  it("validates period_end is after period_start", () => {
    const result = experienceSchema.safeParse({
      title: "기간 테스트",
      period_start: "2024-06-30",
      period_end: "2024-01-01",
    });
    expect(result.success).toBe(false);
    if (!result.success) {
      const periodError = result.error.issues.find((i) =>
        i.path.includes("period_end")
      );
      expect(periodError).toBeDefined();
    }
  });

  it("accepts valid category enum values", () => {
    const categories = [
      "인턴",
      "대외활동",
      "프로젝트",
      "아르바이트",
      "동아리",
      "봉사활동",
      "기타",
    ];
    for (const category of categories) {
      const result = experienceSchema.safeParse({
        title: "Test",
        category,
      });
      expect(result.success).toBe(true);
    }
  });
});
