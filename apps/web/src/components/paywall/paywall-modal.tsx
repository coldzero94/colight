"use client";

import Link from "next/link";
import { Lock } from "lucide-react";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";

interface PaywallModalProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  feature?: string;
  used?: number;
  limit?: number;
}

const featureLabels: Record<string, string> = {
  experience: "경험 등록",
  analysis: "기업 분석",
  question_analysis: "문항 분석",
  draft: "초안 생성",
  review: "첨삭 코칭",
};

export function PaywallModal({
  open,
  onOpenChange,
  feature,
  used,
  limit,
}: PaywallModalProps) {
  const featureLabel = feature ? featureLabels[feature] ?? feature : "";

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-sm">
        <DialogHeader className="items-center text-center">
          <div className="mx-auto mb-2 flex h-12 w-12 items-center justify-center rounded-full bg-white/[0.06]">
            <Lock className="h-6 w-6 text-muted-foreground" />
          </div>
          <DialogTitle>무료 사용 횟수 초과</DialogTitle>
          <DialogDescription>
            {featureLabel && (
              <>
                <span className="font-medium text-foreground">
                  {featureLabel}
                </span>{" "}
                기능의{" "}
              </>
            )}
            무료 사용 횟수를 모두 사용했습니다.
            {used !== undefined && limit !== undefined && (
              <span className="mt-1 block text-xs text-muted-foreground/60">
                {used}/{limit}회 사용
              </span>
            )}
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-2">
          <div className="rounded-lg border p-3">
            <p className="text-sm font-medium">Starter</p>
            <p className="text-xs text-muted-foreground">
              월 9,900원 — 모든 기능 일 5회
            </p>
          </div>
          <div className="rounded-lg border border-primary p-3">
            <p className="text-sm font-medium">Pro</p>
            <p className="text-xs text-muted-foreground">
              월 19,900원 — 무제한 사용
            </p>
          </div>
        </div>

        <DialogFooter className="flex-col gap-2 sm:flex-col">
          <Button asChild className="w-full">
            <Link href="/pricing">플랜 보기</Link>
          </Button>
          <Button
            variant="ghost"
            className="w-full"
            onClick={() => onOpenChange(false)}
          >
            닫기
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
