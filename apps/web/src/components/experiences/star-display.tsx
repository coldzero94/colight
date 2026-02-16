import type { LucideIcon } from "lucide-react";
import { BarChart3, Search, Target, Zap } from "lucide-react";

interface StarDisplayProps {
  situation?: string;
  task?: string;
  action?: string;
  result?: string;
}

const sections = [
  {
    key: "situation",
    label: "Situation",
    icon: Search,
    description: "상황/배경",
    iconColor: "text-blue-300",
  },
  {
    key: "task",
    label: "Task",
    icon: Target,
    description: "과제/목표",
    iconColor: "text-amber-300",
  },
  {
    key: "action",
    label: "Action",
    icon: Zap,
    description: "구체적 행동",
    iconColor: "text-emerald-300",
  },
  {
    key: "result",
    label: "Result",
    icon: BarChart3,
    description: "결과/성과",
    iconColor: "text-violet-300",
  },
] as const;

export function StarDisplay({ situation, task, action, result }: StarDisplayProps) {
  const values = { situation, task, action, result };

  return (
    <div className="space-y-4" data-testid="star-display">
      {sections.map((section) => {
        const content = values[section.key];
        const Icon = section.icon as LucideIcon;
        return (
          <div
            key={section.key}
            className="brand-surface-soft rounded-xl p-4"
          >
            <h4 className="mb-2 flex items-center gap-1.5 text-sm font-semibold text-foreground">
              <span className="inline-flex h-5 w-5 items-center justify-center rounded-md border border-white/10 bg-white/[0.06]">
                <Icon className={`h-3.5 w-3.5 ${section.iconColor}`} aria-hidden="true" />
              </span>
              <span>{section.label}</span>
              <span className="text-xs font-normal text-muted-foreground/60">
                {section.description}
              </span>
            </h4>
            {content ? (
              <p className="text-sm leading-relaxed text-foreground/80 whitespace-pre-wrap">
                {content}
              </p>
            ) : (
              <p className="text-sm text-muted-foreground/60" data-testid="star-empty">
                아직 작성되지 않았습니다
              </p>
            )}
          </div>
        );
      })}
    </div>
  );
}
