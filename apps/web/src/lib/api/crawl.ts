import { apiClient } from "@/lib/api-client";

export interface JobPosting {
  company_name: string;
  position: string;
  department?: string;
  job_type?: string;
  experience_level?: string;
  main_tasks?: string[];
  requirements?: string[];
  preferred?: string[];
  required_skills?: string[];
  soft_skills?: string[];
  company_values_hints?: string[];
  deadline?: string;
}

export async function crawlJobPosting(url: string): Promise<JobPosting> {
  const { data } = await apiClient.post<JobPosting>("/v1/crawl", { url });
  return data;
}
