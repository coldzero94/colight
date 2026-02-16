export default function AuthLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <div className="min-h-screen flex items-center justify-center bg-background px-4">
      <div className="w-full max-w-md">
        <div className="text-center mb-8">
          <h1 className="text-3xl font-bold text-foreground tracking-tight">
            Colight
          </h1>
          <p className="mt-2 text-muted-foreground">AI 자소서 코칭 플랫폼</p>
        </div>
        <div className="rounded-2xl border border-border bg-card p-8 shadow-xl shadow-black/20">
          {children}
        </div>
      </div>
    </div>
  );
}
