# AGENTS.md — legacy workspace platform

## Package manager and core commands
- npm is the package manager; yarn and pnpm lockfiles are not supported.
- `npm run dev` starts the local dev server.
- `npm run start` serves the production build.
- `npm run generate` refreshes generated client bindings.

## Workspace structure
- `src/` — application source; one module per bounded context.
- `src/gateways/` — adapters to external services.
- `src/domain/` — pure domain models, no I/O.
- `src/persistence/` — repositories and migrations.
- `src/presentation/` — screens and UI state.
- `tools/` — operational scripts run by humans or CI.
- `docs/` — architecture records and runbooks.

## Critical constraints
- Migrations are append-only; never edit or delete a committed migration.
- Never commit generated bindings; regenerate before release.
- The platform serves two tenants: north and south.
- All writes pass through the domain layer, never the gateways directly.

## Where to find docs
- `docs/architecture-map.md` — module boundaries and data flow.
- `docs/canonical-examples.md` — reference implementations.
- `docs/known-pitfalls.md` — common failure modes.
- `docs/edge-cases.md` — unusual but supported scenarios.
- `docs/runbooks/` — operational runbooks for humans and agents.

## Which skills to use
- Use the db-migration skill for schema changes.
- Use the deploy-runbook skill for releases.
- Use the incident skill for post-incident reviews.

## Forbidden operations
- Never delete a committed migration file.
- Never bypass the release freeze without a signed exception.
- Never run bulk updates against production directly.
- Never add a new external gateway without an architecture record.

## Versioning policy
- Breaking schema changes require a new minor version plus a migration.
- Feature flags gate new behavior for at least one release cycle.
- Deprecations carry a removal notice for two releases.

## Maintenance rule
- If a section grows past ~10 lines, evaluate extraction to a skill.
- This file loads every session; move detail to docs on demand.

# Migration workflow

Full workflow for schema changes: preconditions, steps, rollback.
Each step is mandatory; skip none.
Follow the order exactly; the rollback window closes at step 24.

## Preconditions
- The migration is reviewed by the platform team before execution.
- A full backup of the source schema exists and is verified.
- The target version is recorded in the migration ledger.
- The change request links the architecture record.
- The rollback window is agreed and published.
- The standby replica is healthy.

## Workflow steps
1. draft the migration file.
   - run from `tools/` with the platform user, never root.
   - record the start and end timestamp in the ledger.
2. register the migration id.
   - run from `tools/` with the platform user, never root.
   - record the start and end timestamp in the ledger.
3. model the domain change.
   - run from `tools/` with the platform user, never root.
   - record the start and end timestamp in the ledger.
4. map old and new schemas.
   - run from `tools/` with the platform user, never root.
   - record the start and end timestamp in the ledger.
5. write the up-script.
   - run from `tools/` with the platform user, never root.
   - record the start and end timestamp in the ledger.
6. write the down-script.
   - run from `tools/` with the platform user, never root.
   - record the start and end timestamp in the ledger.
7. review idempotency.
   - run from `tools/` with the platform user, never root.
   - record the start and end timestamp in the ledger.
8. check tenant isolation.
   - run from `tools/` with the platform user, never root.
   - record the start and end timestamp in the ledger.
9. dry-run against staging.
   - run from `tools/` with the platform user, never root.
   - record the start and end timestamp in the ledger.
10. inspect the staging diff.
   - run from `tools/` with the platform user, never root.
   - record the start and end timestamp in the ledger.
11. freeze the affected tables.
   - run from `tools/` with the platform user, never root.
   - record the start and end timestamp in the ledger.
12. run the migration.
   - run from `tools/` with the platform user, never root.
   - record the start and end timestamp in the ledger.
13. verify row counts.
   - run from `tools/` with the platform user, never root.
   - record the start and end timestamp in the ledger.
14. verify referential integrity.
   - run from `tools/` with the platform user, never root.
   - record the start and end timestamp in the ledger.
15. refresh materialized views.
   - run from `tools/` with the platform user, never root.
   - record the start and end timestamp in the ledger.
16. recompile stored procedures.
   - run from `tools/` with the platform user, never root.
   - record the start and end timestamp in the ledger.
17. update the domain models.
   - run from `tools/` with the platform user, never root.
   - record the start and end timestamp in the ledger.
18. update the gateways.
   - run from `tools/` with the platform user, never root.
   - record the start and end timestamp in the ledger.
19. bump the schema version.
   - run from `tools/` with the platform user, never root.
   - record the start and end timestamp in the ledger.
20. update the architecture record.
   - run from `tools/` with the platform user, never root.
   - record the start and end timestamp in the ledger.
21. annotate the changelog.
   - run from `tools/` with the platform user, never root.
   - record the start and end timestamp in the ledger.
