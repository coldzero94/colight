import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import { CategoryIcon } from "../category-icon";

describe("CategoryIcon", () => {
  it("renders correct icon and color for each category", () => {
    const categories = [
      { name: "동아리", icon: "👥", color: "text-blue-600" },
      { name: "인턴", icon: "💼", color: "text-green-600" },
      { name: "프로젝트", icon: "📁", color: "text-purple-600" },
      { name: "대외활동", icon: "🌐", color: "text-orange-600" },
      { name: "아르바이트", icon: "⏰", color: "text-yellow-600" },
      { name: "봉사활동", icon: "❤️", color: "text-pink-600" },
      { name: "기타", icon: "⋯", color: "text-gray-600" },
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
