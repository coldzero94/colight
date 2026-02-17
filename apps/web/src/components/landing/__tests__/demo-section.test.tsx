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
    expect(screen.getAllByText("URL 입력").length).toBeGreaterThan(0);
    expect(screen.getAllByText("자동 분석").length).toBeGreaterThan(0);
    expect(screen.getAllByText("소재 매칭").length).toBeGreaterThan(0);
    expect(screen.getAllByText("초안 작성").length).toBeGreaterThan(0);
    expect(screen.getAllByText("AI 첨삭").length).toBeGreaterThan(0);
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
