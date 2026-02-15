import { useQuery } from "@tanstack/react-query";
import { getUsage } from "@/lib/api/usage";

export function useUsage() {
  return useQuery({
    queryKey: ["usage"],
    queryFn: getUsage,
    staleTime: 30 * 1000,
  });
}
