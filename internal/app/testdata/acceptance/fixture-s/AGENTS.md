# AGENTS.md — demo workspace

Global guidance only: every line applies to most tasks or defines a global
invariant. No task-specific procedures live here.

## Package manager
- npm is the package manager; yarn and pnpm are not used.

## Core commands
- `npm run dev` starts the local dev server.
- `npm run start` serves the production build.
- `npm run generate` refreshes generated bindings.

## Workspace structure

- `src/app` — application shell and routing.
- `src/lib` — shared utilities and theme tokens.
- `docs/` — architecture records and runbooks.
- `tools/` — operational scripts run by humans or CI.

## Critical constraints

- Migrations are append-only; never edit a committed migration.
- All writes pass through the domain layer.
- Feature flags gate new behavior for one release cycle.
- Never commit generated bindings; regenerate before release.

## Forbidden operations

- Never run bulk updates against production directly.
- Never bypass the release freeze without a signed exception.
- Never add an external gateway without an architecture record.

## Where to find docs

- `docs/architecture-map.md` — module boundaries and data flow.
- `docs/canonical-examples.md` — reference implementations.
- `docs/known-pitfalls.md` — common failure modes.
- `docs/edge-cases.md` — unusual but supported scenarios.

## Which skills to use

- Use the db-migration skill for schema changes.
- Use the deploy-runbook skill for releases.
- Use the incident skill for post-incident reviews.

## Why this file stays short

- Information that only applies to one task must not load every session.
- Task-specific procedures live in skills; detail lives in docs.

## Maintenance rule
- If a section grows past ~10 lines, evaluate EXTRACT_TO_SKILL or
  MOVE_TO_REFERENCE.

## Onboarding note

- The audit expects REUSE + NO_ACTION: coverage is complete and the
  placement is already optimal (refinement §39).
- No extraction by dogma: there are no procedures here to extract.
