"use client";

export default function DashboardError({
  error,
  reset,
}: {
  error: Error;
  reset: () => void;
}) {
  return (
    <div className="flex flex-col items-center justify-center gap-4 py-20">
      <h2 className="text-lg font-semibold text-foreground">
        문제가 발생했습니다
      </h2>
      <p className="text-sm text-muted-foreground">{error.message}</p>
      <button
        onClick={reset}
        className="bg-foreground text-background text-sm font-medium px-4 py-2 rounded-lg hover:bg-foreground/90 transition-colors"
      >
        다시 시도
      </button>
    </div>
  );
}
