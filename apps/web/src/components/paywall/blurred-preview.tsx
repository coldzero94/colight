import Link from "next/link";
import { Lock } from "lucide-react";
import { Button } from "@/components/ui/button";

interface BlurredPreviewProps {
  children: React.ReactNode;
  message?: string;
}

export function BlurredPreview({
  children,
  message = "무료 사용 횟수를 초과했습니다.",
}: BlurredPreviewProps) {
  return (
    <div className="relative">
      <div className="pointer-events-none select-none blur-sm" aria-hidden>
        {children}
      </div>
      <div className="absolute inset-0 flex flex-col items-center justify-center gap-3 bg-white/60">
        <Lock className="h-8 w-8 text-gray-400" />
        <p className="text-sm font-medium text-gray-700">{message}</p>
        <Button asChild size="sm">
          <Link href="/pricing">플랜 업그레이드</Link>
        </Button>
      </div>
    </div>
  );
}
