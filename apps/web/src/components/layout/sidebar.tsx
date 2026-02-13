"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { useAuthStore } from "@/stores/auth-store";

const menuItems = [
  { label: "경험 관리", href: "/experiences", icon: "📋" },
  { label: "기업 분석", href: "/analysis", icon: "🏢" },
  { label: "자소서 코칭", href: "/coaching", icon: "✍️" },
  { label: "대시보드", href: "/dashboard", icon: "📊" },
];

export function Sidebar() {
  const pathname = usePathname();
  const { user } = useAuthStore();

  return (
    <aside className="w-64 h-screen bg-white border-r border-gray-200 flex flex-col">
      <div className="p-5 border-b border-gray-100">
        <Link href="/experiences" className="text-xl font-bold text-gray-900">
          Colight
        </Link>
      </div>

      <nav className="flex-1 py-4 space-y-1 px-3">
        {menuItems.map((item) => {
          const isActive = pathname.startsWith(item.href);
          return (
            <Link
              key={item.href}
              href={item.href}
              className={`flex items-center gap-3 px-3 py-2.5 rounded-lg text-sm transition-colors ${
                isActive
                  ? "bg-gray-100 text-gray-900 font-medium"
                  : "text-gray-600 hover:bg-gray-50 hover:text-gray-900"
              }`}
            >
              <span>{item.icon}</span>
              {item.label}
            </Link>
          );
        })}

        {user?.role === "admin" && (
          <>
            <div className="my-3 border-t border-gray-100" />
            <Link
              href="/admin"
              className={`flex items-center gap-3 px-3 py-2.5 rounded-lg text-sm transition-colors ${
                pathname.startsWith("/admin")
                  ? "bg-gray-100 text-gray-900 font-medium"
                  : "text-gray-600 hover:bg-gray-50 hover:text-gray-900"
              }`}
            >
              <span>🛡️</span>
              어드민
            </Link>
          </>
        )}
      </nav>

      <div className="p-4 border-t border-gray-100">
        <div className="flex items-center gap-3">
          <div className="w-8 h-8 rounded-full bg-gray-200 flex items-center justify-center text-sm font-medium text-gray-600">
            {user?.nickname?.[0] ?? "U"}
          </div>
          <div className="flex-1 min-w-0">
            <p className="text-sm font-medium text-gray-900 truncate">
              {user?.nickname ?? "사용자"}
            </p>
            {user?.role === "admin" && (
              <span className="text-xs text-purple-600">admin</span>
            )}
          </div>
        </div>
      </div>
    </aside>
  );
}
