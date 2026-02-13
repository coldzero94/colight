import { apiClient } from "@/lib/api-client";

export interface Experience {
  id: string;
  user_id: string;
  title: string;
  category: string;
  period_start?: string;
  period_end?: string;
  role: string;
  content: string;
  result: string;
  star_situation: string;
  star_task: string;
  star_action: string;
  star_result: string;
  keywords: string[] | null;
  source: string;
  is_archived: boolean;
  created_at: string;
  updated_at: string;
  weapons?: ExperienceWeapon[];
}

export interface ExperienceWeapon {
  id: string;
  weapon_code: string;
  confidence: number;
  is_primary: boolean;
  reasoning: string;
  user_confirmed: boolean;
  user_modified: boolean;
}

export interface CreateExperienceRequest {
  title: string;
  category?: string;
  period_start?: string;
  period_end?: string;
  role?: string;
  content?: string;
  result?: string;
  star_situation?: string;
  star_task?: string;
  star_action?: string;
  star_result?: string;
  keywords?: string[];
}

export interface UpdateExperienceRequest {
  title?: string;
  category?: string;
  period_start?: string;
  period_end?: string;
  role?: string;
  content?: string;
  result?: string;
  star_situation?: string;
  star_task?: string;
  star_action?: string;
  star_result?: string;
  keywords?: string[];
}

export interface ExperienceListParams {
  sort?: "latest" | "oldest" | "title";
  category?: string;
}

export async function fetchExperiences(
  params?: ExperienceListParams
): Promise<Experience[]> {
  const { data } = await apiClient.get<{ experiences: Experience[] }>(
    "/v1/experiences",
    { params }
  );
  return data.experiences;
}

export async function fetchExperience(id: string): Promise<Experience> {
  const { data } = await apiClient.get<Experience>(`/v1/experiences/${id}`);
  return data;
}

export async function createExperience(
  req: CreateExperienceRequest
): Promise<{ id: string }> {
  const { data } = await apiClient.post<{ id: string }>(
    "/v1/experiences",
    req
  );
  return data;
}

export async function updateExperience(
  id: string,
  req: UpdateExperienceRequest
): Promise<{ id: string }> {
  const { data } = await apiClient.patch<{ id: string }>(
    `/v1/experiences/${id}`,
    req
  );
  return data;
}

export async function deleteExperience(
  id: string
): Promise<{ success: boolean }> {
  const { data} = await apiClient.delete<{ success: boolean }>(
    `/v1/experiences/${id}`
  );
  return data;
}

export interface WeaponTagResponse {
  primary_weapon: {
    code: string;
    confidence: number;
    reasoning: string;
  };
  secondary_weapons: Array<{
    code: string;
    confidence: number;
    reasoning: string;
  }>;
}

export async function tagExperience(id: string): Promise<WeaponTagResponse> {
  const { data } = await apiClient.post<WeaponTagResponse>(
    `/v1/experiences/${id}/tag`
  );
  return data;
}
