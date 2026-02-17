import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import { DashboardSummary } from "../dashboard-summary";
import type { ApplicationStats } from "@/lib/api/applications";

const mockStats: ApplicationStats = {
  total: 8,
  by_status: {
    preparing: 3,
    submitted: 2,
    in_review: 1,
    interview: 1,
    accepted: 1,
    rejected: 0,
  },
  upcoming_deadlines: [
    {
      id: "1",
      company_name: "토스",
      position: "서버",
      status: "preparing",
      deadline: new Date(Date.now() + 48 * 60 * 60 * 1000).toISOString(),
      applied_at: null,
      notes: "",
      tags: [],
      cover_letter_count: 0,
      created_at: "2026-01-01T00:00:00Z",
      updated_at: "2026-01-01T00:00:00Z",
    },
  ],
};

describe("DashboardSummary", () => {
  it("displays total and active application counts", () => {
    render(<DashboardSummary stats={mockStats} />);
    // Total
    expect(screen.getByTestId("stat-total")).toHaveTextContent("8");
    // Active (preparing + submitted + in_review + interview)
    expect(screen.getByTestId("stat-active")).toHaveTextContent("7");
  });

  it("highlights upcoming deadlines", () => {
    render(<DashboardSummary stats={mockStats} />);
    expect(screen.getByTestId("stat-deadlines")).toHaveTextContent("1");
  });

  it("shows completed count", () => {
    render(<DashboardSummary stats={mockStats} />);
    expect(screen.getByTestId("stat-completed")).toHaveTextContent("1");
  });
});
