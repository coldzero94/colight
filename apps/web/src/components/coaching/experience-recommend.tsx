"use client";

import { useState } from "react";
import Link from "next/link";
import { ArrowRight, Sparkles, Target } from "lucide-react";
import { RecommendCard } from "./recommend-card";

interface ExperienceWithScore {
  id: string;
  title: string;
  category: string;
  period_start?: string;
  period_end?: string;
  weapons: Array<{ weapon_name: string; relevance_score?: number }>;
  star_situation: string;
  matchScore: number;
}

interface RequiredWeapons {
  primary: {
    weapon_id: string;
    weapon_name: string;
    reason: string;
  };
  secondary: Array<{
    weapon_id: string;
    weapon_name: string;
    reason: string;
  }>;
}

interface ExperienceRecommendProps {
  experiences: ExperienceWithScore[];
  requiredWeapons: RequiredWeapons;
  onSelect: (selectedIds: string[]) => void;
}

export function ExperienceRecommend({
  experiences,
  requiredWeapons,
  onSelect,
}: ExperienceRecommendProps) {
  const [selectedIds, setSelectedIds] = useState<string[]>([]);

  const handleToggle = (id: string) => {
    setSelectedIds((prev) => {
      if (prev.includes(id)) {
        // Deselect
        return prev.filter((selectedId) => selectedId !== id);
      } else {
        // Select (max 3)
        if (prev.length >= 3) {
          return prev;
        }
        return [...prev, id];
      }
    });
  };

  const handleSubmit = () => {
    onSelect(selectedIds);
  };

  // Empty state
  if (experiences.length === 0) {
    return (
      <div className="rounded-lg border border-border bg-muted p-8 text-center">
        <p className="mb-4 text-muted-foreground">등록된 경험이 없습니다</p>
        <Link
          href="/experiences"
          className="inline-block rounded-lg bg-primary px-4 py-2 text-sm text-primary-foreground hover:bg-primary/90"
        >
          경험 등록하기
        </Link>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div>
        <h2 className="flex items-center gap-2 text-2xl font-bold text-foreground">
          <span className="inline-flex h-8 w-8 items-center justify-center rounded-lg bg-blue-500/15 text-blue-300">
            <Target className="h-4 w-4" aria-hidden="true" />
          </span>
          추천 경험 ({experiences.length}건)
        </h2>
        <p className="mt-1 text-sm text-muted-foreground">
          적합도가 높은 경험을 선택하세요 (최대 3개)
        </p>
      </div>

      <div className="space-y-4">
        {experiences.map((exp) => {
          const isSelected = selectedIds.includes(exp.id);
          const isDisabled = !isSelected && selectedIds.length >= 3;

          return (
            <RecommendCard
              key={exp.id}
              experience={exp}
              selected={isSelected}
              disabled={isDisabled}
              onToggle={() => handleToggle(exp.id)}
            />
          );
        })}
      </div>

      <button
        onClick={handleSubmit}
        disabled={selectedIds.length === 0}
        className="group inline-flex w-full items-center justify-center gap-2 rounded-lg bg-primary px-4 py-3 text-sm font-medium text-primary-foreground transition-all hover:bg-primary/90 hover:shadow-[0_10px_24px_rgba(16,185,129,0.25)] disabled:opacity-50"
      >
        <Sparkles className="h-4 w-4" aria-hidden="true" />
        선택한 경험으로 초안 작성 ({selectedIds.length}개 선택됨)
        <ArrowRight className="h-4 w-4 transition-transform group-hover:translate-x-0.5" aria-hidden="true" />
      </button>
    </div>
  );
}
