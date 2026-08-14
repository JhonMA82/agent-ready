import { createFileRoute } from "@tanstack/react-router";
import { DashboardShell } from "@/components/dashboard-shell";

export const Route = createFileRoute("/dashboard/")({
  component: DefaultDashboard,
});

function DefaultDashboard() {
  return (
    <DashboardShell title="Default dashboard">
      <Metric label="Active users" value="1,204" />
      <Metric label="Sessions" value="8,431" />
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
