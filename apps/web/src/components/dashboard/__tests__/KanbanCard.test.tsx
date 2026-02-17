import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import { DragDropContext, Droppable } from "@hello-pangea/dnd";
import { KanbanCard } from "../kanban-card";
import type { ApplicationDetail } from "@/lib/api/applications";

const baseApp: ApplicationDetail = {
  id: "1",
  company_name: "삼성전자",
  position: "백엔드 개발자",
  status: "preparing",
  deadline: null,
  applied_at: null,
  notes: "",
  tags: [],
  cover_letter_count: 0,
  created_at: "2026-01-01T00:00:00Z",
  updated_at: "2026-01-01T00:00:00Z",
};

function renderCard(app: ApplicationDetail) {
  return render(
    <DragDropContext onDragEnd={() => {}}>
      <Droppable droppableId="test">
        {(provided) => (
          <div ref={provided.innerRef} {...provided.droppableProps}>
            <KanbanCard application={app} index={0} />
            {provided.placeholder}
          </div>
        )}
      </Droppable>
    </DragDropContext>,
  );
}

describe("KanbanCard", () => {
  it("displays company name and position", () => {
    renderCard(baseApp);
    expect(screen.getByText("삼성전자")).toBeInTheDocument();
    expect(screen.getByText("백엔드 개발자")).toBeInTheDocument();
  });

  it("shows cover letter count when > 0", () => {
    renderCard({ ...baseApp, cover_letter_count: 3 });
    expect(screen.getByText(/3개/)).toBeInTheDocument();
  });

  it("shows red deadline badge for <= 3 days", () => {
    const soon = new Date();
    soon.setDate(soon.getDate() + 2);
    renderCard({ ...baseApp, deadline: soon.toISOString() });
    const badge = screen.getByTestId("deadline-badge");
    expect(badge).toHaveClass("text-red-400");
  });

  it("shows amber deadline badge for 4-7 days", () => {
    const later = new Date();
    later.setDate(later.getDate() + 5);
    renderCard({ ...baseApp, deadline: later.toISOString() });
    const badge = screen.getByTestId("deadline-badge");
    expect(badge).toHaveClass("text-amber-400");
  });

  it("shows muted deadline badge for > 7 days", () => {
    const far = new Date();
    far.setDate(far.getDate() + 14);
    renderCard({ ...baseApp, deadline: far.toISOString() });
    const badge = screen.getByTestId("deadline-badge");
    expect(badge).toHaveClass("text-muted-foreground");
  });

  it("does not show deadline badge when no deadline", () => {
    renderCard(baseApp);
    expect(screen.queryByTestId("deadline-badge")).not.toBeInTheDocument();
  });
});
