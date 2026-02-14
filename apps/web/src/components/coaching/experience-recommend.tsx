"use client";

import { useState } from "react";
import Link from "next/link";
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
      <div className="rounded-lg border border-gray-200 bg-gray-50 p-8 text-center">
        <p className="mb-4 text-gray-600">등록된 경험이 없습니다</p>
        <Link
          href="/experiences"
          className="inline-block rounded-lg bg-gray-900 px-4 py-2 text-sm text-white hover:bg-gray-800"
        >
          경험 등록하기
        </Link>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div>
        <h2 className="text-2xl font-bold text-gray-900">
          🎯 추천 경험 ({experiences.length}건)
        </h2>
        <p className="mt-1 text-sm text-gray-600">
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
        className="w-full rounded-lg bg-gray-900 px-4 py-3 text-sm font-medium text-white hover:bg-gray-800 disabled:opacity-50"
      >
        🚀 선택한 경험으로 초안 작성 ({selectedIds.length}개 선택됨)
      </button>
    </div>
  );
}
