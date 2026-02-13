import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { WeaponFilterTabs } from "../weapon-filter-tabs";

const mockCounts: Record<string, number> = {
  W01: 3,
  W02: 2,
  W03: 5,
  W04: 1,
  W05: 4,
  W06: 0,
  W07: 2,
};

describe("WeaponFilterTabs", () => {
  const onChange = vi.fn();

  it("renders all 8 tabs with correct counts", () => {
    render(
      <WeaponFilterTabs
        counts={mockCounts}
        activeWeapon={null}
        onChange={onChange}
      />
    );

    // Total tab
    expect(screen.getByText("전체 (17)")).toBeInTheDocument();

    // 7 weapon tabs
    expect(screen.getByText(/위기극복 \(3\)/)).toBeInTheDocument();
    expect(screen.getByText(/리더십 \(2\)/)).toBeInTheDocument();
    expect(screen.getByText(/팀워크\/협업 \(5\)/)).toBeInTheDocument();
    expect(screen.getByText(/도전정신 \(1\)/)).toBeInTheDocument();
    expect(screen.getByText(/문제해결 \(4\)/)).toBeInTheDocument();
    expect(screen.getByText(/소통\/설득 \(0\)/)).toBeInTheDocument();
    expect(screen.getByText(/성장\/학습 \(2\)/)).toBeInTheDocument();
  });

  it("highlights active weapon tab", () => {
    render(
      <WeaponFilterTabs
        counts={mockCounts}
        activeWeapon="W01"
        onChange={onChange}
      />
    );

    const tabs = screen.getAllByRole("tab");
    // "전체" tab should not be active
    expect(tabs[0]).toHaveAttribute("aria-selected", "false");
    // W01 tab should be active
    expect(tabs[1]).toHaveAttribute("aria-selected", "true");
  });

  it("calls onChange with weapon code on tab click", async () => {
    const user = userEvent.setup();
    render(
      <WeaponFilterTabs
        counts={mockCounts}
        activeWeapon={null}
        onChange={onChange}
      />
    );

    const tabs = screen.getAllByRole("tab");
    await user.click(tabs[2]); // W02 tab

    expect(onChange).toHaveBeenCalledWith("W02");
  });

  it("calls onChange with null when clicking 전체 tab", async () => {
    const user = userEvent.setup();
    render(
      <WeaponFilterTabs
        counts={mockCounts}
        activeWeapon="W01"
        onChange={onChange}
      />
    );

    const allTab = screen.getByText("전체 (17)");
    await user.click(allTab);

    expect(onChange).toHaveBeenCalledWith(null);
  });
});
