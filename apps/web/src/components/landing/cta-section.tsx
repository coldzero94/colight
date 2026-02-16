import Link from "next/link";
import { ArrowRight } from "lucide-react";
import { Button } from "@/components/ui/button";

export function CTASection() {
  return (
    <section className="relative border-t border-border py-20 overflow-hidden">
      {/* Subtle gradient glow */}
      <div className="absolute inset-0 -z-10">
        <div className="absolute bottom-0 left-1/2 -translate-x-1/2 w-[500px] h-[300px] bg-primary/[0.05] rounded-full blur-[100px]" />
      </div>

      <div className="mx-auto max-w-6xl px-4 text-center">
        <h2 className="text-2xl font-bold text-foreground sm:text-3xl">
          지금 무료로 시작하세요
        </h2>
        <p className="mt-3 text-muted-foreground">
          경험 3건, 일 1회 분석·코칭 무료 제공
        </p>
        <div className="mt-8">
          <Button size="lg" asChild>
            <Link href="/signup">
              무료 회원가입
              <ArrowRight className="ml-1 h-4 w-4" />
            </Link>
          </Button>
        </div>
      </div>
    </section>
  );
}
