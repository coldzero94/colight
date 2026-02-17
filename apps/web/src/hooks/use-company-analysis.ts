import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { analyzeCompany, getCompanyData, getRecentAnalyses } from "@/lib/api/analysis";

export function useCompanyData(companyName: string | null) {
  return useQuery({
    queryKey: ["company-data", companyName],
    queryFn: () => getCompanyData(companyName!),
    enabled: !!companyName,
  });
}

export function useAnalyzeCompany() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ companyName, url }: { companyName: string; url?: string }) =>
      analyzeCompany(companyName, url),
    onSuccess: (data, { companyName }) => {
      // Cache the analysis result
      queryClient.setQueryData(["company-analysis", companyName], data);
    },
  });
}

export function useCompanyAnalysis(companyName: string | null) {
  return useQuery({
    queryKey: ["company-analysis", companyName],
    queryFn: () => analyzeCompany(companyName!),
    enabled: !!companyName,
    staleTime: 24 * 60 * 60 * 1000, // 24 hours (server has 365-day cache)
  });
}

export function useRecentAnalyses(limit = 10) {
  return useQuery({
    queryKey: ["recent-analyses", limit],
    queryFn: () => getRecentAnalyses(limit),
    staleTime: 5 * 60 * 1000, // 5 minutes
  });
}
