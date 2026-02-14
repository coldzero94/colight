"use client";

interface CharCounterProps {
  current: number;
  limit: number;
}

export function CharCounter({ current, limit }: CharCounterProps) {
  const percentage = (current / limit) * 100;

  // Color rules based on percentage
  const getColor = () => {
    if (current > limit) return "text-red-600 font-bold";
    if (percentage >= 90) return "text-amber-600";
    if (percentage >= 70) return "text-blue-600";
    return "text-gray-500";
  };

  return (
    <div className="flex items-center justify-between border-t border-gray-200 bg-gray-50 px-4 py-2">
      <div className="text-xs text-gray-500">
        {percentage < 70 && "여유 있음"}
        {percentage >= 70 && percentage < 90 && "적절한 범위"}
        {percentage >= 90 && percentage <= 100 && "거의 다 참"}
        {current > limit && "초과 경고"}
      </div>
      <div className={`text-sm font-medium ${getColor()}`}>
        {current} / {limit}자
        {current > limit && " (초과)"}
      </div>
    </div>
  );
}
