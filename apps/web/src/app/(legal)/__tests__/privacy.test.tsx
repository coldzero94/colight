import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import PrivacyPage from "../privacy/page";

describe("PrivacyPage", () => {
  it("renders page title", () => {
    render(<PrivacyPage />);
    expect(screen.getByText("개인정보처리방침")).toBeInTheDocument();
  });

  it("includes processing purpose section", () => {
    render(<PrivacyPage />);
    expect(
      screen.getByText(/개인정보의 처리 목적/),
    ).toBeInTheDocument();
  });

  it("includes collected items section", () => {
    render(<PrivacyPage />);
    expect(
      screen.getByText(/수집하는 개인정보 항목/),
    ).toBeInTheDocument();
  });

  it("includes third-party provision section", () => {
    render(<PrivacyPage />);
    expect(
      screen.getByText(/개인정보의 제3자 제공/),
    ).toBeInTheDocument();
  });

  it("includes user rights section", () => {
    render(<PrivacyPage />);
    expect(
      screen.getByText(/이용자의 권리와 행사 방법/),
    ).toBeInTheDocument();
  });
});
