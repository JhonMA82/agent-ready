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

## Sync Handoff
- For `sync`, the orchestrator MUST load `incremental-evolution` and `references/sync-flow.md` before any model work. Start with the `agent-ready changes --json` ChangeSet and `agent-ready checkpoint status` stage facts; do not infer them from prose or direct reads.
- Sync classifications, reassessments, recommendations, and verdicts are model-owned. Before returning, the model MUST persist one record in `.agent-ready/state/decisions.jsonl`; Go fact helpers only observe deterministic state. Go does not write semantic state.

## Sync Completion Checklist
Before returning from `sync`, confirm every item:
- [ ] ChangeSet and checkpoint stage facts are recorded.
- [ ] Every changed path is classified relevant or irrelevant with an explicit reason.
- [ ] Relevant changes have a reason-bearing reassessment; irrelevant changes have a reasoned skip.
- [ ] A completed reassessment has categorized ecosystem, productivity, and provider recommendations, or `NO_ADDITIONAL_TOOLS` with a reason.
- [ ] The model has written one `.agent-ready/state/decisions.jsonl` record.

## Adaptive Loop
1. What I know: gather deterministic facts (inspect, state, changes, checkpoint status); label findings with repository-analysis.
2. What I do not know: list the unknowns that block decisions; classify every finding FACT, INFERENCE, or UNKNOWN.
3. Evidence: collect only what the current stage needs; reuse checkpointed evidence before re-reading sources (read `references/token-discipline.md`).
4. Capability: assess tool needs in every audit — the categorized Tool / Capability Assessment (ecosystem, productivity, provider) with observed evidence and reasons is mandatory output (read `references/audit-flow.md`); an absent tool never blocks.
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
- Every successful audit MUST end with a Tool / Capability Assessment that names all three families verbatim — ecosystem, productivity, provider — each with observed evidence and a reason, or `NO_ADDITIONAL_TOOLS` with a reason; never omit or abbreviate a family.
- Apply the Tool Budget minimal-set ordering: rg + fd + jq; then + ast-grep when structural search is needed; then Semble OR Serena when semantic retrieval or navigation is justified; CodeGraph only when the graph capability adds clear value; Headroom only when context compression remains a measured problem. Heavier tools need explicit justification; the final set stays model-owned.

## Output Contract
Return the mode outcome, the recorded decisions with evidence and confidence, the Tool / Capability Assessment naming ecosystem, productivity, and provider, and the checkpoint stage. Resume rules govern the next run (read `references/resume-rules.md`).
