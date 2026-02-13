import type { ExperienceWeapon } from "@/lib/api/experiences";

const weaponLabels: Record<string, { icon: string; name: string }> = {
  W01: { icon: "🛡️", name: "위기극복" },
  W02: { icon: "👑", name: "리더십" },
  W03: { icon: "🤝", name: "팀워크/협업" },
  W04: { icon: "🚀", name: "도전정신" },
  W05: { icon: "🔧", name: "문제해결" },
  W06: { icon: "💬", name: "소통/설득" },
  W07: { icon: "📈", name: "성장/학습" },
};

function getWeaponLabel(code: string) {
  const topCode = code.split("-")[0];
  return weaponLabels[topCode] || { icon: "🏷️", name: code };
}

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
      {visible.map((w) => {
        const label = getWeaponLabel(w.weapon_code);
        return (
          <span
            key={w.id}
            className={`inline-flex items-center gap-0.5 rounded-full px-2 py-0.5 text-xs font-medium ${
              w.is_primary
                ? "bg-gray-900 text-white"
                : "bg-gray-100 text-gray-700"
            }`}
            data-testid={w.is_primary ? "primary-weapon" : "secondary-weapon"}
          >
            <span>{label.icon}</span>
            <span>{label.name}</span>
          </span>
        );
      })}
      {remaining > 0 && (
        <span className="inline-flex items-center rounded-full bg-gray-100 px-2 py-0.5 text-xs text-gray-500">
          +{remaining}
        </span>
      )}
    </div>
  );
}
