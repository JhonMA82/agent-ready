# Tasks: Complete Agent-Ready V1 Ecosystem Intelligence

## Review Workload Forecast

| Field | Value |
|---|---|
| Estimated changed lines | ~2,950 |
| Review budget | 400 changed lines per slice |
| 400-line budget risk | High |
| Chained PRs recommended | Yes |
| Delivery strategy | auto-chain |
| Chain strategy | stacked-to-main |
| Suggested split | PR 1 → 2 → 3 → 4 → 5 → 6 → 7 → 8 → 9 |
| Execution / artifact store | auto / both |

Decision needed before apply: No
Chained PRs recommended: Yes
Chain strategy: stacked-to-main
400-line budget risk: High

First apply boundary: **Slice 1 only**. Order/line estimates: `1/320→2/230→3/360→4/260→5/360→6/260→7/390→8/370→9/400`.

### Suggested Work Units

| Unit | Goal | PR | Test | Runtime | Rollback |
|---|---|---|---|---|---|
| 1 | Ownership-safe upgrades | 1 | `go test ./internal/bootstrap ./internal/lifecycle` | N/A—library | bootstrap/lifecycle |
| 2 | Pruned presence | 2 | `go test ./internal/inventory` | N/A—library | inventory walk |
| 3 | Ecosystem facts | 3 | `go test ./internal/ecosystem ./internal/inventory` | N/A—library | ecosystem attachment |
| 4 | Manager conflicts | 4 | `go test ./internal/ecosystem` | N/A—library | manager resolver |
| 5 | Catalog truth | 5 | `go test ./internal/tools` | N/A—library | catalog metadata |
| 6 | Recommendations | 6 | `go test ./internal/tools` | N/A—library | recommend consumer |
| 7 | Driven audit | 7 | `AGENT_READY_DRIVEN_MODEL=<model> go test ./internal/lifecycle ./internal/app -run 'Test(EmbeddedAssetUpgradeGate|DrivenAudit)'` | unseeded OpenCode audit | audit assets/cohort |
| 8 | Driven sync | 8 | `AGENT_READY_DRIVEN_MODEL=<model> go test ./internal/lifecycle ./internal/app -run 'Test(EmbeddedAssetUpgradeGate|DrivenSync)'` | mutation-driven OpenCode sync | sync assets/cohorts |
| 9 | Driven review | 9 | `AGENT_READY_DRIVEN_MODEL=<model> go test ./internal/lifecycle ./internal/app -run 'Test(EmbeddedAssetUpgradeGate|DrivenReview|Acceptance)' && go test ./...` | grounded/ungrounded OpenCode review | review assets/cohorts |
| 10 | OpenCode version decoupling | 10 | `go test ./internal/opencode ./internal/app && AGENT_READY_DRIVEN_MODEL=<model> go test ./internal/app -run 'Test(DrivenAudit|DrivenSync|DrivenReview)'` | real OpenCode (any >= floor) | opencode preflight + driven tests |

## Phase 1: Ownership Foundation

- [x] 1.1 Test, then implement reconciliation in `internal/bootstrap/{bootstrap.go,bootstrap_test.go}` and `internal/lifecycle/{update.go,update_test.go}`. Parameterize `runEmbeddedAssetUpgradeGate(slice, changedAssets)` for advancement, collisions, new assets, protected state, ordering, and idempotence.

## Phase 2: Deterministic Facts

- [x] 2.1 Test, then add `Presence` and heavy-tree pruning in `internal/inventory/{inventory.go,inventory_test.go}` without changing legacy totals/JSON.
- [x] 2.2 Create tested `internal/ecosystem/{ecosystem.go,ecosystem_test.go}` and inventory attachment for ordered ecosystems, manifests, lockfiles, workspaces, wrappers, frameworks, build/test signals, and stable bytes.
- [x] 2.3 Test, then add manager confidence/conflicts in `internal/ecosystem`, covering wrapper precedence, pyproject ambiguity, and unresolved pnpm+Bun without selection.

## Phase 3: Tool Truth

- [x] 3.1 Test ordered families and independent support states in `internal/tools/{catalog.go,detect.go,tools_test.go,install_test.go}`, preserving V1 and recipe safeguards.
- [x] 3.2 Test, then add grounded, empty, conflict, detect-only, and provider-without-lifecycle cases in `internal/tools/{recommend.go,recommend_test.go}`.

## Phase 4: Embedded Contracts and Driven Proof

- [x] 4.1 RED-test `internal/app/driven_audit_test.go` Git selectors: accept relative/absolute temp roots; fail closed outside-root/missing-Git.
- [x] 4.2 Update `internal/bootstrap/assets/skills/{agent-ready-orchestrator/{SKILL.md,references/audit-flow.md},repository-analysis/{SKILL.md,references/inventory-facts.md}}`; add same-slice unseeded `internal/app/testdata/acceptance/driven/audit/`, parameterized upgrade gate, and driven audit proof.
- [x] 4.3 Update `internal/bootstrap/assets/skills/incremental-evolution/{SKILL.md,references/sync-flow.md}`; add same-slice unseeded `internal/app/testdata/acceptance/driven/sync/{lockfile,prose}/`, parameterized upgrade gate, and reassess/skip runtime proofs.
- [x] 4.4 Update `internal/bootstrap/assets/skills/{targeted-research/{SKILL.md,references/search-strategies.md},artifact-design/{SKILL.md,references/artifact-decisions.md},skill-reviewer/{SKILL.md,references/review-procedure.md}}`; add same-slice unseeded `internal/app/testdata/acceptance/driven/review/{grounded,ungrounded}/`, parameterized upgrade gate, driven-runtime proof, C–P, and full-suite proofs.
- [x] 4.5 User-directed amendment: decouple the harness from the exact OpenCode pin. Replace the exact-version fail-closed preflight in `internal/opencode/{version.go,compatibility.json,version_test.go}` with a minimum-compatible-version floor: installed versions below the floor fail closed with guidance; versions at or above the floor are accepted with the actual version recorded as a fact and a non-blocking warning on drift; keep config/skills schema validation and probe isolation. Adapt driven tests (`internal/app/driven_*.go`) and any embedded assets that hardcode the pin so host runtime updates never block init, audit, sync, or review. Same-slice parameterized upgrade gate if embedded assets change, focused tests, C–P and full-suite regression; ~380 lines; PR 10.

## Scope Guard

Exclude installer/provider/global integration work.