22. run the full test suite.
   - run from `tools/` with the platform user, never root.
   - record the start and end timestamp in the ledger.
23. load-test the new path.
   - run from `tools/` with the platform user, never root.
   - record the start and end timestamp in the ledger.
24. monitor error rates.
   - run from `tools/` with the platform user, never root.
   - record the start and end timestamp in the ledger.
25. record the completion.
   - run from `tools/` with the platform user, never root.
   - record the start and end timestamp in the ledger.
26. update the runbook.
   - run from `tools/` with the platform user, never root.
   - record the start and end timestamp in the ledger.
27. alert the platform channel.
   - run from `tools/` with the platform user, never root.
   - record the start and end timestamp in the ledger.
28. clean up staging artifacts.
   - run from `tools/` with the platform user, never root.
   - record the start and end timestamp in the ledger.
29. mark the change request done.
   - run from `tools/` with the platform user, never root.
   - record the start and end timestamp in the ledger.
30. schedule the next review.
   - run from `tools/` with the platform user, never root.
   - record the start and end timestamp in the ledger.
31. remove the temporary indexes.
   - run from `tools/` with the platform user, never root.
   - record the start and end timestamp in the ledger.
32. vacuum the affected tables.
   - run from `tools/` with the platform user, never root.
   - record the start and end timestamp in the ledger.
33. reset the connection pools.
   - run from `tools/` with the platform user, never root.
   - record the start and end timestamp in the ledger.
34. invalidate cached queries.
   - run from `tools/` with the platform user, never root.
   - record the start and end timestamp in the ledger.
35. replay the transaction log.
   - run from `tools/` with the platform user, never root.
   - record the start and end timestamp in the ledger.
36. verify the standby replica.
   - run from `tools/` with the platform user, never root.
   - record the start and end timestamp in the ledger.
37. compare checksums with the backup.
   - run from `tools/` with the platform user, never root.
   - record the start and end timestamp in the ledger.
38. confirm the rollback window.
   - run from `tools/` with the platform user, never root.
   - record the start and end timestamp in the ledger.
39. document the exact timing.
   - run from `tools/` with the platform user, never root.
   - record the start and end timestamp in the ledger.
40. archive the migration scripts.
   - run from `tools/` with the platform user, never root.
   - record the start and end timestamp in the ledger.
41. link the migration to the release.
   - run from `tools/` with the platform user, never root.
   - record the start and end timestamp in the ledger.
42. update the dependency map.
   - run from `tools/` with the platform user, never root.
   - record the start and end timestamp in the ledger.
43. review the generated bindings.
   - run from `tools/` with the platform user, never root.
   - record the start and end timestamp in the ledger.
44. regenerate client bindings.
   - run from `tools/` with the platform user, never root.
   - record the start and end timestamp in the ledger.
45. smoke-test the admin screens.
   - run from `tools/` with the platform user, never root.
   - record the start and end timestamp in the ledger.
# Release workflow

Complete release procedure: build, gate, freeze, ship, verify.
Run the steps in order; every gate is mandatory.

## Preconditions
- The release branch is cut from main.
- Every change is feature-flagged or release-noted.
- The deploy-runbook skill is loaded before the first step.
- The dashboards cover every new endpoint.

## Release steps
1. verify the release branch.
   - stop on any failed gate; escalate to the release owner.
   - log every action with the exact command used.
   - the freeze window is 60 minutes; never extend it silently.
2. bump the version file.
   - stop on any failed gate; escalate to the release owner.
   - log every action with the exact command used.
   - the freeze window is 60 minutes; never extend it silently.
3. generate the changelog.
   - stop on any failed gate; escalate to the release owner.
   - log every action with the exact command used.
   - the freeze window is 60 minutes; never extend it silently.
4. update the migration ledger.
   - stop on any failed gate; escalate to the release owner.
   - log every action with the exact command used.
   - the freeze window is 60 minutes; never extend it silently.
5. run the full test suite.
   - stop on any failed gate; escalate to the release owner.
   - log every action with the exact command used.
   - the freeze window is 60 minutes; never extend it silently.
6. build the production bundle.
   - stop on any failed gate; escalate to the release owner.
   - log every action with the exact command used.
   - the freeze window is 60 minutes; never extend it silently.
7. sign the bundle manifest.
   - stop on any failed gate; escalate to the release owner.
   - log every action with the exact command used.
   - the freeze window is 60 minutes; never extend it silently.
8. upload the bundle to staging.
   - stop on any failed gate; escalate to the release owner.
   - log every action with the exact command used.
   - the freeze window is 60 minutes; never extend it silently.
9. run the staging smoke suite.
   - stop on any failed gate; escalate to the release owner.
   - log every action with the exact command used.
   - the freeze window is 60 minutes; never extend it silently.
