import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, fireEvent } from "@testing-library/react";

vi.mock("next/navigation", () => ({
  usePathname: () => "/experiences",
}));

const mockMutate = vi.fn();
vi.mock("@/hooks/use-feedback", () => ({
  useFeedback: () => ({
    mutate: mockMutate,
    isPending: false,
  }),
}));

import { FeedbackModal } from "../feedback-modal";

describe("FeedbackModal", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("renders category buttons and textarea when open", () => {
    render(<FeedbackModal open={true} onOpenChange={vi.fn()} />);

    expect(screen.getByText("버그 신고")).toBeInTheDocument();
    expect(screen.getByText("개선 제안")).toBeInTheDocument();
    expect(screen.getByText("기타")).toBeInTheDocument();
    expect(
      screen.getByPlaceholderText("의견을 자유롭게 작성해주세요..."),
    ).toBeInTheDocument();
  });

  it("calls mutate with correct payload on submit", () => {
    render(<FeedbackModal open={true} onOpenChange={vi.fn()} />);

    const textarea = screen.getByPlaceholderText(
      "의견을 자유롭게 작성해주세요...",
    );
    fireEvent.change(textarea, { target: { value: "Great app!" } });

    fireEvent.click(screen.getByText("보내기"));

    expect(mockMutate).toHaveBeenCalledWith(
      {
        category: "improvement",
        content: "Great app!",
        page_url: "/experiences",
      },
      expect.objectContaining({
        onSuccess: expect.any(Function),
        onError: expect.any(Function),
      }),
    );
  });

  it("disables submit button when content is empty", () => {
    render(<FeedbackModal open={true} onOpenChange={vi.fn()} />);

    const submitBtn = screen.getByText("보내기");
    expect(submitBtn).toBeDisabled();
  });

  it("calls onOpenChange when cancel is clicked", () => {
    const onOpenChange = vi.fn();
    render(<FeedbackModal open={true} onOpenChange={onOpenChange} />);

    fireEvent.click(screen.getByText("취소"));
    expect(onOpenChange).toHaveBeenCalledWith(false);
  });
});
