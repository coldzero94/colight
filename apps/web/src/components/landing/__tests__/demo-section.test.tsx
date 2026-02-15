import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import { DemoSection } from "../demo-section";

describe("DemoSection", () => {
  it("renders section heading", () => {
    render(<DemoSection />);
    expect(screen.getByText("한눈에 보는 코칭 프로세스")).toBeInTheDocument();
  });

  it("renders description text", () => {
    render(<DemoSection />);
    expect(
      screen.getByText("채용공고 입력부터 최종 자소서까지, 5단계 AI 코칭")
    ).toBeInTheDocument();
  });

  it("renders placeholder message", () => {
    render(<DemoSection />);
    expect(screen.getByText("서비스 스크린샷 준비 중")).toBeInTheDocument();
  });

  it("renders monitor icon", () => {
    const { container } = render(<DemoSection />);
    const svg = container.querySelector("svg");
    expect(svg).toBeInTheDocument();
  });

  it("has placeholder container with gray background", () => {
    const { container } = render(<DemoSection />);
    const placeholder = container.querySelector(".bg-gray-50");
    expect(placeholder).toBeInTheDocument();
  });
});
