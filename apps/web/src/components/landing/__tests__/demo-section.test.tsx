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

  it("renders all 5 process steps", () => {
    render(<DemoSection />);
    expect(screen.getByText("URL 입력")).toBeInTheDocument();
    expect(screen.getByText("자동 분석")).toBeInTheDocument();
    expect(screen.getByText("소재 매칭")).toBeInTheDocument();
    expect(screen.getByText("초안 작성")).toBeInTheDocument();
    expect(screen.getByText("AI 첨삭")).toBeInTheDocument();
  });

  it("renders step icons as svg elements", () => {
    const { container } = render(<DemoSection />);
    const svgs = container.querySelectorAll("svg");
    expect(svgs.length).toBeGreaterThanOrEqual(5);
  });

  it("has glass container for steps", () => {
    const { container } = render(<DemoSection />);
    const glass = container.querySelector(".glass");
    expect(glass).toBeInTheDocument();
  });
});
