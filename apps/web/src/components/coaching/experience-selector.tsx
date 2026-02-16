"use client";

import { useMemo, useState } from "react";
import { SelectableExpCard } from "./selectable-exp-card";
import { WeaponCoverage } from "./weapon-coverage";

interface Experience {
  id: string;
  title: string;
  category: string;
  period_start?: string;
  period_end?: string;
  star_situation: string;
  star_task: string;
  star_action: string;
  star_result: string;
  weapons: Array<{ code: string; name: string }>;
  matchScore: number;
  matchReasons?: string[];
  isUsed?: boolean;
  keywordMatches?: string[];
}

interface RequiredWeapons {
  primary: { weapon_id: string; weapon_name: string; reason: string };
  secondary: Array<{ weapon_id: string; weapon_name: string; reason: string }>;
}

interface ExperienceSelectorProps {
  experiences: Experience[];
  maxSelect: number;
  requiredWeapons: RequiredWeapons;
  onConfirm: (selectedIds: string[]) => void;
}

export function ExperienceSelector({
  experiences,
  maxSelect,
  requiredWeapons,
  onConfirm,
}: ExperienceSelectorProps) {
  const [selectedIds, setSelectedIds] = useState<string[]>([]);
  const [expandedId, setExpandedId] = useState<string | null>(null);

  // Sort: unused first (by score desc), then used (by score desc)
  const sortedExperiences = useMemo(() => {
    return [...experiences].sort((a, b) => {
      const aUsed = a.isUsed ? 1 : 0;
      const bUsed = b.isUsed ? 1 : 0;
      if (aUsed !== bUsed) return aUsed - bUsed;
      return b.matchScore - a.matchScore;
    });
  }, [experiences]);

  const handleToggle = (id: string) => {
    setSelectedIds((prev) => {
      if (prev.includes(id)) {
        return prev.filter((selectedId) => selectedId !== id);
      } else {
        if (prev.length >= maxSelect) {
          return prev;
        }
        return [...prev, id];
      }
    });
  };

  const handleExpand = (id: string) => {
    setExpandedId(expandedId === id ? null : id);
  };

  const handleConfirm = () => {
    if (selectedIds.length === 0) return;
    onConfirm(selectedIds);
  };

  // Get selected weapons for coverage display
  const selectedWeapons = experiences
    .filter((exp) => selectedIds.includes(exp.id))
    .flatMap((exp) => exp.weapons);

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h2 className="text-2xl font-bold text-foreground">
          경험 선택 ({selectedIds.length}/{maxSelect}개 선택됨)
        </h2>
      </div>

      {/* Weapon coverage - always show */}
      <WeaponCoverage
        selectedWeapons={selectedWeapons}
        requiredWeapons={requiredWeapons}
      />

      {/* Experience cards */}
      <div className="space-y-4">
        {sortedExperiences.map((exp) => {
          const isSelected = selectedIds.includes(exp.id);
          const isDisabled = !isSelected && selectedIds.length >= maxSelect;
          const isExpanded = expandedId === exp.id;

          return (
            <SelectableExpCard
              key={exp.id}
              experience={exp}
              selected={isSelected}
              disabled={isDisabled}
              expanded={isExpanded}
              onToggle={() => handleToggle(exp.id)}
              onExpand={() => handleExpand(exp.id)}
            />
          );
        })}
      </div>

      {/* Confirm button */}
      <button
        onClick={handleConfirm}
        disabled={selectedIds.length === 0}
        className="w-full rounded-lg bg-primary px-4 py-3 text-sm font-medium text-primary-foreground hover:bg-primary/90 disabled:opacity-50"
      >
        선택한 경험으로 초안 작성 시작
      </button>
    </div>
  );
}
