import { describe, it, expect } from "vitest";
import { render, screen, fireEvent } from "@testing-library/react";
import { StructureExample } from "../structure-example";

describe("StructureExample", () => {
  const exampleText = "This is an example text for structure guidance";

  it("starts collapsed (example text not in DOM)", () => {
    render(<StructureExample example={exampleText} />);
    expect(screen.queryByText(exampleText)).not.toBeInTheDocument();
  });

  it("shows toggle button with heading text", () => {
    render(<StructureExample example={exampleText} />);
    const button = screen.getByRole("button");
    expect(button).toBeInTheDocument();
    expect(button).toHaveTextContent("좋은 구조 예시");
  });

  it("expands on button click to show example", () => {
    render(<StructureExample example={exampleText} />);
    const button = screen.getByRole("button");
    fireEvent.click(button);
    expect(screen.getByText(exampleText)).toBeVisible();
  });

  it("collapses on second button click", () => {
    render(<StructureExample example={exampleText} />);
    const button = screen.getByRole("button");

    fireEvent.click(button);
    expect(screen.getByText(exampleText)).toBeVisible();

    fireEvent.click(button);
    expect(screen.queryByText(exampleText)).not.toBeInTheDocument();
  });

  it("heading text remains the same between states", () => {
    render(<StructureExample example={exampleText} />);
    const button = screen.getByRole("button");

    expect(button).toHaveTextContent("좋은 구조 예시");
    fireEvent.click(button);
    expect(button).toHaveTextContent("좋은 구조 예시");
  });
});
