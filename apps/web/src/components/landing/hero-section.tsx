import Link from "next/link";
import { ArrowRight } from "lucide-react";
import { Button } from "@/components/ui/button";

export function HeroSection() {
  return (
    <section className="mx-auto max-w-6xl px-4 py-24 text-center lg:py-32">
      <h1 className="text-4xl font-bold tracking-tight text-gray-900 sm:text-5xl lg:text-6xl">
        취업 준비,
        <br />
        이제 <span className="text-blue-600">AI 코치</span>와 함께
      </h1>
      <p className="mx-auto mt-6 max-w-2xl text-lg text-gray-600">
        채용공고 URL 하나로 기업 분석부터 경험 매칭, STAR 구조 초안 생성,
        4점 첨삭 코칭까지. 당신만의 AI 취업 코치가 합격까지 이끌어드립니다.
      </p>
      <div className="mt-10 flex items-center justify-center gap-4">
        <Button size="lg" asChild>
          <Link href="/signup">
            무료로 시작하기
            <ArrowRight className="ml-1 h-4 w-4" />
          </Link>
        </Button>
        <Button variant="outline" size="lg" asChild>
          <Link href="/login">로그인</Link>
        </Button>
      </div>
      <p className="mt-4 text-sm text-gray-400">
        가입 즉시 무료 체험 — 신용카드 불필요
      </p>
    </section>
  );
}
