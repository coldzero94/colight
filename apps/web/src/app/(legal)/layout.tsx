import Link from "next/link";

export default function LegalLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <div className="min-h-screen bg-background">
      <header className="border-b border-border px-6 py-4">
        <Link href="/" className="text-xl font-bold text-foreground tracking-tight">
          Colight
        </Link>
      </header>
      <main className="mx-auto max-w-3xl px-6 py-12">{children}</main>
    </div>
  );
}
