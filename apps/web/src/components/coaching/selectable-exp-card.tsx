"use client";

import { ChevronDown, ChevronUp } from "lucide-react";

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

interface SelectableExpCardProps {
  experience: Experience;
  selected: boolean;
  disabled: boolean;
  expanded: boolean;
  onToggle: () => void;
  onExpand: () => void;
}

function getScoreColor(score: number): string {
  if (score >= 70) return "text-green-400";
  if (score >= 40) return "text-blue-400";
  return "text-muted-foreground";
}

export function SelectableExpCard({
  experience,
  selected,
  disabled,
  expanded,
  onToggle,
  onExpand,
}: SelectableExpCardProps) {
  const formatPeriod = () => {
    if (!experience.period_start) return "";
    if (!experience.period_end) return experience.period_start;
    return `${experience.period_start} ~ ${experience.period_end}`;
  };

  const scoreColor = getScoreColor(experience.matchScore);

  return (
    <div
      className={`rounded-lg border p-4 transition-colors ${
        selected
          ? "border-blue-500 bg-blue-500/10"
          : "border-border bg-card hover:border-border/80"
      } ${disabled ? "opacity-50" : ""}`}
    >
      <div className="flex items-start gap-3">
        <input
          type="checkbox"
          checked={selected}
          disabled={disabled}
          onChange={onToggle}
          className="mt-1 h-5 w-5 rounded border-border text-blue-400 focus:ring-2 focus:ring-blue-500"
        />

        <div className="flex-1">
          <div className="flex items-start justify-between">
            <div className="flex-1">
              <div className="flex items-center gap-2">
                <h3 className="text-base font-semibold text-foreground">
                  {experience.title}
                </h3>
                {experience.isUsed && (
                  <span
                    data-testid="used-badge"
                    className="rounded-full bg-yellow-500/10 px-2 py-0.5 text-xs font-medium text-yellow-400"
                  >
                    이미 사용됨
                  </span>
                )}
              </div>
              {formatPeriod() && (
                <p className="mt-1 text-xs text-muted-foreground">{formatPeriod()}</p>
              )}
            </div>
            <div className="ml-4 text-right">
              <div data-testid="match-score" className={`text-lg font-bold ${scoreColor}`}>
                {experience.matchScore}%
              </div>
              <button
                type="button"
                onClick={onExpand}
                data-testid="expand-button"
                className="mt-1 flex items-center gap-1 text-xs text-muted-foreground hover:text-foreground/80"
              >
                {expanded ? (
                  <>
                    <span>접기</span>
                    <ChevronUp className="h-3 w-3" />
                  </>
                ) : (
                  <>
                    <span>자세히</span>
                    <ChevronDown className="h-3 w-3" />
                  </>
                )}
              </button>
            </div>
          </div>

          {/* Weapons */}
          <div className="mt-2 flex flex-wrap gap-1">
            {experience.weapons.map((weapon, index) => (
              <span
                key={index}
                className="rounded-full bg-white/[0.06] px-2 py-1 text-xs font-medium text-foreground/80"
              >
                {weapon.name}
              </span>
            ))}
          </div>

          {/* Match reasons */}
          {experience.matchReasons && experience.matchReasons.length > 0 && (
            <div data-testid="match-reasons" className="mt-2 space-y-0.5">
              {experience.matchReasons.map((reason, index) => (
                <p key={index} className="text-xs text-muted-foreground">
                  · {reason}
                </p>
              ))}
            </div>
          )}

          {/* Preview or Full STAR */}
          {!expanded ? (
            <p className="mt-3 line-clamp-2 text-sm text-muted-foreground">
              {experience.star_situation}
            </p>
          ) : (
            <div className="mt-4 space-y-3 rounded-lg bg-white/[0.02] p-4 text-sm">
              <div>
                <span className="font-semibold text-blue-400">[상황]</span>
                <p className="mt-1 text-foreground/80">{experience.star_situation}</p>
              </div>
              <div>
                <span className="font-semibold text-yellow-400">[과제]</span>
                <p className="mt-1 text-foreground/80">{experience.star_task}</p>
              </div>
              <div>
                <span className="font-semibold text-green-400">[행동]</span>
                <p className="mt-1 text-foreground/80">{experience.star_action}</p>
              </div>
              <div>
                <span className="font-semibold text-purple-400">[결과]</span>
                <p className="mt-1 text-foreground/80">{experience.star_result}</p>
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
