# Example Notes — excellent-simple

Read this when calibrating PASS: a small, single-purpose skill can score full marks. Use this shape as the baseline when judging other candidates.

## Scenario (declared facts)
A Go service repo: single module, `go build` and `go test` available, PRs reviewed before merge. Exploration collected these facts; the skill references nothing else.

## Score sheet
| Criterion | Score | Justification |
|---|---|---|
| necessity 25 | 25 | Bump gating is distinct work the audit cannot skip |
| repository_specificity 20 | 20 | go.mod/go.sum, build and test commands, PR flow |
| discovery_description 15 | 15 | Trigger first, one line, <= 250 chars |
| procedural_value 15 | 15 | Activation, hard rules, ordered steps, verdict output |
| progressive_disclosure 10 | 10 | Minimal body; rubric read only at self-score time |
| evidence_grounding 10 | 10 | Every command and file exists in the declared scenario |
| validation 5 | 5 | Frontmatter valid; references resolve |
| Total | 100 | PASS |

## What to copy
Trigger-first description; imperative steps; every claim tied to a declared fact; no invented tooling.
