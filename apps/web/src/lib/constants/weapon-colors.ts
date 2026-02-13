// Weapon category color and icon mappings
// Weapon icons match seed.go (source of truth)
export const WEAPON_CONFIG = {
  W01: {
    name: "위기극복",
    icon: "🛡️",
    bgColor: "bg-red-100",
    textColor: "text-red-700",
    borderColor: "border-red-200",
  },
  W02: {
    name: "리더십",
    icon: "👑",
    bgColor: "bg-amber-100",
    textColor: "text-amber-700",
    borderColor: "border-amber-200",
  },
  W03: {
    name: "팀워크/협업",
    icon: "🤝",
    bgColor: "bg-emerald-100",
    textColor: "text-emerald-700",
    borderColor: "border-emerald-200",
  },
  W04: {
    name: "도전정신",
    icon: "🚀",
    bgColor: "bg-violet-100",
    textColor: "text-violet-700",
    borderColor: "border-violet-200",
  },
  W05: {
    name: "문제해결",
    icon: "🔧",
    bgColor: "bg-blue-100",
    textColor: "text-blue-700",
    borderColor: "border-blue-200",
  },
  W06: {
    name: "소통/설득",
    icon: "💬",
    bgColor: "bg-pink-100",
    textColor: "text-pink-700",
    borderColor: "border-pink-200",
  },
  W07: {
    name: "성장/학습",
    icon: "📈",
    bgColor: "bg-indigo-100",
    textColor: "text-indigo-700",
    borderColor: "border-indigo-200",
  },
} as const;

export type WeaponCode = keyof typeof WEAPON_CONFIG;
