"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { ClipboardList, Mic, Building2, PenTool, LayoutDashboard, Shield } from "lucide-react";
import { useAuthStore } from "@/stores/auth-store";

const menuItems = [
  { label: "경험 관리", href: "/experiences", icon: ClipboardList },
  { label: "경험 인터뷰", href: "/interview", icon: Mic },
  { label: "기업 분석", href: "/analysis", icon: Building2 },
  { label: "자소서 코칭", href: "/coaching", icon: PenTool },
  { label: "대시보드", href: "/dashboard", icon: LayoutDashboard },
];

export function Sidebar() {
  const pathname = usePathname();
  const { user } = useAuthStore();

  return (
    <aside className="w-64 h-screen bg-sidebar border-r border-sidebar-border flex flex-col">
      <div className="p-5 border-b border-sidebar-border">
        <Link href="/experiences" className="text-xl font-bold font-display text-foreground tracking-tight">
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

        {user?.role === "admin" && (
          <>
            <div className="my-3 border-t border-sidebar-border" />
            <Link
              href="/admin"
              className={`flex items-center gap-3 px-3 py-2.5 rounded-lg text-sm transition-all duration-200 ${
                pathname.startsWith("/admin")
                  ? "bg-primary/10 text-primary font-medium border-l-2 border-primary"
                  : "text-muted-foreground hover:bg-white/[0.04] hover:text-foreground"
              }`}
            >
              <Shield className="h-4 w-4" />
              어드민
            </Link>
          </>
        )}
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
            {user?.role === "admin" && (
              <span className="text-xs text-primary/80">admin</span>
            )}
          </div>
        </div>
      </div>
    </aside>
  );
}
