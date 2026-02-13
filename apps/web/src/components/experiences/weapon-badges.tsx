import type { ExperienceWeapon } from "@/lib/api/experiences";
import { WeaponBadge } from "./weapon-badge";

interface WeaponBadgesProps {
  weapons: ExperienceWeapon[];
  maxVisible?: number;
}

export function WeaponBadges({ weapons, maxVisible }: WeaponBadgesProps) {
  if (!weapons || weapons.length === 0) return null;

  const sorted = [...weapons].sort((a, b) =>
    a.is_primary === b.is_primary ? 0 : a.is_primary ? -1 : 1
  );
  const visible = maxVisible ? sorted.slice(0, maxVisible) : sorted;
  const remaining = maxVisible ? sorted.length - maxVisible : 0;

  return (
    <div className="flex flex-wrap gap-1.5" data-testid="weapon-badges">
      {visible.map((w) => (
        <WeaponBadge
          key={w.id}
          code={w.weapon_code}
          isPrimary={w.is_primary}
          size="sm"
        />
      ))}
      {remaining > 0 && (
        <span className="inline-flex items-center rounded-full bg-gray-100 px-2 py-0.5 text-xs text-gray-500">
          +{remaining}
        </span>
      )}
    </div>
  );
}
