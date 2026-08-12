# Gap Analysis — Agent-Ready V1 Context Placement & Token Optimization Refinement

Status: **implemented** (2026-08-11; uncommitted working tree) — all five phases complete; `go build ./...`, `go vet ./...`, `go test -count=1 ./...` green; DoD §48 items 1–17 of 19 satisfied (the two unverified items are the tanstack real-repo run and the manual fixture, covered by fixture-q).
Source document: `OPENCODE_AGENT_READY_V1_CONTEXT_PLACEMENT_REFINEMENT.md` (repo root, untracked)
Route: delegated direct implementation (no SDD ceremony)

## 1. Verdict

V1 already implements roughly 35% of the refinement document. Solid existing
coverage: `NO_ACTION` as a first-class outcome, the three tool-assessment
families, `BoilerplateFacts`, the skill-quality rubric, and the
"duplicate skill" anti-pattern.

Missing entirely is the core of the document: the context-placement hierarchy,
the five new decision verbs, the Context Placement Gate, the placement
analysis contract and cost model, extraction support in `skill-creator`,
placement checks in `skill-reviewer`, `repository_kind` classification,
richer RTK signals, and the five regression fixtures.

Notable finding: **`ast-grep` has no candidate in `tools recommend`** even
though the Tool Budget names it. The Go side must emit a
`structured_search_need` signal.

## 2. Coverage map (document section -> V1 state)

Legend: ✓ present, ◐ partial, ✗ missing.

| § | Requirement | State | V1 evidence |
|---|---|---|---|
| 0 | NO_ACTION correct on dashboard-like repo | ✓ | fixture-e (no-artifact-spam) |
| 1–2 | Placement principle + hierarchy | ✗ | not present |
| 3 | AGENTS.md always-on rule | ✗ | only "no context dumping" |
| 4 | Skill rule (task-specific, no framework generics) | ◐ | Seven Questions cover partially |
| 5 | References/docs rule | ✗ | progressive-disclosure exists, no hierarchy |
| 6 | Script rule | ◐ | Q6 (deterministic script) |
| 7 | Five new decision verbs | ✗ | only six verbs today |
| 8 | Context Placement Analysis contract | ✗ | not present |
| 9 | Qualitative cost model | ✗ | not present |
| 10 | Usage frequency | ✗ | not present |
| 11 | Duplication checks | ◐ | anti-pattern "duplicate skill" + Q5/Q6 |
| 12–13 | tanstack-shadcn case + candidate skill | ✗ | no fixture, no placement evaluation |
| 14 | RTK in mandatory matrix | ◐ | candidate exists; absent from audit contract |
| 15 | RTK beyond dist/build/coverage | ✗ | `hasOutputDirs()` is the only signal |
| 16 | RTK output example (CONSIDER/NOT_JUSTIFIED) | ✗ | states not modeled |
| 17 | Tool assessment completeness | ◐ | ast-grep without Go candidate; RTK unnamed in audit-flow |
| 18 | Provider decision completeness | ◐ | audit-flow names only codegraph, context7 |
| 19 | repository_kind | ✗ | only BoilerplateFacts |
| 20–21 | Boilerplate audit + candidates | ◐ | BoilerplateFacts exists; no extension model |
| 22 | Boilerplate placement questions | ✗ | not present |
| 23 | Repository profile fields | ◐ | yaml is model-owned; no field contract |
| 24 | Orchestrator: coverage != REUSE | ✗ | not present |
| 25 | Context Placement Gate | ✗ | not present |
| 26 | skill-creator extraction without duplication | ✗ | authoring covers CREATE from scratch only |
| 27 | skill-reviewer placement checks | ✗ | not present |
| 28 | General placement review | ✗ | review only gates skills |
| 29 | NO_ACTION backed by coverage+placement+tools | ◐ | placement pillar missing |
| 30 | Stage terminology | ◐ | stage machine exists; placement stage missing |
| 31 | Suggested audit output | ✗ | no output template |
| 32 | Token optimization metrics | ✗ | not present |
| 33 | placement_change provenance | ✗ | provenance.jsonl exists; no placement schema |
| 34 | Sync detects placement changes | ✗ | generic ChangeSet only |
| 35 | Artifact graph relations | ✗ | artifact-graph.yaml model-owned, no contract |
| 36 | Tool reassessment triggers | ◐ | 8 triggers; complexity/provider triggers missing |
| 37–41 | Five regression fixtures | ✗ | fixtures C–P exist, none of the five |
| 42–43 | Rubric: context_placement 10 + rebalance | ✗ | rubric 25/20/15/15/10/10/5 |
| 44 | Boilerplate assessment output | ✗ | not present |
| 45 | No tool recommendation inflation | ◐ | Tool Budget exists; RTK not individualized |
| 46 | `tools recommend --json` signals | ◐ | 9 capabilities; placement/output/ast-grep signals missing |
| 47 | No fake token estimates | ✓ | token-discipline + bounded certainty |
| 48 | DoD | — | ~7 of 19 already satisfied |

Summary: ~9 sections ✓/acceptable, ~14 ◐ partial, ~26 ✗ missing.

## 3. Per-file change plan

### Skills (model-side, 10–12 files)

