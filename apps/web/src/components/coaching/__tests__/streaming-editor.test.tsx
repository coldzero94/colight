import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import { StreamingEditor } from "../streaming-editor";

describe("StreamingEditor", () => {
  it("renders content text", () => {
    const content = "This is the coaching content";
    render(<StreamingEditor content={content} isStreaming={false} />);
    expect(screen.getByText(content)).toBeInTheDocument();
  });

  it("applies animate-pulse when streaming", () => {
    const { container } = render(
      <StreamingEditor content="Streaming content" isStreaming={true} />
    );
    const element = container.querySelector(".animate-pulse");
    expect(element).toBeInTheDocument();
  });

  it("does not apply animate-pulse when not streaming", () => {
    const { container } = render(
      <StreamingEditor content="Static content" isStreaming={false} />
    );
    const element = container.querySelector(".animate-pulse");
    expect(element).not.toBeInTheDocument();
  });

  it("shows placeholder when empty", () => {
    render(<StreamingEditor content="" isStreaming={false} />);
    expect(
      screen.getByText("초안이 생성되면 여기에 표시됩니다...")
    ).toBeInTheDocument();
  });

  it("does not show placeholder when content exists", () => {
    render(<StreamingEditor content="Some content" isStreaming={false} />);
    expect(
      screen.queryByText("초안이 생성되면 여기에 표시됩니다...")
    ).not.toBeInTheDocument();
  });
});
