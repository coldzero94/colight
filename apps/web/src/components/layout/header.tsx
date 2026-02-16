"use client";

import { usePathname } from "next/navigation";
import { MobileNav } from "./mobile-nav";
import { UserDropdown } from "./user-dropdown";

const pageTitles: Record<string, string> = {
  "/experiences": "경험 관리",
  "/interview": "경험 인터뷰",
  "/analysis": "기업 분석",
  "/coaching": "자소서 코칭",
  "/dashboard": "대시보드",
};

export function Header() {
  const pathname = usePathname();

  const title =
    Object.entries(pageTitles).find(([path]) =>
      pathname.startsWith(path)
    )?.[1] ?? "";

  return (
    <header className="relative z-30 h-14 border-b border-border bg-background/78 backdrop-blur-xl flex items-center justify-between px-4 shadow-[0_1px_0_rgba(255,255,255,0.04)]">
      <div className="pointer-events-none absolute inset-x-0 bottom-0 h-px bg-gradient-to-r from-transparent via-primary/30 to-transparent" />
      <div className="flex items-center gap-3">
        <MobileNav />
        <span className="hidden h-5 w-px bg-white/15 sm:block" />
        <h1 className="inline-flex items-center gap-2 text-lg font-semibold text-foreground">
          <span className="brand-chip inline-flex h-5 w-5 items-center justify-center rounded-full">
            <span className="h-1.5 w-1.5 rounded-full bg-white/90" />
          </span>
          {title}
        </h1>
      </div>
      <UserDropdown />
    </header>
  );
}
