import axios from "axios";
import { useAuthStore } from "@/stores/auth-store";

export const apiClient = axios.create({
  baseURL: process.env.NEXT_PUBLIC_API_URL || "http://localhost:9000",
  headers: { "Content-Type": "application/json" },
});

// Generate a short trace ID for request correlation
function generateTraceId(): string {
  if (typeof crypto !== "undefined" && crypto.randomUUID) {
    return crypto.randomUUID().slice(0, 8);
  }
  return Math.random().toString(36).slice(2, 10);
}

// Request interceptor — attach token + trace ID
apiClient.interceptors.request.use((config) => {
  const { accessToken } = useAuthStore.getState();
  if (accessToken) {
    config.headers.Authorization = `Bearer ${accessToken}`;
  }

  const traceId = generateTraceId();
  config.headers["X-Trace-ID"] = traceId;

  return config;
});

// Response interceptor — log errors + refresh on 401
apiClient.interceptors.response.use(
  (response) => response,
  async (error) => {
    const originalRequest = error?.config as
      | ({ _retry?: boolean; url?: string; headers?: Record<string, string> })
      | undefined;

    const status = error?.response?.status;
    const requestUrl = originalRequest?.url ?? "";
    const isRefreshEndpoint = requestUrl.includes("/v1/auth/refresh");

    // Log API errors in development
    if (process.env.NODE_ENV === "development" && status) {
      const traceId = originalRequest?.headers?.["X-Trace-ID"] ?? "unknown";
      console.error(`[API] ${status} ${requestUrl} trace_id=${traceId}`);
    }

    // Never attempt refresh for refresh endpoint itself.
    // Otherwise, refresh 401 can recurse indefinitely and crash the renderer.
    if (
      status === 401 &&
      originalRequest &&
      !originalRequest._retry &&
      !isRefreshEndpoint
    ) {
      originalRequest._retry = true;
      const success = await useAuthStore.getState().refreshAccessToken();

      if (success) {
        const { accessToken } = useAuthStore.getState();
        originalRequest.headers = originalRequest.headers ?? {};
        if (accessToken) {
          originalRequest.headers.Authorization = `Bearer ${accessToken}`;
        }
        return apiClient(originalRequest);
      }
    }

    return Promise.reject(error);
  }
);
