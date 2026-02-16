import { Skeleton } from "@/components/ui/skeleton";

interface TaggingStatusProps {
  status: "idle" | "tagging" | "success" | "error";
  onRetry: () => void;
}

export function TaggingStatus({ status, onRetry }: TaggingStatusProps) {
  if (status === "idle" || status === "success") {
    return null;
  }

  if (status === "tagging") {
    return (
      <div className="flex items-center gap-3 rounded-lg border border-border bg-white/[0.02] p-4">
        <div className="flex gap-2">
          <Skeleton className="h-6 w-16 rounded-full" />
          <Skeleton className="h-6 w-20 rounded-full" />
          <Skeleton className="h-6 w-14 rounded-full" />
        </div>
        <p className="text-sm text-muted-foreground">
          AI가 경험을 분석하고 있어요... <span className="text-xs">(약 3~5초)</span>
        </p>
      </div>
    );
  }

  // error
  return (
    <div className="flex items-center justify-between rounded-lg border border-red-500/20 bg-red-500/10 p-4">
      <p className="text-sm text-red-400">무기 분석에 실패했습니다.</p>
      <button
        onClick={onRetry}
        className="text-sm font-medium text-red-400 hover:text-red-300 transition-colors"
      >
        다시 시도
      </button>
    </div>
  );
}