| File | Change |
|---|---|
| `skills/agent-ready-orchestrator/SKILL.md` | Rule "coverage is not sufficient to conclude REUSE"; six placement questions; RTK + ast-grep named in productivity; five providers named in provider family |
| `skills/agent-ready-orchestrator/references/audit-flow.md` | `context_placement` stage; internal stage names (§30); RTK explicit status vocabulary RECOMMENDED/CONSIDER/NOT_JUSTIFIED/ALREADY_AVAILABLE; compact provider output; audit output template (§31) |
| `skills/artifact-design/SKILL.md` | Five new verbs; Context Placement Gate before REUSE; no fake token estimates; frequency discipline |
| `skills/artifact-design/references/artifact-decisions.md` | New decision table rows; placement analysis contract (§8); cost model (§9); frequency (§10); duplication order (§11); gate flow (§25) |
| `skills/skill-creator/SKILL.md` + `authoring-procedure.md` | Extraction mode: preserve semantics, router in source, no dual copy, placement provenance (§26, §33) |
| `skills/skill-reviewer/SKILL.md` + `review-procedure.md` | Checks context_savings / duplication_after_extraction / discoverability_preserved / always_on_guidance_not_removed; new rejection grounds (§27, §43) |
| `skills/repository-analysis/SKILL.md` + `inventory-facts.md` | repository_profile.kind + confidence; boilerplate audit; frequency inference default unknown (§19–20, §22–23, §10) |
| `skills/incremental-evolution/SKILL.md` + `sync-flow.md` | Placement-change detection; new reassessment triggers (§36); provenance + graph relations (§33–35) |
| `references/skill-system/skill-quality-rubric.md` | context_placement weight 10; rebalance to 20/15/15/15/10/10/10/5 (§42) |
| `references/skill-system/anti-patterns.md` | Placement anti-patterns (dual copy after extraction, hidden always-on rules, script hiding semantic decisions) |

### Go (3–5 files)

| File | Change |
|---|---|
| `internal/inventory/inventory.go` (+test) | Optional `agents_md` fact: `{path, lines}` when a root AGENTS.md exists |
| `internal/tools/recommend.go` (+test) | New candidate `ast-grep` (signal `structured_search_need`, source surface >= threshold); new candidate for `context_placement_pressure` (AGENTS.md > 300 lines; no catalog entry); RTK trigger enriched with build/test scripts + CI (keeps `rtk` signal value); append new candidates at the end to minimize index churn |
| `internal/bootstrap/content_test.go` | Update locked rubric weights to the new contract |

### Regression fixtures (Q–U, ~10–15 files)

| Fixture | Seeds | Expects |
|---|---|---|
| q tanstack-shadcn-boilerplate | package.json (tanstack/start/react/shadcn), AGENTS.md with 13-step screen procedure, route examples, scripts/generate-routes | kind: boilerplate evidence; placement decision REUSE-with-reason or EXTRACT-with-savings; shadcn external-skill reuse; no generic react skill |
| r long-agents | AGENTS.md 500+ lines (50 global / 150 migration / 100 release / 200 examples) | COMPACT + EXTRACT migration + EXTRACT release + MOVE examples; recommend emits context_placement_pressure |
| s short-optimal-agents | AGENTS.md ~60 lines, well-placed global guidance | REUSE + NO_ACTION; no context_placement_pressure |
| t deterministic-procedure | AGENTS long deterministic procedure + existing equivalent script | REPLACE_WITH_SCRIPT / COMPACT; no redundant skill |
| u external-canonical-skill | Guidance for capability covered by canonical external skill | REUSE_EXTERNAL_SKILL; no wrapper |

### Docs (3 files)

`docs/how-it-works.md` (placement stage + RTK/ast-grep in assessment), `docs/usage.md` (new decision verbs, only if they appear in documented output), `CHANGELOG.md` entry.

## 4. Implementation phases

1. **Skills contract** — placement hierarchy, verbs, gate, contracts, rubric rebalance, extraction/review support, sync extension
2. **Go signals** — inventory `agents_md` fact; recommend candidates (ast-grep, context_placement_pressure, RTK enrichment) + tests; content_test rubric weights
3. **Regression fixtures** — Q–U + acceptance harness extension
4. **Docs + changelog**
5. **Verification** — `go build ./...`, `go vet ./...`, `go test ./...`

One writer per phase, sequential, with a short handoff between phases.

## 5. Decisions adopted

1. **placement_review (§28)**: integrated into the existing `review` stage — no new agent (document forbids more agents).
2. **Rubric rebalance (§42)**: normative for future candidates; shipped skills are not retroactively re-scored unless revisited via UPDATE.
3. **context_placement_pressure threshold**: Go signal at AGENTS.md > 300 lines; frequency defaults to `unknown` (rule §10 — never invent frequency).
4. **RTK states (§14–16)**: model-side contract in audit-flow; Go only adds richer deterministic signals, no new verdict schema (§46: "Go solo reporta señales").
5. **Go schema**: additive only — `agents_md` fact and new candidates; `agent-ready.recommend/v1` stays, no new schema version.

## 6. Locked contracts to keep green

- `internal/bootstrap/content_test.go` — rubric weights (update to new contract), skill content strings (preserve or update with justification).
- `internal/bootstrap/bootstrap_test.go` — manifest vs embedded walk (no add/remove of asset files → no manifest change).
- `internal/tools/recommend_test.go` — candidate order assertions (append new candidates last; update broken indices).
- `internal/app/acceptance_test.go` + driven tests — fixture cohorts (extend for Q–U).

## 7. Effort estimate

~25–35 files, ~600–1100 changed lines — above the 400-line review budget; delivery decision (chained PRs vs size:exception) is pending user choice at the end of implementation.
