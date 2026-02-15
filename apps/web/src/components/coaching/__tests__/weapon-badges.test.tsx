import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import { WeaponBadges } from "../weapon-badges";

describe("WeaponBadges", () => {
  const primaryWeapon = {
    weapon_id: "w1",
    weapon_name: "프로젝트 리딩",
    reason: "팀 리더십 경험이 풍부함",
  };

  const secondaryWeapons = [
    {
      weapon_id: "w2",
      weapon_name: "데이터 분석",
      reason: "SQL 및 Python 활용 능력",
    },
    {
      weapon_id: "w3",
      weapon_name: "문제 해결",
      reason: "복잡한 기술 이슈 해결 경험",
    },
  ];

  it("renders primary weapon with trophy icon", () => {
    render(
      <WeaponBadges primary={primaryWeapon} secondary={secondaryWeapons} />
    );
    expect(screen.getByText("🏆")).toBeInTheDocument();
    expect(screen.getByText("프로젝트 리딩")).toBeInTheDocument();
  });

  it("renders secondary weapons", () => {
    render(
      <WeaponBadges primary={primaryWeapon} secondary={secondaryWeapons} />
    );
    expect(screen.getByText("데이터 분석")).toBeInTheDocument();
    expect(screen.getByText("문제 해결")).toBeInTheDocument();
  });

  it("displays reasons for weapons", () => {
    render(
      <WeaponBadges primary={primaryWeapon} secondary={secondaryWeapons} />
    );
    expect(
      screen.getByText(/팀 리더십 경험이 풍부함/)
    ).toBeInTheDocument();
    expect(
      screen.getByText(/SQL 및 Python 활용 능력/)
    ).toBeInTheDocument();
    expect(
      screen.getByText(/복잡한 기술 이슈 해결 경험/)
    ).toBeInTheDocument();
  });

  it("handles empty secondary weapons", () => {
    render(<WeaponBadges primary={primaryWeapon} secondary={[]} />);
    expect(screen.getByText("프로젝트 리딩")).toBeInTheDocument();
    expect(screen.queryByText("데이터 분석")).not.toBeInTheDocument();
  });

  it("distinguishes primary from secondary visually", () => {
    const { container } = render(
      <WeaponBadges primary={primaryWeapon} secondary={secondaryWeapons} />
    );
    // Primary weapon has data-primary="true" attribute
    const primaryBadge = container.querySelector("[data-primary='true']");
    expect(primaryBadge).toBeInTheDocument();
    expect(primaryBadge).toHaveClass("bg-amber-100");
  });
});
