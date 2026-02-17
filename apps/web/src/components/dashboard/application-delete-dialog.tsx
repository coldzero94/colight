"use client";

import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { useDeleteApplication } from "@/hooks/use-applications";
import type { ApplicationDetail } from "@/lib/api/applications";
import { toast } from "sonner";

interface ApplicationDeleteDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  application: ApplicationDetail | null;
}

export function ApplicationDeleteDialog({
  open,
  onOpenChange,
  application,
}: ApplicationDeleteDialogProps) {
  const deleteMutation = useDeleteApplication();

  const handleDelete = () => {
    if (!application) return;

    deleteMutation.mutate(application.id, {
      onSuccess: () => {
        toast.success("지원 현황이 삭제되었습니다.");
        onOpenChange(false);
      },
      onError: () => {
        toast.error("삭제에 실패했습니다.");
      },
    });
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-sm">
        <DialogHeader>
          <DialogTitle>지원 현황 삭제</DialogTitle>
          <DialogDescription>
            <strong>{application?.company_name}</strong> 지원 현황을
            삭제하시겠습니까? 이 작업은 되돌릴 수 없습니다.
          </DialogDescription>
        </DialogHeader>
        <DialogFooter>
          <Button
            type="button"
            variant="ghost"
            onClick={() => onOpenChange(false)}
          >
            취소
          </Button>
          <Button
            variant="destructive"
            onClick={handleDelete}
            disabled={deleteMutation.isPending}
          >
            {deleteMutation.isPending ? "삭제 중..." : "삭제"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
