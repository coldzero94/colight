import { describe, it, expect, vi, beforeEach } from "vitest";
import { submitFeedback } from "../feedback";

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

describe("feedback API", () => {
  describe("submitFeedback", () => {
    it("calls POST /v1/feedback with bug category", async () => {
      const mockResponse = {
        id: "fb-123",
        created_at: "2024-01-15T10:00:00Z",
      };
      mockPost.mockResolvedValue({ data: mockResponse });

      const input = {
        category: "bug" as const,
        content: "로그인이 안 됩니다.",
        page_url: "/login",
      };

      const result = await submitFeedback(input);

      expect(mockPost).toHaveBeenCalledWith("/v1/feedback", input);
      expect(result).toEqual(mockResponse);
    });

    it("calls POST /v1/feedback with improvement category", async () => {
      const mockResponse = {
        id: "fb-456",
        created_at: "2024-01-15T11:00:00Z",
      };
      mockPost.mockResolvedValue({ data: mockResponse });

      const input = {
        category: "improvement" as const,
        content: "경험 카드 디자인을 개선해주세요.",
        page_url: "/experiences",
      };

      const result = await submitFeedback(input);

      expect(mockPost).toHaveBeenCalledWith("/v1/feedback", input);
      expect(result).toEqual(mockResponse);
    });

    it("calls POST /v1/feedback without page_url", async () => {
      const mockResponse = {
        id: "fb-789",
        created_at: "2024-01-15T12:00:00Z",
      };
      mockPost.mockResolvedValue({ data: mockResponse });

      const input = {
        category: "other" as const,
        content: "전반적으로 좋은 서비스입니다.",
      };

      const result = await submitFeedback(input);

      expect(mockPost).toHaveBeenCalledWith("/v1/feedback", input);
      expect(result).toEqual(mockResponse);
    });
  });
});
