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
    <div className="flex gap-2 overflow-x-auto pb-2" role="tablist" aria-label="무기 역량 필터">
      <button
        role="tab"
        aria-selected={activeWeapon === null}
        onClick={() => onChange(null)}
        className={`shrink-0 rounded-full px-4 py-2 text-sm font-medium transition-colors ${
          activeWeapon === null
            ? "bg-primary text-primary-foreground"
            : "bg-white/[0.06] text-muted-foreground hover:bg-white/[0.08]"
        }`}
      >
        전체 ({totalCount})
      </button>

      {WEAPON_CODES.map((code) => {
        const config = WEAPON_CONFIG[code];
        const Icon = config.icon;
        const count = counts[code] || 0;
        const isActive = activeWeapon === code;

        return (
          <button
            key={code}
            role="tab"
            aria-selected={isActive}
            onClick={() => onChange(code)}
            className={`shrink-0 rounded-full px-3 py-2 text-sm font-medium transition-all ${
              isActive
                ? `${config.bgColor} ${config.textColor} ${config.borderColor} border`
                : "bg-white/[0.06] text-muted-foreground hover:bg-white/[0.08] hover:text-foreground"
            }`}
          >
            <span className="inline-flex items-center gap-1.5">
              <span className={`inline-flex h-5 w-5 items-center justify-center rounded-md ${isActive ? config.iconBgColor : "bg-white/[0.08]"}`}>
                <Icon className={`h-3.5 w-3.5 ${isActive ? config.iconColor : "text-muted-foreground"}`} aria-hidden="true" />
              </span>
              <span>{config.name} ({count})</span>
            </span>
          </button>
        );
      })}
    </div>
  );
}
