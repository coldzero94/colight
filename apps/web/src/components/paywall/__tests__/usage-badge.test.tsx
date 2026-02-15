import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import { UsageBadge } from "../usage-badge";

describe("UsageBadge", () => {
  it("shows remaining/limit for free tier", () => {
    render(<UsageBadge used={1} limit={3} />);
    expect(screen.getByText("2/3")).toBeInTheDocument();
  });

  it("shows green when usage is low", () => {
    const { container } = render(<UsageBadge used={0} limit={3} />);
    expect(container.firstChild).toHaveClass("bg-green-100");
  });

  it("shows yellow when usage is at 70%+", () => {
    const { container } = render(<UsageBadge used={3} limit={4} />);
    expect(container.firstChild).toHaveClass("bg-yellow-100");
  });

  it("shows red when limit reached", () => {
    const { container } = render(<UsageBadge used={3} limit={3} />);
    expect(container.firstChild).toHaveClass("bg-red-100");
  });

  it("shows unlimited for paid tier (limit=-1)", () => {
    render(<UsageBadge used={0} limit={-1} />);
    expect(screen.getByText("무제한")).toBeInTheDocument();
  });
});
