import { useMutation } from "@tanstack/react-query";
import {
  submitFeedback,
  type FeedbackInput,
  type FeedbackResponse,
} from "@/lib/api/feedback";

export function useFeedback() {
  return useMutation<FeedbackResponse, Error, FeedbackInput>({
    mutationFn: submitFeedback,
  });
}
