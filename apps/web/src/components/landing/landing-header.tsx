"use client";

import Link from "next/link";
import { Button } from "@/components/ui/button";
import { ColightLogo } from "@/components/common/colight-logo";

export function LandingHeader() {
  return (
    <header className="sticky top-0 z-40 border-b border-white/[0.06] bg-background/60 backdrop-blur-xl">
      <div className="mx-auto flex h-16 max-w-6xl items-center justify-between px-6">
        <Link href="/" className="flex items-center gap-2">
          <ColightLogo size={28} />
          <span className="text-lg font-bold font-display text-foreground tracking-tight">
            Colight
          </span>
        </Link>
        <div className="flex items-center gap-3">
          <Button variant="ghost" size="sm" asChild>
            <Link href="/login">로그인</Link>
          </Button>
          <Button size="sm" asChild>
            <Link href="/signup">시작하기</Link>
          </Button>
        </div>
      </div>
    </header>
  );
}
