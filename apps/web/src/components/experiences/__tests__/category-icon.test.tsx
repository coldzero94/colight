import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import { CategoryIcon } from "../category-icon";

describe("CategoryIcon", () => {
  it("renders correct icon and color for each category", () => {
    const categories = [
      { name: "동아리", icon: "👥", color: "text-blue-400" },
      { name: "인턴", icon: "💼", color: "text-primary" },
      { name: "프로젝트", icon: "📁", color: "text-purple-400" },
      { name: "대외활동", icon: "🌐", color: "text-orange-400" },
      { name: "아르바이트", icon: "⏰", color: "text-yellow-400" },
      { name: "봉사활동", icon: "❤️", color: "text-pink-400" },
      { name: "기타", icon: "⋯", color: "text-muted-foreground" },
    ];

    for (const cat of categories) {
      const { unmount } = render(<CategoryIcon category={cat.name} />);
      const el = screen.getByTestId("category-icon");
      expect(el).toHaveTextContent(cat.icon);
      expect(el.className).toContain(cat.color);
      unmount();
    }
  });
});
