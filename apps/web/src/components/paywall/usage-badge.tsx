"use client";

import { cn } from "@/lib/utils";

interface UsageBadgeProps {
  used: number;
  limit: number;
  className?: string;
}

function getColor(used: number, limit: number) {
  if (limit <= 0) return "bg-green-100 text-green-700";
  const ratio = used / limit;
  if (ratio >= 1) return "bg-red-100 text-red-700";
  if (ratio >= 0.7) return "bg-yellow-100 text-yellow-700";
  return "bg-green-100 text-green-700";
}

export function UsageBadge({ used, limit, className }: UsageBadgeProps) {
  if (limit < 0) {
    return (
      <span
        className={cn(
          "inline-flex items-center rounded-md px-2 py-0.5 text-xs font-medium bg-green-100 text-green-700",
          className,
        )}
      >
        무제한
      </span>
    );
  }

  const remaining = Math.max(0, limit - used);

  return (
    <span
      className={cn(
        "inline-flex items-center rounded-md px-2 py-0.5 text-xs font-medium",
        getColor(used, limit),
        className,
      )}
    >
      {remaining}/{limit}
    </span>
  );
}
