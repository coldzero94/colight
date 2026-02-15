import Link from "next/link";
import { ArrowRight } from "lucide-react";
import { Button } from "@/components/ui/button";

export function CTASection() {
  return (
    <section className="bg-gray-900 py-20">
      <div className="mx-auto max-w-6xl px-4 text-center">
        <h2 className="text-2xl font-bold text-white sm:text-3xl">
          지금 무료로 시작하세요
        </h2>
        <p className="mt-3 text-gray-400">
          경험 3건, 일 1회 분석·코칭 무료 제공
        </p>
        <div className="mt-8">
          <Button
            size="lg"
            className="bg-white text-gray-900 hover:bg-gray-100"
            asChild
          >
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
