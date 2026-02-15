import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import { WeaponCoverage } from "../weapon-coverage";

describe("WeaponCoverage", () => {
  const requiredWeapons = {
    primary: { weapon_id: "w1", weapon_name: "프로젝트 리딩", reason: "리더십" },
    secondary: [
      { weapon_id: "w2", weapon_name: "데이터 분석", reason: "분석력" },
      { weapon_id: "w3", weapon_name: "문제 해결", reason: "해결력" },
    ],
  };

  const selectedWeapons = [
    { code: "w1", name: "프로젝트 리딩" },
    { code: "w2", name: "데이터 분석" },
    { code: "w3", name: "문제 해결" },
  ];

  it("shows checkmarks for covered weapons", () => {
    render(
      <WeaponCoverage
        selectedWeapons={selectedWeapons}
        requiredWeapons={requiredWeapons}
      />
    );
    expect(screen.getAllByText("✓").length).toBe(3);
  });

  it("displays primary weapon with label", () => {
    render(
      <WeaponCoverage
        selectedWeapons={selectedWeapons}
        requiredWeapons={requiredWeapons}
      />
    );
    expect(screen.getByText("프로젝트 리딩")).toBeInTheDocument();
    expect(screen.getByText("(주 무기)")).toBeInTheDocument();
  });

  it("displays secondary weapons with labels", () => {
    render(
      <WeaponCoverage
        selectedWeapons={selectedWeapons}
        requiredWeapons={requiredWeapons}
      />
    );
    expect(screen.getByText("데이터 분석")).toBeInTheDocument();
    expect(screen.getByText("문제 해결")).toBeInTheDocument();
    expect(screen.getAllByText("(부 무기)").length).toBe(2);
  });

  it("shows full coverage summary message", () => {
    render(
      <WeaponCoverage
        selectedWeapons={selectedWeapons}
        requiredWeapons={requiredWeapons}
      />
    );
    expect(
      screen.getByText(/필수 무기가 모두 커버되었습니다/)
    ).toBeInTheDocument();
  });

  it("shows uncovered indicators when no weapons selected", () => {
    render(
      <WeaponCoverage
        selectedWeapons={[]}
        requiredWeapons={requiredWeapons}
      />
    );
    // All should show "○" instead of "✓"
    expect(screen.getAllByText("○").length).toBe(3);
    expect(screen.queryByText("✓")).not.toBeInTheDocument();
    expect(
      screen.getByText(/주 무기를 포함한 경험을 선택하세요/)
    ).toBeInTheDocument();
  });

  it("shows partial coverage when only primary is covered", () => {
    render(
      <WeaponCoverage
        selectedWeapons={[{ code: "w1", name: "프로젝트 리딩" }]}
        requiredWeapons={requiredWeapons}
      />
    );
    expect(screen.getAllByText("✓").length).toBe(1);
    expect(screen.getAllByText("○").length).toBe(2);
    expect(
      screen.getByText("부 무기를 추가로 선택하면 더 좋습니다")
    ).toBeInTheDocument();
  });
});
