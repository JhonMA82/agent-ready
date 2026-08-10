# Driven Sync Cohorts

Unseeded repository content for the driven sync proof (`TestDrivenSync` in
`internal/app/driven_sync_test.go`). The harness copies each cohort into a
fresh temporary Git repository, installs the harness with `agent-ready init`,
saves a deterministic checkpoint baseline, applies the cohort's post-baseline
mutation, and runs the real `opencode run --command agent-ready sync`. No
conclusions, expected outputs, or verdicts are seeded anywhere in these
fixtures; the structural oracle observes only events, state, and artifacts the
sync itself produces.

- `lockfile/` — a Go module with a package.json but no lockfile at baseline;
  the test adds `package-lock.json` after the baseline, so the sync MUST
  reassess tool capabilities with reasons.
- `prose/` — a Go module with a README; the test edits only the README prose
  after the baseline, so the sync MUST record why reassessment was
  unnecessary.
