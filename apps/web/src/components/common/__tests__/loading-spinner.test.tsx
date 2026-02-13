import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import { LoadingSpinner } from "../loading-spinner";

describe("LoadingSpinner", () => {
  it("renders without text", () => {
    const { container } = render(<LoadingSpinner />);
    // Spinner div should exist with animate-spin class
    const spinner = container.querySelector(".animate-spin");
    expect(spinner).toBeInTheDocument();
  });

  it("renders with text", () => {
    render(<LoadingSpinner text="로딩 중..." />);
    expect(screen.getByText("로딩 중...")).toBeInTheDocument();
  });

  it("does not render text when not provided", () => {
    render(<LoadingSpinner />);
    expect(screen.queryByText("로딩 중...")).not.toBeInTheDocument();
  });

  it("applies small size class", () => {
    const { container } = render(<LoadingSpinner size="sm" />);
    const spinner = container.querySelector(".animate-spin");
    expect(spinner).toHaveClass("h-4", "w-4", "border-2");
  });

  it("applies medium size class by default", () => {
    const { container } = render(<LoadingSpinner />);
    const spinner = container.querySelector(".animate-spin");
    expect(spinner).toHaveClass("h-8", "w-8", "border-4");
  });

  it("applies large size class", () => {
    const { container } = render(<LoadingSpinner size="lg" />);
    const spinner = container.querySelector(".animate-spin");
    expect(spinner).toHaveClass("h-12", "w-12", "border-4");
  });
});
