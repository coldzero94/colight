import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { ViewToggle } from "../view-toggle";

describe("ViewToggle", () => {
  it("toggles between grid and list view", async () => {
    const user = userEvent.setup();
    const onChange = vi.fn();

    render(<ViewToggle mode="grid" onChange={onChange} />);

    const listBtn = screen.getByLabelText("리스트 뷰");
    await user.click(listBtn);
    expect(onChange).toHaveBeenCalledWith("list");

    onChange.mockClear();

    const gridBtn = screen.getByLabelText("그리드 뷰");
    await user.click(gridBtn);
    expect(onChange).toHaveBeenCalledWith("grid");
  });
});
