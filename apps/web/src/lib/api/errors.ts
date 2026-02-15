import { AxiosError } from "axios";

interface ApiErrorPayload {
  error: {
    message: string;
    code: string;
    used?: number;
    limit?: number;
    upgrade_url?: string;
  };
}

/** Check if an error is a usage limit exceeded error (USAGE_001). */
export function isUsageLimitError(
  error: unknown,
): error is AxiosError<ApiErrorPayload> {
  if (!(error instanceof AxiosError)) return false;
  const data = error.response?.data as ApiErrorPayload | undefined;
  return data?.error?.code === "USAGE_001";
}

/** Extract usage info from a usage limit error. */
export function getUsageLimitInfo(error: AxiosError<ApiErrorPayload>) {
  const data = error.response?.data;
  return {
    message: data?.error?.message ?? "무료 사용 횟수를 초과했습니다.",
    used: data?.error?.used ?? 0,
    limit: data?.error?.limit ?? 0,
  };
}
