"use client";

import { useEffect } from "react";
import { useRouter } from "next/navigation";
import { useAuthStore } from "@/stores/auth-store";
import { hasRole, type UserRole } from "@/lib/auth-utils";

interface AdminGuardProps {
  children: React.ReactNode;
  requiredRole?: UserRole;
}

export function AdminGuard({
  children,
  requiredRole = "admin",
}: AdminGuardProps) {
  const { user } = useAuthStore();
  const router = useRouter();

  const allowed = user ? hasRole(user.role, requiredRole) : false;

  useEffect(() => {
    if (user && !allowed) {
      router.replace("/experiences");
    }
  }, [user, allowed, router]);

  if (!user || !allowed) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="animate-spin h-8 w-8 border-4 border-border border-t-foreground rounded-full" />
      </div>
    );
  }

  return <>{children}</>;
}
