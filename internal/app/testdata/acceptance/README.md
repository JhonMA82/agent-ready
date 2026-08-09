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
