# Example Notes — excellent-complex

Read this when calibrating PASS for a larger candidate: progressive disclosure plus a reference that resolves one decision. Use this shape when a skill needs detail beyond the body.

## Scenario (declared facts)
A Go service repo with a `migrations/` directory and a standard SQL toolchain.

## Score sheet
| Criterion | Score | Justification |
|---|---|---|
| necessity 25 | 25 | Migration risk review is distinct work the audit cannot skip |
| repository_specificity 20 | 20 | migrations/ path, rollback and lock rules, PR flow |
| discovery_description 15 | 15 | Trigger first, one line, <= 250 chars |
| procedural_value 15 | 15 | Activation, hard rules, ordered steps, table output |
| progressive_disclosure 10 | 10 | Body minimal; checklist loads only when a migration is reviewed |
| evidence_grounding 10 | 10 | Checklist items map to declared repo facts |
| validation 5 | 5 | Frontmatter valid; `references/checklist.md` resolves |
| Total | 100 | PASS |

## What to copy
A body that names its reference and the decision it resolves; detail lives in the reference, not the body.
