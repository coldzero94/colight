"use client";

import type { Experience } from "@/lib/api/experiences";
import { ExperienceCard } from "./experience-card";

interface ExperienceListProps {
  experiences: Experience[];
  viewMode: "grid" | "list";
}

export function ExperienceList({ experiences, viewMode }: ExperienceListProps) {
  if (viewMode === "list") {
    return (
      <div className="space-y-2" data-testid="experience-list-view">
        {experiences.map((exp) => (
          <ExperienceCard key={exp.id} experience={exp} viewMode="list" />
        ))}
      </div>
    );
  }

  return (
    <div
      className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4"
      data-testid="experience-grid-view"
    >
      {experiences.map((exp) => (
        <ExperienceCard key={exp.id} experience={exp} viewMode="grid" />
      ))}
    </div>
  );
}
