"use client";

import { useState, useSyncExternalStore } from "react";
import { createPortal } from "react-dom";
import Link from "next/link";
import { usePathname } from "next/navigation";
import { ClipboardList, Mic, Building2, PenTool, LayoutDashboard, Shield, Menu, X, User } from "lucide-react";
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

const emptySubscribe = () => () => {};

export function MobileNav() {
  const [open, setOpen] = useState(false);
  const mounted = useSyncExternalStore(emptySubscribe, () => true, () => false);
  const pathname = usePathname();
  const { user } = useAuthStore();

  return (
    <>
      <button
        onClick={() => setOpen(true)}
        className="brand-chip lg:hidden rounded-lg p-2 text-primary hover:brightness-110 transition-all"
        aria-label="메뉴 열기"
      >
        <Menu className="h-5 w-5" />
      </button>

      {mounted && createPortal(
        <>
          {/* Overlay */}
          {open && (
            <div
              className="fixed inset-0 bg-black/60 backdrop-blur-sm z-40 lg:hidden"
              onClick={() => setOpen(false)}
            />
          )}

          {/* Sheet */}
          <div
            className={`brand-surface fixed top-0 left-0 h-full w-72 bg-sidebar/95 border-r border-sidebar-border z-50 transition-transform duration-200 lg:hidden ${
              open ? "translate-x-0" : "-translate-x-full"
            }`}
          >
            <div className="p-5 border-b border-sidebar-border flex items-center justify-between">
              <div className="flex items-center gap-2">
                <ColightLogo size={24} />
                <span className="text-xl font-bold font-display text-foreground tracking-tight">Colight</span>
              </div>
              <button
                onClick={() => setOpen(false)}
                className="p-1 text-muted-foreground hover:text-foreground transition-colors"
              >
                <X className="h-5 w-5" />
              </button>
            </div>

            <nav className="py-4 px-3 space-y-1">
              {menuItems.map((item) => {
                const isActive = pathname.startsWith(item.href);
                return (
                  <Link
                    key={item.href}
                    href={item.href}
                    onClick={() => setOpen(false)}
                    className={`group flex items-center gap-3 px-3 py-2.5 rounded-xl text-sm transition-all duration-200 ${
                      isActive
                        ? "brand-chip text-primary font-medium"
                        : "text-muted-foreground hover:bg-white/[0.06] hover:text-foreground"
                    }`}
                  >
                    <span className={`inline-flex h-6 w-6 items-center justify-center rounded-lg ${
                      isActive ? "bg-black/20" : "bg-white/[0.04] group-hover:bg-white/[0.08]"
                    }`}>
                      <item.icon className="h-3.5 w-3.5" />
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
                    onClick={() => setOpen(false)}
                    className="group flex items-center gap-3 px-3 py-2.5 rounded-xl text-sm text-muted-foreground hover:bg-white/[0.06] hover:text-foreground transition-all duration-200"
                  >
                    <span className="inline-flex h-6 w-6 items-center justify-center rounded-lg bg-white/[0.04] group-hover:bg-white/[0.08]">
                      <Shield className="h-3.5 w-3.5" />
                    </span>
                    어드민
                  </Link>
                </>
              )}
            </nav>
          </div>
        </>,
        document.body
      )}
    </>
  );
}
