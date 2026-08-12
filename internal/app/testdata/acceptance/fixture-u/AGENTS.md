# AGENTS.md — shadcn component workspace

## Package manager
- npm is the package manager.

## Core commands
- `npm run dev` starts the local dev server.
- `npm run start` serves the built component library.

## Workspace structure
- `components/` — shared components; primitives under `components/ui/`.
- `components.json` — shadcn registry config, kept in sync.

## shadcn component work
- Use the canonical external shadcn skill for component work: it covers
  the add/update workflow, variant conventions, and registry sync.
- This file only routes to that skill; it does not duplicate the
  capability guidance here (router, not content).
- React patterns follow the canonical React skill; no repository-specific
  React guidance is maintained in this workspace.

## Constraints
- Never commit generated registry output.
- Keep `components.json` in sync when adding a component.
- If the canonical skill is unavailable, do not write a local wrapper;
  ask the orchestrator for the external reference instead.
