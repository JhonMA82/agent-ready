import { createFileRoute } from "@tanstack/react-router";
import { DashboardShell } from "@/components/dashboard-shell";

export const Route = createFileRoute("/dashboard/operations")({
  component: OperationsDashboard,
});

function OperationsDashboard() {
  return (
    <DashboardShell title="Operations">
      <Metric label="Uptime" value="99.98%" />
      <Metric label="Incidents" value="2" />
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
