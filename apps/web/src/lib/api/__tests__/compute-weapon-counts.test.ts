import { describe, it, expect } from "vitest";

// Import only the pure function, mock the api-client dependency
vi.mock("@/lib/api-client", () => ({
  apiClient: { get: vi.fn(), post: vi.fn(), patch: vi.fn(), delete: vi.fn() },
}));

import { computeWeaponCounts, type Experience } from "../experiences";

function makeExperience(
  weapons: Array<{ weapon_code: string }>
): Experience {
  return {
    id: "exp-1",
    user_id: "u-1",
    title: "test",
    category: "인턴",
    role: "",
    content: "",
    result: "",
    star_situation: "",
    star_task: "",
    star_action: "",
    star_result: "",
    keywords: null,
    source: "",
    is_archived: false,
    created_at: "",
    updated_at: "",
    weapons: weapons.map((w, i) => ({
      id: `w-${i}`,
      weapon_code: w.weapon_code,
      confidence: 0.9,
      is_primary: i === 0,
      reasoning: "",
      user_confirmed: false,
      user_modified: false,
    })),
  };
}

describe("computeWeaponCounts", () => {
  it("returns empty object for empty array", () => {
    expect(computeWeaponCounts([])).toEqual({});
  });

  it("counts weapons from experiences", () => {
    const experiences = [
      makeExperience([{ weapon_code: "W01" }, { weapon_code: "W02" }]),
      makeExperience([{ weapon_code: "W01" }, { weapon_code: "W03" }]),
    ];

    const counts = computeWeaponCounts(experiences);
    expect(counts).toEqual({ W01: 2, W02: 1, W03: 1 });
  });

  it("deduplicates weapons within same experience", () => {
    const experiences = [
      makeExperience([
        { weapon_code: "W01_primary" },
        { weapon_code: "W01_secondary" },
      ]),
    ];

    const counts = computeWeaponCounts(experiences);
    expect(counts).toEqual({ W01: 1 });
  });

  it("skips experiences without weapons", () => {
    const exp: Experience = {
      id: "1",
      user_id: "u-1",
      title: "test",
      category: "인턴",
      role: "",
      content: "",
      result: "",
      star_situation: "",
      star_task: "",
      star_action: "",
      star_result: "",
      keywords: null,
      source: "",
      is_archived: false,
      created_at: "",
      updated_at: "",
    };

    expect(computeWeaponCounts([exp])).toEqual({});
  });

  it("uses first 3 chars of weapon_code", () => {
    const experiences = [
      makeExperience([{ weapon_code: "W05_extra_info" }]),
    ];

    const counts = computeWeaponCounts(experiences);
    expect(counts).toEqual({ W05: 1 });
  });
});
