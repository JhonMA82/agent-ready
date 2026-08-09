---
name: agent-ready-orchestrator
description: "Trigger: /agent-ready audit, sync, review, or status. Coordinate the repository-local adaptive audit loop; stop with evidence or NO_ACTION, never with ungrounded artifacts."
---
Coordinate the repository-local `/agent-ready` run: dispatch the requested mode, then run the adaptive loop. The loop is a map, not a rigid sequence: revisit any step when new evidence or a stop condition requires it.

## Mode Dispatch
- audit: run the full adaptive loop, exploration plan to checkpoint (read `references/audit-flow.md`).
- sync: reconcile installed harness state with repository facts; selective update only.
- review: gate a candidate skill with skill-reviewer before acceptance.
- status: report checkpoint and state facts; decide resume or NO_ACTION (read `references/resume-rules.md`).

## Adaptive Loop
1. What I know: gather deterministic facts (inspect, state, changes, checkpoint status); label findings with repository-analysis.
2. What I do not know: list the unknowns that block decisions; classify every finding FACT, INFERENCE, or UNKNOWN.
3. Evidence: collect only what the current stage needs; reuse checkpointed evidence before re-reading sources (read `references/token-discipline.md`).
4. Capability: reason about available tools; an absent tool never blocks (tool management is out of scope).
5. Research: close evidence gaps with targeted-research (read `references/audit-flow.md`).
6. Ask: when only the user can resolve an unknown, stop with ASK_USER (read `references/stop-conditions.md`).
7. Confidence: record confidence with every decision; label each artifact with its evidence.
8. Artifact value: propose only evidence-backed artifacts and record avoided ones; when nothing scores above threshold, return NO_ACTION and create nothing.
9. Review: gate every candidate with skill-reviewer before creation.
10. Stop: apply the stop conditions and checkpoint the completed stage.

## Hard Rules
- Never dump repository content into context; load the smallest useful context.
- Never accept "N skills generated" as evidence; evidence is labeled facts and recorded decisions.
- No model, agent, or subtask overrides; modes dispatch through `$ARGUMENTS`.
- Record every decision, stop reason, and verdict in state (decisions.jsonl, provenance.jsonl).

## Output Contract
Return the mode outcome, the recorded decisions with evidence and confidence, and the checkpoint stage. Resume rules govern the next run (read `references/resume-rules.md`).
