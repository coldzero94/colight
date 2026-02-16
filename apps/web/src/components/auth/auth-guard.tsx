"use client";

import { useEffect } from "react";
import { useRouter } from "next/navigation";
import { useAuthStore } from "@/stores/auth-store";

export function AuthGuard({ children }: { children: React.ReactNode }) {
  const { accessToken, user, isLoading, fetchUser, _hasHydrated } = useAuthStore();
  const router = useRouter();

  useEffect(() => {
    // Wait for hydration to complete before checking auth
    if (!_hasHydrated) return;

    if (!accessToken) {
      router.replace("/login");
      return;
    }
    if (!user && !isLoading) {
      fetchUser();
    }
  }, [accessToken, user, isLoading, fetchUser, router, _hasHydrated]);

  // Show loading while hydrating
  if (!_hasHydrated || !accessToken || isLoading) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="animate-spin h-8 w-8 border-4 border-gray-300 border-t-gray-900 rounded-full" />
      </div>
    );
  }

  return <>{children}</>;
}
