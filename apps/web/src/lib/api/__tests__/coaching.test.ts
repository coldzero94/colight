import { describe, it, expect, vi, beforeEach } from "vitest";
import {
  analyzeQuestion,
  generateDraft,
  getCoachingSessions,
  getCoverLetter,
  updateCoverLetter,
  createVersion,
  getVersions,
} from "../coaching";

// Mock apiClient
vi.mock("@/lib/api-client", () => ({
  apiClient: {
    post: vi.fn(),
    get: vi.fn(),
    patch: vi.fn(),
  },
}));

import { apiClient } from "@/lib/api-client";

const mockPost = vi.mocked(apiClient.post);
const mockGet = vi.mocked(apiClient.get);
const mockPatch = vi.mocked(apiClient.patch);

beforeEach(() => {
  vi.clearAllMocks();
});

describe("coaching API", () => {
  describe("analyzeQuestion", () => {
    it("calls POST /v1/coaching/question-analysis with correct payload", async () => {
      const mockResult = {
        surface_question: "test",
        real_intents: [],
        required_weapons: { primary: {}, secondary: [] },
        writing_structure: { total_chars: 800, sections: [] },
        key_keywords: [],
      };
      mockPost.mockResolvedValue({ data: mockResult });

      const result = await analyzeQuestion({
        application_id: "app-123",
        question_text: "테스트 문항입니다 (최소 10자)",
        char_limit: 800,
      });

      expect(mockPost).toHaveBeenCalledWith("/v1/coaching/question-analysis", {
        application_id: "app-123",
        question_text: "테스트 문항입니다 (최소 10자)",
        char_limit: 800,
      });
      expect(result).toEqual(mockResult);
    });
  });

  describe("generateDraft", () => {
    it("calls POST /v1/coaching/draft with correct payload", async () => {
      const mockResult = { draft: "초안 내용..." };
      mockPost.mockResolvedValue({ data: mockResult });

      const result = await generateDraft({
        application_id: "app-123",
        experience_ids: ["exp-1", "exp-2"],
        question_text: "테스트 문항입니다 (최소 10자)",
        char_limit: 800,
      });

      expect(mockPost).toHaveBeenCalledWith("/v1/coaching/draft", {
        application_id: "app-123",
        experience_ids: ["exp-1", "exp-2"],
        question_text: "테스트 문항입니다 (최소 10자)",
        char_limit: 800,
      });
      expect(result).toEqual(mockResult);
    });
  });

  describe("getCoachingSessions", () => {
    it("calls GET /v1/coaching/sessions with cover_letter_id", async () => {
      const mockSessions = { sessions: [] };
      mockGet.mockResolvedValue({ data: mockSessions });

      const result = await getCoachingSessions("cl-123");

      expect(mockGet).toHaveBeenCalledWith("/v1/coaching/sessions", {
        params: { cover_letter_id: "cl-123" },
      });
      expect(result).toEqual(mockSessions);
    });
  });

  describe("getCoverLetter", () => {
    it("calls GET /v1/coaching/cover-letters/:id", async () => {
      const mockCL = { id: "cl-123", current_content: "test" };
      mockGet.mockResolvedValue({ data: mockCL });

      const result = await getCoverLetter("cl-123");

      expect(mockGet).toHaveBeenCalledWith("/v1/coaching/cover-letters/cl-123");
      expect(result).toEqual(mockCL);
    });
  });

  describe("updateCoverLetter", () => {
    it("calls PATCH /v1/coaching/cover-letters/:id", async () => {
      mockPatch.mockResolvedValue({ data: { updated_at: "now" } });

      const result = await updateCoverLetter("cl-123", "new content");

      expect(mockPatch).toHaveBeenCalledWith(
        "/v1/coaching/cover-letters/cl-123",
        { content: "new content" }
      );
      expect(result).toEqual({ updated_at: "now" });
    });
  });

  describe("createVersion", () => {
    it("calls POST /v1/coaching/cover-letters/:id/versions", async () => {
      const mockVersion = { version: { id: "v-1", version_number: 1 } };
      mockPost.mockResolvedValue({ data: mockVersion });

      const result = await createVersion("cl-123", "version content");

      expect(mockPost).toHaveBeenCalledWith(
        "/v1/coaching/cover-letters/cl-123/versions",
        { content: "version content" }
      );
      expect(result).toEqual(mockVersion);
    });
  });

  describe("getVersions", () => {
    it("calls GET /v1/coaching/cover-letters/:id/versions", async () => {
      const mockVersions = { versions: [] };
      mockGet.mockResolvedValue({ data: mockVersions });

      const result = await getVersions("cl-123");

      expect(mockGet).toHaveBeenCalledWith(
        "/v1/coaching/cover-letters/cl-123/versions"
      );
      expect(result).toEqual(mockVersions);
    });
  });
});
