import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { ExperienceActions } from "../experience-actions";

const mockPush = vi.fn();
vi.mock("next/navigation", () => ({
  useRouter: () => ({ push: mockPush }),
}));

describe("ExperienceActions", () => {
  it("navigates to edit page on edit click", async () => {
    const user = userEvent.setup();
    const onDeleteClick = vi.fn();

    render(
      <ExperienceActions experienceId="test-123" onDeleteClick={onDeleteClick} />
    );

    const editBtn = screen.getByText("수정");
    await user.click(editBtn);
    expect(mockPush).toHaveBeenCalledWith("/experiences/test-123/edit");
  });
});
