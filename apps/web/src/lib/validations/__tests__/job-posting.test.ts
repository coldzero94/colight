import { describe, it, expect } from "vitest";
import { jobPostingSchema } from "../job-posting";

describe("JobPostingSchema", () => {
  it("validates complete job posting data", () => {
    const result = jobPostingSchema.safeParse({
      company_name: "네이버",
      position: "백엔드 개발자",
      department: "서비스개발팀",
      job_type: "정규직",
      experience_level: "경력 5년 이상",
      location: "경기 성남시 분당구",
      main_tasks: ["API 설계", "DB 설계"],
      requirements: ["Go 경험", "PostgreSQL"],
      preferred: ["Docker", "K8s"],
      required_skills: ["Go", "PostgreSQL"],
      soft_skills: ["커뮤니케이션"],
      company_values_hints: ["혁신"],
      deadline: "2026-06-30",
    });
    expect(result.success).toBe(true);
    if (result.success) {
      expect(result.data.company_name).toBe("네이버");
      expect(result.data.main_tasks).toHaveLength(2);
    }
  });

  it("validates minimal required fields only", () => {
    const result = jobPostingSchema.safeParse({
      company_name: "테스트 회사",
      position: "개발자",
    });
    expect(result.success).toBe(true);
    if (result.success) {
      // Defaults should be applied
      expect(result.data.main_tasks).toEqual([]);
      expect(result.data.requirements).toEqual([]);
      expect(result.data.required_skills).toEqual([]);
    }
  });

  it("rejects missing company_name", () => {
    const result = jobPostingSchema.safeParse({
      position: "개발자",
    });
    expect(result.success).toBe(false);
  });

  it("rejects missing position", () => {
    const result = jobPostingSchema.safeParse({
      company_name: "테스트 회사",
    });
    expect(result.success).toBe(false);
  });

  it("rejects empty company_name", () => {
    const result = jobPostingSchema.safeParse({
      company_name: "",
      position: "개발자",
    });
    expect(result.success).toBe(false);
    if (!result.success) {
      const companyError = result.error.issues.find(
        (i) => i.path[0] === "company_name"
      );
      expect(companyError?.message).toContain("회사명은 필수");
    }
  });

  it("rejects invalid job_type enum value", () => {
    const result = jobPostingSchema.safeParse({
      company_name: "회사",
      position: "개발자",
      job_type: "잘못된값",
    });
    expect(result.success).toBe(false);
  });

  it("accepts valid job_type enum values", () => {
    const validTypes = ["정규직", "계약직", "인턴", "파견직", "기타"];
    for (const jt of validTypes) {
      const result = jobPostingSchema.safeParse({
        company_name: "회사",
        position: "개발자",
        job_type: jt,
      });
      expect(result.success).toBe(true);
    }
  });

  it("rejects invalid source_url", () => {
    const result = jobPostingSchema.safeParse({
      company_name: "회사",
      position: "개발자",
      source_url: "not-a-url",
    });
    expect(result.success).toBe(false);
  });

  it("accepts valid source_url", () => {
    const result = jobPostingSchema.safeParse({
      company_name: "회사",
      position: "개발자",
      source_url: "https://www.jobkorea.co.kr/Recruit/GI_Read/12345",
    });
    expect(result.success).toBe(true);
  });
});
