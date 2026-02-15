import Link from "next/link";

export function LandingFooter() {
  return (
    <footer className="border-t border-gray-200 bg-white py-10">
      <div className="mx-auto max-w-6xl px-4">
        <div className="flex flex-col items-center justify-between gap-4 sm:flex-row">
          <p className="text-sm font-medium text-gray-900">Colight</p>
          <div className="flex gap-6">
            <Link
              href="/privacy"
              className="text-sm text-gray-500 hover:text-gray-900"
            >
              개인정보처리방침
            </Link>
            <Link
              href="/terms"
              className="text-sm text-gray-500 hover:text-gray-900"
            >
              이용약관
            </Link>
          </div>
          <p className="text-sm text-gray-400">&copy; 2026 Colight</p>
        </div>
      </div>
    </footer>
  );
}
