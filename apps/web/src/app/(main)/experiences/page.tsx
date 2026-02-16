"use client";

import { useEffect, useMemo, useState } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import { ClipboardList, Search } from "lucide-react";
import { EmptyState } from "@/components/common/empty-state";
import { LoadingSpinner } from "@/components/common/loading-spinner";
import { ExperienceList } from "@/components/experiences/experience-list";
import { ViewToggle } from "@/components/experiences/view-toggle";
import { SortSelect } from "@/components/experiences/sort-select";
import { WeaponFilterTabs } from "@/components/experiences/weapon-filter-tabs";
import { WeaponRadarChart } from "@/components/experiences/weapon-radar-chart";
import { useExperiences } from "@/hooks/use-experiences";
import { computeWeaponCounts } from "@/lib/api/experiences";
import {
  WEAPON_CONFIG,
  type WeaponCode,
} from "@/lib/constants/weapon-colors";

export default function ExperiencesPage() {
  const router = useRouter();
  const searchParams = useSearchParams();

  const sort = searchParams.get("sort") || "latest";
  const category = searchParams.get("category") || "";
  const weapon = searchParams.get("weapon") || "";

  const [viewMode, setViewMode] = useState<"grid" | "list">(() => {
    if (typeof window !== "undefined") {
      return (localStorage.getItem("exp-view") as "grid" | "list") || "grid";
    }
    return "grid";
  });

  // Fetch all experiences (unfiltered by weapon) for counts
  const { data: allExperiences, isLoading: isLoadingAll } = useExperiences({
    sort: sort as "latest" | "oldest" | "title",
    category: category || undefined,
  });

  // Fetch filtered experiences when weapon filter is active
  const { data: filteredExperiences, isLoading: isLoadingFiltered } =
    useExperiences({
      sort: sort as "latest" | "oldest" | "title",
      category: category || undefined,
      weapon: weapon || undefined,
    });

  const isLoading = isLoadingAll || (weapon && isLoadingFiltered);
  const experiences = weapon ? filteredExperiences : allExperiences;

  const weaponCounts = useMemo(
    () => computeWeaponCounts(allExperiences || []),
    [allExperiences]
  );

  useEffect(() => {
    localStorage.setItem("exp-view", viewMode);
  }, [viewMode]);

  const handleSortChange = (newSort: string) => {
    const params = new URLSearchParams(searchParams.toString());
    params.set("sort", newSort);
    router.push(`/experiences?${params.toString()}`);
  };

  const handleWeaponChange = (newWeapon: string | null) => {
    const params = new URLSearchParams(searchParams.toString());
    if (newWeapon) {
      params.set("weapon", newWeapon);
    } else {
      params.delete("weapon");
    }
    router.push(`/experiences?${params.toString()}`);
  };

  if (isLoading) {
    return (
      <div className="flex justify-center py-20">
        <LoadingSpinner />
      </div>
    );
  }

  if (!allExperiences || allExperiences.length === 0) {
    return (
      <div className="space-y-6">
        <div className="flex items-center justify-between">
          <h1 className="text-2xl font-bold font-display text-foreground">내 경험</h1>
        </div>
        <EmptyState
          icon={<ClipboardList className="h-7 w-7" />}
          title="아직 등록된 경험이 없습니다"
          description="첫 번째 경험을 등록해보세요!"
          action={{
            label: "경험 등록",
            onClick: () => router.push("/experiences/new"),
          }}
        />
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold font-display text-foreground">내 경험</h1>
        <button
          onClick={() => router.push("/experiences/new")}
          className="rounded-lg bg-primary px-4 py-2 text-sm font-medium text-primary-foreground hover:bg-primary/90 transition-colors"
        >
          + 경험 추가
        </button>
      </div>

      <WeaponRadarChart counts={weaponCounts} />

      <WeaponFilterTabs
        counts={weaponCounts}
        activeWeapon={weapon || null}
        onChange={handleWeaponChange}
      />

      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <SortSelect value={sort} onChange={handleSortChange} />
        </div>
        <ViewToggle mode={viewMode} onChange={setViewMode} />
      </div>

      {!experiences || experiences.length === 0 ? (
        <EmptyState
          icon={<Search className="h-7 w-7" />}
          title={`아직 ${WEAPON_CONFIG[weapon as WeaponCode]?.name || ""} 역량의 경험이 없습니다`}
          description="다른 무기를 선택하거나 경험을 등록해보세요."
          action={{
            label: "전체 보기",
            onClick: () => handleWeaponChange(null),
          }}
        />
      ) : (
        <ExperienceList experiences={experiences} viewMode={viewMode} />
      )}
    </div>
  );
}
