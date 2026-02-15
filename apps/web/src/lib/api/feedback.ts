import { apiClient } from "@/lib/api-client";

export interface FeedbackInput {
  category: "bug" | "improvement" | "other";
  content: string;
  page_url?: string;
}

export interface FeedbackResponse {
  id: string;
  created_at: string;
}

export async function submitFeedback(
  input: FeedbackInput,
): Promise<FeedbackResponse> {
  const { data } = await apiClient.post<FeedbackResponse>(
    "/v1/feedback",
    input,
  );
  return data;
}
