import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import TermsPage from "../terms/page";

describe("TermsPage", () => {
  it("renders page title", () => {
    render(<TermsPage />);
    expect(screen.getByText("이용약관")).toBeInTheDocument();
  });

  it("includes service scope section", () => {
    render(<TermsPage />);
    expect(
      screen.getByText(/서비스의 제공/),
    ).toBeInTheDocument();
  });

  it("includes AI disclaimer section", () => {
    render(<TermsPage />);
    expect(
      screen.getByText(/AI 서비스 면책/),
    ).toBeInTheDocument();
  });

  it("includes copyright section", () => {
    render(<TermsPage />);
    expect(screen.getByText(/제7조.*저작권/)).toBeInTheDocument();
  });

  it("includes dispute resolution section", () => {
    render(<TermsPage />);
    expect(screen.getByText(/분쟁 해결/)).toBeInTheDocument();
  });
});
