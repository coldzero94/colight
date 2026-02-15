import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import { StarHighlighter } from "../star-highlighter";

describe("StarHighlighter", () => {
  it("renders plain text without tags", () => {
    const plainText = "This is plain text without any STAR tags";
    render(<StarHighlighter text={plainText} />);
    expect(screen.getByText(plainText)).toBeInTheDocument();
  });

  it("highlights [상황] tag with blue background", () => {
    const text = "Text with [상황] situation tag";
    render(<StarHighlighter text={text} />);
    const situationTag = screen.getByText("[상황]");
    expect(situationTag).toBeInTheDocument();
    expect(situationTag).toHaveClass("bg-blue-50");
  });

  it("highlights [과제] tag with amber background", () => {
    const text = "Text with [과제] task tag";
    render(<StarHighlighter text={text} />);
    const taskTag = screen.getByText("[과제]");
    expect(taskTag).toBeInTheDocument();
    expect(taskTag).toHaveClass("bg-amber-50");
  });

  it("highlights [행동] tag with green background", () => {
    const text = "Text with [행동] action tag";
    render(<StarHighlighter text={text} />);
    const actionTag = screen.getByText("[행동]");
    expect(actionTag).toBeInTheDocument();
    expect(actionTag).toHaveClass("bg-green-50");
  });

  it("highlights [결과] tag with purple background", () => {
    const text = "Text with [결과] result tag";
    render(<StarHighlighter text={text} />);
    const resultTag = screen.getByText("[결과]");
    expect(resultTag).toBeInTheDocument();
    expect(resultTag).toHaveClass("bg-purple-50");
  });

  it("highlights multiple STAR tags", () => {
    const text = "[상황] context [과제] task [행동] action [결과] result";
    render(<StarHighlighter text={text} />);
    expect(screen.getByText("[상황]")).toBeInTheDocument();
    expect(screen.getByText("[과제]")).toBeInTheDocument();
    expect(screen.getByText("[행동]")).toBeInTheDocument();
    expect(screen.getByText("[결과]")).toBeInTheDocument();
  });

  it("preserves text between tags", () => {
    const text = "Start [상황] middle [결과] end";
    render(<StarHighlighter text={text} />);
    expect(screen.getByText(/Start/)).toBeInTheDocument();
    expect(screen.getByText(/middle/)).toBeInTheDocument();
    expect(screen.getByText(/end/)).toBeInTheDocument();
  });
});
