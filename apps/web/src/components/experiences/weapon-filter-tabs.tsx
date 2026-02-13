import { WEAPON_CONFIG, type WeaponCode } from "@/lib/constants/weapon-colors";

const WEAPON_CODES = Object.keys(WEAPON_CONFIG) as WeaponCode[];

interface WeaponFilterTabsProps {
  counts: Record<string, number>;
  activeWeapon: string | null;
  onChange: (weapon: string | null) => void;
}

export function WeaponFilterTabs({
  counts,
  activeWeapon,
  onChange,
}: WeaponFilterTabsProps) {
  const totalCount = Object.values(counts).reduce((sum, n) => sum + n, 0);

  return (
    <div className="flex gap-2 overflow-x-auto pb-2" role="tablist">
      <button
        role="tab"
        aria-selected={activeWeapon === null}
        onClick={() => onChange(null)}
        className={`shrink-0 rounded-full px-3 py-1.5 text-sm font-medium transition-colors ${
          activeWeapon === null
            ? "bg-gray-900 text-white"
            : "bg-gray-100 text-gray-600 hover:bg-gray-200"
        }`}
      >
        전체 ({totalCount})
      </button>

      {WEAPON_CODES.map((code) => {
        const config = WEAPON_CONFIG[code];
        const count = counts[code] || 0;
        const isActive = activeWeapon === code;

        return (
          <button
            key={code}
            role="tab"
            aria-selected={isActive}
            onClick={() => onChange(code)}
            className={`shrink-0 rounded-full px-3 py-1.5 text-sm font-medium transition-colors ${
              isActive
                ? `${config.bgColor} ${config.textColor} ${config.borderColor} border`
                : "bg-gray-100 text-gray-600 hover:bg-gray-200"
            }`}
          >
            {config.icon} {config.name} ({count})
          </button>
        );
      })}
    </div>
  );
}
