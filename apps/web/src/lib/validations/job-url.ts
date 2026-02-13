import { z } from "zod/v4";

// Zod schema for job posting URL validation
export const jobUrlSchema = z.object({
  url: z.string().url("유효한 URL을 입력해주세요").min(1, "URL을 입력해주세요"),
});

export type JobUrlFormValues = z.infer<typeof jobUrlSchema>;

// Supported job posting domains
export type SupportedDomain = "jobkorea" | "catch" | "wanted" | "saramin";

export interface DomainInfo {
  domain: SupportedDomain | "unknown";
  label: string;
  supported: boolean;
  method: "goquery" | "playwright" | "api" | "ai-fallback";
}

// Domain detection map
const DOMAIN_MAP: Record<string, DomainInfo> = {
  "jobkorea.co.kr": {
    domain: "jobkorea",
    label: "잡코리아",
    supported: true,
    method: "goquery",
  },
  "catch.co.kr": {
    domain: "catch",
    label: "캐치",
    supported: true,
    method: "goquery",
  },
  "wanted.co.kr": {
    domain: "wanted",
    label: "원티드",
    supported: false,
    method: "playwright",
  },
  "saramin.co.kr": {
    domain: "saramin",
    label: "사람인",
    supported: false,
    method: "api",
  },
};

/**
 * Detects the job posting domain from a URL
 */
export function detectDomain(url: string): DomainInfo {
  try {
    const urlObj = new URL(url);
    const hostname = urlObj.hostname;

    // Check each domain pattern
    for (const [pattern, info] of Object.entries(DOMAIN_MAP)) {
      if (hostname.includes(pattern)) {
        return info;
      }
    }

    // Unknown domain
    return {
      domain: "unknown",
      label: "기타",
      supported: false,
      method: "ai-fallback",
    };
  } catch {
    return {
      domain: "unknown",
      label: "기타",
      supported: false,
      method: "ai-fallback",
    };
  }
}
