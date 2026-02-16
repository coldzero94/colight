"use client";

import { useState } from "react";
import dynamic from "next/dynamic";
import { ChevronDown, ChevronUp } from "lucide-react";
import {
  WEAPON_CONFIG,
  type WeaponCode,
} from "@/lib/constants/weapon-colors";

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

interface WeaponRadarChartProps {
  counts: Record<string, number>;
}

const WEAPON_CODES: WeaponCode[] = [
  "W01",
  "W02",
  "W03",
  "W04",
  "W05",
  "W06",
  "W07",
];

export function WeaponRadarChart({ counts }: WeaponRadarChartProps) {
  const [isOpen, setIsOpen] = useState(true);

  const maxCount = Math.max(...WEAPON_CODES.map((c) => counts[c] || 0), 1);

  const data = WEAPON_CODES.map((code) => ({
    weapon: WEAPON_CONFIG[code].name,
    count: counts[code] || 0,
    fullMark: maxCount,
  }));

  const missingWeapons = WEAPON_CODES.filter((code) => !counts[code])
    .map((code) => WEAPON_CONFIG[code].name);

  return (
    <div className="rounded-lg border border-border bg-card">
      <button
        onClick={() => setIsOpen(!isOpen)}
        className="flex w-full items-center justify-between px-4 py-3 text-sm font-medium text-foreground hover:bg-white/[0.04]"
      >
        <span>무기 분포</span>
        {isOpen ? (
          <ChevronUp className="h-4 w-4 text-muted-foreground/60" />
        ) : (
          <ChevronDown className="h-4 w-4 text-muted-foreground/60" />
        )}
      </button>

      {isOpen && (
        <div className="border-t border-border px-4 pb-4">
          <div className="h-64 w-full">
            <ResponsiveContainer width="100%" height="100%">
              <RadarChart data={data} cx="50%" cy="50%" outerRadius="70%">
                <PolarGrid stroke="rgba(255,255,255,0.1)" />
                <PolarAngleAxis dataKey="weapon" tick={{ fontSize: 12, fill: "rgba(255,255,255,0.6)" }} />
                <Radar
                  name="보유 경험"
                  dataKey="count"
                  stroke="#34d399"
                  fill="#34d399"
                  fillOpacity={0.25}
                />
              </RadarChart>
            </ResponsiveContainer>
          </div>

          {missingWeapons.length > 0 && (
            <p className="mt-2 text-center text-xs text-yellow-400">
              보완 추천: {missingWeapons.join(", ")} 역량의 경험을 추가해보세요
            </p>
          )}
        </div>
      )}
    </div>
  );
}
