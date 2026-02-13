import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import { ExperienceCard } from "../experience-card";
import type { Experience } from "@/lib/api/experiences";

// Mock next/link
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
  id: "test-id",
  user_id: "user-1",
  title: "인턴 경험",
  category: "인턴",
  period_start: "2024-01-01",
  period_end: "2024-06-30",
  role: "개발자",
  content: "",
  result: "",
  star_situation: "스타트업에서 인턴을 했다",
  star_task: "",
  star_action: "",
  star_result: "",
  keywords: null,
  source: "manual",
  is_archived: false,
  created_at: "2024-01-01T00:00:00Z",
  updated_at: "2024-01-01T00:00:00Z",
  weapons: [
    {
      id: "w1",
      weapon_code: "W01",
      confidence: 0.9,
      is_primary: true,
      reasoning: "",
      user_confirmed: false,
      user_modified: false,
    },
    {
      id: "w2",
      weapon_code: "W02",
      confidence: 0.8,
      is_primary: false,
      reasoning: "",
      user_confirmed: false,
      user_modified: false,
    },
    {
      id: "w3",
      weapon_code: "W03",
      confidence: 0.7,
      is_primary: false,
      reasoning: "",
      user_confirmed: false,
      user_modified: false,
    },
    {
      id: "w4",
      weapon_code: "W04",
      confidence: 0.6,
      is_primary: false,
      reasoning: "",
      user_confirmed: false,
      user_modified: false,
    },
  ],
};

import { vi } from "vitest";

describe("ExperienceCard", () => {
  it("renders title, category badge, and period in grid view", () => {
    render(<ExperienceCard experience={mockExperience} viewMode="grid" />);

    expect(screen.getByText("인턴 경험")).toBeInTheDocument();
    expect(screen.getByTestId("experience-card-grid")).toBeInTheDocument();
    expect(screen.getByText("2024.01 ~ 2024.06")).toBeInTheDocument();
  });

  it("renders horizontal layout in list view", () => {
    render(<ExperienceCard experience={mockExperience} viewMode="list" />);

    expect(screen.getByTestId("experience-card-list")).toBeInTheDocument();
    expect(screen.getByText("인턴 경험")).toBeInTheDocument();
  });

  it("truncates weapon badges to max 3 with +N indicator", () => {
    render(<ExperienceCard experience={mockExperience} viewMode="grid" />);

    // Should show max 3 weapons and a "+1" indicator
    expect(screen.getByText("+1")).toBeInTheDocument();
  });

  it("navigates to detail page on click", () => {
    render(<ExperienceCard experience={mockExperience} viewMode="grid" />);

    const link = screen.getByTestId("experience-card-grid");
    expect(link).toHaveAttribute("href", "/experiences/test-id");
  });
});
