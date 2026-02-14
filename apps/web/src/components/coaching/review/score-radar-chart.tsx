"use client";

import dynamic from "next/dynamic";
import type { ReviewScores } from "@/lib/api/coaching";

const RadarChart = dynamic(
  () => import("recharts").then((m) => m.RadarChart),
  { ssr: false }
);
const PolarGrid = dynamic(
  () => import("recharts").then((m) => m.PolarGrid),
  { ssr: false }
);
const PolarAngleAxis = dynamic(
  () => import("recharts").then((m) => m.PolarAngleAxis),
  { ssr: false }
);
const Radar = dynamic(() => import("recharts").then((m) => m.Radar), {
  ssr: false,
});
const ResponsiveContainer = dynamic(
  () => import("recharts").then((m) => m.ResponsiveContainer),
  { ssr: false }
);

interface ScoreRadarChartProps {
  scores: ReviewScores;
  previousScores?: ReviewScores;
}

const DIMENSION_LABELS: Record<string, string> = {
  specificity: "구체성",
  job_fit: "직무적합",
  company_fit: "기업적합",
  authenticity: "진정성",
};

export function ScoreRadarChart({
  scores,
  previousScores,
}: ScoreRadarChartProps) {
  const data = Object.entries(DIMENSION_LABELS).map(([key, label]) => ({
    dimension: label,
    current: scores[key as keyof ReviewScores],
    ...(previousScores
      ? { previous: previousScores[key as keyof ReviewScores] }
      : {}),
  }));

  return (
    <div className="h-64 w-full">
      <ResponsiveContainer width="100%" height="100%">
        <RadarChart data={data} cx="50%" cy="50%" outerRadius="70%">
          <PolarGrid />
          <PolarAngleAxis dataKey="dimension" tick={{ fontSize: 12 }} />
          {previousScores && (
            <Radar
              name="이전"
              dataKey="previous"
              stroke="#9CA3AF"
              fill="#9CA3AF"
              fillOpacity={0.15}
            />
          )}
          <Radar
            name="현재"
            dataKey="current"
            stroke="#3B82F6"
            fill="#3B82F6"
            fillOpacity={0.25}
          />
        </RadarChart>
      </ResponsiveContainer>
    </div>
  );
}
