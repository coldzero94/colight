import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import { WeaponBadge } from "../weapon-badge";

describe("WeaponBadge", () => {
  it("renders correct color for each weapon code W01-W07", () => {
    const weapons = [
      { code: "W01", expectedText: "위기극복", icon: "🛡️" },
      { code: "W02", expectedText: "리더십", icon: "👑" },
      { code: "W03", expectedText: "팀워크/협업", icon: "🤝" },
      { code: "W04", expectedText: "도전정신", icon: "🚀" },
      { code: "W05", expectedText: "문제해결", icon: "🔧" },
      { code: "W06", expectedText: "소통/설득", icon: "💬" },
      { code: "W07", expectedText: "성장/학습", icon: "📈" },
    ];

    weapons.forEach((w) => {
      const { container } = render(<WeaponBadge code={w.code} />);
      expect(screen.getByText(w.expectedText)).toBeInTheDocument();
      expect(screen.getByText(w.icon)).toBeInTheDocument();
      expect(container.querySelector(`[data-weapon-badge="${w.code}"]`)).toBeInTheDocument();
    });
  });

  it("renders primary weapon with emphasis styling", () => {
    render(<WeaponBadge code="W01" isPrimary={true} />);

    const badge = screen.getByTestId("primary-weapon");
    expect(badge).toHaveClass("border-2"); // Thicker border for primary
    expect(screen.getByText("(주)")).toBeInTheDocument(); // Primary label
  });

  it("renders correct icon for each weapon code", () => {
    const iconMap = {
      W01: "🛡️",
      W02: "👑",
      W03: "🤝",
      W04: "🚀",
      W05: "🔧",
      W06: "💬",
      W07: "📈",
    };

    Object.entries(iconMap).forEach(([code, icon]) => {
      render(<WeaponBadge code={code} />);
      expect(screen.getAllByText(icon).length).toBeGreaterThan(0);
    });
  });

  it("shows confidence when showConfidence is true", () => {
    render(<WeaponBadge code="W01" showConfidence={true} confidence={0.85} />);
    expect(screen.getByText("85%")).toBeInTheDocument();
  });
});
