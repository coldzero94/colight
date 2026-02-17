import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
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

  it("changes panel content when clicking a step", async () => {
    const user = userEvent.setup();
    render(<DemoSection />);

    expect(screen.getByText("채용공고 URL 붙여넣기")).toBeInTheDocument();

    await user.click(screen.getAllByRole("button", { name: /AI 첨삭/ })[0]);
    expect(screen.getByText("최종 첨삭 리포트")).toBeInTheDocument();

    await user.click(screen.getAllByRole("button", { name: /소재 매칭/ })[0]);
    expect(screen.getByText("STAR 경험 목록")).toBeInTheDocument();
  });
});