10. review the staging dashboards.
   - stop on any failed gate; escalate to the release owner.
   - log every action with the exact command used.
   - the freeze window is 60 minutes; never extend it silently.
11. open the release freeze.
   - stop on any failed gate; escalate to the release owner.
   - log every action with the exact command used.
   - the freeze window is 60 minutes; never extend it silently.
12. notify the platform channel.
   - stop on any failed gate; escalate to the release owner.
   - log every action with the exact command used.
   - the freeze window is 60 minutes; never extend it silently.
13. promote the bundle to production.
   - stop on any failed gate; escalate to the release owner.
   - log every action with the exact command used.
   - the freeze window is 60 minutes; never extend it silently.
14. verify the production checksums.
   - stop on any failed gate; escalate to the release owner.
   - log every action with the exact command used.
   - the freeze window is 60 minutes; never extend it silently.
15. watch the error-rate dashboards.
   - stop on any failed gate; escalate to the release owner.
   - log every action with the exact command used.
   - the freeze window is 60 minutes; never extend it silently.
16. confirm the freeze window.
   - stop on any failed gate; escalate to the release owner.
   - log every action with the exact command used.
   - the freeze window is 60 minutes; never extend it silently.
17. close the release freeze.
   - stop on any failed gate; escalate to the release owner.
   - log every action with the exact command used.
   - the freeze window is 60 minutes; never extend it silently.
18. annotate the release tag.
   - stop on any failed gate; escalate to the release owner.
   - log every action with the exact command used.
   - the freeze window is 60 minutes; never extend it silently.
19. post the summary to the team channel.
   - stop on any failed gate; escalate to the release owner.
   - log every action with the exact command used.
   - the freeze window is 60 minutes; never extend it silently.
20. archive the release artifacts.
   - stop on any failed gate; escalate to the release owner.
   - log every action with the exact command used.
   - the freeze window is 60 minutes; never extend it silently.
21. schedule the post-release review.
   - stop on any failed gate; escalate to the release owner.
   - log every action with the exact command used.
   - the freeze window is 60 minutes; never extend it silently.
22. record the release in the ledger.
   - stop on any failed gate; escalate to the release owner.
   - log every action with the exact command used.
   - the freeze window is 60 minutes; never extend it silently.
# Examples and edge cases

Canonical examples and unusual scenarios that agents must recognize.

## Edge cases
1. migration applies twice.
   - trigger: the condition that exposes the case.
   - impact: what breaks or degrades.
   - handling: the documented recovery path.
   - verification: the check that proves recovery.
2. migration fails halfway.
   - trigger: the condition that exposes the case.
   - impact: what breaks or degrades.
   - handling: the documented recovery path.
   - verification: the check that proves recovery.
3. tenant data mixed in one table.
   - trigger: the condition that exposes the case.
   - impact: what breaks or degrades.
   - handling: the documented recovery path.
   - verification: the check that proves recovery.
4. gateway timeout on retry.
   - trigger: the condition that exposes the case.
   - impact: what breaks or degrades.
   - handling: the documented recovery path.
   - verification: the check that proves recovery.
5. generated bindings out of date.
   - trigger: the condition that exposes the case.
   - impact: what breaks or degrades.
   - handling: the documented recovery path.
   - verification: the check that proves recovery.
6. release freeze missed.
   - trigger: the condition that exposes the case.
   - impact: what breaks or degrades.
   - handling: the documented recovery path.
   - verification: the check that proves recovery.
7. rollback after a failed release.
   - trigger: the condition that exposes the case.
   - impact: what breaks or degrades.
   - handling: the documented recovery path.
   - verification: the check that proves recovery.
8. bulk update touches the wrong tenant.
   - trigger: the condition that exposes the case.
   - impact: what breaks or degrades.
   - handling: the documented recovery path.
   - verification: the check that proves recovery.
9. duplicate transaction replay.
   - trigger: the condition that exposes the case.
   - impact: what breaks or degrades.
   - handling: the documented recovery path.
   - verification: the check that proves recovery.
10. standby replica lag.
   - trigger: the condition that exposes the case.
   - impact: what breaks or degrades.
   - handling: the documented recovery path.
   - verification: the check that proves recovery.
11. feature flag stuck on.
   - trigger: the condition that exposes the case.
   - impact: what breaks or degrades.
   - handling: the documented recovery path.
   - verification: the check that proves recovery.
12. deprecated endpoint still called.
   - trigger: the condition that exposes the case.
   - impact: what breaks or degrades.
   - handling: the documented recovery path.
   - verification: the check that proves recovery.
13. cache invalidation misses.
   - trigger: the condition that exposes the case.
   - impact: what breaks or degrades.
   - handling: the documented recovery path.
   - verification: the check that proves recovery.
