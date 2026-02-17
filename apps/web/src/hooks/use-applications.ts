import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
  getApplications,
  getApplicationStats,
  updateApplicationStatus,
  createApplication,
  updateApplication,
  deleteApplication,
  linkAnalysis,
  type ApplicationDetail,
  type ApplicationStatus,
  type CreateApplicationInput,
  type UpdateApplicationInput,
} from "@/lib/api/applications";

const APPLICATIONS_KEY = ["applications"] as const;
const APPLICATION_STATS_KEY = ["applications", "stats"] as const;

export function useApplications() {
  return useQuery({
    queryKey: [...APPLICATIONS_KEY],
    queryFn: () => getApplications().then((res) => res.applications),
  });
}

export function useApplicationStats() {
  return useQuery({
    queryKey: [...APPLICATION_STATS_KEY],
    queryFn: getApplicationStats,
  });
}

export function useUpdateApplicationStatus() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, status }: { id: string; status: ApplicationStatus }) =>
      updateApplicationStatus(id, status),
    onMutate: async ({ id, status }) => {
      await queryClient.cancelQueries({ queryKey: APPLICATIONS_KEY });

      const previous =
        queryClient.getQueryData<ApplicationDetail[]>(APPLICATIONS_KEY);

      queryClient.setQueryData<ApplicationDetail[]>(
        APPLICATIONS_KEY,
        (old) =>
          old?.map((app) => (app.id === id ? { ...app, status } : app)) ?? [],
      );

      return { previous };
    },
    onError: (_err, _vars, context) => {
      if (context?.previous) {
        queryClient.setQueryData(APPLICATIONS_KEY, context.previous);
      }
    },
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: APPLICATIONS_KEY });
      queryClient.invalidateQueries({ queryKey: APPLICATION_STATS_KEY });
    },
  });
}

export function useCreateApplication() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (input: CreateApplicationInput) => createApplication(input),
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: APPLICATIONS_KEY });
      queryClient.invalidateQueries({ queryKey: APPLICATION_STATS_KEY });
    },
  });
}

export function useUpdateApplication() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, input }: { id: string; input: UpdateApplicationInput }) =>
      updateApplication(id, input),
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: APPLICATIONS_KEY });
      queryClient.invalidateQueries({ queryKey: APPLICATION_STATS_KEY });
    },
  });
}

export function useDeleteApplication() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (id: string) => deleteApplication(id),
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: APPLICATIONS_KEY });
      queryClient.invalidateQueries({ queryKey: APPLICATION_STATS_KEY });
    },
  });
}

export function useLinkAnalysis() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ appId, analysisId }: { appId: string; analysisId: string }) =>
      linkAnalysis(appId, analysisId),
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: APPLICATIONS_KEY });
    },
  });
}
