import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import { ExperienceList } from "../experience-list";
import type { Experience } from "@/lib/api/experiences";

vi.mock("next/link", () => ({
  default: ({
    children,
    href,
    ...props
  }: {
    children: React.ReactNode;
    href: string;
    [key: string]: unknown;
  }) => (
    <a href={href} {...props}>
      {children}
    </a>
  ),
}));

const mockExperience: Experience = {
  id: "test-1",
  user_id: "user-1",
  title: "프로젝트 경험",
  category: "프로젝트",
  period_start: "2024-01-01",
  period_end: "2024-06-30",
  role: "팀장",
  content: "",
  result: "",
  star_situation: "팀 프로젝트를 진행했다",
  star_task: "",
  star_action: "",
  star_result: "",
  keywords: null,
  source: "manual",
  is_archived: false,
  created_at: "2024-01-01T00:00:00Z",
  updated_at: "2024-01-01T00:00:00Z",
  weapons: [],
};

describe("ExperienceList", () => {
  it("renders experience cards in grid view", () => {
    const experiences = [mockExperience, { ...mockExperience, id: "test-2", title: "인턴 경험" }];
    render(<ExperienceList experiences={experiences} viewMode="grid" />);

    expect(screen.getByTestId("experience-grid-view")).toBeInTheDocument();
    expect(screen.getByText("프로젝트 경험")).toBeInTheDocument();
    expect(screen.getByText("인턴 경험")).toBeInTheDocument();
  });

  it("renders experience cards in list view", () => {
    const experiences = [mockExperience];
    render(<ExperienceList experiences={experiences} viewMode="list" />);

    expect(screen.getByTestId("experience-list-view")).toBeInTheDocument();
    expect(screen.getByText("프로젝트 경험")).toBeInTheDocument();
  });

  it("renders empty when no experiences", () => {
    const { container } = render(
      <ExperienceList experiences={[]} viewMode="grid" />
    );
    // Grid container should exist but be empty
    const grid = container.querySelector("[data-testid='experience-grid-view']");
    expect(grid).toBeInTheDocument();
    expect(grid?.children.length).toBe(0);
  });
});
