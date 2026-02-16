import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import { WeaponBadge } from "../weapon-badge";

describe("WeaponBadge", () => {
  it("renders correct label and icon for each weapon code W01-W07", () => {
    const weapons = [
      { code: "W01", expectedText: "위기극복" },
      { code: "W02", expectedText: "리더십" },
      { code: "W03", expectedText: "팀워크/협업" },
      { code: "W04", expectedText: "도전정신" },
      { code: "W05", expectedText: "문제해결" },
      { code: "W06", expectedText: "소통/설득" },
      { code: "W07", expectedText: "성장/학습" },
    ];

    weapons.forEach((w) => {
      const { container, unmount } = render(<WeaponBadge code={w.code} />);
      expect(screen.getByText(w.expectedText)).toBeInTheDocument();
      expect(screen.getByTestId(`weapon-icon-${w.code}`)).toBeInTheDocument();
      expect(container.querySelector(`[data-weapon-badge="${w.code}"]`)).toBeInTheDocument();
      unmount();
    });
  });

  it("renders primary weapon with emphasis styling", () => {
    render(<WeaponBadge code="W01" isPrimary={true} />);

    const badge = screen.getByTestId("primary-weapon");
    expect(badge).toHaveClass("border-2"); // Thicker border for primary
    expect(screen.getByText("(주)")).toBeInTheDocument(); // Primary label
  });

  it("renders an icon element for each weapon code", () => {
    const codes = ["W01", "W02", "W03", "W04", "W05", "W06", "W07"];

    codes.forEach((code) => {
      const { unmount } = render(<WeaponBadge code={code} />);
      expect(screen.getByTestId(`weapon-icon-${code}`)).toBeInTheDocument();
      unmount();
    });
  });

  it("shows confidence when showConfidence is true", () => {
    render(<WeaponBadge code="W01" showConfidence={true} confidence={0.85} />);
    expect(screen.getByText("85%")).toBeInTheDocument();
  });
});