14. connection pool exhaustion.
   - trigger: the condition that exposes the case.
   - impact: what breaks or degrades.
   - handling: the documented recovery path.
   - verification: the check that proves recovery.
15. vacuum blocks the migration.
   - trigger: the condition that exposes the case.
   - impact: what breaks or degrades.
   - handling: the documented recovery path.
   - verification: the check that proves recovery.
16. checksum mismatch on restore.
   - trigger: the condition that exposes the case.
   - impact: what breaks or degrades.
   - handling: the documented recovery path.
   - verification: the check that proves recovery.
17. staging differs from production.
   - trigger: the condition that exposes the case.
   - impact: what breaks or degrades.
   - handling: the documented recovery path.
   - verification: the check that proves recovery.
18. changelog entry missing.
   - trigger: the condition that exposes the case.
   - impact: what breaks or degrades.
   - handling: the documented recovery path.
   - verification: the check that proves recovery.
19. architecture record out of date.
   - trigger: the condition that exposes the case.
   - impact: what breaks or degrades.
   - handling: the documented recovery path.
   - verification: the check that proves recovery.
20. runbook contradicts the code.
   - trigger: the condition that exposes the case.
   - impact: what breaks or degrades.
   - handling: the documented recovery path.
   - verification: the check that proves recovery.
21. migration id collision.
   - trigger: the condition that exposes the case.
   - impact: what breaks or degrades.
   - handling: the documented recovery path.
   - verification: the check that proves recovery.
22. down-script deletes data.
   - trigger: the condition that exposes the case.
   - impact: what breaks or degrades.
   - handling: the documented recovery path.
   - verification: the check that proves recovery.
23. monitoring gap after release.
   - trigger: the condition that exposes the case.
   - impact: what breaks or degrades.
   - handling: the documented recovery path.
   - verification: the check that proves recovery.
24. error-rate spike during freeze.
   - trigger: the condition that exposes the case.
   - impact: what breaks or degrades.
   - handling: the documented recovery path.
   - verification: the check that proves recovery.
25. lockfile drift between branches.
   - trigger: the condition that exposes the case.
   - impact: what breaks or degrades.
   - handling: the documented recovery path.
   - verification: the check that proves recovery.
26. node version mismatch.
   - trigger: the condition that exposes the case.
   - impact: what breaks or degrades.
   - handling: the documented recovery path.
   - verification: the check that proves recovery.
27. generated manifest unreadable.
   - trigger: the condition that exposes the case.
   - impact: what breaks or degrades.
   - handling: the documented recovery path.
   - verification: the check that proves recovery.
28. sidebar entry points to a dead route.
   - trigger: the condition that exposes the case.
   - impact: what breaks or degrades.
   - handling: the documented recovery path.
   - verification: the check that proves recovery.
29. preset schema changed without a variant.
   - trigger: the condition that exposes the case.
   - impact: what breaks or degrades.
   - handling: the documented recovery path.
   - verification: the check that proves recovery.
30. canonical example diverged.
   - trigger: the condition that exposes the case.
   - impact: what breaks or degrades.
   - handling: the documented recovery path.
   - verification: the check that proves recovery.
31. docs describe a removed module.
   - trigger: the condition that exposes the case.
   - impact: what breaks or degrades.
   - handling: the documented recovery path.
   - verification: the check that proves recovery.
32. skill references a deleted file.
   - trigger: the condition that exposes the case.
   - impact: what breaks or degrades.
   - handling: the documented recovery path.
   - verification: the check that proves recovery.
33. release tag points at the wrong commit.
   - trigger: the condition that exposes the case.
   - impact: what breaks or degrades.
   - handling: the documented recovery path.
   - verification: the check that proves recovery.
34. backup restored to staging.
   - trigger: the condition that exposes the case.
   - impact: what breaks or degrades.
   - handling: the documented recovery path.
   - verification: the check that proves recovery.
35. tenant-specific defaults leaking.
   - trigger: the condition that exposes the case.
   - impact: what breaks or degrades.
   - handling: the documented recovery path.
   - verification: the check that proves recovery.
36. timezone drift in the ledger.
   - trigger: the condition that exposes the case.
   - impact: what breaks or degrades.
   - handling: the documented recovery path.
   - verification: the check that proves recovery.
37. frozen tables block the nightly job.
   - trigger: the condition that exposes the case.
   - impact: what breaks or degrades.
   - handling: the documented recovery path.
   - verification: the check that proves recovery.
38. quote-style change in scripts.
   - trigger: the condition that exposes the case.
   - impact: what breaks or degrades.
   - handling: the documented recovery path.
   - verification: the check that proves recovery.
39. empty result set treated as failure.
   - trigger: the condition that exposes the case.
   - impact: what breaks or degrades.
   - handling: the documented recovery path.
   - verification: the check that proves recovery.
