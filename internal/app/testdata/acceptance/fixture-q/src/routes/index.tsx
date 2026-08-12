import { createFileRoute } from "@tanstack/react-router";

export const Route = createFileRoute("/")({
  component: DashboardIndex,
});

function DashboardIndex() {
  return <section aria-label="dashboard summary">Dashboard</section>;
}
