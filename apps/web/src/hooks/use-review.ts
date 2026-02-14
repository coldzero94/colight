import { useMutation } from "@tanstack/react-query";
import { requestReview, type ReviewRequest, type ReviewResult } from "@/lib/api/coaching";

export function useReview() {
  return useMutation<ReviewResult, Error, ReviewRequest>({
    mutationFn: requestReview,
  });
}
