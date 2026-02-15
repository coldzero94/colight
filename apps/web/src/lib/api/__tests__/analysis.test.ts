import { describe, it, expect, vi, beforeEach } from "vitest";
import { analyzeCompany, getCompanyData } from "../analysis";

// Mock apiClient
vi.mock("@/lib/api-client", () => ({
  apiClient: {
    post: vi.fn(),
    get: vi.fn(),
  },
}));

import { apiClient } from "@/lib/api-client";

const mockPost = vi.mocked(apiClient.post);
const mockGet = vi.mocked(apiClient.get);

beforeEach(() => {
  vi.clearAllMocks();
});

describe("analysis API", () => {
  describe("analyzeCompany", () => {
    it("calls POST /v1/analyze-company with company name", async () => {
      const mockAnalysis = {
        company_name: "카카오",
        core_values: [
          { keyword: "혁신", description: "끊임없는 기술 혁신" },
        ],
        talent_traits: [
          { trait: "주도성", description: "자기주도적 문제 해결" },
        ],
        recent_trends: [
          { title: "AI 사업 확장", summary: "AI 기술 투자 증가" },
        ],
        strategy_keywords: ["AI", "플랫폼"],
        avoid_expressions: ["구태의연", "보수적"],
        source: "ai_generated" as const,
      };
      mockPost.mockResolvedValue({ data: mockAnalysis });

      const result = await analyzeCompany("카카오");

      expect(mockPost).toHaveBeenCalledWith("/v1/analyze-company", {
        company_name: "카카오",
      });
      expect(result).toEqual(mockAnalysis);
    });
  });

  describe("getCompanyData", () => {
    it("calls GET /v1/company-data with company name param", async () => {
      const mockData = {
        basic_info: {
          corp_name: "주식회사 카카오",
          corp_code: "00123456",
          stock_code: "035720",
          ceo: "홍길동",
          industry: "IT서비스",
          address: "경기도 성남시",
        },
        news: [
          {
            title: "카카오, AI 신사업 진출",
            link: "https://example.com/news1",
            description: "AI 분야 투자 확대",
            pub_date: "2024-01-15",
            source: "뉴스1",
          },
        ],
      };
      mockGet.mockResolvedValue({ data: mockData });

      const result = await getCompanyData("카카오");

      expect(mockGet).toHaveBeenCalledWith("/v1/company-data", {
        params: { name: "카카오" },
      });
      expect(result).toEqual(mockData);
    });
  });
});
