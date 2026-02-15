import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import { BlurredPreview } from "../blurred-preview";

describe("BlurredPreview", () => {
  it("renders children with blur", () => {
    const { container } = render(
      <BlurredPreview>
        <p>결과 내용</p>
      </BlurredPreview>,
    );
    const blurred = container.querySelector(".blur-sm");
    expect(blurred).toBeInTheDocument();
  });

  it("shows default message", () => {
    render(
      <BlurredPreview>
        <p>content</p>
      </BlurredPreview>,
    );
    expect(
      screen.getByText("무료 사용 횟수를 초과했습니다."),
    ).toBeInTheDocument();
  });

  it("shows custom message", () => {
    render(
      <BlurredPreview message="커스텀 메시지">
        <p>content</p>
      </BlurredPreview>,
    );
    expect(screen.getByText("커스텀 메시지")).toBeInTheDocument();
  });

  it("has upgrade link to pricing", () => {
    render(
      <BlurredPreview>
        <p>content</p>
      </BlurredPreview>,
    );
    const link = screen.getByRole("link", { name: "플랜 업그레이드" });
    expect(link).toHaveAttribute("href", "/pricing");
  });
});
