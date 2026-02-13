import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import { StarDisplay } from "../star-display";

describe("StarDisplay", () => {
  it("renders all STAR sections with content", () => {
    render(
      <StarDisplay
        situation="상황 내용"
        task="과제 내용"
        action="행동 내용"
        result="결과 내용"
      />
    );

    expect(screen.getByText("상황 내용")).toBeInTheDocument();
    expect(screen.getByText("과제 내용")).toBeInTheDocument();
    expect(screen.getByText("행동 내용")).toBeInTheDocument();
    expect(screen.getByText("결과 내용")).toBeInTheDocument();
    expect(screen.getByText("Situation")).toBeInTheDocument();
    expect(screen.getByText("Task")).toBeInTheDocument();
    expect(screen.getByText("Action")).toBeInTheDocument();
    expect(screen.getByText("Result")).toBeInTheDocument();
  });

  it("shows placeholder for empty sections", () => {
    render(<StarDisplay />);

    const placeholders = screen.getAllByTestId("star-empty");
    expect(placeholders).toHaveLength(4);
    placeholders.forEach((el) => {
      expect(el).toHaveTextContent("아직 작성되지 않았습니다");
    });
  });
});
