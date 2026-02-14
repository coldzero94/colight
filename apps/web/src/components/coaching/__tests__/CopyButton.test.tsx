import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import { CopyButton } from "../copy-button";

describe("CopyButton", () => {
  it("renders copy button", () => {
    render(<CopyButton content="Test content" />);
    expect(screen.getByRole("button", { name: /복사/ })).toBeInTheDocument();
  });

  it("renders with icon", () => {
    render(<CopyButton content="Test" />);
    const button = screen.getByRole("button", { name: /복사/ });
    expect(button).toHaveTextContent("📋");
  });

  it("accepts removeStarTags prop", () => {
    render(<CopyButton content="[상황]\nTest" removeStarTags={true} />);
    expect(screen.getByRole("button")).toBeInTheDocument();
  });

  it("accepts removeStarTags false", () => {
    render(<CopyButton content="Test" removeStarTags={false} />);
    expect(screen.getByRole("button")).toBeInTheDocument();
  });

  // TODO: E2E tests for clipboard API interactions
  // - copies plain text to clipboard on click
  // - strips HTML tags from copied content
  // - removes STAR tags when option enabled
  // - shows success toast with character count
  // These require browser clipboard API which is not available in jsdom
});
