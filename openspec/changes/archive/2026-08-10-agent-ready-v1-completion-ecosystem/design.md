# Design: Complete Agent-Ready V1 Ecosystem Intelligence

## Technical Approach

In reversible slices, make upgrades ownership-safe, add bounded ecosystem facts, enrich the catalog, then drive audit, sync, and review contracts. V1 schemas and Go-facts/model-decisions remain authoritative.

## Architecture Decisions

| Decision | Choice and rationale | Rejected tradeoff |
|---|---|---|
| Upgrade ownership | `lifecycle.UpdatePlan` compares ownership, current bytes, and target bytes. It creates absent assets and reports collisions/modifications; obsolete unchanged assets remain owned. | Blind replacement overwrites users; deletion is unnecessary. |
| Detection boundary | `internal/ecosystem` is a pure table-driven detector emitting evidence, confidence, and conflicts. | Skill tables vary; primary selection violates model ownership. |
| JSON evolution | Keep V1 fields/types; add optional ordered slices and family IDs backed by one `Catalog`. | V2 breaks readers; another catalog competes. |
| Support truth | Each capability is `supported | unsupported | unknown`; only tested recipes support install. | Booleans conflate support states. |
| Driven oracle | Execute the local command through OpenCode; validate evidence-backed shape and provenance, never model verdicts. | Seeded conclusions and Go verdicts break authority. |

## Data Flow and Failure Behavior

```text
WalkDir -> paths + pruned presence -> ecosystem.Detect -> inspect JSON
                                      -> tools.Recommend -> model assessment
Catalog -> tools.Status -> legacy map + ordered families -------^
old manifest + current bytes + target -> UpdatePlan -> safeio transaction
unseeded temp repo -> agent-ready init -> opencode run --command agent-ready
                  -> JSONL events + newly written state -> structural oracle
```

Traversal records pruned roots, ignores descendants, and sorts collections. Read/parse errors fail inspection; unknown signals remain evidence. Upgrade conflicts produce no writes.

## Executable Runtime Contract

Outside `testing.Short()`, `internal/app/driven_audit_test.go` uses `AGENT_READY_DRIVEN_MODEL` to build and initialize `agent-ready` in a `t.TempDir()` Git fixture. With isolated HOME/XDG, it runs `opencode run --dir <repo> --model "$AGENT_READY_DRIVEN_MODEL" --format json --command agent-ready audit|sync|review`.

Inputs are repository files plus post-baseline mutations, never seeded conclusions. The structural oracle observes OpenCode JSONL events and new state/artifacts, requires fact-helper use, and checks three families plus evidence/reason or reasoned `NO_ADDITIONAL_TOOLS`. Sync proves lockfile reassessment or prose-only skip; review proves grounded pass and ungrounded rejection without fixed wording or verdict identity.

## Interfaces and Affected Source

- Modify `internal/{bootstrap,lifecycle,inventory,tools}`; create `internal/ecosystem` for evidence, conflicts, and candidates.
- Modify the embedded contracts named per slice below under `internal/bootstrap/assets/skills/`.
- Extend `internal/app/acceptance_test.go`; create `internal/app/driven_audit_test.go` and unseeded `internal/app/testdata/acceptance/driven/` cohorts. This follows the existing fixture root: C–P stay at current sibling paths; the driver reads only `driven/`.

## Autonomous Auto-Chain Slices

Each slice targets under 400 authored changed lines, includes its behavior tests, and is independently rollback-safe.

`internal/lifecycle/update_test.go` defines table-driven `runEmbeddedAssetUpgradeGate` with `t.TempDir()` and `{slice, changedAssets}`. It matches declarations to changed embedded paths and, per path, proves unchanged advancement and modified preservation/reporting. New paths also prove absent installation and unmanaged collision. Every case checks protected state and idempotence; omissions fail the slice.

| Slice | Deliverable and same-slice gate |
|---|---|
| 1 | Reconciled ownership/manifest; cross-version conflict, protection, and idempotence tests. |
| 2 | Pruned walk/presence facts; temp-tree and legacy JSON tests. |
| 3 | Mixed ecosystem, wrapper, framework, and build/test tables; stability tests. |
| 4 | Manager confidence/conflicts, including pnpm+Bun and pyproject ambiguity. |
| 5 | Single-catalog states/families; compatibility, ordering, and support-truth tests. |
| 6 | Evidence-only recommendations; grounded, empty, conflict tests. |
| 7 | Initial-audit assets: `agent-ready-orchestrator/SKILL.md`, `agent-ready-orchestrator/references/audit-flow.md`, `repository-analysis/SKILL.md`, and `repository-analysis/references/inventory-facts.md`; run their declared-asset upgrade gate plus one unseeded driven `audit` cohort. |
| 8 | Relevant-sync assets: `incremental-evolution/SKILL.md` and `incremental-evolution/references/sync-flow.md`; run their declared-asset upgrade gate plus driven baseline-to-lockfile reassessment and baseline-to-prose skip cohorts. |
| 9 | External Verification/reviewer assets: `targeted-research/SKILL.md`, `targeted-research/references/search-strategies.md`, `artifact-design/SKILL.md`, `artifact-design/references/artifact-decisions.md`, `skill-reviewer/SKILL.md`, and `skill-reviewer/references/review-procedure.md`; run their declared-asset upgrade gate plus driven grounded-pass and ungrounded-rejection `review` cohorts, C–P, and `go test ./...`. |

Slice 7–9 paths are relative to `internal/bootstrap/assets/skills/`. Each runs ownership and runtime gates. Additional embedded paths must join that slice's declaration and matrix. Rollback removes only its assets, driver cases, and fixtures.

## Threat Matrix and Rollout

| Boundary | Applicability | Safe/failure behavior and planned RED test |
|---|---|---|
| Documentation-like paths | N/A: no executable classification changes. | None. |
| Git repository selection | Applicable: driven process uses `--dir` and helper `git -C`. | Accept relative/absolute temp roots; fail closed for outside-root or missing Git roots; RED table test covers all selectors. |
| Commit state | N/A: no commits or index operations. | None. |
| Push state | N/A: no pushes. | None. |
| PR commands | N/A: no PR automation. | None. |

No migration or flag is required. Reverse-order rollback never deletes user-modified assets or model state. Installer expansion, provider lifecycle, and global OpenCode/RTK integration remain deferred to named follow-ups.

## Open Questions

None.
