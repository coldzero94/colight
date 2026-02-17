import { apiClient } from "@/lib/api-client";

export type ApplicationStatus =
  | "preparing"
  | "submitted"
  | "in_review"
  | "interview"
  | "accepted"
  | "rejected";

export interface ApplicationDetail {
  id: string;
  company_name: string;
  position: string;
  status: ApplicationStatus;
  deadline: string | null;
  applied_at: string | null;
  notes: string;
  tags: string[];
  analysis_id?: string;
  cover_letter_count: number;
  created_at: string;
  updated_at: string;
}

export interface CreateApplicationInput {
  company_name: string;
  position?: string;
  job_url?: string;
  deadline?: string;
  notes?: string;
  tags?: string[];
}

export interface UpdateApplicationInput {
  company_name?: string;
  position?: string;
  job_url?: string;
  deadline?: string;
  applied_at?: string;
  notes?: string;
  tags?: string[];
}

export interface ApplicationStats {
  total: number;
  by_status: Record<string, number>;
  upcoming_deadlines: ApplicationDetail[];
}

export async function getApplications(): Promise<{
  applications: ApplicationDetail[];
}> {
  const { data } = await apiClient.get<{ applications: ApplicationDetail[] }>(
    "/v1/applications",
  );
  return data;
}

export async function updateApplicationStatus(
  id: string,
  status: ApplicationStatus,
): Promise<ApplicationDetail> {
  const { data } = await apiClient.patch<ApplicationDetail>(
    `/v1/applications/${id}/status`,
    { status },
  );
  return data;
}

export async function getApplicationStats(): Promise<ApplicationStats> {
  const { data } = await apiClient.get<ApplicationStats>(
    "/v1/applications/stats",
  );
  return data;
}

export async function createApplication(
  input: CreateApplicationInput,
): Promise<ApplicationDetail> {
  const { data } = await apiClient.post<ApplicationDetail>(
    "/v1/applications",
    input,
  );
  return data;
}

export async function updateApplication(
  id: string,
  input: UpdateApplicationInput,
): Promise<ApplicationDetail> {
  const { data } = await apiClient.patch<ApplicationDetail>(
    `/v1/applications/${id}`,
    input,
  );
  return data;
}

export async function deleteApplication(id: string): Promise<void> {
  await apiClient.delete(`/v1/applications/${id}`);
}

export async function linkAnalysis(
  appId: string,
  analysisId: string,
): Promise<ApplicationDetail> {
  const { data } = await apiClient.post<ApplicationDetail>(
    `/v1/applications/${appId}/link-analysis`,
    { analysis_id: analysisId },
  );
  return data;
}
