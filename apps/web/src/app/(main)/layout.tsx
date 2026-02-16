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
      <div className="app-shell relative flex h-screen overflow-hidden">
        <div className="app-noise pointer-events-none absolute inset-0 -z-10" />
        <div className="pointer-events-none absolute -left-20 top-28 -z-10 h-72 w-72 rounded-full bg-primary/12 blur-[120px]" />
        <div className="pointer-events-none absolute -right-16 bottom-24 -z-10 h-72 w-72 rounded-full bg-cyan-400/10 blur-[120px]" />

        <div className="hidden lg:block relative z-20">
          <Sidebar />
        </div>

        <div className="relative z-10 flex flex-1 flex-col overflow-hidden">
          <Header />
          <main className="relative z-0 flex-1 overflow-auto p-6">{children}</main>
        </div>
      </div>
      <FeedbackButton />
    </AuthGuard>
  );
}
