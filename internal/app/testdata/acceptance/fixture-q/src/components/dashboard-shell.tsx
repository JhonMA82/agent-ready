import { cn } from "@/lib/utils";

// Shared dashboard layout shell: every current dashboard composes through it
// (semantic tokens only, responsive stack below md / grid above).
export function DashboardShell({
  title,
  children,
  className,
}: {
  title: string;
  children: React.ReactNode;
  className?: string;
}) {
  return (
    <section
      aria-label={title}
      className={cn("grid gap-6 p-6 md:grid-cols-2", className)}
    >
      <h1 className="text-xl font-semibold text-foreground">{title}</h1>
      {children}
    </section>
  );
}
