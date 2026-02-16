"use client";

export default function GlobalError({
  reset,
}: {
  error: Error & { digest?: string };
  reset: () => void;
}) {
  return (
    <html lang="ko">
      <body>
        <div className="flex min-h-screen flex-col items-center justify-center gap-4 p-4 text-center">
          <h1 className="text-xl font-bold text-foreground">
            서버에 문제가 발생했습니다
          </h1>
          <p className="text-sm text-muted-foreground">
            잠시 후 다시 시도해주세요.
          </p>
          <div className="flex gap-2">
            <button
              onClick={reset}
              className="rounded-lg bg-primary px-4 py-2 text-sm font-medium text-primary-foreground hover:bg-primary/90 transition-colors"
            >
              다시 시도
            </button>
            {/* eslint-disable-next-line @next/next/no-html-link-for-pages -- global-error has no Next.js router context */}
            <a
              href="/"
              className="rounded-lg border border-border px-4 py-2 text-sm font-medium text-foreground/80 hover:bg-white/[0.04] transition-colors"
            >
              홈으로
            </a>
          </div>
        </div>
      </body>
    </html>
  );
}
