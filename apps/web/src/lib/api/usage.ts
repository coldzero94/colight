import { apiClient } from "@/lib/api-client";

export interface FeatureUsage {
  allowed: boolean;
  used: number;
  limit: number;
  remaining: number;
}

export interface UsageResponse {
  plan: "free" | "starter" | "pro" | "season";
  features: Record<string, FeatureUsage>;
}

export async function getUsage(): Promise<UsageResponse> {
  const { data } = await apiClient.get<UsageResponse>("/v1/usage");
  return data;
}
