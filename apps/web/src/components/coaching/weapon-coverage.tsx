"use client";

interface Weapon {
  code: string;
  name: string;
}

interface RequiredWeapons {
  primary: { weapon_id: string; weapon_name: string; reason: string };
  secondary: Array<{ weapon_id: string; weapon_name: string; reason: string }>;
}

interface WeaponCoverageProps {
  selectedWeapons: Weapon[];
  requiredWeapons: RequiredWeapons;
}

export function WeaponCoverage({
  selectedWeapons,
  requiredWeapons,
}: WeaponCoverageProps) {
  const selectedCodes = selectedWeapons.map((w) => w.code);

  const isPrimaryCovered = selectedCodes.includes(requiredWeapons.primary.weapon_id);
  const coveredSecondary = requiredWeapons.secondary.filter((s) =>
    selectedCodes.includes(s.weapon_id)
  );

  return (
    <div data-testid="weapon-coverage" className="rounded-lg border border-border bg-card p-4">
      <h4 className="mb-3 text-sm font-semibold text-foreground">
        선택한 경험의 무기 커버리지
      </h4>

      <div className="space-y-2">
        {/* Primary weapon */}
        <div className="flex items-center gap-2">
          <span
            className={`inline-flex h-5 w-5 items-center justify-center rounded-full text-xs ${
              isPrimaryCovered
                ? "bg-primary/10 text-primary"
                : "bg-white/[0.06] text-muted-foreground/60"
            }`}
          >
            {isPrimaryCovered ? "✓" : "○"}
          </span>
          <span className="text-sm font-medium text-foreground/80">
            {requiredWeapons.primary.weapon_name} <span className="text-xs text-muted-foreground">(주 무기)</span>
          </span>
        </div>

        {/* Secondary weapons */}
        {requiredWeapons.secondary.map((weapon, index) => {
          const isCovered = coveredSecondary.some((s) => s.weapon_id === weapon.weapon_id);
          return (
            <div key={index} className="flex items-center gap-2">
              <span
                className={`inline-flex h-5 w-5 items-center justify-center rounded-full text-xs ${
                  isCovered
                    ? "bg-blue-500/10 text-blue-400"
                    : "bg-white/[0.06] text-muted-foreground/60"
                }`}
              >
                {isCovered ? "✓" : "○"}
              </span>
              <span className="text-sm text-muted-foreground">
                {weapon.weapon_name} <span className="text-xs text-muted-foreground">(부 무기)</span>
              </span>
            </div>
          );
        })}
      </div>

      {/* Summary */}
      <div className="mt-3 text-xs text-muted-foreground">
        {isPrimaryCovered && coveredSecondary.length > 0 ? (
          <span className="text-green-400">✅ 필수 무기가 모두 커버되었습니다</span>
        ) : !isPrimaryCovered ? (
          <span className="text-yellow-400">⚠️ 주 무기를 포함한 경험을 선택하세요</span>
        ) : (
          <span>부 무기를 추가로 선택하면 더 좋습니다</span>
        )}
      </div>
    </div>
  );
}
