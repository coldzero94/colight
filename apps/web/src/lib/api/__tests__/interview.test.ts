import { describe, it, expect, vi, beforeEach } from "vitest";
import {
  generateQuestion,
  extractSTAR,
  saveInterviewExperience,
  type ChatMessage,
} from "../interview";

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

describe("interview API", () => {
  describe("generateQuestion", () => {
    it("calls POST /v1/interview/question with correct payload", async () => {
      const mockResponse = {
        question: "어떤 프로젝트를 진행하셨나요?",
        stage: "warmup",
        next_stage: "memory",
        is_complete: false,
      };
      mockPost.mockResolvedValue({ data: mockResponse });

      const result = await generateQuestion({
        stage: "warmup",
        messages: [
          { role: "assistant", content: "안녕하세요!" },
          { role: "user", content: "안녕하세요." },
        ],
      });

      expect(mockPost).toHaveBeenCalledWith("/v1/interview/question", {
        stage: "warmup",
        messages: [
          { role: "assistant", content: "안녕하세요!" },
          { role: "user", content: "안녕하세요." },
        ],
      });
      expect(result).toEqual(mockResponse);
    });
  });

  describe("extractSTAR", () => {
    it("calls POST /v1/interview/extract with messages", async () => {
      const mockResponse = {
        title: "프로젝트 리딩 경험",
        category: "직무",
        content: "팀을 이끌어 프로젝트를 완수했습니다.",
        result: "성공적으로 출시",
        star_situation: "신규 프로젝트 시작",
        star_task: "팀 리딩 필요",
        star_action: "효과적인 커뮤니케이션",
        star_result: "기한 내 출시 성공",
        keywords: ["리더십", "프로젝트관리"],
      };
      mockPost.mockResolvedValue({ data: mockResponse });

      const messages: ChatMessage[] = [
        { role: "assistant", content: "어떤 프로젝트였나요?" },
        { role: "user", content: "신규 서비스 개발 프로젝트였습니다." },
        { role: "assistant", content: "어려웠던 점은?" },
        { role: "user", content: "팀원 간 의견 조율이 어려웠습니다." },
      ];

      const result = await extractSTAR(messages);

      expect(mockPost).toHaveBeenCalledWith("/v1/interview/extract", {
        messages,
      });
      expect(result).toEqual(mockResponse);
    });
  });

  describe("saveInterviewExperience", () => {
    it("calls POST /v1/interview/save with correct payload", async () => {
      const mockResponse = {
        experience_id: "exp-123",
        tagged: true,
      };
      mockPost.mockResolvedValue({ data: mockResponse });

      const input = {
        title: "프로젝트 리딩 경험",
        category: "직무",
        content: "팀을 이끌어 프로젝트를 완수했습니다.",
        result: "성공적으로 출시",
        star_situation: "신규 프로젝트 시작",
        star_task: "팀 리딩 필요",
        star_action: "효과적인 커뮤니케이션",
        star_result: "기한 내 출시 성공",
        keywords: ["리더십", "프로젝트관리"],
      };

      const result = await saveInterviewExperience(input);

      expect(mockPost).toHaveBeenCalledWith("/v1/interview/save", input);
      expect(result).toEqual(mockResponse);
    });
  });
});
