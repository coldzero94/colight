"use client";

import Link from "next/link";
import { Sparkles } from "lucide-react";
import { Button } from "@/components/ui/button";
import { ColightLogo } from "@/components/common/colight-logo";

export function LandingHeader() {
  return (
    <header className="sticky top-0 z-40 px-4 pt-3 sm:px-6">
      <div className="glass-strong mx-auto flex h-16 w-full max-w-6xl items-center justify-between rounded-2xl px-4 sm:px-6">
        <Link href="/" className="flex items-center gap-2">
          <ColightLogo size={30} className="text-primary" />
          <span className="text-lg font-bold font-display text-foreground tracking-tight">
            Colight
          </span>
        </Link>
        <div className="hidden items-center gap-5 text-sm text-muted-foreground md:flex">
          <span className="inline-flex items-center gap-1 text-primary/90">
            <Sparkles className="h-3.5 w-3.5" />
            AI Writing OS
          </span>
          <span>기업 분석</span>
          <span>경험 매칭</span>
          <span>자소서 코칭</span>
        </div>
        <div className="flex items-center gap-3">
          <Button variant="ghost" size="sm" asChild>
            <Link href="/login">로그인</Link>
          </Button>
          <Button size="sm" className="glow-sm" asChild>
            <Link href="/signup">시작하기</Link>
          </Button>
        </div>
      </div>
    </header>
  );
}
