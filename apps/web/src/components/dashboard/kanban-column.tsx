"use client";

import { Droppable } from "@hello-pangea/dnd";
import { KanbanCard } from "./kanban-card";
import type { ApplicationDetail, ApplicationStatus } from "@/lib/api/applications";

export interface ColumnConfig {
  status: ApplicationStatus;
  label: string;
  icon: string;
  color: string;
  bgColor: string;
}

interface KanbanColumnProps {
  config: ColumnConfig;
  applications: ApplicationDetail[];
}

export function KanbanColumn({ config, applications }: KanbanColumnProps) {
  return (
    <div className="flex w-[280px] shrink-0 flex-col rounded-lg bg-white/[0.02] lg:w-auto lg:flex-1">
      <div className="flex items-center gap-2 px-3 py-2.5 border-b border-border">
        <span>{config.icon}</span>
        <h3 className="text-sm font-semibold text-foreground/80">{config.label}</h3>
        <span
          data-testid={`column-count-${config.status}`}
          className={`ml-auto inline-flex h-5 min-w-5 items-center justify-center rounded-full px-1.5 text-xs font-medium ${config.bgColor} ${config.color}`}
        >
          {applications.length}
        </span>
      </div>

      <Droppable droppableId={config.status}>
        {(provided, snapshot) => (
          <div
            ref={provided.innerRef}
            {...provided.droppableProps}
            className={`flex-1 space-y-2 p-2 min-h-[120px] transition-colors rounded-b-lg ${
              snapshot.isDraggingOver ? "bg-blue-500/5" : ""
            }`}
          >
            {applications.map((app, index) => (
              <KanbanCard key={app.id} application={app} index={index} />
            ))}
            {provided.placeholder}

            {applications.length === 0 && !snapshot.isDraggingOver && (
              <div className="flex h-full min-h-[80px] items-center justify-center rounded-lg border-2 border-dashed border-border">
                <p className="text-xs text-muted-foreground/60">여기에 드래그하세요</p>
              </div>
            )}
          </div>
        )}
      </Droppable>
    </div>
  );
}
