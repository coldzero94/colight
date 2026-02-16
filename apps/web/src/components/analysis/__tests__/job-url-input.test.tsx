import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { JobUrlInput } from "../job-url-input";

describe("JobUrlInput", () => {
  it("renders domain badge for jobkorea URL", async () => {
    const user = userEvent.setup();
    const onSubmit = vi.fn();

    render(<JobUrlInput onSubmit={onSubmit} isLoading={false} />);

    const input = screen.getByPlaceholderText(/URL을 입력/i);
    await user.type(input, "https://www.jobkorea.co.kr/Recruit/GI_Read/12345");

    expect(screen.getByText("잡코리아")).toBeInTheDocument();
  });

  it("renders domain badge for catch URL", async () => {
    const user = userEvent.setup();
    const onSubmit = vi.fn();

    render(<JobUrlInput onSubmit={onSubmit} isLoading={false} />);

    const input = screen.getByPlaceholderText(/URL을 입력/i);
    await user.type(input, "https://www.catch.co.kr/NCS/RecruitInfoDetail/12345");

    expect(screen.getByText("캐치")).toBeInTheDocument();
  });

  it("renders domain badge for wanted URL", async () => {
    const user = userEvent.setup();
    const onSubmit = vi.fn();

    render(<JobUrlInput onSubmit={onSubmit} isLoading={false} />);

    const input = screen.getByPlaceholderText(/URL을 입력/i);
    await user.type(input, "https://www.wanted.co.kr/wd/12345");

    expect(screen.getByText("원티드")).toBeInTheDocument();
  });

  it("shows no badge for unknown domain", async () => {
    const user = userEvent.setup();
    const onSubmit = vi.fn();

    render(<JobUrlInput onSubmit={onSubmit} isLoading={false} />);

    const input = screen.getByPlaceholderText(/URL을 입력/i);
    await user.type(input, "https://www.example.com/job/12345");

    expect(screen.queryByText("기타")).not.toBeInTheDocument();
  });

  it("shows error for invalid URL", async () => {
    const user = userEvent.setup();
    const onSubmit = vi.fn();

    render(<JobUrlInput onSubmit={onSubmit} isLoading={false} />);

    const input = screen.getByPlaceholderText(/URL을 입력/i);
    await user.type(input, "not-a-url");

    const submitBtn = screen.getByText("분석 시작");
    await user.click(submitBtn);

    expect(screen.getByText(/유효한 URL/i)).toBeInTheDocument();
    expect(onSubmit).not.toHaveBeenCalled();
  });

  it("shows validation error for empty submission", async () => {
    const user = userEvent.setup();
    const onSubmit = vi.fn();

    render(<JobUrlInput onSubmit={onSubmit} isLoading={false} />);

    const submitBtn = screen.getByText("분석 시작");
    await user.click(submitBtn);

    expect(screen.getByText(/URL을 입력/i)).toBeInTheDocument();
    expect(onSubmit).not.toHaveBeenCalled();
  });

  it("shows loading state when submitting", async () => {
    const onSubmit = vi.fn();

    render(<JobUrlInput onSubmit={onSubmit} isLoading={true} />);

    const submitBtn = screen.getByText("분석 중...");
    expect(submitBtn).toBeDisabled();
  });

  it("calls onSubmit with valid URL", async () => {
    const user = userEvent.setup();
    const onSubmit = vi.fn();

    render(<JobUrlInput onSubmit={onSubmit} isLoading={false} />);

    const input = screen.getByPlaceholderText(/URL을 입력/i);
    await user.type(input, "https://www.jobkorea.co.kr/Recruit/GI_Read/12345");

    const submitBtn = screen.getByText("분석 시작");
    await user.click(submitBtn);

    expect(onSubmit).toHaveBeenCalledWith("https://www.jobkorea.co.kr/Recruit/GI_Read/12345");
  });
});
