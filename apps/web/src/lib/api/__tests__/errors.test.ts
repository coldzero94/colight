import { describe, it, expect } from "vitest";
import { AxiosError, type InternalAxiosRequestConfig } from "axios";
import {
  isUsageLimitError,
  getUsageLimitInfo,
  type ApiErrorPayload,
} from "../errors";

/** Helper to create a typed AxiosError with response data. */
function createAxiosError(
  message: string,
  response?: {
    data: ApiErrorPayload | Record<string, unknown>;
    status: number;
    statusText: string;
  },
): AxiosError<ApiErrorPayload> {
  const error = new AxiosError<ApiErrorPayload>(message);
  if (response) {
    error.response = {
      ...response,
      data: response.data as ApiErrorPayload,
      headers: {},
      config: {} as InternalAxiosRequestConfig,
    };
  }
  return error;
}

describe("errors API utilities", () => {
  describe("isUsageLimitError", () => {
    it("returns true for AxiosError with USAGE_001 code", () => {
      const error = createAxiosError("Usage limit exceeded", {
        data: {
          error: {
            message: "무료 사용 횟수를 초과했습니다.",
            code: "USAGE_001",
            used: 5,
            limit: 5,
          },
        },
        status: 429,
        statusText: "Too Many Requests",
      });

      expect(isUsageLimitError(error)).toBe(true);
    });

    it("returns false for AxiosError with different code", () => {
      const error = createAxiosError("Bad request", {
        data: {
          error: {
            message: "잘못된 요청입니다.",
            code: "BAD_REQUEST",
          },
        },
        status: 400,
        statusText: "Bad Request",
      });

      expect(isUsageLimitError(error)).toBe(false);
    });

    it("returns false for non-AxiosError", () => {
      const error = new Error("Generic error");

      expect(isUsageLimitError(error)).toBe(false);
    });

    it("returns false for AxiosError without response", () => {
      const error = createAxiosError("Network error");

      expect(isUsageLimitError(error)).toBe(false);
    });

    it("returns false for AxiosError with malformed response", () => {
      const error = createAxiosError("Server error", {
        data: {},
        status: 500,
        statusText: "Internal Server Error",
      });

      expect(isUsageLimitError(error)).toBe(false);
    });
  });

  describe("getUsageLimitInfo", () => {
    it("extracts message, used, and limit from error response", () => {
      const error = createAxiosError("Usage limit exceeded", {
        data: {
          error: {
            message: "무료 사용 횟수를 초과했습니다.",
            code: "USAGE_001",
            used: 5,
            limit: 5,
          },
        },
        status: 429,
        statusText: "Too Many Requests",
      });

      const info = getUsageLimitInfo(error);

      expect(info).toEqual({
        message: "무료 사용 횟수를 초과했습니다.",
        used: 5,
        limit: 5,
      });
    });

    it("returns default values when data is missing", () => {
      const error = createAxiosError("Usage limit exceeded", {
        data: {
          error: {
            message: "",
            code: "USAGE_001",
          },
        },
        status: 429,
        statusText: "Too Many Requests",
      });

      const info = getUsageLimitInfo(error);

      expect(info).toEqual({
        message: "",
        used: 0,
        limit: 0,
      });
    });

    it("returns default values when response is undefined", () => {
      const error = createAxiosError("Network error");

      const info = getUsageLimitInfo(error);

      expect(info).toEqual({
        message: "무료 사용 횟수를 초과했습니다.",
        used: 0,
        limit: 0,
      });
    });
  });
});
