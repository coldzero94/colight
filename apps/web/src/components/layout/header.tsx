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
    <header className="relative z-30 h-14 border-b border-border bg-background/85 backdrop-blur-md flex items-center justify-between px-4 shadow-[0_1px_0_rgba(255,255,255,0.04)]">
      <div className="flex items-center gap-3">
        <MobileNav />
        <h1 className="text-lg font-semibold text-foreground">{title}</h1>
      </div>
      <UserDropdown />
    </header>
  );
}
