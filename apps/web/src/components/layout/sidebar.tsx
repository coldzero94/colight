"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { ClipboardList, Mic, Building2, PenTool, LayoutDashboard, Shield, User } from "lucide-react";
import { useAuthStore } from "@/stores/auth-store";
import { ColightLogo } from "@/components/common/colight-logo";

const menuItems = [
  { label: "경험 관리", href: "/experiences", icon: ClipboardList },
  { label: "경험 인터뷰", href: "/interview", icon: Mic },
  { label: "기업 분석", href: "/analysis", icon: Building2 },
  { label: "자소서 코칭", href: "/coaching", icon: PenTool },
  { label: "대시보드", href: "/dashboard", icon: LayoutDashboard },
  { label: "마이페이지", href: "/my-page", icon: User },
];

export function Sidebar() {
  const pathname = usePathname();
  const { user } = useAuthStore();

  return (
    <aside className="brand-surface relative flex h-screen w-64 flex-col border-r border-sidebar-border/90 bg-sidebar/90">
      <div className="pointer-events-none absolute inset-y-0 right-0 w-px bg-gradient-to-b from-transparent via-primary/30 to-transparent" />

      <div className="border-b border-sidebar-border/80 px-5 py-4">
        <Link href="/experiences" className="group flex items-center gap-2.5 text-xl font-bold font-display text-foreground tracking-tight">
          <ColightLogo size={26} />
          <div>
            <span className="block leading-none transition-opacity group-hover:opacity-90">Colight</span>
            <span className="text-[10px] font-medium uppercase tracking-[0.22em] text-primary/80">
              Together We Light
            </span>
          </div>
        </Link>
      </div>

      <nav className="flex-1 space-y-1 px-3 py-4">
        {menuItems.map((item) => {
          const isActive = pathname.startsWith(item.href);
          return (
            <Link
              key={item.href}
              href={item.href}
              className={`group flex items-center gap-3 rounded-xl px-3 py-2.5 text-sm transition-all duration-200 ${
                isActive
                  ? "brand-chip text-primary font-medium shadow-[0_8px_18px_rgba(76,141,255,0.18)]"
                  : "text-muted-foreground hover:bg-white/[0.06] hover:text-foreground hover:translate-x-0.5"
              }`}
            >
              <span className={`inline-flex h-6 w-6 items-center justify-center rounded-lg transition-colors ${
                isActive ? "bg-black/20" : "bg-white/[0.04] group-hover:bg-white/[0.08]"
              }`}>
                <item.icon className={`h-3.5 w-3.5 transition-colors ${isActive ? "text-primary" : "text-muted-foreground group-hover:text-foreground"}`} />
              </span>
              {item.label}
            </Link>
          );
        })}

        {(user?.role === "admin" || user?.role === "super_admin") && (
          <>
            <div className="my-3 border-t border-sidebar-border" />
            <Link
              href="/admin"
              className={`group flex items-center gap-3 rounded-xl px-3 py-2.5 text-sm transition-all duration-200 ${
                pathname.startsWith("/admin")
                  ? "brand-chip text-primary font-medium shadow-[0_8px_18px_rgba(76,141,255,0.18)]"
                  : "text-muted-foreground hover:bg-white/[0.06] hover:text-foreground hover:translate-x-0.5"
              }`}
            >
              <span className={`inline-flex h-6 w-6 items-center justify-center rounded-lg transition-colors ${
                pathname.startsWith("/admin") ? "bg-black/20" : "bg-white/[0.04] group-hover:bg-white/[0.08]"
              }`}>
                <Shield className={`h-3.5 w-3.5 transition-colors ${pathname.startsWith("/admin") ? "text-primary" : "text-muted-foreground group-hover:text-foreground"}`} />
              </span>
              어드민
            </Link>
          </>
        )}
      </nav>

      <div className="border-t border-sidebar-border/80 p-4">
        <div className="flex items-center gap-3">
          <div className="brand-chip flex h-8 w-8 items-center justify-center rounded-full text-sm font-medium text-primary">
            {user?.nickname?.[0] ?? "U"}
          </div>
          <div className="flex-1 min-w-0">
            <p className="text-sm font-medium text-foreground truncate">
              {user?.nickname ?? "사용자"}
            </p>
            {(user?.role === "admin" || user?.role === "super_admin") && (
              <span className="text-xs text-primary/80">{user.role === "super_admin" ? "super admin" : "admin"}</span>
            )}
          </div>
        </div>
      </div>
    </aside>
  );
}
