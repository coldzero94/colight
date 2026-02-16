"use client";

import { Droppable } from "@hello-pangea/dnd";
import type { LucideIcon } from "lucide-react";
import { KanbanCard } from "./kanban-card";
import type { ApplicationDetail, ApplicationStatus } from "@/lib/api/applications";

export interface ColumnConfig {
  status: ApplicationStatus;
  label: string;
  icon: LucideIcon;
  color: string;
  bgColor: string;
}

interface KanbanColumnProps {
  config: ColumnConfig;
  applications: ApplicationDetail[];
}

export function KanbanColumn({ config, applications }: KanbanColumnProps) {
  const Icon = config.icon;

  return (
    <div className="brand-surface-soft flex w-[280px] shrink-0 flex-col rounded-xl lg:w-auto lg:flex-1">
      <div className="flex items-center gap-2 border-b border-white/10 px-3 py-2.5">
        <span
          className={`inline-flex h-6 w-6 items-center justify-center rounded-md border border-white/10 ${config.bgColor}`}
        >
          <Icon className={`h-3.5 w-3.5 ${config.color}`} aria-hidden="true" />
        </span>
        <h3 className="text-sm font-semibold text-foreground/85">{config.label}</h3>
        <span
          data-testid={`column-count-${config.status}`}
          className={`ml-auto inline-flex h-5 min-w-5 items-center justify-center rounded-full border border-white/12 px-1.5 text-xs font-medium ${config.bgColor} ${config.color}`}
        >
          {applications.length}
        </span>
      </div>

      <Droppable droppableId={config.status}>
        {(provided, snapshot) => (
          <div
            ref={provided.innerRef}
            {...provided.droppableProps}
            className={`min-h-[140px] flex-1 space-y-2 rounded-b-xl p-2 transition-colors ${
              snapshot.isDraggingOver ? "bg-primary/10" : ""
            }`}
          >
            {applications.map((app, index) => (
              <KanbanCard key={app.id} application={app} index={index} />
            ))}
            {provided.placeholder}

            {applications.length === 0 && !snapshot.isDraggingOver && (
              <div className="flex h-full min-h-[92px] items-center justify-center rounded-lg border border-dashed border-white/16 bg-white/[0.02]">
                <p className="text-xs text-muted-foreground/70">여기에 드래그하세요</p>
              </div>
            )}
          </div>
        )}
      </Droppable>
    </div>
  );
}
