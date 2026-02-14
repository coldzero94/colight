import { z } from "zod/v4";

export const jobTypes = [
  "정규직",
  "계약직",
  "인턴",
  "파견직",
  "기타",
] as const;

export type JobType = (typeof jobTypes)[number];

export const jobPostingSchema = z.object({
  company_name: z.string().min(1, "회사명은 필수입니다."),
  position: z.string().min(1, "포지션명은 필수입니다."),
  department: z.string().optional(),
  job_type: z.enum(jobTypes).optional(),
  experience_level: z.string().optional(),
  education: z.string().optional(),
  location: z.string().optional(),
  salary: z.string().optional(),
  main_tasks: z.array(z.string()).default([]),
  requirements: z.array(z.string()).default([]),
  preferred: z.array(z.string()).default([]),
  required_skills: z.array(z.string()).default([]),
  soft_skills: z.array(z.string()).default([]),
  company_values_hints: z.array(z.string()).default([]),
  deadline: z.string().optional(),
  source_url: z.string().url("유효한 URL이어야 합니다.").optional(),
  source_domain: z.string().optional(),
});

export type JobPosting = z.infer<typeof jobPostingSchema>;
