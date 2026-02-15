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
          <h1 className="text-xl font-bold text-gray-900">
            서버에 문제가 발생했습니다
          </h1>
          <p className="text-sm text-gray-500">
            잠시 후 다시 시도해주세요.
          </p>
          <div className="flex gap-2">
            <button
              onClick={reset}
              className="rounded-lg bg-gray-900 px-4 py-2 text-sm font-medium text-white hover:bg-gray-800 transition-colors"
            >
              다시 시도
            </button>
            <a
              href="/"
              className="rounded-lg border border-gray-200 px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50 transition-colors"
            >
              홈으로
            </a>
          </div>
        </div>
      </body>
    </html>
  );
}
