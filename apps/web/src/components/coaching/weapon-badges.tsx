"use client";

import { Award } from "lucide-react";

interface WeaponInfo {
  weapon_id: string;
  weapon_name: string;
  reason: string;
}

interface WeaponBadgesProps {
  primary: WeaponInfo;
  secondary: WeaponInfo[];
}

export function WeaponBadges({ primary, secondary }: WeaponBadgesProps) {
  return (
    <div className="rounded-lg border border-border bg-card p-6">
      <h3 className="mb-4 text-lg font-semibold text-foreground">필요 무기</h3>

      <div className="flex flex-wrap gap-2">
        {/* Primary weapon - highlighted */}
        <div
          data-primary="true"
          className="inline-flex items-center gap-1.5 rounded-full bg-amber-500/10 px-4 py-2 text-sm font-semibold text-amber-300 ring-2 ring-amber-500/50"
        >
          <span className="inline-flex h-5 w-5 items-center justify-center rounded-md bg-amber-500/20">
            <Award className="h-3.5 w-3.5" aria-hidden="true" />
          </span>
          <span>{primary.weapon_name}</span>
        </div>

        {/* Secondary weapons */}
        {secondary.map((weapon, index) => (
          <div
            key={index}
            className="inline-flex items-center gap-1 rounded-full bg-white/[0.06] px-4 py-2 text-sm font-medium text-foreground/80"
          >
            <span>{weapon.weapon_name}</span>
          </div>
        ))}
      </div>

      {/* Weapon reasons */}
      <div className="mt-4 space-y-2 text-xs text-muted-foreground">
        <div>
          <span className="font-medium">주 무기:</span> {primary.reason}
        </div>
        {secondary.map((weapon, index) => (
          <div key={index}>
            <span className="font-medium">부 무기:</span> {weapon.reason}
          </div>
        ))}
      </div>
    </div>
  );
}
