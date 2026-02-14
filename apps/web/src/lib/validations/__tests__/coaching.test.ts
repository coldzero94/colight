import { describe, it, expect } from "vitest";
import { questionAnalysisSchema } from "../coaching";

describe("questionAnalysisSchema", () => {
  it("validates question_text minimum length (10 chars)", () => {
    const result = questionAnalysisSchema.safeParse({
      application_id: "123e4567-e89b-12d3-a456-426614174000",
      question_text: "본인이 팀 프로젝트에서 어려움을 극복한 경험을 구체적으로 기술하세요.",
      char_limit: 800,
    });
    expect(result.success).toBe(true);
  });

  it("rejects question_text under 10 chars", () => {
    const result = questionAnalysisSchema.safeParse({
      application_id: "123e4567-e89b-12d3-a456-426614174000",
      question_text: "짧은 문항",
      char_limit: 800,
    });
    expect(result.success).toBe(false);
    if (!result.success) {
      const questionError = result.error.issues.find((i) =>
        i.path.includes("question_text")
      );
      expect(questionError).toBeDefined();
      expect(questionError?.message).toContain("최소 10자");
    }
  });

  it("validates char_limit range (200-2000)", () => {
    const validLimits = [200, 800, 1500, 2000];
    for (const limit of validLimits) {
      const result = questionAnalysisSchema.safeParse({
        application_id: "123e4567-e89b-12d3-a456-426614174000",
        question_text: "본인이 팀 프로젝트에서 어려움을 극복한 경험을 구체적으로 기술하세요.",
        char_limit: limit,
      });
      expect(result.success).toBe(true);
    }
  });

  it("rejects char_limit below 200", () => {
    const result = questionAnalysisSchema.safeParse({
      application_id: "123e4567-e89b-12d3-a456-426614174000",
      question_text: "본인이 팀 프로젝트에서 어려움을 극복한 경험을 구체적으로 기술하세요.",
      char_limit: 199,
    });
    expect(result.success).toBe(false);
    if (!result.success) {
      const limitError = result.error.issues.find((i) =>
        i.path.includes("char_limit")
      );
      expect(limitError).toBeDefined();
      expect(limitError?.message).toContain("200자");
    }
  });

  it("rejects char_limit above 2000", () => {
    const result = questionAnalysisSchema.safeParse({
      application_id: "123e4567-e89b-12d3-a456-426614174000",
      question_text: "본인이 팀 프로젝트에서 어려움을 극복한 경험을 구체적으로 기술하세요.",
      char_limit: 2001,
    });
    expect(result.success).toBe(false);
    if (!result.success) {
      const limitError = result.error.issues.find((i) =>
        i.path.includes("char_limit")
      );
      expect(limitError).toBeDefined();
      expect(limitError?.message).toContain("2000자");
    }
  });

  it("requires valid UUID for application_id", () => {
    const result = questionAnalysisSchema.safeParse({
      application_id: "invalid-uuid",
      question_text: "본인이 팀 프로젝트에서 어려움을 극복한 경험을 구체적으로 기술하세요.",
      char_limit: 800,
    });
    expect(result.success).toBe(false);
    if (!result.success) {
      const idError = result.error.issues.find((i) =>
        i.path.includes("application_id")
      );
      expect(idError).toBeDefined();
    }
  });

  it("rejects missing required fields", () => {
    const result = questionAnalysisSchema.safeParse({});
    expect(result.success).toBe(false);
    if (!result.success) {
      expect(result.error.issues.length).toBeGreaterThan(0);
    }
  });
});
