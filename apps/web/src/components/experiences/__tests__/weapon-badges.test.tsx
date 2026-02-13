import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import { WeaponBadges } from "../weapon-badges";
import type { ExperienceWeapon } from "@/lib/api/experiences";

const mockWeapons: ExperienceWeapon[] = [
  {
    id: "1",
    weapon_code: "W01",
    confidence: 0.9,
    is_primary: true,
    reasoning: "",
    user_confirmed: false,
    user_modified: false,
  },
  {
    id: "2",
    weapon_code: "W03",
    confidence: 0.7,
    is_primary: false,
    reasoning: "",
    user_confirmed: false,
    user_modified: false,
  },
];

describe("WeaponBadges", () => {
  it("renders primary weapon with emphasis", () => {
    render(<WeaponBadges weapons={mockWeapons} />);

    const primaryBadge = screen.getByTestId("primary-weapon");
    expect(primaryBadge).toBeInTheDocument();
    expect(primaryBadge).toHaveTextContent("위기극복");
    expect(primaryBadge.className).toContain("border-2"); // Thicker border for primary
    expect(screen.getByText("(주)")).toBeInTheDocument(); // Primary label
  });

  it("renders secondary weapons as smaller badges", () => {
    render(<WeaponBadges weapons={mockWeapons} />);

    const secondaryBadges = screen.getAllByTestId("secondary-weapon");
    expect(secondaryBadges).toHaveLength(1);
    expect(secondaryBadges[0]).toHaveTextContent("팀워크/협업");
    expect(secondaryBadges[0].className).toContain("border"); // Regular border for secondary
  });
});
