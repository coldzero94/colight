"use client";

import { useState } from "react";
import { BarChart3, Layout, Target, Tag } from "lucide-react";
import type { AdviceItem } from "@/lib/api/coaching";

const CATEGORY_CONFIG: Record<
  string,
  { label: string; icon: typeof BarChart3 }
> = {
  metric: { label: "수치/데이터 보강", icon: BarChart3 },
  structure: { label: "구조 개선", icon: Layout },
  detail: { label: "구체성 강화", icon: Target },
  keyword: { label: "키워드 활용", icon: Tag },
};

interface AdvicePanelProps {
  advice: AdviceItem[];
}

export function AdvicePanel({ advice }: AdvicePanelProps) {
  const [dismissed, setDismissed] = useState<Set<number>>(new Set());

  if (advice.length === 0) return null;

  const sorted = [...advice].sort((a, b) => a.priority - b.priority);

  const toggleDismiss = (index: number) => {
    setDismissed((prev) => {
      const next = new Set(prev);
      if (next.has(index)) {
        next.delete(index);
      } else {
        next.add(index);
      }
      return next;
    });
  };

  return (
    <div className="brand-surface-soft rounded-2xl p-5">
      <h3 className="mb-3 text-sm font-semibold text-foreground/90">
        개선 포인트
      </h3>
      <ul className="space-y-3">
        {sorted.map((item, idx) => {
          const config = CATEGORY_CONFIG[item.category] ?? {
            label: item.category,
            icon: Target,
          };
          const Icon = config.icon;
          const isDismissed = dismissed.has(idx);

          return (
            <li
              key={idx}
              role="listitem"
              className={`flex items-start gap-3 rounded-xl p-3 transition-colors ${
                isDismissed
                  ? "opacity-50"
                  : "bg-background/50"
              }`}
            >
              <input
                type="checkbox"
                checked={isDismissed}
                onChange={() => toggleDismiss(idx)}
                className="mt-1 h-4 w-4 shrink-0 rounded border-border/50 accent-primary"
              />
              <Icon className="mt-0.5 h-4 w-4 shrink-0 text-primary/70" />
              <div className="min-w-0">
                <span className="text-xs font-medium text-primary/80">
                  {config.label}
                </span>
                <p
                  className={`mt-0.5 text-sm text-muted-foreground ${
                    isDismissed ? "line-through" : ""
                  }`}
                >
                  {item.content}
                </p>
              </div>
            </li>
          );
        })}
      </ul>
    </div>
  );
}
