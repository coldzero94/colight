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
    <div data-testid="weapon-coverage" className="rounded-lg border border-gray-200 bg-white p-4">
      <h4 className="mb-3 text-sm font-semibold text-gray-900">
        선택한 경험의 무기 커버리지
      </h4>

      <div className="space-y-2">
        {/* Primary weapon */}
        <div className="flex items-center gap-2">
          <span
            className={`inline-flex h-5 w-5 items-center justify-center rounded-full text-xs ${
              isPrimaryCovered
                ? "bg-green-100 text-green-600"
                : "bg-gray-100 text-gray-400"
            }`}
          >
            {isPrimaryCovered ? "✓" : "○"}
          </span>
          <span className="text-sm font-medium text-gray-700">
            {requiredWeapons.primary.weapon_name} <span className="text-xs text-gray-500">(주 무기)</span>
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
                    ? "bg-blue-100 text-blue-600"
                    : "bg-gray-100 text-gray-400"
                }`}
              >
                {isCovered ? "✓" : "○"}
              </span>
              <span className="text-sm text-gray-600">
                {weapon.weapon_name} <span className="text-xs text-gray-500">(부 무기)</span>
              </span>
            </div>
          );
        })}
      </div>

      {/* Summary */}
      <div className="mt-3 text-xs text-gray-500">
        {isPrimaryCovered && coveredSecondary.length > 0 ? (
          <span className="text-green-600">✅ 필수 무기가 모두 커버되었습니다</span>
        ) : !isPrimaryCovered ? (
          <span className="text-amber-600">⚠️ 주 무기를 포함한 경험을 선택하세요</span>
        ) : (
          <span>부 무기를 추가로 선택하면 더 좋습니다</span>
        )}
      </div>
    </div>
  );
}
