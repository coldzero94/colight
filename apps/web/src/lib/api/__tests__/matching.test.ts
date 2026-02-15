import { describe, it, expect, vi, beforeEach } from "vitest";
import { matchExperiences } from "../matching";

// Mock apiClient
vi.mock("@/lib/api-client", () => ({
  apiClient: {
    post: vi.fn(),
  },
}));

import { apiClient } from "@/lib/api-client";

const mockPost = vi.mocked(apiClient.post);

beforeEach(() => {
  vi.clearAllMocks();
});

describe("matching API", () => {
  describe("matchExperiences", () => {
    it("calls POST /v1/match with company name", async () => {
      const mockResponse = {
        company_name: "네이버",
        matches: [
          {
            experience_id: "exp-1",
            overall_fit: 85,
            job_relevance: 90,
            talent_fit: 80,
            uniqueness: 85,
            reasoning: "기술 역량과 인재상이 잘 부합합니다.",
            suggested_angle: "혁신적인 문제 해결 경험을 강조하세요.",
          },
          {
            experience_id: "exp-2",
            overall_fit: 75,
            job_relevance: 80,
            talent_fit: 70,
            uniqueness: 75,
            reasoning: "협업 경험이 두드러집니다.",
            suggested_angle: "팀워크와 소통 능력을 부각하세요.",
          },
        ],
        total: 2,
      };
      mockPost.mockResolvedValue({ data: mockResponse });

      const result = await matchExperiences("네이버");

      expect(mockPost).toHaveBeenCalledWith("/v1/match", {
        company_name: "네이버",
      });
      expect(result).toEqual(mockResponse);
    });
  });
});
