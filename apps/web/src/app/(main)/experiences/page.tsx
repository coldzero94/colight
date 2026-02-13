"use client";

import { useEffect, useState } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import { EmptyState } from "@/components/common/empty-state";
import { LoadingSpinner } from "@/components/common/loading-spinner";
import { ExperienceList } from "@/components/experiences/experience-list";
import { ViewToggle } from "@/components/experiences/view-toggle";
import { SortSelect } from "@/components/experiences/sort-select";
import { useExperiences } from "@/hooks/use-experiences";

export default function ExperiencesPage() {
  const router = useRouter();
  const searchParams = useSearchParams();

  const sort = searchParams.get("sort") || "latest";
  const category = searchParams.get("category") || "";

  const [viewMode, setViewMode] = useState<"grid" | "list">(() => {
    if (typeof window !== "undefined") {
      return (localStorage.getItem("exp-view") as "grid" | "list") || "grid";
    }
    return "grid";
  });

  const { data: experiences, isLoading } = useExperiences({
    sort: sort as "latest" | "oldest" | "title",
    category: category || undefined,
  });

  useEffect(() => {
    localStorage.setItem("exp-view", viewMode);
  }, [viewMode]);

  const handleSortChange = (newSort: string) => {
    const params = new URLSearchParams(searchParams.toString());
    params.set("sort", newSort);
    router.push(`/experiences?${params.toString()}`);
  };

  if (isLoading) {
    return (
      <div className="flex justify-center py-20">
        <LoadingSpinner />
      </div>
    );
  }

  if (!experiences || experiences.length === 0) {
    return (
      <div className="space-y-6">
        <div className="flex items-center justify-between">
          <h1 className="text-2xl font-bold text-gray-900">내 경험</h1>
        </div>
        <EmptyState
          icon={<span>📋</span>}
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
        <h1 className="text-2xl font-bold text-gray-900">내 경험</h1>
        <button
          onClick={() => router.push("/experiences/new")}
          className="rounded-lg bg-gray-900 px-4 py-2 text-sm font-medium text-white hover:bg-gray-800 transition-colors"
        >
          + 경험 추가
        </button>
      </div>

      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <SortSelect value={sort} onChange={handleSortChange} />
        </div>
        <ViewToggle mode={viewMode} onChange={setViewMode} />
      </div>

      <ExperienceList experiences={experiences} viewMode={viewMode} />
    </div>
  );
}
