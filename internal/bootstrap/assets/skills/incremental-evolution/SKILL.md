---
name: incremental-evolution
description: "Trigger: syncing installed harness state after repository changes. Interpret the ChangeSet, update only affected evidence and artifacts, and never re-run a full audit."
---
Keep the harness aligned with the repository through selective updates. Load `references/sync-flow.md` when a ChangeSet arrives; reuse `changes` and `checkpoint status` facts without re-reading sources.

## Activation Contract
Run when `/agent-ready sync` is dispatched or when `changes` reports paths changed since the checkpoint baseline.

## Hard Rules
- Interpret the ChangeSet first: only changed paths are update candidates.
- Update only the evidence and artifacts the ChangeSet affects; skip everything else.
- Never re-run a full audit: completed evidence is reused from the checkpoint.
- Not every dependency change requires an artifact change; decide per path.
- Every update records the ChangeSet entries that justify it.

## Execution Steps
1. Read the ChangeSet from `changes` and the current stage from `checkpoint status`.
2. Map each changed path to the evidence and artifacts it affects.
3. Decide per affected item: update it, reuse it, or leave it unchanged.
4. Apply the selective updates and record each with its justifying ChangeSet entry.

## Output Contract
Return the list of updated artifacts, each tied to its ChangeSet entries; unchanged paths stay unlisted.
