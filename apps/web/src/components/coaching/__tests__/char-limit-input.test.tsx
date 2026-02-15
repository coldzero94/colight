import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import type { UseFormRegisterReturn } from "react-hook-form";
import { CharLimitInput } from "../char-limit-input";

describe("CharLimitInput", () => {
  const mockRegistration: UseFormRegisterReturn = {
    name: "charLimit",
    onChange: () => Promise.resolve(),
    onBlur: () => Promise.resolve(),
    ref: () => {},
  };

  it("renders number input with default value", () => {
    render(
      <CharLimitInput
        registration={mockRegistration}
        defaultValue={1000}
      />
    );
    const input = screen.getByRole("spinbutton");
    expect(input).toBeInTheDocument();
    expect(input).toHaveValue(1000);
  });

  it("renders without default value", () => {
    render(<CharLimitInput registration={mockRegistration} />);
    const input = screen.getByRole("spinbutton");
    expect(input).toBeInTheDocument();
    expect(input).toHaveValue(null);
  });

  it("displays error message with role alert", () => {
    const errorMessage = "글자 수 제한은 필수입니다";
    render(
      <CharLimitInput
        registration={mockRegistration}
        error={errorMessage}
      />
    );
    const error = screen.getByRole("alert");
    expect(error).toBeInTheDocument();
    expect(error).toHaveTextContent(errorMessage);
  });

  it("does not display error when error prop is undefined", () => {
    render(<CharLimitInput registration={mockRegistration} />);
    expect(screen.queryByRole("alert")).not.toBeInTheDocument();
  });
});
