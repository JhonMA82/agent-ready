// legacy/deprecated: obsolete implementation pattern
// Old-style dashboard: inline styles, duplicated markup, no shared shell, no
// theme tokens. Do not use as a reference for new work.
import { createFileRoute } from "@tanstack/react-router";

export const Route = createFileRoute("/dashboard/legacy")({
  component: LegacyDashboard,
});

function LegacyDashboard() {
  return (
    <div style={{ display: "flex", flexDirection: "column", gap: 12 }}>
      <h2 style={{ fontSize: 18, color: "#111827" }}>Legacy dashboard</h2>
      <div style={{ padding: 12, border: "1px solid #d1d5db" }}>
        <span>old metric tile</span>
      </div>
      <div style={{ padding: 12, border: "1px solid #d1d5db" }}>
        <span>old metric tile</span>
      </div>
    </div>
  );
}
