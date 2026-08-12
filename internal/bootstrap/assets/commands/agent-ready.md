---
description: Run repository-ready orchestration (audit, sync, review, or status)
---
Route every mode through `agent-ready-orchestrator` and execute the flow selected by `$ARGUMENTS`: audit, sync, review, or status. For `sync`, the orchestrator MUST load `incremental-evolution` and `references/sync-flow.md` before any model work.
