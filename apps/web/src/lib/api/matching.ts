import { apiClient } from "@/lib/api-client";

export interface MatchResult {
  experience_id: string;
  overall_fit: number; // 0-100
  job_relevance: number; // 0-100
  talent_fit: number; // 0-100
  uniqueness: number; // 0-100
  reasoning: string;
  suggested_angle: string;
}

export interface MatchResponse {
  company_name: string;
  matches: MatchResult[];
  total: number;
}

export async function matchExperiences(companyName: string): Promise<MatchResponse> {
  const { data } = await apiClient.post<MatchResponse>("/v1/match", {
    company_name: companyName,
  });
  return data;
}
