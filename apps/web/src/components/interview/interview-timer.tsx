"use client";

import { Clock } from "lucide-react";

interface InterviewTimerProps {
  seconds: number;
}

export function InterviewTimer({ seconds }: InterviewTimerProps) {
  const mins = Math.floor(seconds / 60);
  const secs = seconds % 60;
  const display = `${String(mins).padStart(2, "0")}:${String(secs).padStart(2, "0")}`;

  return (
    <div className="flex items-center gap-1.5 text-sm text-gray-500">
      <Clock className="h-4 w-4" />
      <span aria-label="경과 시간">{display}</span>
    </div>
  );
}
