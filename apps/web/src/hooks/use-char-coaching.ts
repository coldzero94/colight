import { useMutation } from "@tanstack/react-query";
import { requestCharCoaching, type CharCoachingRequest, type CharCoachingResult } from "@/lib/api/coaching";

export function useCharCoaching() {
  return useMutation<CharCoachingResult, Error, CharCoachingRequest>({
    mutationFn: requestCharCoaching,
  });
}
