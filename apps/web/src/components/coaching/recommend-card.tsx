"use client";

interface ExperienceWithScore {
  id: string;
  title: string;
  category: string;
  period_start?: string;
  period_end?: string;
  weapons: Array<{ weapon_name: string }>;
  star_situation: string;
  matchScore: number;
}

interface RecommendCardProps {
  experience: ExperienceWithScore;
  selected: boolean;
  disabled: boolean;
  onToggle: () => void;
}

export function RecommendCard({
  experience,
  selected,
  disabled,
  onToggle,
}: RecommendCardProps) {
  const formatPeriod = () => {
    if (!experience.period_start) return "";
    if (!experience.period_end) return experience.period_start;
    return `${experience.period_start} ~ ${experience.period_end}`;
  };

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
            <div>
              <h3 className="text-base font-semibold text-foreground">
                {experience.title}
              </h3>
              {formatPeriod() && (
                <p className="mt-1 text-xs text-muted-foreground">{formatPeriod()}</p>
              )}
            </div>
            <div className="ml-4 text-right">
              <div className="text-lg font-bold text-blue-400">
                적합도 {experience.matchScore}%
              </div>
            </div>
          </div>

          {/* Weapons */}
          <div className="mt-2 flex flex-wrap gap-1">
            {experience.weapons.map((weapon, index) => (
              <span
                key={index}
                className="rounded-full bg-white/[0.06] px-2 py-1 text-xs font-medium text-foreground/80"
              >
                {weapon.weapon_name}
              </span>
            ))}
          </div>

          {/* Preview */}
          <p className="mt-3 line-clamp-2 text-sm text-muted-foreground">
            {experience.star_situation}
          </p>
        </div>
      </div>
    </div>
  );
}
