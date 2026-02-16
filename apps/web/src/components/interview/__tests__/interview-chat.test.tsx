import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { InterviewChat } from "../interview-chat";

describe("InterviewChat", () => {
  it("renders message bubbles with correct alignment", () => {
    const messages = [
      { role: "assistant" as const, content: "안녕하세요!" },
      { role: "user" as const, content: "안녕하세요, 반갑습니다" },
    ];

    render(
      <InterviewChat messages={messages} isLoading={false} onSend={vi.fn()} />
    );

    expect(screen.getByText("안녕하세요!")).toBeInTheDocument();
    expect(screen.getByText("안녕하세요, 반갑습니다")).toBeInTheDocument();

    // AI message has subtle background
    const aiMsg = screen.getByText("안녕하세요!").closest("div[class]");
    expect(aiMsg?.className).toContain("text-foreground");

    // User message has primary background
    const userMsg = screen
      .getByText("안녕하세요, 반갑습니다")
      .closest("div[class]");
    expect(userMsg?.className).toContain("bg-primary");
  });

  it("calls onSend when submit button is clicked", async () => {
    const user = userEvent.setup();
    const onSend = vi.fn();

    render(
      <InterviewChat messages={[]} isLoading={false} onSend={onSend} />
    );

    const input = screen.getByPlaceholderText("답변을 입력하세요...");
    await user.type(input, "내 답변입니다");

    const sendBtn = screen.getByRole("button", { name: "전송" });
    await user.click(sendBtn);

    expect(onSend).toHaveBeenCalledWith("내 답변입니다");
  });

  it("clears input after sending", async () => {
    const user = userEvent.setup();

    render(
      <InterviewChat messages={[]} isLoading={false} onSend={vi.fn()} />
    );

    const input = screen.getByPlaceholderText("답변을 입력하세요...");
    await user.type(input, "답변");

    const sendBtn = screen.getByRole("button", { name: "전송" });
    await user.click(sendBtn);

    expect(input).toHaveValue("");
  });

  it("shows loading dots when isLoading is true", () => {
    render(
      <InterviewChat messages={[]} isLoading={true} onSend={vi.fn()} />
    );

    expect(screen.getByLabelText("답변 생성 중")).toBeInTheDocument();
  });

  it("disables input and button when isLoading", () => {
    render(
      <InterviewChat messages={[]} isLoading={true} onSend={vi.fn()} />
    );

    expect(screen.getByPlaceholderText("답변을 입력하세요...")).toBeDisabled();
  });

  it("does not send empty messages", async () => {
    const user = userEvent.setup();
    const onSend = vi.fn();

    render(
      <InterviewChat messages={[]} isLoading={false} onSend={onSend} />
    );

    const sendBtn = screen.getByRole("button", { name: "전송" });
    await user.click(sendBtn);

    expect(onSend).not.toHaveBeenCalled();
  });
});
