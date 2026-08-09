---
name: migration-safety-review
description: "Trigger: a PR adds or edits a SQL migration. Review the migration for data loss, lock risk, and rollback before merge."
---
Review SQL migrations before merge. Load `references/checklist.md` when a migration must be reviewed; read `../../references/skill-system/skill-quality-rubric.md` when self-scoring.

## Activation Contract
Run when a PR adds or edits a file under the repo's `migrations/` directory.

## Hard Rules
- Never approve a migration without a rollback path.
- Never review from the diff alone; read the full migration file.
- Blocking findings block merge; record them with the verdict.

## Execution Steps
1. Load the checklist reference.
2. Read each migration file in the PR.
3. Apply the checklist; record PASS or FAIL per item with evidence.
4. Return the verdict with the per-item table.

## Output Contract
Verdict (approve / changes-requested) plus the checklist table: per-item PASS/FAIL and the evidence line for each.
