import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import { KanbanBoard } from "../kanban-board";
import type { ApplicationDetail } from "@/lib/api/applications";

const mockApps: ApplicationDetail[] = [
  {
    id: "1",
    company_name: "삼성전자",
    position: "BE",
    status: "preparing",
    deadline: null,
    applied_at: null,
    notes: "",
    tags: [],
    cover_letter_count: 2,
    created_at: "2026-01-01T00:00:00Z",
    updated_at: "2026-01-01T00:00:00Z",
  },
  {
    id: "2",
    company_name: "네이버",
    position: "FE",
    status: "preparing",
    deadline: null,
    applied_at: null,
    notes: "",
    tags: [],
    cover_letter_count: 1,
    created_at: "2026-01-02T00:00:00Z",
    updated_at: "2026-01-02T00:00:00Z",
  },
  {
    id: "3",
    company_name: "카카오",
    position: "서버",
    status: "submitted",
    deadline: null,
    applied_at: null,
    notes: "",
    tags: [],
    cover_letter_count: 0,
    created_at: "2026-01-03T00:00:00Z",
    updated_at: "2026-01-03T00:00:00Z",
  },
];

describe("KanbanBoard", () => {
  it("renders 6 columns", () => {
    render(<KanbanBoard applications={mockApps} onStatusChange={vi.fn()} />);
    expect(screen.getByText("준비중")).toBeInTheDocument();
    expect(screen.getByText("제출완료")).toBeInTheDocument();
    expect(screen.getByText("서류검토")).toBeInTheDocument();
    expect(screen.getByText("면접")).toBeInTheDocument();
    expect(screen.getByText("합격")).toBeInTheDocument();
    expect(screen.getByText("불합격")).toBeInTheDocument();
  });

  it("displays card count per column", () => {
    render(<KanbanBoard applications={mockApps} onStatusChange={vi.fn()} />);
    // preparing column: 2 cards
    const preparingCount = screen.getByTestId("column-count-preparing");
    expect(preparingCount).toHaveTextContent("2");
    // submitted column: 1 card
    const submittedCount = screen.getByTestId("column-count-submitted");
    expect(submittedCount).toHaveTextContent("1");
  });

  it("shows empty hint for empty columns", () => {
    render(<KanbanBoard applications={mockApps} onStatusChange={vi.fn()} />);
    // interview, accepted, rejected should show empty hints
    const emptyHints = screen.getAllByText("여기에 드래그하세요");
    expect(emptyHints.length).toBeGreaterThanOrEqual(3);
  });
});
