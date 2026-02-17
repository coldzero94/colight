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
import { useUpdateApplication } from "@/hooks/use-applications";
import type { ApplicationDetail } from "@/lib/api/applications";
import { toast } from "sonner";

interface ApplicationEditModalProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  application: ApplicationDetail | null;
}

export function ApplicationEditModal({
  open,
  onOpenChange,
  application,
}: ApplicationEditModalProps) {
  // Re-mount form when application changes via key prop on EditForm
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>지원 현황 수정</DialogTitle>
          <DialogDescription>지원 정보를 수정합니다.</DialogDescription>
        </DialogHeader>
        {application && (
          <EditForm
            key={application.id}
            application={application}
            onClose={() => onOpenChange(false)}
          />
        )}
      </DialogContent>
    </Dialog>
  );
}

function EditForm({
  application,
  onClose,
}: {
  application: ApplicationDetail;
  onClose: () => void;
}) {
  const [companyName, setCompanyName] = useState(application.company_name);
  const [position, setPosition] = useState(application.position);
  const [deadline, setDeadline] = useState(
    application.deadline
      ? new Date(application.deadline).toISOString().slice(0, 16)
      : "",
  );
  const [notes, setNotes] = useState(application.notes);
  const [tagsInput, setTagsInput] = useState(application.tags.join(", "));

  const updateMutation = useUpdateApplication();

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();

    const tags = tagsInput
      .split(",")
      .map((t) => t.trim())
      .filter(Boolean);

    updateMutation.mutate(
      {
        id: application.id,
        input: {
          company_name: companyName.trim(),
          position: position.trim(),
          deadline: deadline || undefined,
          notes: notes.trim(),
          tags,
        },
      },
      {
        onSuccess: () => {
          toast.success("지원 현황이 수정되었습니다.");
          onClose();
        },
        onError: () => {
          toast.error("수정에 실패했습니다.");
        },
      },
    );
  };

  return (
    <form onSubmit={handleSubmit} className="space-y-4">
      <div className="space-y-2">
        <Label htmlFor="edit-company-name">회사명</Label>
        <Input
          id="edit-company-name"
          value={companyName}
          onChange={(e) => setCompanyName(e.target.value)}
        />
      </div>
      <div className="space-y-2">
        <Label htmlFor="edit-position">포지션</Label>
        <Input
          id="edit-position"
          value={position}
          onChange={(e) => setPosition(e.target.value)}
        />
      </div>
      <div className="space-y-2">
        <Label htmlFor="edit-deadline">마감일</Label>
        <Input
          id="edit-deadline"
          type="datetime-local"
          value={deadline}
          onChange={(e) => setDeadline(e.target.value)}
        />
      </div>
      <div className="space-y-2">
        <Label htmlFor="edit-notes">메모</Label>
        <Textarea
          id="edit-notes"
          value={notes}
          onChange={(e) => setNotes(e.target.value)}
          rows={3}
        />
      </div>
      <div className="space-y-2">
        <Label htmlFor="edit-tags">태그 (쉼표로 구분)</Label>
        <Input
          id="edit-tags"
          value={tagsInput}
          onChange={(e) => setTagsInput(e.target.value)}
          placeholder="예: 관심, 대기업"
        />
      </div>
      <DialogFooter>
        <Button type="button" variant="ghost" onClick={onClose}>
          취소
        </Button>
        <Button type="submit" disabled={updateMutation.isPending}>
          {updateMutation.isPending ? "저장 중..." : "저장"}
        </Button>
      </DialogFooter>
    </form>
  );
}
