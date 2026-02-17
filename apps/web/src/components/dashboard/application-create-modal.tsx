"use client";

import { useState } from "react";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { useCreateApplication } from "@/hooks/use-applications";
import { toast } from "sonner";

/** Convert datetime-local value ("2026-02-20T14:30") to RFC3339 for the API. */
function toRFC3339(datetimeLocal: string): string | undefined {
  if (!datetimeLocal) return undefined;
  return new Date(datetimeLocal).toISOString();
}

interface ApplicationCreateModalProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export function ApplicationCreateModal({
  open,
  onOpenChange,
}: ApplicationCreateModalProps) {
  const [companyName, setCompanyName] = useState("");
  const [position, setPosition] = useState("");
  const [jobUrl, setJobUrl] = useState("");
  const [deadline, setDeadline] = useState("");
  const [notes, setNotes] = useState("");
  const [tagsInput, setTagsInput] = useState("");

  const createMutation = useCreateApplication();

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();

    if (!companyName.trim()) {
      toast.error("회사명을 입력해주세요.");
      return;
    }

    const tags = tagsInput
      .split(",")
      .map((t) => t.trim())
      .filter(Boolean);

    createMutation.mutate(
      {
        company_name: companyName.trim(),
        position: position.trim() || undefined,
        job_url: jobUrl.trim() || undefined,
        deadline: toRFC3339(deadline),
        notes: notes.trim() || undefined,
        tags: tags.length > 0 ? tags : undefined,
      },
      {
        onSuccess: () => {
          toast.success("지원 현황이 추가되었습니다.");
          resetForm();
          onOpenChange(false);
        },
        onError: () => {
          toast.error("추가에 실패했습니다.");
        },
      },
    );
  };

  const resetForm = () => {
    setCompanyName("");
    setPosition("");
    setJobUrl("");
    setDeadline("");
    setNotes("");
    setTagsInput("");
  };

  return (
    <Dialog
      open={open}
      onOpenChange={(v) => {
        if (!v) resetForm();
        onOpenChange(v);
      }}
    >
      <DialogContent className="border-white/[0.08] sm:max-w-md">
        <DialogHeader>
          <DialogTitle>지원 현황 추가</DialogTitle>
          <DialogDescription>
            새로운 지원 현황을 추가합니다.
          </DialogDescription>
        </DialogHeader>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="company-name">
              회사명 <span className="text-destructive">*</span>
            </Label>
            <Input
              id="company-name"
              value={companyName}
              onChange={(e) => setCompanyName(e.target.value)}
              placeholder="예: 삼성전자"
              autoFocus
            />
          </div>
          <div className="space-y-2">
            <Label htmlFor="position">포지션</Label>
            <Input
              id="position"
              value={position}
              onChange={(e) => setPosition(e.target.value)}
              placeholder="예: 백엔드 개발자"
            />
          </div>
          <div className="space-y-2">
            <Label htmlFor="job-url">채용 공고 URL</Label>
            <Input
              id="job-url"
              type="url"
              value={jobUrl}
              onChange={(e) => setJobUrl(e.target.value)}
              placeholder="https://..."
            />
          </div>
          <div className="space-y-2">
            <Label htmlFor="deadline">마감일</Label>
            <Input
              id="deadline"
              type="datetime-local"
              value={deadline}
              onChange={(e) => setDeadline(e.target.value)}
              className="dark-date-input"
            />
          </div>
          <div className="space-y-2">
            <Label htmlFor="notes">메모</Label>
            <Textarea
              id="notes"
              value={notes}
              onChange={(e) => setNotes(e.target.value)}
              placeholder="메모를 입력하세요"
              rows={3}
            />
          </div>
          <div className="space-y-2">
            <Label htmlFor="tags">태그 (쉼표로 구분)</Label>
            <Input
              id="tags"
              value={tagsInput}
              onChange={(e) => setTagsInput(e.target.value)}
              placeholder="예: 관심, 대기업, 백엔드"
            />
          </div>
          <DialogFooter>
            <Button
              type="button"
              variant="ghost"
              onClick={() => onOpenChange(false)}
            >
              취소
            </Button>
            <Button type="submit" disabled={createMutation.isPending}>
              {createMutation.isPending ? "추가 중..." : "추가"}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
