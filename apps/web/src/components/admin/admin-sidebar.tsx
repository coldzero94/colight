"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";

const menuItems = [
  { label: "대시보드", href: "/admin", icon: "📊" },
  { label: "사용자 관리", href: "/admin/users", icon: "👥" },
  { label: "프롬프트 관리", href: "/admin/prompts", icon: "📝" },
];

export function AdminSidebar() {
  const pathname = usePathname();

  return (
    <aside className="w-60 bg-gray-900 text-white flex flex-col">
      <div className="p-4 border-b border-gray-700">
        <h2 className="text-lg font-bold">Colight Admin</h2>
      </div>

      <nav className="flex-1 py-4">
        {menuItems.map((item) => {
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
