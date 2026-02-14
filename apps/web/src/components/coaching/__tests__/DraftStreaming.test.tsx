import { describe, it, expect, vi } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { DraftStreaming } from "../draft-streaming";

describe("DraftStreaming", () => {
  it("shows loading indicator when streaming starts", () => {
    const onComplete = vi.fn();

    render(
      <DraftStreaming
        content=""
        isStreaming={true}
        charLimit={800}
        onComplete={onComplete}
      />
    );

    // Loading indicator should be visible
    expect(screen.getByTestId("streaming-indicator")).toBeInTheDocument();
  });

  it("renders streamed text in real-time", () => {
    const onComplete = vi.fn();

    const { rerender } = render(
      <DraftStreaming
        content="[상황]\n"
        isStreaming={true}
        charLimit={800}
        onComplete={onComplete}
      />
    );

    // Initial content
    expect(screen.getByText(/\[상황\]/)).toBeInTheDocument();

    // Update content (simulating streaming)
    rerender(
      <DraftStreaming
        content="[상황]\n저는 팀 프로젝트에서..."
        isStreaming={true}
        charLimit={800}
        onComplete={onComplete}
      />
    );

    // Updated content should appear
    expect(screen.getByText(/저는 팀 프로젝트에서/)).toBeInTheDocument();
  });

  it("highlights STAR tags with correct colors", () => {
    const onComplete = vi.fn();

    render(
      <DraftStreaming
        content="[상황]\n설명\n[과제]\n정의\n[행동]\n실행\n[결과]\n성과"
        isStreaming={false}
        charLimit={800}
        onComplete={onComplete}
      />
    );

    // STAR tags should be highlighted
    const situationTag = screen.getByText("[상황]");
    expect(situationTag).toHaveClass("text-blue-700");

    const taskTag = screen.getByText("[과제]");
    expect(taskTag).toHaveClass("text-amber-700");

    const actionTag = screen.getByText("[행동]");
    expect(actionTag).toHaveClass("text-green-700");

    const resultTag = screen.getByText("[결과]");
    expect(resultTag).toHaveClass("text-purple-700");
  });

  it("updates character counter during streaming", () => {
    const onComplete = vi.fn();

    const { rerender } = render(
      <DraftStreaming
        content="Test"
        isStreaming={true}
        charLimit={800}
        onComplete={onComplete}
      />
    );

    // Initial counter (4 chars)
    expect(screen.getByText(/4.*\/ 800/)).toBeInTheDocument();

    // Update with more content
    rerender(
      <DraftStreaming
        content="Test content with more text"
        isStreaming={true}
        charLimit={800}
        onComplete={onComplete}
      />
    );

    // Counter should update to 27 chars
    expect(screen.getByText(/27.*\/ 800/)).toBeInTheDocument();
  });

  it("shows edit button after streaming completes", async () => {
    const user = userEvent.setup();
    const onComplete = vi.fn();

    const { rerender } = render(
      <DraftStreaming
        content="[상황]\n초안 내용"
        isStreaming={true}
        charLimit={800}
        onComplete={onComplete}
      />
    );

    // Edit button should not be visible during streaming
    expect(screen.queryByRole("button", { name: /편집/ })).not.toBeInTheDocument();

    // Complete streaming
    rerender(
      <DraftStreaming
        content="[상황]\n초안 내용\n[결과]\n완료"
        isStreaming={false}
        charLimit={800}
        onComplete={onComplete}
      />
    );

    // Edit button should appear
    await waitFor(() => {
      expect(screen.getByRole("button", { name: /편집/ })).toBeInTheDocument();
    });

    // Click edit button
    const editButton = screen.getByRole("button", { name: /편집/ });
    await user.click(editButton);

    expect(onComplete).toHaveBeenCalled();
  });
});
