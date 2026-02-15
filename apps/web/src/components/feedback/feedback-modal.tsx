"use client";

import { useState } from "react";
import { usePathname } from "next/navigation";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { useFeedback } from "@/hooks/use-feedback";
import type { FeedbackInput } from "@/lib/api/feedback";

interface FeedbackModalProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

const categories: { value: FeedbackInput["category"]; label: string }[] = [
  { value: "bug", label: "버그 신고" },
  { value: "improvement", label: "개선 제안" },
  { value: "other", label: "기타" },
];

export function FeedbackModal({ open, onOpenChange }: FeedbackModalProps) {
  const pathname = usePathname();
  const [category, setCategory] =
    useState<FeedbackInput["category"]>("improvement");
  const [content, setContent] = useState("");
  const { mutate, isPending } = useFeedback();

  function handleSubmit() {
    if (!content.trim()) return;

    mutate(
      { category, content: content.trim(), page_url: pathname },
      {
        onSuccess: () => {
          toast.success("피드백이 전송되었습니다. 감사합니다!");
          setContent("");
          setCategory("improvement");
          onOpenChange(false);
        },
        onError: () => {
          toast.error("피드백 전송에 실패했습니다. 다시 시도해주세요.");
        },
      },
    );
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>피드백 보내기</DialogTitle>
          <DialogDescription>
            서비스 개선을 위한 의견을 보내주세요.
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-4">
          <div className="flex gap-2">
            {categories.map((cat) => (
              <button
                key={cat.value}
                type="button"
                onClick={() => setCategory(cat.value)}
                className={`rounded-full px-3 py-1 text-sm transition-colors ${
                  category === cat.value
                    ? "bg-primary text-primary-foreground"
                    : "bg-muted text-muted-foreground hover:bg-muted/80"
                }`}
              >
                {cat.label}
              </button>
            ))}
          </div>

          <textarea
            value={content}
            onChange={(e) => setContent(e.target.value)}
            placeholder="의견을 자유롭게 작성해주세요..."
            rows={4}
            className="w-full resize-none rounded-md border border-input bg-background px-3 py-2 text-sm placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring"
          />
        </div>

        <DialogFooter>
          <Button
            variant="outline"
            onClick={() => onOpenChange(false)}
            disabled={isPending}
          >
            취소
          </Button>
          <Button
            onClick={handleSubmit}
            disabled={isPending || !content.trim()}
          >
            {isPending ? "전송 중..." : "보내기"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
