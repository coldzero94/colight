import { Monitor } from "lucide-react";

export function DemoSection() {
  return (
    <section className="border-t border-border py-20">
      <div className="mx-auto max-w-6xl px-4">
        <h2 className="text-center text-2xl font-bold text-foreground sm:text-3xl">
          한눈에 보는 코칭 프로세스
        </h2>
        <p className="mx-auto mt-3 max-w-xl text-center text-muted-foreground">
          채용공고 입력부터 최종 자소서까지, 5단계 AI 코칭
        </p>
        <div className="mx-auto mt-12 max-w-4xl">
          <div className="flex aspect-video items-center justify-center rounded-2xl border border-border bg-card">
            <div className="text-center">
              <Monitor className="mx-auto h-12 w-12 text-muted-foreground/30" />
              <p className="mt-3 text-sm text-muted-foreground/50">
                서비스 스크린샷 준비 중
              </p>
            </div>
          </div>
        </div>
      </div>
    </section>
  );
}
