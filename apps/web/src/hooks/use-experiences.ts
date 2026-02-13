import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
  fetchExperiences,
  fetchExperience,
  createExperience,
  updateExperience,
  deleteExperience,
  type ExperienceListParams,
  type CreateExperienceRequest,
  type UpdateExperienceRequest,
} from "@/lib/api/experiences";

const EXPERIENCES_KEY = ["experiences"] as const;

export function useExperiences(params?: ExperienceListParams) {
  return useQuery({
    queryKey: [...EXPERIENCES_KEY, params],
    queryFn: () => fetchExperiences(params),
  });
}

export function useExperience(id: string) {
  return useQuery({
    queryKey: [...EXPERIENCES_KEY, id],
    queryFn: () => fetchExperience(id),
    enabled: !!id,
  });
}

export function useCreateExperience() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: CreateExperienceRequest) => createExperience(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: EXPERIENCES_KEY });
    },
  });
}

export function useUpdateExperience() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: UpdateExperienceRequest }) =>
      updateExperience(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: EXPERIENCES_KEY });
    },
  });
}

export function useDeleteExperience() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (id: string) => deleteExperience(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: EXPERIENCES_KEY });
    },
  });
}
