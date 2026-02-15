import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { CompanySelect } from "../company-select";

const mockCompanies = [
  {
    id: "company-1",
    company_name: "삼성전자",
    position: "소프트웨어 엔지니어",
    created_at: "2024-01-01T00:00:00Z",
  },
  {
    id: "company-2",
    company_name: "네이버",
    position: "백엔드 개발자",
    created_at: "2024-01-02T00:00:00Z",
  },
  {
    id: "company-3",
    company_name: "카카오",
    position: "프론트엔드 개발자",
    created_at: "2024-01-03T00:00:00Z",
  },
];

describe("CompanySelect", () => {
  it("renders select with company options", () => {
    render(
      <CompanySelect
        companies={mockCompanies}
        value=""
        onChange={vi.fn()}
      />
    );

    expect(screen.getByRole("combobox")).toBeInTheDocument();
    expect(screen.getByText("기업을 선택하세요")).toBeInTheDocument();
    expect(screen.getByText("삼성전자 - 소프트웨어 엔지니어")).toBeInTheDocument();
    expect(screen.getByText("네이버 - 백엔드 개발자")).toBeInTheDocument();
    expect(screen.getByText("카카오 - 프론트엔드 개발자")).toBeInTheDocument();
  });

  it("calls onChange when selection changes", async () => {
    const user = userEvent.setup();
    const onChange = vi.fn();

    render(
      <CompanySelect
        companies={mockCompanies}
        value=""
        onChange={onChange}
      />
    );

    const select = screen.getByRole("combobox");
    await user.selectOptions(select, "company-1");

    expect(onChange).toHaveBeenCalledWith("company-1");
  });

  it("displays selected value", () => {
    render(
      <CompanySelect
        companies={mockCompanies}
        value="company-2"
        onChange={vi.fn()}
      />
    );

    const select = screen.getByRole("combobox") as HTMLSelectElement;
    expect(select.value).toBe("company-2");
  });

  it("displays error message when provided", () => {
    render(
      <CompanySelect
        companies={mockCompanies}
        value=""
        onChange={vi.fn()}
        error="기업을 선택해주세요"
      />
    );

    const errorAlert = screen.getByRole("alert");
    expect(errorAlert).toBeInTheDocument();
    expect(errorAlert).toHaveTextContent("기업을 선택해주세요");
  });

  it("does not display error when error prop is not provided", () => {
    render(
      <CompanySelect
        companies={mockCompanies}
        value=""
        onChange={vi.fn()}
      />
    );

    expect(screen.queryByRole("alert")).not.toBeInTheDocument();
  });

  it("renders empty select with only placeholder when no companies", () => {
    render(
      <CompanySelect companies={[]} value="" onChange={vi.fn()} />
    );

    expect(screen.getByRole("combobox")).toBeInTheDocument();
    expect(screen.getByText("기업을 선택하세요")).toBeInTheDocument();
  });

  it("renders label for accessibility", () => {
    render(
      <CompanySelect
        companies={mockCompanies}
        value=""
        onChange={vi.fn()}
      />
    );

    expect(screen.getByLabelText("기업 선택")).toBeInTheDocument();
  });
});
