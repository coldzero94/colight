import { ArrowRight } from "lucide-react";

interface EmptyStateProps {
  icon?: React.ReactNode;
  title: string;
  description?: string;
  action?: {
    label: string;
    onClick: () => void;
  };
}

export function EmptyState({ icon, title, description, action }: EmptyStateProps) {
  return (
    <div className="relative overflow-hidden rounded-2xl border border-border bg-card/70 px-6 py-16 text-center">
      <div className="pointer-events-none absolute inset-0 bg-[radial-gradient(circle_at_top,_rgba(56,189,248,0.2),_transparent_62%)]" />
      <div className="relative flex flex-col items-center justify-center gap-4">
        {icon && (
          <div className="inline-flex h-16 w-16 items-center justify-center rounded-2xl border border-white/15 bg-background/80 text-primary shadow-[0_0_30px_rgba(56,189,248,0.28)] animate-pulse-glow">
            {icon}
          </div>
        )}
        <h2 className="text-lg font-semibold text-foreground">{title}</h2>
        {description && <p className="max-w-md text-sm text-muted-foreground">{description}</p>}
        {action && (
          <button
            onClick={action.onClick}
            className="mt-2 inline-flex items-center gap-2 rounded-lg bg-primary px-4 py-2 text-sm font-medium text-primary-foreground transition-colors hover:bg-primary/90"
          >
            {action.label}
            <ArrowRight className="h-4 w-4" aria-hidden="true" />
          </button>
        )}
      </div>
    </div>
  );
}
