"use client";

export default function CoachingEditError({
  error,
  reset,
}: {
  error: Error;
  reset: () => void;
}) {
  return (
    <div className="flex flex-col items-center justify-center gap-4 py-20">
      <h2 className="text-lg font-semibold text-foreground">
        에디터를 불러올 수 없습니다
      </h2>
      <p className="text-sm text-muted-foreground">{error.message}</p>
      <button
        onClick={reset}
        className="rounded-lg bg-primary px-4 py-2 text-sm font-medium text-primary-foreground hover:bg-primary/90"
      >
        다시 시도
      </button>
    </div>
  );
}
