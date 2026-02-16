const categoryConfig: Record<
  string,
  { icon: string; color: string; label: string }
> = {
  동아리: { icon: "👥", color: "text-blue-400 bg-blue-500/10", label: "동아리" },
  인턴: { icon: "💼", color: "text-primary bg-primary/10", label: "인턴" },
  프로젝트: {
    icon: "📁",
    color: "text-purple-400 bg-purple-500/10",
    label: "프로젝트",
  },
  대외활동: {
    icon: "🌐",
    color: "text-orange-400 bg-orange-500/10",
    label: "대외활동",
  },
  아르바이트: {
    icon: "⏰",
    color: "text-yellow-400 bg-yellow-500/10",
    label: "아르바이트",
  },
  봉사활동: { icon: "❤️", color: "text-pink-400 bg-pink-500/10", label: "봉사활동" },
  기타: { icon: "⋯", color: "text-muted-foreground bg-white/[0.06]", label: "기타" },
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
