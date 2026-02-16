import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import { CategoryIcon } from "../category-icon";

describe("CategoryIcon", () => {
  it("renders correct icon and color for each category", () => {
    const categories = [
      { name: "동아리", color: "text-blue-400" },
      { name: "인턴", color: "text-primary" },
      { name: "프로젝트", color: "text-violet-400" },
      { name: "대외활동", color: "text-orange-400" },
      { name: "아르바이트", color: "text-yellow-400" },
      { name: "봉사활동", color: "text-pink-400" },
      { name: "기타", color: "text-muted-foreground" },
    ];

    for (const cat of categories) {
      const { unmount } = render(<CategoryIcon category={cat.name} />);
      const el = screen.getByTestId("category-icon");
      expect(el.className).toContain(cat.color);
      expect(el.querySelector("svg")).toBeInTheDocument();
      unmount();
    }
  });
});
