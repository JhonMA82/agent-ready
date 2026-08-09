# Resume Rules

How a new `/agent-ready` invocation continues previous work. Read this reference at run start, when `checkpoint status` reports an incomplete audit, or before re-collecting any evidence. It resolves what to redo and what to skip.

## Entry decision

| State at start | Action |
|---|---|
| No checkpoint and no pending plan | Start a new audit at exploration_plan |
| Incomplete checkpoint at stage S | Resume at stage S; re-analyze only inventory paths whose hash changed |
| Completed checkpoint | Report the completed outcome; do not re-run |
| Nothing changed and no pending plan | Return NO_ACTION; create no artifacts |

## Rules

- Resume re-analyzes only changed paths; `changes` diffs inventory hashes and never reads model state.
- Completed evidence is never re-collected; read it from state files (decisions.jsonl, provenance.jsonl, artifact-graph.yaml).
- A resumed run writes a new checkpoint envelope when the stage completes.
- The checkpoint is the source of truth for progress; an interruption never restarts the audit at stage 1.
