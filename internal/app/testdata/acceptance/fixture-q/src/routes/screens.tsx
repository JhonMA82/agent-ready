import { createFileRoute } from "@tanstack/react-router";

export const Route = createFileRoute("/screens")({
  component: ScreensPage,
});

function ScreensPage() {
  return (
    <section aria-label="screens list">
      <h1>Screens</h1>
    </section>
  );
}
