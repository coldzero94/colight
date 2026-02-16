import { AuthGuard } from "@/components/auth/auth-guard";
import { Sidebar } from "@/components/layout/sidebar";
import { Header } from "@/components/layout/header";
import { FeedbackButton } from "@/components/feedback/feedback-button";

export default function MainLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <AuthGuard>
      <div className="flex h-screen">
        <div className="hidden lg:block">
          <Sidebar />
        </div>

        <div className="flex flex-1 flex-col overflow-hidden">
          <Header />
          <main className="relative z-0 flex-1 overflow-auto p-6">{children}</main>
        </div>
      </div>
      <FeedbackButton />
    </AuthGuard>
  );
}
