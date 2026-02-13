import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { SortSelect } from "../sort-select";

describe("SortSelect", () => {
  it("changes sort order", async () => {
    const user = userEvent.setup();
    const onChange = vi.fn();

    render(<SortSelect value="latest" onChange={onChange} />);

    const select = screen.getByLabelText("정렬");
    await user.selectOptions(select, "oldest");
    expect(onChange).toHaveBeenCalledWith("oldest");
  });
});
