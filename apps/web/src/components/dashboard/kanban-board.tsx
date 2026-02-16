"use client";

import { useCallback, useMemo } from "react";
import { DragDropContext, type DropResult } from "@hello-pangea/dnd";
import {
  CheckCircle2,
  ClipboardList,
  FileText,
  Mic,
  Send,
  XCircle,
} from "lucide-react";
import { KanbanColumn, type ColumnConfig } from "./kanban-column";
import type { ApplicationDetail, ApplicationStatus } from "@/lib/api/applications";

const COLUMNS: ColumnConfig[] = [
  {
    status: "preparing",
    label: "준비중",
    icon: FileText,
    color: "text-muted-foreground",
    bgColor: "bg-white/[0.06]",
  },
  {
    status: "submitted",
    label: "제출완료",
    icon: Send,
    color: "text-blue-400",
    bgColor: "bg-blue-500/10",
  },
  {
    status: "in_review",
    label: "서류검토",
    icon: ClipboardList,
    color: "text-amber-400",
    bgColor: "bg-amber-500/10",
  },
  {
    status: "interview",
    label: "면접",
    icon: Mic,
    color: "text-purple-400",
    bgColor: "bg-purple-500/10",
  },
  {
    status: "accepted",
    label: "합격",
    icon: CheckCircle2,
    color: "text-primary",
    bgColor: "bg-primary/10",
  },
  {
    status: "rejected",
    label: "불합격",
    icon: XCircle,
    color: "text-red-400",
    bgColor: "bg-red-500/10",
  },
];

interface KanbanBoardProps {
  applications: ApplicationDetail[];
  onStatusChange: (id: string, status: ApplicationStatus) => void;
}

export function KanbanBoard({ applications, onStatusChange }: KanbanBoardProps) {
  const grouped = useMemo(() => {
    const map: Record<string, ApplicationDetail[]> = {};
    for (const col of COLUMNS) {
      map[col.status] = [];
    }
    for (const app of applications) {
      if (map[app.status]) {
        map[app.status].push(app);
      }
    }
    return map;
  }, [applications]);

  const handleDragEnd = useCallback(
    (result: DropResult) => {
      const { draggableId, destination } = result;
      if (!destination) return;

      const newStatus = destination.droppableId as ApplicationStatus;
      const app = applications.find((a) => a.id === draggableId);
      if (!app || app.status === newStatus) return;

      onStatusChange(draggableId, newStatus);
    },
    [applications, onStatusChange],
  );

  return (
    <DragDropContext onDragEnd={handleDragEnd}>
      <div className="brand-surface rounded-2xl p-3 md:p-4">
        <div className="flex gap-3 overflow-x-auto pb-2 snap-x snap-mandatory lg:snap-none lg:overflow-visible">
          {COLUMNS.map((col) => (
            <div key={col.status} className="snap-start">
              <KanbanColumn config={col} applications={grouped[col.status]} />
            </div>
          ))}
        </div>
      </div>
    </DragDropContext>
  );
}
