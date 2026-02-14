import { z } from "zod/v4";

export const questionAnalysisSchema = z.object({
  application_id: z.string().uuid("기업을 선택해주세요"),
  question_text: z
    .string()
    .min(10, "문항은 최소 10자 이상 입력해주세요")
    .max(500, "문항은 500자까지 입력 가능합니다"),
  char_limit: z
    .number()
    .min(200, "최소 200자 이상이어야 합니다")
    .max(2000, "최대 2000자까지 가능합니다"),
});

export type QuestionAnalysisInput = z.infer<typeof questionAnalysisSchema>;
