export type UserRole = "user" | "manager" | "admin" | "super_admin";

export const ROLE_LEVEL: Record<UserRole, number> = {
  user: 1,
  manager: 2,
  admin: 3,
  super_admin: 4,
} as const;

/** Returns true if `userRole` is at or above `requiredRole`. */
export function hasRole(userRole: UserRole, requiredRole: UserRole): boolean {
  return ROLE_LEVEL[userRole] >= ROLE_LEVEL[requiredRole];
}

export function isManager(role: UserRole): boolean {
  return hasRole(role, "manager");
}

export function isAdmin(role: UserRole): boolean {
  return hasRole(role, "admin");
}

export function isSuperAdmin(role: UserRole): boolean {
  return hasRole(role, "super_admin");
}
