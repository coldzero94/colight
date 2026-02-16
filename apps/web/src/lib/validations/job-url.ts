import { z } from "zod/v4";

// Zod schema for job posting URL validation
export const jobUrlSchema = z.object({
  url: z.string().url("유효한 URL을 입력해주세요").min(1, "URL을 입력해주세요"),
});

export type JobUrlFormValues = z.infer<typeof jobUrlSchema>;

export interface DomainInfo {
  domain: string;
  label: string;
}

// Domain detection map — all domains are supported via universal crawling pipeline
const DOMAIN_MAP: Record<string, DomainInfo> = {
  "jobkorea.co.kr": { domain: "jobkorea", label: "잡코리아" },
  "catch.co.kr": { domain: "catch", label: "캐치" },
  "wanted.co.kr": { domain: "wanted", label: "원티드" },
  "saramin.co.kr": { domain: "saramin", label: "사람인" },
  "programmers.co.kr": { domain: "programmers", label: "프로그래머스" },
  "jumpit.saramin.co.kr": { domain: "jumpit", label: "점핏" },
  "rocketpunch.com": { domain: "rocketpunch", label: "로켓펀치" },
  "linkedin.com": { domain: "linkedin", label: "LinkedIn" },
};

/**
 * Detects the job posting domain from a URL
 */
export function detectDomain(url: string): DomainInfo | null {
  try {
    const urlObj = new URL(url);
    const hostname = urlObj.hostname;

    for (const [pattern, info] of Object.entries(DOMAIN_MAP)) {
      if (hostname.includes(pattern)) {
        return info;
      }
    }

    return null;
  } catch {
    return null;
  }
}
