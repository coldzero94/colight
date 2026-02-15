import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

vi.mock("next/dynamic", () => ({
  default: (_loader: any, _opts?: any) => {
    // Return a passthrough component that renders children
    const Mock = (props: any) => (
      <div data-testid="recharts-mock">
        {props.children}
      </div>
    );
    Mock.displayName = "DynamicMock";
    return Mock;
  },
}));

import { WeaponRadarChart } from "../weapon-radar-chart";

describe("WeaponRadarChart", () => {
  const fullCounts: Record<string, number> = {
    W01: 3,
    W02: 2,
    W03: 5,
    W04: 1,
    W05: 4,
    W06: 0,
    W07: 2,
  };

  it("renders chart title", () => {
    render(<WeaponRadarChart counts={fullCounts} />);
    expect(screen.getByText("무기 분포")).toBeInTheDocument();
  });

  it("shows recommendation for missing weapons", () => {
    render(<WeaponRadarChart counts={fullCounts} />);
    expect(screen.getByText(/보완 추천/)).toBeInTheDocument();
    expect(screen.getByText(/소통\/설득/)).toBeInTheDocument();
  });

  it("shows no recommendation when all weapons present", () => {
    const allPresent = {
      W01: 1, W02: 1, W03: 1, W04: 1, W05: 1, W06: 1, W07: 1,
    };
    render(<WeaponRadarChart counts={allPresent} />);
    expect(screen.queryByText(/보완 추천/)).not.toBeInTheDocument();
  });

  it("lists multiple missing weapons", () => {
    const partial = { W01: 2, W03: 1 };
    render(<WeaponRadarChart counts={partial} />);
    const text = screen.getByText(/보완 추천/).textContent!;
    expect(text).toContain("리더십");
    expect(text).toContain("도전정신");
    expect(text).toContain("문제해결");
    expect(text).toContain("소통/설득");
    expect(text).toContain("성장/학습");
  });

  it("handles all-zero counts", () => {
    render(<WeaponRadarChart counts={{}} />);
    expect(screen.getByText(/보완 추천/)).toBeInTheDocument();
    const text = screen.getByText(/보완 추천/).textContent!;
    expect(text).toContain("위기극복");
    expect(text).toContain("성장/학습");
  });

  it("collapses and expands on click", async () => {
    const user = userEvent.setup();
    render(<WeaponRadarChart counts={fullCounts} />);

    // Initially open — recommendation visible
    expect(screen.getByText(/보완 추천/)).toBeInTheDocument();

    // Collapse
    await user.click(screen.getByText("무기 분포"));
    expect(screen.queryByText(/보완 추천/)).not.toBeInTheDocument();

    // Expand
    await user.click(screen.getByText("무기 분포"));
    expect(screen.getByText(/보완 추천/)).toBeInTheDocument();
  });
});
