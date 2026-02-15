import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import { InterviewTimer } from "../interview-timer";

describe("InterviewTimer", () => {
  it("formats 0 seconds as 00:00", () => {
    render(<InterviewTimer seconds={0} />);
    expect(screen.getByText("00:00")).toBeInTheDocument();
  });

  it("formats 65 seconds as 01:05", () => {
    render(<InterviewTimer seconds={65} />);
    expect(screen.getByText("01:05")).toBeInTheDocument();
  });

  it("formats 600 seconds as 10:00", () => {
    render(<InterviewTimer seconds={600} />);
    expect(screen.getByText("10:00")).toBeInTheDocument();
  });

  it("formats 125 seconds as 02:05", () => {
    render(<InterviewTimer seconds={125} />);
    expect(screen.getByText("02:05")).toBeInTheDocument();
  });
});
