import type { LucideIcon } from "lucide-react";
import {
  Briefcase,
  Clock3,
  Ellipsis,
  FolderKanban,
  Globe2,
  Heart,
  Users,
} from "lucide-react";

const categoryConfig: Record<
  string,
  { icon: LucideIcon; color: string; iconColor: string; label: string }
> = {
  동아리: {
    icon: Users,
    color: "text-blue-400 bg-blue-500/10 border-blue-500/20",
    iconColor: "text-blue-300",
    label: "동아리",
  },
  인턴: {
    icon: Briefcase,
    color: "text-primary bg-primary/10 border-primary/25",
    iconColor: "text-primary",
    label: "인턴",
  },
  프로젝트: {
    icon: FolderKanban,
    color: "text-violet-400 bg-violet-500/10 border-violet-500/20",
    iconColor: "text-violet-300",
    label: "프로젝트",
  },
  대외활동: {
    icon: Globe2,
    color: "text-orange-400 bg-orange-500/10 border-orange-500/20",
    iconColor: "text-orange-300",
    label: "대외활동",
  },
  아르바이트: {
    icon: Clock3,
    color: "text-yellow-400 bg-yellow-500/10 border-yellow-500/20",
    iconColor: "text-yellow-300",
    label: "아르바이트",
  },
  봉사활동: {
    icon: Heart,
    color: "text-pink-400 bg-pink-500/10 border-pink-500/20",
    iconColor: "text-pink-300",
    label: "봉사활동",
  },
  기타: {
    icon: Ellipsis,
    color: "text-muted-foreground bg-white/[0.06] border-white/10",
    iconColor: "text-muted-foreground",
    label: "기타",
  },
};

interface CategoryIconProps {
  category: string;
  showLabel?: boolean;
}

export function CategoryIcon({ category, showLabel = false }: CategoryIconProps) {
  const config = categoryConfig[category] || categoryConfig["기타"];
  const Icon = config.icon;

  return (
    <span
      className={`inline-flex items-center gap-1.5 rounded-full border px-2.5 py-0.5 text-xs font-medium ${config.color}`}
      data-testid="category-icon"
      data-category={category}
    >
      <span className="inline-flex h-4 w-4 items-center justify-center rounded-sm bg-black/20">
        <Icon className={`h-3 w-3 ${config.iconColor}`} aria-hidden="true" />
      </span>
      {showLabel && <span>{config.label}</span>}
    </span>
  );
}

export function getCategoryConfig(category: string) {
  return categoryConfig[category] || categoryConfig["기타"];
}
