import Link from "next/link";
import { ArrowRight } from "lucide-react";
import { Button } from "@/components/ui/button";

export function CTASection() {
  return (
    <section className="relative overflow-hidden py-28 lg:py-36">
      {/* Background */}
      <div className="absolute inset-0 -z-10">
        <div className="absolute inset-0 bg-gradient-to-t from-primary/[0.06] to-transparent" />
        <div className="absolute bottom-0 left-1/2 -translate-x-1/2 w-[600px] h-[400px] bg-primary/[0.15] rounded-full blur-[150px]" />
      </div>

      <div className="mx-auto max-w-3xl px-6 text-center">
        <h2 className="font-display text-3xl font-bold tracking-tight text-foreground sm:text-4xl lg:text-5xl">
          지금 무료로 시작하세요
        </h2>
        <p className="mx-auto mt-5 max-w-xl text-lg leading-relaxed text-muted-foreground">
          경험 3건, 일 1회 분석·코칭 무료 제공.
          <br className="hidden sm:block" />
          신용카드 없이 바로 시작할 수 있습니다.
        </p>
        <div className="mt-10">
          <Button size="lg" className="glow-md h-13 px-10 text-base" asChild>
            <Link href="/signup">
              무료 회원가입
              <ArrowRight className="ml-2 h-4 w-4" />
            </Link>
          </Button>
        </div>

        {/* Trust signals */}
        <div className="mt-12 flex flex-wrap items-center justify-center gap-8 text-sm text-muted-foreground/50">
          <span>설치 불필요</span>
          <span className="hidden sm:block">·</span>
          <span>3분 안에 첫 코칭</span>
          <span className="hidden sm:block">·</span>
          <span>개인정보 안전 보호</span>
        </div>
      </div>
    </section>
  );
}
