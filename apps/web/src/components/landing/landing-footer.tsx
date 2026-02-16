import Link from "next/link";

export function LandingFooter() {
  return (
    <footer className="border-t border-white/[0.06] py-12">
      <div className="mx-auto max-w-6xl px-6">
        <div className="flex flex-col items-center justify-between gap-6 sm:flex-row">
          <div className="flex items-center gap-2">
            <div className="flex h-6 w-6 items-center justify-center rounded-md bg-primary">
              <span className="text-xs font-bold text-primary-foreground">C</span>
            </div>
            <span className="text-sm font-medium font-display text-foreground">Colight</span>
          </div>
          <div className="flex gap-8">
            <Link
              href="/privacy"
              className="text-sm text-muted-foreground/60 hover:text-foreground transition-colors duration-200"
            >
              개인정보처리방침
            </Link>
            <Link
              href="/terms"
              className="text-sm text-muted-foreground/60 hover:text-foreground transition-colors duration-200"
            >
              이용약관
            </Link>
          </div>
          <p className="text-sm text-muted-foreground/40">&copy; 2026 Colight</p>
        </div>
      </div>
    </footer>
  );
}
