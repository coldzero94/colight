import { useState } from "react";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { tagExperience } from "@/lib/api/experiences";

type TaggingState = "idle" | "tagging" | "success" | "error";

export function useWeaponTagging() {
  const [state, setState] = useState<TaggingState>("idle");
  const queryClient = useQueryClient();

  const tagMutation = useMutation({
    mutationFn: (experienceId: string) => tagExperience(experienceId),
    onMutate: () => {
      setState("tagging");
    },
    onSuccess: (_, experienceId) => {
      setState("success");
      // Invalidate experience queries to refetch with new weapons
      queryClient.invalidateQueries({ queryKey: ["experiences"] });
      queryClient.invalidateQueries({
        queryKey: ["experiences", experienceId],
      });
    },
    onError: () => {
      setState("error");
    },
  });

  const triggerTagging = (experienceId: string) => {
    setState("tagging");
    tagMutation.mutate(experienceId);
  };

  const reset = () => {
    setState("idle");
  };

  return {
    state,
    triggerTagging,
    reset,
    error: tagMutation.error,
  };
}
