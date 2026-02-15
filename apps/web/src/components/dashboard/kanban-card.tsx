"use client";

import { Draggable } from "@hello-pangea/dnd";
import type { ApplicationDetail } from "@/lib/api/applications";

interface KanbanCardProps {
  application: ApplicationDetail;
  index: number;
}

function getDeadlineDays(deadline: string): number {
  const now = new Date();
  const dl = new Date(deadline);
  return Math.ceil((dl.getTime() - now.getTime()) / (1000 * 60 * 60 * 24));
}

function getDeadlineColor(days: number): string {
  if (days <= 0) return "bg-red-50 text-red-600";
  if (days <= 3) return "bg-red-50 text-red-600";
  if (days <= 7) return "bg-amber-50 text-amber-600";
  return "bg-gray-50 text-gray-500";
}

function DeadlineBadge({ deadline }: { deadline: string }) {
  const days = getDeadlineDays(deadline);
  const color = getDeadlineColor(days);
  const label = days <= 0 ? "마감" : `D-${days}`;

  return (
    <span
      data-testid="deadline-badge"
      className={`inline-flex items-center gap-1 rounded px-1.5 py-0.5 text-xs font-medium ${color}`}
    >
      {label}
    </span>
  );
}

export function KanbanCard({ application, index }: KanbanCardProps) {
  return (
    <Draggable draggableId={application.id} index={index}>
      {(provided, snapshot) => (
        <div
          ref={provided.innerRef}
          {...provided.draggableProps}
          {...provided.dragHandleProps}
          className={`rounded-lg border bg-white p-3 transition-shadow ${
            snapshot.isDragging
              ? "shadow-lg ring-2 ring-blue-200"
              : "shadow-sm hover:shadow-md"
          }`}
        >
          <div className="flex items-start justify-between gap-2">
            <h4 className="text-sm font-semibold text-gray-900 truncate">
              {application.company_name}
            </h4>
            <svg
              className="h-4 w-4 shrink-0 text-gray-300"
              fill="currentColor"
              viewBox="0 0 6 10"
            >
              <circle cx="1" cy="1" r="1" />
              <circle cx="5" cy="1" r="1" />
              <circle cx="1" cy="5" r="1" />
              <circle cx="5" cy="5" r="1" />
              <circle cx="1" cy="9" r="1" />
              <circle cx="5" cy="9" r="1" />
            </svg>
          </div>

          <p className="mt-0.5 text-xs text-gray-500 truncate">
            {application.position}
          </p>

          <div className="mt-2 flex items-center gap-2 text-xs">
            {application.cover_letter_count > 0 && (
              <span className="inline-flex items-center gap-1 text-gray-500">
                <svg
                  className="h-3.5 w-3.5"
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={2}
                    d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"
                  />
                </svg>
                {application.cover_letter_count}개
              </span>
            )}
            {application.deadline && (
              <DeadlineBadge deadline={application.deadline} />
            )}
          </div>
        </div>
      )}
    </Draggable>
  );
}
