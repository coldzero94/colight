import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { StarField } from "../star-field";

describe("StarField", () => {
  const baseProps = {
    label: "Situation",
    icon: "📍",
    placeholder: "상황을 설명해주세요",
    guide: "어떤 상황이었나요?",
    maxLength: 1000,
    registration: {
      name: "star_situation",
      onChange: vi.fn(),
      onBlur: vi.fn(),
      ref: vi.fn(),
    },
  };

  it("renders label and icon", () => {
    render(<StarField {...baseProps} />);

    expect(screen.getByText("Situation")).toBeInTheDocument();
    expect(screen.getByText("📍")).toBeInTheDocument();
  });

  it("renders guide text", () => {
    render(<StarField {...baseProps} />);

    expect(screen.getByText("어떤 상황이었나요?")).toBeInTheDocument();
  });

  it("shows character counter starting at 0", () => {
    render(<StarField {...baseProps} />);

    expect(screen.getByTestId("char-counter")).toHaveTextContent("0/1000");
  });

  it("updates character counter on input", async () => {
    const user = userEvent.setup();
    render(<StarField {...baseProps} />);

    const textarea = screen.getByPlaceholderText("상황을 설명해주세요");
    await user.type(textarea, "Hello");

    expect(screen.getByTestId("char-counter")).toHaveTextContent("5/1000");
  });

  it("shows error message when provided", () => {
    render(<StarField {...baseProps} error="Situation은 1000자 이내로 입력해주세요." />);

    expect(
      screen.getByText("Situation은 1000자 이내로 입력해주세요.")
    ).toBeInTheDocument();
  });

  it("shows red counter when exceeding max length", () => {
    render(<StarField {...baseProps} maxLength={3} defaultValue="abcde" />);

    const counter = screen.getByTestId("char-counter");
    expect(counter).toHaveTextContent("5/3");
    expect(counter.className).toContain("text-red-500");
  });

  it("shows gray counter within limit", () => {
    render(<StarField {...baseProps} defaultValue="ab" />);

    const counter = screen.getByTestId("char-counter");
    expect(counter).toHaveTextContent("2/1000");
    expect(counter.className).toContain("text-gray-400");
  });
});
