import type { LucideIcon } from "lucide-react";
import {
  Crown,
  Handshake,
  MessageSquareQuote,
  Rocket,
  ShieldAlert,
  TrendingUp,
  Wrench,
} from "lucide-react";

interface WeaponConfigItem {
  name: string;
  icon: LucideIcon;
  bgColor: string;
  textColor: string;
  borderColor: string;
  iconColor: string;
  iconBgColor: string;
}

// Weapon category color and icon mappings
// Dark-first: uses translucent bg + lighter text for dark backgrounds
export const WEAPON_CONFIG = {
  W01: {
    name: "위기극복",
    icon: ShieldAlert,
    bgColor: "bg-red-500/15",
    textColor: "text-red-400",
    borderColor: "border-red-500/30",
    iconColor: "text-red-300",
    iconBgColor: "bg-red-500/20",
  },
  W02: {
    name: "리더십",
    icon: Crown,
    bgColor: "bg-amber-500/15",
    textColor: "text-amber-400",
    borderColor: "border-amber-500/30",
    iconColor: "text-amber-300",
    iconBgColor: "bg-amber-500/20",
  },
  W03: {
    name: "팀워크/협업",
    icon: Handshake,
    bgColor: "bg-emerald-500/15",
    textColor: "text-emerald-400",
    borderColor: "border-emerald-500/30",
    iconColor: "text-emerald-300",
    iconBgColor: "bg-emerald-500/20",
  },
  W04: {
    name: "도전정신",
    icon: Rocket,
    bgColor: "bg-violet-500/15",
    textColor: "text-violet-400",
    borderColor: "border-violet-500/30",
    iconColor: "text-violet-300",
    iconBgColor: "bg-violet-500/20",
  },
  W05: {
    name: "문제해결",
    icon: Wrench,
    bgColor: "bg-blue-500/15",
    textColor: "text-blue-400",
    borderColor: "border-blue-500/30",
    iconColor: "text-blue-300",
    iconBgColor: "bg-blue-500/20",
  },
  W06: {
    name: "소통/설득",
    icon: MessageSquareQuote,
    bgColor: "bg-pink-500/15",
    textColor: "text-pink-400",
    borderColor: "border-pink-500/30",
    iconColor: "text-pink-300",
    iconBgColor: "bg-pink-500/20",
  },
  W07: {
    name: "성장/학습",
    icon: TrendingUp,
    bgColor: "bg-indigo-500/15",
    textColor: "text-indigo-400",
    borderColor: "border-indigo-500/30",
    iconColor: "text-indigo-300",
    iconBgColor: "bg-indigo-500/20",
  },
} as const satisfies Record<string, WeaponConfigItem>;

export type WeaponCode = keyof typeof WEAPON_CONFIG;
