"use client";

import { useMutation } from "@tanstack/react-query";
import {
  generateDraft,
  type GenerateDraftRequest,
} from "@/lib/api/coaching";

export function useDraftCoaching() {
  return useMutation({
    mutationFn: (req: GenerateDraftRequest) => generateDraft(req),
  });
}
