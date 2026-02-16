"use client";

import { experienceCategories } from "@/lib/validations/experience";
import type { UseFormRegisterReturn } from "react-hook-form";

interface CategorySelectProps {
  registration: UseFormRegisterReturn;
  error?: string;
}

export function CategorySelect({ registration, error }: CategorySelectProps) {
  return (
    <div className="space-y-1">
      <label htmlFor="category" className="text-sm font-medium text-foreground/80">카테고리</label>
      <select
        id="category"
        {...registration}
        className="w-full rounded-lg border border-border bg-input px-3 py-2 text-sm text-foreground focus:outline-none focus:ring-2 focus:ring-ring"
      >
        <option value="">선택해주세요</option>
        {experienceCategories.map((cat) => (
          <option key={cat} value={cat}>
            {cat}
          </option>
        ))}
      </select>
      {error && <p role="alert" className="text-xs text-red-500">{error}</p>}
    </div>
  );
}
