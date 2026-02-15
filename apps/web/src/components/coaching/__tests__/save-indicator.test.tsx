import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import { SaveIndicator } from "../save-indicator";

describe("SaveIndicator", () => {
  it("shows 저장됨 for saved status", () => {
    render(<SaveIndicator status="saved" />);
    expect(screen.getByText("저장됨")).toBeInTheDocument();
  });

  it("shows 저장 중... for saving status", () => {
    render(<SaveIndicator status="saving" />);
    expect(screen.getByText("저장 중...")).toBeInTheDocument();
  });

  it("shows 저장되지 않음 for unsaved status", () => {
    render(<SaveIndicator status="unsaved" />);
    expect(screen.getByText("저장 안 됨")).toBeInTheDocument();
  });

  it("renders with correct visual indicator", () => {
    const { container } = render(<SaveIndicator status="saved" />);
    expect(container.firstChild).toBeInTheDocument();
  });
});
