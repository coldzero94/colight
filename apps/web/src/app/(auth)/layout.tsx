export default function AuthLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <div className="min-h-screen relative flex items-center justify-center bg-background px-4">
      <div className="absolute inset-0 -z-10 overflow-hidden">
        <div className="absolute top-1/3 left-1/2 -translate-x-1/2 -translate-y-1/2 w-[400px] h-[300px] bg-primary/[0.08] rounded-full blur-[100px] animate-pulse-glow" />
      </div>
      <div className="w-full max-w-md">
        <div className="text-center mb-8">
          <h1 className="text-3xl font-bold font-display text-foreground tracking-tight">
            Colight
          </h1>
          <p className="mt-2 text-muted-foreground">AI 자소서 코칭 플랫폼</p>
        </div>
        <div className="rounded-2xl border border-white/[0.08] bg-card p-8 shadow-xl shadow-black/20">
          {children}
        </div>
      </div>
    </div>
  );
}
