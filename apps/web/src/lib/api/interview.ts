import { apiClient } from "@/lib/api-client";

export type InterviewStage =
  | "warmup"
  | "memory"
  | "challenge"
  | "solution"
  | "outcome";

export interface ChatMessage {
  role: "user" | "assistant";
  content: string;
}

export interface GenerateQuestionRequest {
  stage: InterviewStage;
  messages: ChatMessage[];
}

export interface GenerateQuestionResponse {
  question: string;
  stage: InterviewStage;
  next_stage: InterviewStage | "";
  is_complete: boolean;
}

export const STAGE_LABELS: Record<InterviewStage, string> = {
  warmup: "가볍게",
  memory: "기억에 남는 순간",
  challenge: "어려웠던 점",
  solution: "해결법",
  outcome: "결과/배운 점",
};

export const STAGES: InterviewStage[] = [
  "warmup",
  "memory",
  "challenge",
  "solution",
  "outcome",
];

export interface ExtractSTARResult {
  title: string;
  category: string;
  content: string;
  result: string;
  star_situation: string;
  star_task: string;
  star_action: string;
  star_result: string;
  keywords: string[];
}

export interface SaveExperienceResponse {
  experience_id: string;
  tagged: boolean;
}

export async function generateQuestion(
  req: GenerateQuestionRequest,
): Promise<GenerateQuestionResponse> {
  const { data } = await apiClient.post<GenerateQuestionResponse>(
    "/v1/interview/question",
    req,
  );
  return data;
}

export async function extractSTAR(
  messages: ChatMessage[],
): Promise<ExtractSTARResult> {
  const { data } = await apiClient.post<ExtractSTARResult>(
    "/v1/interview/extract",
    { messages },
  );
  return data;
}

export async function saveInterviewExperience(
  input: ExtractSTARResult,
): Promise<SaveExperienceResponse> {
  const { data } = await apiClient.post<SaveExperienceResponse>(
    "/v1/interview/save",
    input,
  );
  return data;
}
