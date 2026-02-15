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
  cover_letter_count: number;
  created_at: string;
  updated_at: string;
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
