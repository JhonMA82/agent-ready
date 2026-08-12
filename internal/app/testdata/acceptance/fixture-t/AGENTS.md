# AGENTS.md — presets workspace

Global guidance: package manager, core commands, structure. The procedures
below are deterministic; the repository owns equivalent scripts.

## Package manager
- npm is the package manager.

## Core commands
- `npm run dev` starts the local dev server.
- `npm run start` serves the built presets.
- `npm run generate-routes` regenerates routes from presets.
- `npm run validate-presets` validates generated presets.

## Workspace structure
- `presets/` — preset definitions and templates.
- `generated/` — generated output, never edited by hand.

## Generate routes (deterministic procedure)
1. Ensure the working tree is clean: `git status --porcelain` is empty.
2. Read `presets/manifest.json` and list the preset ids in order.
3. For each preset id, resolve its template under `presets/templates/`.
4. Substitute the `{{id}}`, `{{title}}`, and `{{schema}}` placeholders.
5. Write the rendered file to `generated/routes/<id>.json`.
6. Validate JSON syntax with `node -e` on every generated file.
7. Re-run the sort check with `npm run validate-presets`.
8. If validation fails, fix the offending preset and restart from step 1.
9. When the tree is clean and all presets validate, commit the change.

## Validate presets (deterministic procedure)
1. Load every file under `generated/routes/`.
2. Fail on the first file that does not parse as JSON.
3. Check each id matches `^[a-z0-9-]+$`.
4. Check every referenced template exists under `presets/templates/`.
5. Check the sort order matches `presets/manifest.json`.
6. Print a one-line summary and exit 0.

## Determinism note
- These procedures require no semantic judgment: fixed steps, mechanically
  checkable results, automated validation.
- The equivalent scripts already exist (scripts/generate-routes,
  scripts/validate-presets); instruction-based repetition is more fragile.

## Constraints
- Never edit generated output by hand.
- Keep the procedures in sync with the scripts; the script is the source.

## Router
- Run `npm run generate-routes` to regenerate routes; run
  `npm run validate-presets` to validate presets.
