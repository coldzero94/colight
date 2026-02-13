const categoryConfig: Record<
  string,
  { icon: string; color: string; label: string }
> = {
  동아리: { icon: "👥", color: "text-blue-600 bg-blue-50", label: "동아리" },
  인턴: { icon: "💼", color: "text-green-600 bg-green-50", label: "인턴" },
  프로젝트: {
    icon: "📁",
    color: "text-purple-600 bg-purple-50",
    label: "프로젝트",
  },
  대외활동: {
    icon: "🌐",
    color: "text-orange-600 bg-orange-50",
    label: "대외활동",
  },
  아르바이트: {
    icon: "⏰",
    color: "text-yellow-600 bg-yellow-50",
    label: "아르바이트",
  },
  봉사활동: { icon: "❤️", color: "text-pink-600 bg-pink-50", label: "봉사활동" },
  기타: { icon: "⋯", color: "text-gray-600 bg-gray-50", label: "기타" },
};

interface CategoryIconProps {
  category: string;
  showLabel?: boolean;
}

export function CategoryIcon({ category, showLabel = false }: CategoryIconProps) {
  const config = categoryConfig[category] || categoryConfig["기타"];

  return (
    <span
      className={`inline-flex items-center gap-1 rounded-full px-2 py-0.5 text-xs font-medium ${config.color}`}
      data-testid="category-icon"
      data-category={category}
    >
      <span>{config.icon}</span>
      {showLabel && <span>{config.label}</span>}
    </span>
  );
}

export function getCategoryConfig(category: string) {
  return categoryConfig[category] || categoryConfig["기타"];
}
