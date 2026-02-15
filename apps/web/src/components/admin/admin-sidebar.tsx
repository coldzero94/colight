"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { useAuthStore } from "@/stores/auth-store";
import { hasRole, type UserRole } from "@/lib/auth-utils";

interface MenuItem {
  label: string;
  href: string;
  icon: string;
  minRole: UserRole;
}

const menuItems: MenuItem[] = [
  { label: "대시보드", href: "/admin", icon: "📊", minRole: "admin" },
  { label: "사용자 관리", href: "/admin/users", icon: "👥", minRole: "admin" },
  {
    label: "프롬프트 관리",
    href: "/admin/prompts",
    icon: "📝",
    minRole: "admin",
  },
  {
    label: "시스템 설정",
    href: "/admin/settings",
    icon: "⚙️",
    minRole: "super_admin",
  },
  {
    label: "사용량 모니터링",
    href: "/admin/usage",
    icon: "📈",
    minRole: "admin",
  },
  { label: "감사 로그", href: "/admin/logs", icon: "📋", minRole: "admin" },
];

export function AdminSidebar() {
  const pathname = usePathname();
  const { user } = useAuthStore();
  const userRole = user?.role ?? "user";

  const visibleItems = menuItems.filter((item) =>
    hasRole(userRole, item.minRole)
  );

  return (
    <aside className="w-60 bg-gray-900 text-white flex flex-col">
      <div className="p-4 border-b border-gray-700">
        <h2 className="text-lg font-bold">Colight Admin</h2>
      </div>

      <nav className="flex-1 py-4">
        {visibleItems.map((item) => {
          const isActive =
            item.href === "/admin"
              ? pathname === "/admin"
              : pathname.startsWith(item.href);

          return (
            <Link
              key={item.href}
              href={item.href}
              className={`flex items-center gap-3 px-4 py-2.5 text-sm transition-colors ${
                isActive
                  ? "bg-gray-700 text-white"
                  : "text-gray-300 hover:bg-gray-800 hover:text-white"
              }`}
            >
              <span>{item.icon}</span>
              {item.label}
            </Link>
          );
        })}
      </nav>

      <div className="p-4 border-t border-gray-700">
        <Link
          href="/experiences"
          className="flex items-center gap-2 text-sm text-gray-400 hover:text-white transition-colors"
        >
          ← 메인 앱으로
        </Link>
      </div>
    </aside>
  );
}
