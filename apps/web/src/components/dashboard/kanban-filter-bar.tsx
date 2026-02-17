"use client";

import { useMemo } from "react";
import { Search, Tag, X } from "lucide-react";
import { Input } from "@/components/ui/input";
import type { ApplicationDetail } from "@/lib/api/applications";

export interface KanbanFilters {
  query: string;
  tags: string[];
}

interface KanbanFilterBarProps {
  applications: ApplicationDetail[];
  filters: KanbanFilters;
  onFiltersChange: (filters: KanbanFilters) => void;
}

export function KanbanFilterBar({
  applications,
  filters,
  onFiltersChange,
}: KanbanFilterBarProps) {
  const allTags = useMemo(() => {
    const tagSet = new Set<string>();
    for (const app of applications) {
      for (const tag of app.tags) {
        tagSet.add(tag);
      }
    }
    return Array.from(tagSet).sort();
  }, [applications]);

  const toggleTag = (tag: string) => {
    const newTags = filters.tags.includes(tag)
      ? filters.tags.filter((t) => t !== tag)
      : [...filters.tags, tag];
    onFiltersChange({ ...filters, tags: newTags });
  };

  const hasActiveFilters = filters.query !== "" || filters.tags.length > 0;

  return (
    <div className="brand-surface-soft rounded-xl px-4 py-3 space-y-3">
      <div className="flex items-center gap-3">
        <div className="relative flex-1">
          <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground/50" />
          <Input
            placeholder="회사명 또는 포지션 검색..."
            value={filters.query}
            onChange={(e) =>
              onFiltersChange({ ...filters, query: e.target.value })
            }
            className="pl-9 bg-white/[0.04] border-white/12"
          />
        </div>
        {hasActiveFilters && (
          <button
            onClick={() => onFiltersChange({ query: "", tags: [] })}
            className="inline-flex items-center gap-1 rounded-lg border border-white/12 px-2.5 py-1.5 text-xs text-muted-foreground transition-colors hover:text-foreground"
          >
            <X className="h-3 w-3" />
            초기화
          </button>
        )}
      </div>

      {allTags.length > 0 && (
        <div className="flex items-center gap-2 flex-wrap">
          <Tag className="h-3.5 w-3.5 text-muted-foreground/50 shrink-0" />
          {allTags.map((tag) => {
            const active = filters.tags.includes(tag);
            return (
              <button
                key={tag}
                onClick={() => toggleTag(tag)}
                className={`rounded-md px-2 py-0.5 text-xs font-medium transition-colors ${
                  active
                    ? "bg-primary/20 text-primary border border-primary/30"
                    : "bg-white/[0.04] text-muted-foreground border border-white/12 hover:bg-white/[0.08]"
                }`}
              >
                {tag}
              </button>
            );
          })}
        </div>
      )}
    </div>
  );
}

/** Filters applications client-side based on active filters. */
export function filterApplications(
  applications: ApplicationDetail[],
  filters: KanbanFilters,
): ApplicationDetail[] {
  let result = applications;

  if (filters.query) {
    const q = filters.query.toLowerCase();
    result = result.filter(
      (app) =>
        app.company_name.toLowerCase().includes(q) ||
        app.position.toLowerCase().includes(q),
    );
  }

  if (filters.tags.length > 0) {
    result = result.filter((app) =>
      filters.tags.some((tag) => app.tags.includes(tag)),
    );
  }

  return result;
}
