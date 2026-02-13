import { z } from "zod/v4";

export const experienceCategories = [
  "인턴",
  "대외활동",
  "프로젝트",
  "아르바이트",
  "동아리",
  "봉사활동",
  "기타",
] as const;

export type ExperienceCategory = (typeof experienceCategories)[number];

export const experienceSchema = z
  .object({
    title: z
      .string()
      .min(2, "제목은 최소 2자 이상이어야 합니다.")
      .max(100, "제목은 100자 이내로 입력해주세요."),
    category: z.string().optional(),
    period_start: z.string().optional(),
    period_end: z.string().optional(),
    role: z.string().max(50, "역할은 50자 이내로 입력해주세요.").optional(),
    star_situation: z
      .string()
      .max(1000, "Situation은 1000자 이내로 입력해주세요.")
      .optional(),
    star_task: z
      .string()
      .max(1000, "Task는 1000자 이내로 입력해주세요.")
      .optional(),
    star_action: z
      .string()
      .max(2000, "Action은 2000자 이내로 입력해주세요.")
      .optional(),
    star_result: z
      .string()
      .max(1000, "Result는 1000자 이내로 입력해주세요.")
      .optional(),
    content: z
      .string()
      .max(5000, "자유 입력은 5000자 이내로 입력해주세요.")
      .optional(),
  })
  .refine(
    (data) => {
      if (data.period_start && data.period_end) {
        return new Date(data.period_end) >= new Date(data.period_start);
      }
      return true;
    },
    {
      message: "종료일은 시작일 이후여야 합니다.",
      path: ["period_end"],
    }
  );

export type ExperienceFormValues = z.infer<typeof experienceSchema>;
