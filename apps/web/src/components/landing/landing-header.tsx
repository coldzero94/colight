"use client";

import Link from "next/link";
import { Button } from "@/components/ui/button";

export function LandingHeader() {
  return (
    <header className="sticky top-0 z-40 border-b border-white/[0.06] bg-background/60 backdrop-blur-xl">
      <div className="mx-auto flex h-16 max-w-6xl items-center justify-between px-6">
        <Link href="/" className="flex items-center gap-2">
          <div className="flex h-7 w-7 items-center justify-center rounded-lg bg-primary">
            <span className="text-sm font-bold text-primary-foreground">C</span>
          </div>
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
