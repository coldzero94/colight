import { describe, it, expect, vi, beforeEach } from "vitest";
import type { InternalAxiosRequestConfig, AxiosResponse } from "axios";

// vi.hoisted ensures the variable is available during vi.mock hoisting
const mockGetState = vi.hoisted(() =>
  vi.fn(() => ({
    accessToken: null as string | null,
    refreshAccessToken: vi.fn(),
  }))
);

vi.mock("@/stores/auth-store", () => ({
  useAuthStore: {
    getState: mockGetState,
  },
}));

import { apiClient } from "../api-client";

/**
 * Axios interceptor handlers are internal — extract with proper types.
 * These helpers narrow the handler types for test assertions.
 */
function getRequestHandler() {
  const handler = apiClient.interceptors.request.handlers![0];
  return {
    fulfilled: handler.fulfilled as (
      config: InternalAxiosRequestConfig,
    ) => InternalAxiosRequestConfig,
  };
}

function getResponseHandler() {
  const handler = apiClient.interceptors.response.handlers![0];
  return {
    fulfilled: handler.fulfilled as (response: AxiosResponse) => AxiosResponse,
    rejected: handler.rejected as (error: unknown) => Promise<never>,
  };
}

describe("apiClient", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe("configuration", () => {
    it("uses default base URL when NEXT_PUBLIC_API_URL is not set", () => {
      expect(apiClient.defaults.baseURL).toBe("http://localhost:9000");
    });

    it("has Content-Type application/json header", () => {
      expect(apiClient.defaults.headers["Content-Type"]).toBe(
        "application/json"
      );
    });
  });

  describe("request interceptor", () => {
    it("attaches Bearer token when accessToken exists", () => {
      mockGetState.mockReturnValue({
        accessToken: "test-access-token",
        refreshAccessToken: vi.fn(),
      });

      const { fulfilled } = getRequestHandler();
      const config = { headers: {} } as InternalAxiosRequestConfig;
      const result = fulfilled(config);

      expect(result.headers.Authorization).toBe("Bearer test-access-token");
    });

    it("does not attach Authorization header when accessToken is null", () => {
      mockGetState.mockReturnValue({
        accessToken: null,
        refreshAccessToken: vi.fn(),
      });

      const { fulfilled } = getRequestHandler();
      const config = { headers: {} } as InternalAxiosRequestConfig;
      const result = fulfilled(config);

      expect(result.headers.Authorization).toBeUndefined();
    });
  });

  describe("response interceptor", () => {
    it("returns response as-is on success", () => {
      const { fulfilled } = getResponseHandler();
      const response = {
        data: { data: "success" },
        status: 200,
        statusText: "OK",
        headers: {},
        config: {} as InternalAxiosRequestConfig,
      } as AxiosResponse;

      const result = fulfilled(response);

      expect(result).toEqual(response);
    });

    it("does not retry on 401 if refresh fails", async () => {
      const mockRefresh = vi.fn().mockResolvedValue(false);

      mockGetState.mockReturnValue({
        accessToken: "old-token",
        refreshAccessToken: mockRefresh,
      });

      const error = {
        config: { headers: {} },
        response: { status: 401 },
      };

      const { rejected } = getResponseHandler();

      await expect(rejected(error)).rejects.toEqual(error);
      expect(mockRefresh).toHaveBeenCalledOnce();
    });

    it("does not retry on 401 if already retried", async () => {
      const mockRefresh = vi.fn();

      mockGetState.mockReturnValue({
        accessToken: "token",
        refreshAccessToken: mockRefresh,
      });

      const error = {
        config: { headers: {}, _retry: true },
        response: { status: 401 },
      };

      const { rejected } = getResponseHandler();

      await expect(rejected(error)).rejects.toEqual(error);
      expect(mockRefresh).not.toHaveBeenCalled();
    });

    it("does not retry on non-401 errors", async () => {
      const mockRefresh = vi.fn();

      mockGetState.mockReturnValue({
        accessToken: "token",
        refreshAccessToken: mockRefresh,
      });

      const error = {
        config: { headers: {} },
        response: { status: 500 },
      };

      const { rejected } = getResponseHandler();

      await expect(rejected(error)).rejects.toEqual(error);
      expect(mockRefresh).not.toHaveBeenCalled();
    });

    it("does not retry when refresh endpoint itself returns 401", async () => {
      const mockRefresh = vi.fn();

      mockGetState.mockReturnValue({
        accessToken: "expired-token",
        refreshAccessToken: mockRefresh,
      });

      const error = {
        config: { headers: {}, url: "/v1/auth/refresh" },
        response: { status: 401 },
      };

      const { rejected } = getResponseHandler();

      await expect(rejected(error)).rejects.toEqual(error);
      expect(mockRefresh).not.toHaveBeenCalled();
    });
  });
});
