import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import { StreamingProgress } from "../streaming-progress";

describe("StreamingProgress", () => {
  it("shows streaming indicator when isStreaming", () => {
    render(
      <StreamingProgress charCount={500} charLimit={1000} isStreaming={true} />
    );
    expect(screen.getByText("초안 생성 중...")).toBeInTheDocument();
    expect(screen.getByTestId("streaming-indicator")).toBeInTheDocument();
  });

  it("shows 작성 완료 when not streaming", () => {
    render(
      <StreamingProgress charCount={500} charLimit={1000} isStreaming={false} />
    );
    expect(screen.getByText("작성 완료")).toBeInTheDocument();
    expect(screen.queryByTestId("streaming-indicator")).not.toBeInTheDocument();
  });

  it("displays char count and limit", () => {
    render(
      <StreamingProgress charCount={750} charLimit={1000} isStreaming={false} />
    );
    expect(screen.getByText("750 / 1000자")).toBeInTheDocument();
  });

  it("shows red text and 초과 when over limit", () => {
    const { container } = render(
      <StreamingProgress charCount={1100} charLimit={1000} isStreaming={false} />
    );
    expect(screen.getByText(/1100 \/ 1000자/)).toBeInTheDocument();
    expect(screen.getByText(/\(초과\)/)).toBeInTheDocument();
    const element = container.querySelector(".text-red-600");
    expect(element).toBeInTheDocument();
  });

  it("shows normal color when under limit", () => {
    const { container } = render(
      <StreamingProgress charCount={500} charLimit={1000} isStreaming={false} />
    );
    const redElement = container.querySelector(".text-red-600");
    // The count element should use text-gray-900, not red
    expect(screen.getByText("500 / 1000자")).toHaveClass("text-gray-900");
    expect(redElement).not.toBeInTheDocument();
  });

  it("displays progress bar", () => {
    const { container } = render(
      <StreamingProgress charCount={500} charLimit={1000} isStreaming={false} />
    );
    // Progress bar is a div with bg-blue-500 class
    const progressBar = container.querySelector(".bg-blue-500");
    expect(progressBar).toBeInTheDocument();
  });
});
