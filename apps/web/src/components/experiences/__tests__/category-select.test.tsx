import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import { CategorySelect } from "../category-select";

describe("CategorySelect", () => {
  const mockRegistration = {
    name: "category",
    onChange: vi.fn(),
    onBlur: vi.fn(),
    ref: vi.fn(),
  };

  it("renders all category options", () => {
    render(<CategorySelect registration={mockRegistration} />);

    expect(screen.getByText("선택해주세요")).toBeInTheDocument();
    expect(screen.getByText("인턴")).toBeInTheDocument();
    expect(screen.getByText("대외활동")).toBeInTheDocument();
    expect(screen.getByText("프로젝트")).toBeInTheDocument();
    expect(screen.getByText("아르바이트")).toBeInTheDocument();
    expect(screen.getByText("동아리")).toBeInTheDocument();
    expect(screen.getByText("봉사활동")).toBeInTheDocument();
    expect(screen.getByText("기타")).toBeInTheDocument();
  });

  it("renders label", () => {
    render(<CategorySelect registration={mockRegistration} />);
    expect(screen.getByText("카테고리")).toBeInTheDocument();
  });

  it("shows error message when provided", () => {
    render(
      <CategorySelect
        registration={mockRegistration}
        error="카테고리를 선택해주세요"
      />
    );

    expect(screen.getByText("카테고리를 선택해주세요")).toBeInTheDocument();
  });

  it("does not show error when not provided", () => {
    const { container } = render(
      <CategorySelect registration={mockRegistration} />
    );

    expect(container.querySelector(".text-red-500")).not.toBeInTheDocument();
  });
});
