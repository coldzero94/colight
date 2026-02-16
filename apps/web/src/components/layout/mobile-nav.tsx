"use client";

import { useState, useEffect } from "react";
import { createPortal } from "react-dom";
import Link from "next/link";
import { usePathname } from "next/navigation";
import { ClipboardList, Mic, Building2, PenTool, LayoutDashboard, Shield, Menu, X } from "lucide-react";
import { useAuthStore } from "@/stores/auth-store";

const menuItems = [
  { label: "경험 관리", href: "/experiences", icon: ClipboardList },
  { label: "경험 인터뷰", href: "/interview", icon: Mic },
  { label: "기업 분석", href: "/analysis", icon: Building2 },
  { label: "자소서 코칭", href: "/coaching", icon: PenTool },
  { label: "대시보드", href: "/dashboard", icon: LayoutDashboard },
];

export function MobileNav() {
  const [open, setOpen] = useState(false);
  const [mounted, setMounted] = useState(false);
  const pathname = usePathname();
  const { user } = useAuthStore();

  useEffect(() => {
    setMounted(true);
  }, []);

  return (
    <>
      <button
        onClick={() => setOpen(true)}
        className="lg:hidden p-2 text-muted-foreground hover:text-foreground transition-colors"
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
            className={`fixed top-0 left-0 h-full w-72 bg-sidebar border-r border-sidebar-border z-50 transition-transform duration-200 lg:hidden ${
              open ? "translate-x-0" : "-translate-x-full"
            }`}
          >
            <div className="p-5 border-b border-sidebar-border flex items-center justify-between">
              <span className="text-xl font-bold text-foreground tracking-tight">Colight</span>
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
                    className={`flex items-center gap-3 px-3 py-2.5 rounded-lg text-sm transition-all duration-200 ${
                      isActive
                        ? "bg-primary/10 text-primary font-medium"
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
                    onClick={() => setOpen(false)}
                    className="flex items-center gap-3 px-3 py-2.5 rounded-lg text-sm text-muted-foreground hover:bg-white/[0.04] hover:text-foreground transition-all duration-200"
                  >
                    <Shield className="h-4 w-4" />
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
