import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import { CharCounter } from "../char-counter";

describe("CharCounter", () => {
  it("shows 여유 있음 when under 70%", () => {
    render(<CharCounter current={500} limit={1000} />);
    expect(screen.getByText("여유 있음")).toBeInTheDocument();
    expect(screen.getByText("500 / 1000자")).toBeInTheDocument();
  });

  it("shows 적절한 범위 when 70-90%", () => {
    render(<CharCounter current={750} limit={1000} />);
    expect(screen.getByText("적절한 범위")).toBeInTheDocument();
    expect(screen.getByText("750 / 1000자")).toBeInTheDocument();
  });

  it("shows 거의 다 참 when 90-100%", () => {
    render(<CharCounter current={950} limit={1000} />);
    expect(screen.getByText("거의 다 참")).toBeInTheDocument();
    expect(screen.getByText("950 / 1000자")).toBeInTheDocument();
  });

  it("shows 초과 경고 when over 100%", () => {
    render(<CharCounter current={1100} limit={1000} />);
    expect(screen.getByText("초과 경고")).toBeInTheDocument();
    expect(screen.getByText(/1100 \/ 1000자/)).toBeInTheDocument();
    expect(screen.getByText(/\(초과\)/)).toBeInTheDocument();
  });

  it("handles exact limit boundary", () => {
    render(<CharCounter current={1000} limit={1000} />);
    expect(screen.getByText("거의 다 참")).toBeInTheDocument();
    expect(screen.getByText("1000 / 1000자")).toBeInTheDocument();
  });
});
