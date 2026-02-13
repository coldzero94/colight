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
      <div className="flex items-center gap-3 rounded-lg border border-gray-100 bg-gray-50 p-4">
        <div className="flex gap-2">
          <Skeleton className="h-6 w-16 rounded-full" />
          <Skeleton className="h-6 w-20 rounded-full" />
          <Skeleton className="h-6 w-14 rounded-full" />
        </div>
        <p className="text-sm text-gray-500">
          AI가 경험을 분석하고 있어요... <span className="text-xs">(약 3~5초)</span>
        </p>
      </div>
    );
  }

  // error
  return (
    <div className="flex items-center justify-between rounded-lg border border-red-100 bg-red-50 p-4">
      <p className="text-sm text-red-600">무기 분석에 실패했습니다.</p>
      <button
        onClick={onRetry}
        className="text-sm font-medium text-red-700 hover:text-red-800 transition-colors"
      >
        다시 시도
      </button>
    </div>
  );
}
