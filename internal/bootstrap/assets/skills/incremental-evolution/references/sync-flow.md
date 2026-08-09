# Sync Flow

How a ChangeSet becomes selective updates. Read this reference when `changes` reports a diff or when sync must decide what to touch. It resolves how to interpret the ChangeSet and what to update.

## ChangeSet interpretation (normative)

| ChangeSet entry | Meaning | Sync response |
|---|---|---|
| `added` | Path exists now, not at baseline | Evaluate evidence; update only if it affects a shipped artifact |
| `changed` | Path content differs from baseline | Update the evidence and artifacts that cite this path |
| `removed` | Path gone since baseline | Re-check dependents; removal decisions stay within sync scope (no lifecycle approval flows) |
| `first_run` | No baseline existed | Treat as initial inventory, not a change signal; no updates from it alone |

## Rules

- Selective updates only: a changed path updates the evidence that cites it and nothing else; artifacts whose evidence is unchanged are not touched.
- No full re-audit: sync never re-runs completed stages; it reuses checkpointed evidence (R14) and reads `changes` hashes, never model state.
- Not every dependency requires change: a dependency bump updates the evidence that cites it, not every artifact in the harness.
- Unchanged paths never appear in sync output; they are already aligned with the checkpoint baseline.
- Nothing changed and no pending plan: return NO_ACTION with zero artifacts.
