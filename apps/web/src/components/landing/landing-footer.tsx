import Link from "next/link";

export function LandingFooter() {
  return (
    <footer className="border-t border-border py-10">
      <div className="mx-auto max-w-6xl px-4">
        <div className="flex flex-col items-center justify-between gap-4 sm:flex-row">
          <p className="text-sm font-medium text-foreground">Colight</p>
          <div className="flex gap-6">
            <Link
              href="/privacy"
              className="text-sm text-muted-foreground hover:text-foreground transition-colors"
            >
              개인정보처리방침
            </Link>
            <Link
              href="/terms"
              className="text-sm text-muted-foreground hover:text-foreground transition-colors"
            >
              이용약관
            </Link>
          </div>
          <p className="text-sm text-muted-foreground/60">&copy; 2026 Colight</p>
        </div>
      </div>
    </footer>
  );
}
