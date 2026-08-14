import { createFileRoute } from "@tanstack/react-router";
import { DashboardShell } from "@/components/dashboard-shell";

export const Route = createFileRoute("/dashboard/finance")({
  component: FinanceDashboard,
});

function FinanceDashboard() {
  return (
    <DashboardShell title="Finance">
      <Metric label="Revenue" value="€ 92,410" />
      <Metric label="Expenses" value="€ 41,208" />
    </DashboardShell>
  );
}

function Metric({ label, value }: { label: string; value: string }) {
  return (
    <div className="rounded-lg border p-4">
      <p className="text-sm text-muted">{label}</p>
      <p className="text-2xl font-semibold text-foreground">{value}</p>
    </div>
  );
}
