"use client";

import { useRouter } from "next/navigation";
import Link from "next/link";
import { useAuthStore } from "@/stores/auth-store";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";

export function UserDropdown() {
  const { user, logout } = useAuthStore();
  const router = useRouter();

  const handleLogout = () => {
    logout();
    router.replace("/login");
  };

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <button
          className="relative z-40 flex items-center gap-2 rounded-lg px-2 py-1.5 transition-all duration-200 hover:bg-white/[0.05]"
          aria-label="사용자 메뉴"
        >
          <div className="flex h-7 w-7 items-center justify-center rounded-full bg-primary/20 text-xs font-medium text-primary ring-1 ring-primary/30">
            {user?.nickname?.[0] ?? "U"}
          </div>
          <span className="hidden text-sm text-foreground/80 sm:block">
            {user?.nickname ?? "사용자"}
          </span>
        </button>
      </DropdownMenuTrigger>

      <DropdownMenuContent
        align="end"
        className="z-[80] w-52 rounded-xl border-border bg-card/95 p-1 shadow-xl shadow-black/30 backdrop-blur-md"
      >
        <DropdownMenuLabel className="px-3 py-2">
          <p className="truncate text-sm font-medium text-foreground">
            {user?.nickname ?? "사용자"}
          </p>
          {user?.email && (
            <p className="truncate text-xs text-muted-foreground">{user.email}</p>
          )}
        </DropdownMenuLabel>
        <DropdownMenuSeparator />

        {user?.role === "admin" && (
          <DropdownMenuItem asChild className="cursor-pointer rounded-lg px-3 py-2 text-sm text-foreground/80">
            <Link href="/admin">어드민</Link>
          </DropdownMenuItem>
        )}

        <DropdownMenuItem
          onClick={handleLogout}
          className="cursor-pointer rounded-lg px-3 py-2 text-sm text-destructive focus:text-destructive"
        >
          로그아웃
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  );
}
