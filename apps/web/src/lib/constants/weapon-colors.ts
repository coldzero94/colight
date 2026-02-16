// Weapon category color and icon mappings
// Weapon icons match seed.go (source of truth)
// Dark-first: uses translucent bg + lighter text for dark backgrounds
export const WEAPON_CONFIG = {
  W01: {
    name: "위기극복",
    icon: "🛡️",
    bgColor: "bg-red-500/15",
    textColor: "text-red-400",
    borderColor: "border-red-500/30",
  },
  W02: {
    name: "리더십",
    icon: "👑",
    bgColor: "bg-amber-500/15",
    textColor: "text-amber-400",
    borderColor: "border-amber-500/30",
  },
  W03: {
    name: "팀워크/협업",
    icon: "🤝",
    bgColor: "bg-emerald-500/15",
    textColor: "text-emerald-400",
    borderColor: "border-emerald-500/30",
  },
  W04: {
    name: "도전정신",
    icon: "🚀",
    bgColor: "bg-violet-500/15",
    textColor: "text-violet-400",
    borderColor: "border-violet-500/30",
  },
  W05: {
    name: "문제해결",
    icon: "🔧",
    bgColor: "bg-blue-500/15",
    textColor: "text-blue-400",
    borderColor: "border-blue-500/30",
  },
  W06: {
    name: "소통/설득",
    icon: "💬",
    bgColor: "bg-pink-500/15",
    textColor: "text-pink-400",
    borderColor: "border-pink-500/30",
  },
  W07: {
    name: "성장/학습",
    icon: "📈",
    bgColor: "bg-indigo-500/15",
    textColor: "text-indigo-400",
    borderColor: "border-indigo-500/30",
  },
} as const;

export type WeaponCode = keyof typeof WEAPON_CONFIG;
