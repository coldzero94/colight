"use client";

import { useState } from "react";
import { ChevronDown, ChevronUp, Briefcase } from "lucide-react";
import Link from "next/link";
import { useExperiences } from "@/hooks/use-experiences";

interface ExperiencePickerProps {
  selectedIds: string[];
  onSelectionChange: (ids: string[]) => void;
  maxSelect?: number;
}

export function ExperiencePicker({
  selectedIds,
  onSelectionChange,
  maxSelect = 3,
}: ExperiencePickerProps) {
  const [isOpen, setIsOpen] = useState(false);
  const [expandedId, setExpandedId] = useState<string | null>(null);
  const { data: experiences, isLoading } = useExperiences();

  const handleToggle = (id: string) => {
    if (selectedIds.includes(id)) {
      onSelectionChange(selectedIds.filter((sid) => sid !== id));
    } else if (selectedIds.length < maxSelect) {
      onSelectionChange([...selectedIds, id]);
    }
  };

  const list = experiences ?? [];

  return (
    <div className="space-y-2">
      <button
        type="button"
        onClick={() => setIsOpen(!isOpen)}
        className="flex w-full items-center justify-between rounded-xl border border-border/50 px-4 py-3 text-left transition-colors hover:border-border/80"
      >
        <div className="flex items-center gap-2">
          <Briefcase className="h-4 w-4 text-muted-foreground" />
          <span className="text-sm font-medium text-foreground/80">
            경험 선택{" "}
            <span className="text-muted-foreground">(선택사항)</span>
          </span>
          {selectedIds.length > 0 && (
            <span className="rounded-full bg-primary/20 px-2 py-0.5 text-xs font-medium text-primary">
              {selectedIds.length}/{maxSelect}
            </span>
          )}
        </div>
        {isOpen ? (
          <ChevronUp className="h-4 w-4 text-muted-foreground" />
        ) : (
          <ChevronDown className="h-4 w-4 text-muted-foreground" />
        )}
      </button>

      {isOpen && (
        <div className="space-y-2 rounded-xl border border-border/30 p-3">
          {isLoading ? (
            <p className="py-4 text-center text-sm text-muted-foreground">
              경험 불러오는 중...
            </p>
          ) : list.length === 0 ? (
            <div className="py-4 text-center">
              <p className="text-sm text-muted-foreground">
                등록된 경험이 없습니다
              </p>
              <Link
                href="/experiences"
                className="mt-2 inline-block text-sm text-primary hover:underline"
              >
                경험 등록하러 가기
              </Link>
            </div>
          ) : (
            <>
              <p className="text-xs text-muted-foreground">
                AI가 선택한 경험을 반영해 맞춤 분석을 제공합니다 (최대{" "}
                {maxSelect}개)
              </p>
              <div className="max-h-80 space-y-2 overflow-y-auto">
                {list
                  .filter((exp) => !exp.is_archived)
                  .map((exp) => {
                    const isSelected = selectedIds.includes(exp.id);
                    const isDisabled =
                      !isSelected && selectedIds.length >= maxSelect;
                    const isExpanded = expandedId === exp.id;

                    return (
                      <div
                        key={exp.id}
                        className={`rounded-lg border p-3 transition-colors ${
                          isSelected
                            ? "border-blue-500 bg-blue-500/10"
                            : "border-border/50 bg-card"
                        } ${isDisabled ? "opacity-50" : ""}`}
                      >
                        <div className="flex items-start gap-3">
                          <input
                            type="checkbox"
                            checked={isSelected}
                            disabled={isDisabled}
                            onChange={() => handleToggle(exp.id)}
                            className="mt-0.5 h-4 w-4 rounded border-border text-blue-400 focus:ring-2 focus:ring-blue-500"
                          />
                          <div className="min-w-0 flex-1">
                            <div className="flex items-center gap-2">
                              <h4 className="truncate text-sm font-medium text-foreground">
                                {exp.title}
                              </h4>
                              <span className="shrink-0 text-xs text-muted-foreground">
                                {exp.category}
                              </span>
                            </div>

                            {/* Weapon badges */}
                            {exp.weapons && exp.weapons.length > 0 && (
                              <div className="mt-1.5 flex flex-wrap gap-1">
                                {exp.weapons.map((w) => (
                                  <span
                                    key={w.id}
                                    className="rounded-full bg-white/[0.06] px-2 py-0.5 text-xs text-foreground/70"
                                  >
                                    {w.weapon_code}
                                  </span>
                                ))}
                              </div>
                            )}

                            {/* Situation preview */}
                            {exp.star_situation && (
                              <p
                                className={`mt-1.5 text-xs text-muted-foreground ${
                                  isExpanded ? "" : "line-clamp-1"
                                }`}
                              >
                                {exp.star_situation}
                              </p>
                            )}

                            {isExpanded && (
                              <div className="mt-2 space-y-1.5 text-xs">
                                {exp.star_task && (
                                  <p className="text-foreground/70">
                                    <span className="font-medium text-yellow-400">
                                      [과제]
                                    </span>{" "}
                                    {exp.star_task}
                                  </p>
                                )}
                                {exp.star_action && (
                                  <p className="text-foreground/70">
                                    <span className="font-medium text-green-400">
                                      [행동]
                                    </span>{" "}
                                    {exp.star_action}
                                  </p>
                                )}
                                {exp.star_result && (
                                  <p className="text-foreground/70">
                                    <span className="font-medium text-purple-400">
                                      [결과]
                                    </span>{" "}
                                    {exp.star_result}
                                  </p>
                                )}
                              </div>
                            )}

                            <button
                              type="button"
                              onClick={() =>
                                setExpandedId(isExpanded ? null : exp.id)
                              }
                              className="mt-1 text-xs text-muted-foreground hover:text-foreground/80"
                            >
                              {isExpanded ? "접기" : "자세히"}
                            </button>
                          </div>
                        </div>
                      </div>
                    );
                  })}
              </div>
            </>
          )}
        </div>
      )}
    </div>
  );
}
