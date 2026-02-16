"use client";

import Link from "next/link";
import type { Experience } from "@/lib/api/experiences";
import { CategoryIcon } from "./category-icon";
import { WeaponBadges } from "./weapon-badges";

interface ExperienceCardProps {
  experience: Experience;
  viewMode: "grid" | "list";
}

export function ExperienceCard({ experience, viewMode }: ExperienceCardProps) {
  const period = formatPeriod(experience.period_start, experience.period_end);

  if (viewMode === "list") {
    return (
      <Link
        href={`/experiences/${experience.id}`}
        className="flex items-center gap-4 rounded-lg border border-border bg-card p-4 hover:shadow-md transition-shadow cursor-pointer"
        data-testid="experience-card-list"
      >
        <div className="flex items-center gap-3 min-w-0 flex-1">
          <CategoryIcon category={experience.category} />
          <div className="min-w-0 flex-1">
            <h3 className="text-sm font-semibold text-foreground truncate">
              {experience.title}
            </h3>
            {period && (
              <p className="text-xs text-muted-foreground/60 mt-0.5">{period}</p>
            )}
          </div>
        </div>
        <div className="hidden md:block flex-1 min-w-0">
          {experience.star_situation && (
            <p className="text-xs text-muted-foreground truncate">
              {experience.star_situation}
            </p>
          )}
        </div>
        <div className="flex items-center gap-3">
          {experience.weapons && experience.weapons.length > 0 && (
            <WeaponBadges weapons={experience.weapons} maxVisible={3} />
          )}
        </div>
      </Link>
    );
  }

  return (
    <Link
      href={`/experiences/${experience.id}`}
      className="flex flex-col rounded-lg border border-border bg-card p-4 hover:shadow-md transition-shadow cursor-pointer h-full"
      data-testid="experience-card-grid"
    >
      <div className="flex items-center justify-between mb-2">
        <CategoryIcon category={experience.category} showLabel />
        {period && <span className="text-xs text-muted-foreground/60">{period}</span>}
      </div>
      <h3 className="text-lg font-semibold text-foreground line-clamp-2 mb-2">
        {experience.title}
      </h3>
      {experience.star_situation && (
        <p className="text-sm text-muted-foreground line-clamp-2 mb-3 flex-1">
          {experience.star_situation}
        </p>
      )}
      {experience.weapons && experience.weapons.length > 0 && (
        <div className="mt-auto pt-2">
          <WeaponBadges weapons={experience.weapons} maxVisible={3} />
        </div>
      )}
    </Link>
  );
}

function formatPeriod(start?: string, end?: string): string {
  if (!start && !end) return "";
  const fmt = (d: string) => d.slice(0, 7).replace("-", ".");
  if (start && end) return `${fmt(start)} ~ ${fmt(end)}`;
  if (start) return `${fmt(start)} ~`;
  return `~ ${fmt(end!)}`;
}
