import { createRootRoute } from "@tanstack/react-router";

export const Route = createRootRoute({
  component: () => (
    <main className="flex min-h-screen flex-col">
      <aside className="w-64 border-r" />
      <section className="flex-1 p-6" />
    </main>
  ),
});
