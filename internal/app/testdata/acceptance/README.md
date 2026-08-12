# Acceptance Fixtures — agent-ready-northstar-audit

The fixtures under this directory implement the spec section 40 acceptance
subset for the audit change. Each `fixture-<letter>/` is a miniature
repository whose seeded files model a repository state and whose `expect/`
subtree carries the canonical evidence files that a correct audit must
produce or recognize. The process harness in
`internal/app/acceptance_test.go` copies each fixture into a fresh temporary
Git repository, runs the built binary, and asserts JSON facts plus evidence
shapes. Evidence is asserted by content, never by counts alone:
"N skills generated" is never accepted as evidence (spec 40, R16).

## Fixture-ID mapping

| ID | Fixture | Evidence class (spec 26 / 40) | Seeded evidence |
|---|---|---|---|
| C | adaptive-analysis | exploration plan; FACT/INFERENCE/UNKNOWN labels | `expect/exploration-plan.yaml`, `expect/evidence-labels.md` |
| D | discovery | evidence-backed proposal before any artifact | `expect/proposal.md`, seeded `decisions.jsonl` |
| E | no-artifact-spam | avoided artifacts recorded; "N skills generated" never accepted | seeded `decisions.jsonl` with NO_ACTION records; absent skills dir |
| F | rubric-creation | >=85 PASS -> skill created, reviewer gate passed | seeded skill + `NOTES.md` rubric breakdown >=85 |
| G | rubric-rejection | <70 REJECT + rubric justification in state | seeded weak skill + `decisions.jsonl` REJECT with justification |

Fixtures H–P (resume, isolation, degradation, decision-evidence) arrive with
PR12; O (lifecycle approval flows) is deliberately absent per the
specification deferral.

## PR12 additions

| ID | Fixture | Evidence class (spec 26 / 40) | Seeded evidence |
|---|---|---|---|
| H | ask-user | ASK_USER after no-new-evidence iterations | seeded `decisions.jsonl` ASK_USER record |
| I | stop-with-concerns | STOP_WITH_CONCERNS + recorded reasons | seeded `decisions.jsonl` STOP_WITH_CONCERNS record |
| J | incremental-sync | selective update, changed paths only | seeded UPDATE record, no full re-audit marker |
| K | no-action-sync | NO_ACTION, zero artifacts | seeded NO_ACTION record; absent `.opencode/` |
| L | resume | stage-3 checkpoint, unchanged sources -> resume, no re-collection | checkpoints built by the harness via `checkpoint save` |
| M | isolation | no global writes; trees byte-identical | snapshot equality around read-only helpers |
| N | tool-degradation | no Tool Manager -> capability reasoning, no block | seeded capability-reasoning record |
| P | decision-evidence | decisions.jsonl/provenance.jsonl record every decision + rationale | seeded decisions + provenance records |

## Context-placement additions (refinement §37–41)

| ID | Fixture | Evidence class | Seeded evidence |
|---|---|---|---|
| Q | tanstack-shadcn-boilerplate | repository_kind boilerplate (extension points, generated files, scaffolding); screen-creation placement analysis (REUSE vs EXTRACT_TO_SKILL vs COMPACT); shadcn external-skill reuse, no generic React/TanStack skill; RTK on build scripts, no pressure signal (AGENTS < 300 lines) | `expect/repository-kind.md`, `expect/placement-screen-creation.md`, `expect/external-skill-reuse.md` |
| R | long-agents | AGENTS.md 500+ lines (50 global / 150 migration / 100 release / 200 examples) -> COMPACT + EXTRACT migration + EXTRACT release + MOVE examples, router preserved; recommend emits `context_placement_pressure` (no catalog id) | `expect/placement-analysis.md` |
| S | short-optimal-agents | ~60-line well-placed global AGENTS.md -> REUSE + NO_ACTION, no extraction by dogma, no pressure signal | `expect/placement-reuse.md` |
| T | deterministic-procedure | deterministic procedure replaced by existing equivalent scripts (generate-routes, validate-presets) -> REPLACE_WITH_SCRIPT / COMPACT, no redundant skill | `expect/decision-replace-with-script.md` |
| U | external-canonical-skill | capability covered by canonical external skill -> REUSE_EXTERNAL_SKILL, no local wrapper | `expect/decision-reuse-external-skill.md` |
