# AGENTS.md — tanstack-shadcn admin starter

This repository is a boilerplate: downstream users copy it and then extend
the scaffold (new screens, features, presets). Guidance below is short on
purpose — it loads every session — so it covers only what applies to most
tasks and routes to more specific context.

## Package manager and core commands

- npm is the package manager; never commit yarn or pnpm lockfiles.
- `npm run dev` starts the dev server.
- `npm run build` produces the production bundle.
- `npm run start` serves the built bundle.
- `npm run generate-routes` regenerates route manifests from `src/routes`.

## Workspace structure

- `src/routes` — file-based routes; one file is one route.
- `src/features` — feature modules; one folder per feature.
- `src/components` — shared presentational components.
- `src/lib` — utilities, theme tokens, and the shadcn `cn` helper.

## Critical constraints

- Do not edit generated files under `generated/`; edit sources, then run
  `npm run generate-routes` to regenerate them.
- Routes are server-first: SSR by default, client islands only where
  required by the feature.
- Keep theme tokens in `src/lib/theme`; never hardcode colors or radii.

## Adding a dashboard screen — 13-step procedure

1. Choose the route path from the sidebar entry the screen belongs to.
2. Create the route file under `src/routes` using the file-based naming
   convention (`<path>.tsx` or `<path>.<detail>.tsx` for modals).
3. Add the route to the sidebar registration in `src/components/sidebar`.
4. Place feature-specific code in `src/features/<feature>/`, never in the
   route file itself.
5. Give the route an `onLoad` that prefetches the feature's data source.
6. Handle loading state with the shared `ScreenLoader` component.
7. Handle error state with the shared `ScreenErrorBoundary` component.
8. Use theme tokens for every style decision; verify contrast against the
   token palette.
9. Match the responsive conventions: stack below `md`, grid above.
10. Reference the canonical screen example for the layout being built.
11. Run the validation commands listed below before committing.
12. Regenerate routes with `npm run generate-routes` if the route manifest
    changed.
13. Verify the screen in both server render and client navigation.

## SSR / loading / error state expectations

- Every route must render on the server without accessing `window`.
- Data fetching happens in the route `onLoad`, not in the component.
- Loading UI must not flash on server-first navigation.
- Error UI must offer a retry that re-runs the route load.

## Theme token usage

- Colors: `--color-*` tokens from `src/lib/theme/tokens.css`.
- Typography: `text-sm`/`text-base` from the token scale, never raw px.
- Radii and shadows come from the design-token presets.
- New variants of an existing component need a preset, not a new token.

## Responsive conventions

- Layouts stack on small screens and become multi-column at `md`.
- Tables degrade to card lists below `sm`.
- Sidebar collapses to a drawer below `lg`.
- Touch targets stay at least 44px on every breakpoint.

## Canonical screen examples

- `src/routes/screens.tsx` — list screen with filters and pagination.
- `src/routes/index.tsx` — dashboard summary screen.
- `src/features/screens/` — feature module layout to copy for new features.
- `src/routes/dashboard/` — dashboard screens (default, finance, operations)
  compose through `DashboardShell` with theme tokens; `legacy.tsx` is
  deprecated and must not be copied. For new UI/features, select the closest
  canonical example before implementation.

## Validation commands

- `npm run generate-routes` — regenerate the route manifest.
- `npm run build` — type-check and produce the production bundle.
- `npm test` — run the unit suite (no output-volume workflow today).

## What downstream users should edit

- `src/features/**` — add features here.
- `src/routes/**` — add screens here.
- `components.json` — keep in sync when adding shadcn components.
- Anything else in the scaffold should stay as shipped unless an
  architecture decision changes it.
