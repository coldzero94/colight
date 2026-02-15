import { describe, it, expect, vi, beforeEach } from "vitest";
import { getUsage } from "../usage";

// Mock apiClient
vi.mock("@/lib/api-client", () => ({
  apiClient: {
    get: vi.fn(),
  },
}));

import { apiClient } from "@/lib/api-client";

const mockGet = vi.mocked(apiClient.get);

beforeEach(() => {
  vi.clearAllMocks();
});

describe("usage API", () => {
  describe("getUsage", () => {
    it("calls GET /v1/usage and returns usage data", async () => {
      const mockUsage = {
        plan: "free" as const,
        features: {
          experience_tagging: {
            allowed: true,
            used: 3,
            limit: 5,
            remaining: 2,
          },
          company_analysis: {
            allowed: true,
            used: 2,
            limit: 3,
            remaining: 1,
          },
          draft_coaching: {
            allowed: true,
            used: 1,
            limit: 3,
            remaining: 2,
          },
        },
      };
      mockGet.mockResolvedValue({ data: mockUsage });

      const result = await getUsage();

      expect(mockGet).toHaveBeenCalledWith("/v1/usage");
      expect(result).toEqual(mockUsage);
    });

    it("handles pro plan with unlimited features", async () => {
      const mockUsage = {
        plan: "pro" as const,
        features: {
          experience_tagging: {
            allowed: true,
            used: 50,
            limit: -1,
            remaining: -1,
          },
          company_analysis: {
            allowed: true,
            used: 30,
            limit: -1,
            remaining: -1,
          },
          draft_coaching: {
            allowed: true,
            used: 20,
            limit: -1,
            remaining: -1,
          },
        },
      };
      mockGet.mockResolvedValue({ data: mockUsage });

      const result = await getUsage();

      expect(mockGet).toHaveBeenCalledWith("/v1/usage");
      expect(result).toEqual(mockUsage);
    });
  });
});
