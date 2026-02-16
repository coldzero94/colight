import Link from "next/link";
import { ColightLogo } from "@/components/common/colight-logo";

export function LandingFooter() {
  return (
    <footer className="relative border-t border-white/[0.08] py-12">
      <div className="pointer-events-none absolute inset-0 -z-10 bg-gradient-to-t from-black/35 to-transparent" />
      <div className="mx-auto max-w-6xl px-6">
        <div className="flex flex-col items-center justify-between gap-6 sm:flex-row">
          <div className="flex items-center gap-2">
            <ColightLogo size={24} className="text-primary" />
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
