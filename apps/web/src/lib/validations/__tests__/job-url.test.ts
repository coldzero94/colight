import { describe, it, expect } from "vitest";
import { jobUrlSchema, detectDomain } from "../job-url";

describe("JobUrlSchema", () => {
  it("validates correct URL", () => {
    const result = jobUrlSchema.safeParse({
      url: "https://www.jobkorea.co.kr/Recruit/GI_Read/12345",
    });
    expect(result.success).toBe(true);
  });

  it("rejects empty URL", () => {
    const result = jobUrlSchema.safeParse({ url: "" });
    expect(result.success).toBe(false);
    if (!result.success) {
      expect(result.error.issues[0].message).toContain("URL을 입력해주세요");
    }
  });

  it("rejects invalid URL format", () => {
    const result = jobUrlSchema.safeParse({ url: "not-a-url" });
    expect(result.success).toBe(false);
    if (!result.success) {
      expect(result.error.issues[0].message).toContain("유효한 URL");
    }
  });
});

describe("detectDomain", () => {
  it("detects jobkorea domain", () => {
    const info = detectDomain("https://www.jobkorea.co.kr/Recruit/GI_Read/12345");
    expect(info).not.toBeNull();
    expect(info!.domain).toBe("jobkorea");
    expect(info!.label).toBe("잡코리아");
  });

  it("detects catch domain", () => {
    const info = detectDomain("https://www.catch.co.kr/NCS/RecruitInfoDetail/12345");
    expect(info).not.toBeNull();
    expect(info!.domain).toBe("catch");
    expect(info!.label).toBe("캐치");
  });

  it("detects wanted domain", () => {
    const info = detectDomain("https://www.wanted.co.kr/wd/12345");
    expect(info).not.toBeNull();
    expect(info!.domain).toBe("wanted");
    expect(info!.label).toBe("원티드");
  });

  it("detects saramin domain", () => {
    const info = detectDomain("https://www.saramin.co.kr/zf_user/jobs/relay/view?rec_idx=12345");
    expect(info).not.toBeNull();
    expect(info!.domain).toBe("saramin");
    expect(info!.label).toBe("사람인");
  });

  it("detects programmers domain", () => {
    const info = detectDomain("https://career.programmers.co.kr/job_positions/12345");
    expect(info).not.toBeNull();
    expect(info!.domain).toBe("programmers");
    expect(info!.label).toBe("프로그래머스");
  });

  it("detects linkedin domain", () => {
    const info = detectDomain("https://www.linkedin.com/jobs/view/12345");
    expect(info).not.toBeNull();
    expect(info!.domain).toBe("linkedin");
    expect(info!.label).toBe("LinkedIn");
  });

  it("returns null for unknown domains", () => {
    const info = detectDomain("https://www.example.com/job/12345");
    expect(info).toBeNull();
  });

  it("handles invalid URL gracefully", () => {
    const info = detectDomain("not-a-url");
    expect(info).toBeNull();
  });
});
