import { WEAPON_CONFIG, type WeaponCode } from "@/lib/constants/weapon-colors";

interface WeaponBadgeProps {
  code: string;
  isPrimary?: boolean;
  size?: "sm" | "md" | "lg";
  showConfidence?: boolean;
  confidence?: number;
}

export function WeaponBadge({
  code,
  isPrimary = false,
  size = "md",
  showConfidence = false,
  confidence,
}: WeaponBadgeProps) {
  const weaponCode = code.substring(0, 3) as WeaponCode; // Handle W01-A -> W01
  const config = WEAPON_CONFIG[weaponCode];

  if (!config) {
    return null;
  }

  const sizeClasses = {
    sm: "text-xs px-2 py-0.5",
    md: "text-sm px-2.5 py-1",
    lg: "text-base px-3 py-1.5",
  };

  const borderWidth = isPrimary ? "border-2" : "border";

  return (
    <span
      className={`inline-flex items-center gap-1 rounded-full ${config.bgColor} ${config.textColor} ${config.borderColor} ${borderWidth} ${sizeClasses[size]} font-medium`}
      data-testid={isPrimary ? "primary-weapon" : "secondary-weapon"}
      data-weapon-badge={code}
    >
      <span>{config.icon}</span>
      <span>{config.name}</span>
      {isPrimary && <span className="text-xs">(주)</span>}
      {showConfidence && confidence !== undefined && (
        <span className="text-xs opacity-70">{Math.round(confidence * 100)}%</span>
      )}
    </span>
  );
}
