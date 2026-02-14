import { useEffect, useRef, useState } from "react";
import { matchExperiences, type MatchResponse } from "@/lib/api/matching";

interface UseAutoMatchingOptions {
  companyName: string | null;
  hasMatching: boolean;
  experienceCount: number;
  enabled?: boolean;
}

export function useAutoMatching({
  companyName,
  hasMatching,
  experienceCount,
  enabled = true,
}: UseAutoMatchingOptions) {
  const [isMatching, setIsMatching] = useState(false);
  const [matchingResult, setMatchingResult] = useState<MatchResponse | null>(
    null
  );
  const [error, setError] = useState<Error | null>(null);
  const triggeredRef = useRef(false);

  const triggerMatching = async () => {
    if (!companyName) return;
    setIsMatching(true);
    setError(null);
    try {
      const result = await matchExperiences(companyName);
      setMatchingResult(result);
    } catch (err) {
      setError(err instanceof Error ? err : new Error("매칭 실패"));
    } finally {
      setIsMatching(false);
    }
  };

  useEffect(() => {
    if (
      enabled &&
      companyName &&
      !hasMatching &&
      experienceCount > 0 &&
      !triggeredRef.current
    ) {
      triggeredRef.current = true;
      triggerMatching();
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [companyName, hasMatching, experienceCount, enabled]);

  return { isMatching, matchingResult, error, triggerMatching };
}
