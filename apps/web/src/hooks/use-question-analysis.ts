"use client";

import { useMutation } from "@tanstack/react-query";
import {
  analyzeQuestion,
  type QuestionAnalysisRequest,
} from "@/lib/api/coaching";

export function useQuestionAnalysis() {
  return useMutation({
    mutationFn: (req: QuestionAnalysisRequest) => analyzeQuestion(req),
  });
}
