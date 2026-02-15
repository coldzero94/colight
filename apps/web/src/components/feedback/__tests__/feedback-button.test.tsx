import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { FeedbackButton } from "../feedback-button";

// Mock FeedbackModal component
vi.mock("../feedback-modal", () => ({
  FeedbackModal: ({ open, onOpenChange }: { open: boolean; onOpenChange: (open: boolean) => void }) => (
    <div data-testid="feedback-modal" data-open={open}>
      <button onClick={() => onOpenChange(false)}>Close Modal</button>
    </div>
  ),
}));

describe("FeedbackButton", () => {
  it("renders floating button", () => {
    render(<FeedbackButton />);
    const button = screen.getByRole("button", { name: "피드백 보내기" });
    expect(button).toBeInTheDocument();
  });

  it("has correct aria-label for accessibility", () => {
    render(<FeedbackButton />);
    const button = screen.getByLabelText("피드백 보내기");
    expect(button).toBeInTheDocument();
  });

  it("has fixed positioning classes", () => {
    render(<FeedbackButton />);
    const button = screen.getByRole("button", { name: "피드백 보내기" });
    expect(button).toHaveClass("fixed");
  });

  it("shows modal on click", async () => {
    const user = userEvent.setup();
    render(<FeedbackButton />);

    const button = screen.getByRole("button", { name: "피드백 보내기" });
    await user.click(button);

    const modal = screen.getByTestId("feedback-modal");
    expect(modal).toHaveAttribute("data-open", "true");
  });

  it("modal is hidden by default", () => {
    render(<FeedbackButton />);

    const modal = screen.getByTestId("feedback-modal");
    expect(modal).toHaveAttribute("data-open", "false");
  });

  it("can close modal after opening", async () => {
    const user = userEvent.setup();
    render(<FeedbackButton />);

    // Open modal
    const button = screen.getByRole("button", { name: "피드백 보내기" });
    await user.click(button);

    let modal = screen.getByTestId("feedback-modal");
    expect(modal).toHaveAttribute("data-open", "true");

    // Close modal
    const closeButton = screen.getByText("Close Modal");
    await user.click(closeButton);

    modal = screen.getByTestId("feedback-modal");
    expect(modal).toHaveAttribute("data-open", "false");
  });

  it("renders icon inside button", () => {
    const { container } = render(<FeedbackButton />);
    const svg = container.querySelector("svg");
    expect(svg).toBeInTheDocument();
  });
});
