import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { DeleteDialog } from "../delete-dialog";

describe("DeleteDialog", () => {
  it("calls onConfirm after delete confirmation", async () => {
    const user = userEvent.setup();
    const onConfirm = vi.fn();
    const onCancel = vi.fn();

    render(
      <DeleteDialog title="테스트 경험" onConfirm={onConfirm} onCancel={onCancel} />
    );

    expect(screen.getByText(/테스트 경험/)).toBeInTheDocument();

    const deleteBtn = screen.getByText("삭제");
    await user.click(deleteBtn);
    expect(onConfirm).toHaveBeenCalledOnce();
  });
});
