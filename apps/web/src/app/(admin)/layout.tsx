import { AuthGuard } from "@/components/auth/auth-guard";
import { AdminGuard } from "@/components/auth/admin-guard";
import { AdminSidebar } from "@/components/admin/admin-sidebar";

export default function AdminLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <AuthGuard>
      <AdminGuard>
        <div className="flex h-screen">
          <AdminSidebar />
          <main className="flex-1 overflow-auto bg-background p-6">
            {children}
          </main>
        </div>
      </AdminGuard>
    </AuthGuard>
  );
}
