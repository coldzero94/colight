import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { RecommendCard } from "../recommend-card";

const mockExperience = {
  id: "exp-1",
  title: "프로젝트 리더 경험",
  category: "프로젝트",
  period_start: "2024-01",
  period_end: "2024-06",
  weapons: [
    { weapon_name: "문제해결력" },
    { weapon_name: "리더십" },
    { weapon_name: "커뮤니케이션" },
  ],
  star_situation: "팀 프로젝트에서 리더로 활동하며 어려운 문제를 해결했습니다",
  matchScore: 92,
};

describe("RecommendCard", () => {
  it("renders experience title and category", () => {
    render(
      <RecommendCard
        experience={mockExperience}
        selected={false}
        disabled={false}
        onToggle={vi.fn()}
      />
    );

    expect(screen.getByText("프로젝트 리더 경험")).toBeInTheDocument();
  });

  it("renders match score", () => {
    render(
      <RecommendCard
        experience={mockExperience}
        selected={false}
        disabled={false}
        onToggle={vi.fn()}
      />
    );

    expect(screen.getByText("적합도 92%")).toBeInTheDocument();
  });

  it("renders period when available", () => {
    render(
      <RecommendCard
        experience={mockExperience}
        selected={false}
        disabled={false}
        onToggle={vi.fn()}
      />
    );

    expect(screen.getByText("2024-01 ~ 2024-06")).toBeInTheDocument();
  });

  it("does not render period when not available", () => {
    const expWithoutPeriod = { ...mockExperience, period_start: undefined, period_end: undefined };

    render(
      <RecommendCard
        experience={expWithoutPeriod}
        selected={false}
        disabled={false}
        onToggle={vi.fn()}
      />
    );

    expect(screen.queryByText(/~/)).not.toBeInTheDocument();
  });

  it("renders weapon badges", () => {
    render(
      <RecommendCard
        experience={mockExperience}
        selected={false}
        disabled={false}
        onToggle={vi.fn()}
      />
    );

    expect(screen.getByText("문제해결력")).toBeInTheDocument();
    expect(screen.getByText("리더십")).toBeInTheDocument();
    expect(screen.getByText("커뮤니케이션")).toBeInTheDocument();
  });

  it("renders situation preview", () => {
    render(
      <RecommendCard
        experience={mockExperience}
        selected={false}
        disabled={false}
        onToggle={vi.fn()}
      />
    );

    expect(
      screen.getByText("팀 프로젝트에서 리더로 활동하며 어려운 문제를 해결했습니다")
    ).toBeInTheDocument();
  });

  it("calls onToggle when clicked", async () => {
    const user = userEvent.setup();
    const onToggle = vi.fn();

    render(
      <RecommendCard
        experience={mockExperience}
        selected={false}
        disabled={false}
        onToggle={onToggle}
      />
    );

    const checkbox = screen.getByRole("checkbox");
    await user.click(checkbox);

    expect(onToggle).toHaveBeenCalledOnce();
  });

  it("shows selected state styling", () => {
    render(
      <RecommendCard
        experience={mockExperience}
        selected={true}
        disabled={false}
        onToggle={vi.fn()}
      />
    );

    const checkbox = screen.getByRole("checkbox");
    expect(checkbox).toBeChecked();
  });

  it("shows disabled state", () => {
    render(
      <RecommendCard
        experience={mockExperience}
        selected={false}
        disabled={true}
        onToggle={vi.fn()}
      />
    );

    const checkbox = screen.getByRole("checkbox");
    expect(checkbox).toBeDisabled();
  });

  it("does not call onToggle when disabled and clicked", async () => {
    const user = userEvent.setup();
    const onToggle = vi.fn();

    render(
      <RecommendCard
        experience={mockExperience}
        selected={false}
        disabled={true}
        onToggle={onToggle}
      />
    );

    const checkbox = screen.getByRole("checkbox");
    await user.click(checkbox);

    expect(onToggle).not.toHaveBeenCalled();
  });
});
