import { describe, it, expect, vi, beforeEach } from "vitest";
import {
  fetchExperiences,
  fetchExperience,
  createExperience,
  updateExperience,
  deleteExperience,
  tagExperience,
} from "../experiences";

// Mock apiClient
vi.mock("@/lib/api-client", () => ({
  apiClient: {
    get: vi.fn(),
    post: vi.fn(),
    patch: vi.fn(),
    delete: vi.fn(),
  },
}));

import { apiClient } from "@/lib/api-client";

const mockGet = vi.mocked(apiClient.get);
const mockPost = vi.mocked(apiClient.post);
const mockPatch = vi.mocked(apiClient.patch);
const mockDelete = vi.mocked(apiClient.delete);

beforeEach(() => {
  vi.clearAllMocks();
});

describe("experiences API", () => {
  describe("fetchExperiences", () => {
    it("calls GET /v1/experiences with no params", async () => {
      const mockExperiences = {
        experiences: [
          {
            id: "exp-1",
            user_id: "u-1",
            title: "프로젝트 리딩",
            category: "직무",
            role: "팀장",
            content: "내용",
            result: "결과",
            star_situation: "상황",
            star_task: "과제",
            star_action: "행동",
            star_result: "결과",
            keywords: ["리더십"],
            source: "manual",
            is_archived: false,
            created_at: "2024-01-01T00:00:00Z",
            updated_at: "2024-01-01T00:00:00Z",
          },
        ],
      };
      mockGet.mockResolvedValue({ data: mockExperiences });

      const result = await fetchExperiences();

      expect(mockGet).toHaveBeenCalledWith("/v1/experiences", { params: undefined });
      expect(result).toEqual(mockExperiences.experiences);
    });

    it("calls GET /v1/experiences with params", async () => {
      const mockExperiences = { experiences: [] };
      mockGet.mockResolvedValue({ data: mockExperiences });

      const result = await fetchExperiences({
        sort: "latest",
        category: "직무",
        weapon: "W01",
      });

      expect(mockGet).toHaveBeenCalledWith("/v1/experiences", {
        params: { sort: "latest", category: "직무", weapon: "W01" },
      });
      expect(result).toEqual([]);
    });
  });

  describe("fetchExperience", () => {
    it("calls GET /v1/experiences/:id", async () => {
      const mockExperience = {
        id: "exp-1",
        user_id: "u-1",
        title: "프로젝트 리딩",
        category: "직무",
        role: "팀장",
        content: "내용",
        result: "결과",
        star_situation: "상황",
        star_task: "과제",
        star_action: "행동",
        star_result: "결과",
        keywords: ["리더십"],
        source: "manual",
        is_archived: false,
        created_at: "2024-01-01T00:00:00Z",
        updated_at: "2024-01-01T00:00:00Z",
      };
      mockGet.mockResolvedValue({ data: mockExperience });

      const result = await fetchExperience("exp-1");

      expect(mockGet).toHaveBeenCalledWith("/v1/experiences/exp-1");
      expect(result).toEqual(mockExperience);
    });
  });

  describe("createExperience", () => {
    it("calls POST /v1/experiences with correct payload", async () => {
      const mockResponse = { id: "exp-new" };
      mockPost.mockResolvedValue({ data: mockResponse });

      const result = await createExperience({
        title: "새 경험",
        category: "직무",
        role: "개발자",
        content: "내용",
        result: "결과",
        star_situation: "상황",
        star_task: "과제",
        star_action: "행동",
        star_result: "결과",
        keywords: ["개발"],
      });

      expect(mockPost).toHaveBeenCalledWith("/v1/experiences", {
        title: "새 경험",
        category: "직무",
        role: "개발자",
        content: "내용",
        result: "결과",
        star_situation: "상황",
        star_task: "과제",
        star_action: "행동",
        star_result: "결과",
        keywords: ["개발"],
      });
      expect(result).toEqual(mockResponse);
    });
  });

  describe("updateExperience", () => {
    it("calls PATCH /v1/experiences/:id with correct payload", async () => {
      const mockResponse = { id: "exp-1" };
      mockPatch.mockResolvedValue({ data: mockResponse });

      const result = await updateExperience("exp-1", {
        title: "수정된 제목",
        content: "수정된 내용",
      });

      expect(mockPatch).toHaveBeenCalledWith("/v1/experiences/exp-1", {
        title: "수정된 제목",
        content: "수정된 내용",
      });
      expect(result).toEqual(mockResponse);
    });
  });

  describe("deleteExperience", () => {
    it("calls DELETE /v1/experiences/:id", async () => {
      const mockResponse = { success: true };
      mockDelete.mockResolvedValue({ data: mockResponse });

      const result = await deleteExperience("exp-1");

      expect(mockDelete).toHaveBeenCalledWith("/v1/experiences/exp-1");
      expect(result).toEqual(mockResponse);
    });
  });

  describe("tagExperience", () => {
    it("calls POST /v1/experiences/:id/tag", async () => {
      const mockResponse = {
        primary_weapon: {
          code: "W01",
          confidence: 0.95,
          reasoning: "리더십 역량 명확",
        },
        secondary_weapons: [
          {
            code: "W02",
            confidence: 0.75,
            reasoning: "협업 요소 포함",
          },
        ],
      };
      mockPost.mockResolvedValue({ data: mockResponse });

      const result = await tagExperience("exp-1");

      expect(mockPost).toHaveBeenCalledWith("/v1/experiences/exp-1/tag");
      expect(result).toEqual(mockResponse);
    });
  });
});
