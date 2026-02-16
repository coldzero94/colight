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
        className="group brand-surface-soft flex cursor-pointer items-center gap-4 rounded-xl p-4 transition-all hover:-translate-y-0.5 hover:border-primary/35 hover:shadow-[0_16px_32px_rgba(0,0,0,0.24)]"
        data-testid="experience-card-list"
      >
        <div className="flex items-center gap-3 min-w-0 flex-1">
          <CategoryIcon category={experience.category} />
          <div className="min-w-0 flex-1">
            <h3 className="truncate text-sm font-semibold text-foreground">
              {experience.title}
            </h3>
            {period && (
              <p className="text-xs text-muted-foreground/60 mt-0.5">{period}</p>
            )}
          </div>
        </div>
        <div className="hidden md:block flex-1 min-w-0">
          {experience.star_situation && (
            <p className="truncate text-xs text-muted-foreground group-hover:text-muted-foreground/90">
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
      className="group brand-surface-soft flex h-full cursor-pointer flex-col rounded-xl p-4 transition-all hover:-translate-y-0.5 hover:border-primary/35 hover:shadow-[0_16px_32px_rgba(0,0,0,0.24)]"
      data-testid="experience-card-grid"
    >
      <div className="flex items-center justify-between mb-2">
        <CategoryIcon category={experience.category} showLabel />
        {period && (
          <span className="text-xs text-muted-foreground/65 group-hover:text-muted-foreground/85">
            {period}
          </span>
        )}
      </div>
      <h3 className="mb-2 line-clamp-2 text-lg font-semibold text-foreground">
        {experience.title}
      </h3>
      {experience.star_situation && (
        <p className="mb-3 flex-1 line-clamp-2 text-sm text-muted-foreground group-hover:text-muted-foreground/90">
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
