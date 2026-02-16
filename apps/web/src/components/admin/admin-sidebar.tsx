"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import type { LucideIcon } from "lucide-react";
import { LayoutDashboard, Users, FileText, Settings, BarChart3, ScrollText, ArrowLeft } from "lucide-react";
import { useAuthStore } from "@/stores/auth-store";
import { hasRole, type UserRole } from "@/lib/auth-utils";
import { ColightLogo } from "@/components/common/colight-logo";

interface MenuItem {
  label: string;
  href: string;
  icon: LucideIcon;
  minRole: UserRole;
}

const menuItems: MenuItem[] = [
  { label: "대시보드", href: "/admin", icon: LayoutDashboard, minRole: "admin" },
  { label: "사용자 관리", href: "/admin/users", icon: Users, minRole: "admin" },
  { label: "프롬프트 관리", href: "/admin/prompts", icon: FileText, minRole: "admin" },
  { label: "시스템 설정", href: "/admin/settings", icon: Settings, minRole: "super_admin" },
  { label: "사용량 모니터링", href: "/admin/usage", icon: BarChart3, minRole: "admin" },
  { label: "감사 로그", href: "/admin/logs", icon: ScrollText, minRole: "admin" },
];

export function AdminSidebar() {
  const pathname = usePathname();
  const { user } = useAuthStore();
  const userRole = user?.role ?? "user";

  const visibleItems = menuItems.filter((item) =>
    hasRole(userRole, item.minRole)
  );

  return (
    <aside className="w-64 h-screen bg-sidebar border-r border-sidebar-border flex flex-col">
      <div className="p-5 border-b border-sidebar-border">
        <Link href="/experiences" className="flex items-center gap-2 text-xl font-bold font-display text-foreground tracking-tight">
          <ColightLogo size={24} />
          Colight
        </Link>
      </div>

      <nav className="flex-1 py-4 space-y-1 px-3">
        {visibleItems.map((item) => {
          const isActive =
            item.href === "/admin"
              ? pathname === "/admin"
              : pathname.startsWith(item.href);

          return (
            <Link
              key={item.href}
              href={item.href}
              className={`flex items-center gap-3 px-3 py-2.5 rounded-lg text-sm transition-all duration-200 ${
                isActive
                  ? "bg-primary/10 text-primary font-medium border-l-2 border-primary"
                  : "text-muted-foreground hover:bg-white/[0.04] hover:text-foreground"
              }`}
            >
              <item.icon className="h-4 w-4" />
              {item.label}
            </Link>
          );
        })}

        <div className="my-3 border-t border-sidebar-border" />
        <Link
          href="/experiences"
          className="flex items-center gap-3 px-3 py-2.5 rounded-lg text-sm text-muted-foreground hover:bg-white/[0.04] hover:text-foreground transition-all duration-200"
        >
          <ArrowLeft className="h-4 w-4" />
          메인으로 돌아가기
        </Link>
      </nav>

      <div className="p-4 border-t border-sidebar-border">
        <div className="flex items-center gap-3">
          <div className="w-8 h-8 rounded-full bg-primary/20 flex items-center justify-center text-sm font-medium text-primary">
            {user?.nickname?.[0] ?? "U"}
          </div>
          <div className="flex-1 min-w-0">
            <p className="text-sm font-medium text-foreground truncate">
              {user?.nickname ?? "사용자"}
            </p>
            {user?.role && (
              <span className="text-xs text-primary/80">{user.role}</span>
            )}
          </div>
        </div>
      </div>
    </aside>
  );
}
