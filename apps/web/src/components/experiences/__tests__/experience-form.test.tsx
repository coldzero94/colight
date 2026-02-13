import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { ExperienceForm } from "../experience-form";

describe("ExperienceForm", () => {
  const onSubmit = vi.fn();
  const onCancel = vi.fn();

  it("shows validation error for empty title on submit", async () => {
    const user = userEvent.setup();
    render(
      <ExperienceForm mode="create" onSubmit={onSubmit} onCancel={onCancel} />
    );

    const submitButton = screen.getByText("등록");
    await user.click(submitButton);

    expect(
      await screen.findByText("제목은 최소 2자 이상이어야 합니다.")
    ).toBeInTheDocument();
    expect(onSubmit).not.toHaveBeenCalled();
  });

  it("displays character counter for STAR fields", () => {
    render(
      <ExperienceForm mode="create" onSubmit={onSubmit} onCancel={onCancel} />
    );

    const counters = screen.getAllByTestId("char-counter");
    expect(counters.length).toBeGreaterThanOrEqual(4); // S, T, A, R
    expect(counters[0]).toHaveTextContent("0/1000"); // Situation counter
  });

  it("pre-fills form data in edit mode", () => {
    render(
      <ExperienceForm
        mode="edit"
        defaultValues={{
          title: "인턴 경험",
          star_situation: "스타트업에서 일했다",
        }}
        onSubmit={onSubmit}
        onCancel={onCancel}
      />
    );

    const titleInput = screen.getByPlaceholderText("동아리 축제 부스 운영");
    expect(titleInput).toHaveValue("인턴 경험");

    const submitButton = screen.getByText("수정");
    expect(submitButton).toBeInTheDocument();
  });
});
