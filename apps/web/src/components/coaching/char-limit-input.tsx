"use client";

import type { UseFormRegisterReturn } from "react-hook-form";

interface CharLimitInputProps {
  registration: UseFormRegisterReturn;
  defaultValue?: number;
  error?: string;
}

export function CharLimitInput({
  registration,
  defaultValue,
  error,
}: CharLimitInputProps) {
  return (
    <div className="space-y-1">
      <label
        htmlFor="char_limit"
        className="text-sm font-medium text-foreground/80"
      >
        글자수 제한
      </label>
      <div className="flex items-center gap-2">
        <input
          id="char_limit"
          type="number"
          min={200}
          max={2000}
          step={50}
          defaultValue={defaultValue}
          {...registration}
          className="w-32 rounded-lg border border-border bg-input px-3 py-2 text-sm text-foreground focus:outline-none focus:ring-2 focus:ring-ring"
        />
        <span className="text-sm text-muted-foreground">자</span>
      </div>
      {error && (
        <p role="alert" className="text-xs text-red-500">
          {error}
        </p>
      )}
    </div>
  );
}
