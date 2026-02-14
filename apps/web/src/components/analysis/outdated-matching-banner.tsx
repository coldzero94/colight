interface OutdatedMatchingBannerProps {
  reason: string;
  onRefresh: () => void;
  isRefreshing?: boolean;
}

export function OutdatedMatchingBanner({
  reason,
  onRefresh,
  isRefreshing = false,
}: OutdatedMatchingBannerProps) {
  if (!reason) return null;

  return (
    <div
      className="mb-4 rounded-lg border border-yellow-200 bg-yellow-50 p-4"
      role="alert"
    >
      <div className="flex items-center justify-between">
        <div>
          <p className="text-sm font-medium text-yellow-800">
            매칭 결과가 최신이 아닐 수 있습니다
          </p>
          <p className="mt-1 text-sm text-yellow-600">{reason}</p>
        </div>
        <button
          onClick={onRefresh}
          disabled={isRefreshing}
          className="ml-4 rounded-md bg-yellow-100 px-3 py-1.5 text-sm font-medium text-yellow-800 hover:bg-yellow-200 disabled:opacity-50"
        >
          {isRefreshing ? "매칭 중..." : "다시 매칭하기"}
        </button>
      </div>
    </div>
  );
}
