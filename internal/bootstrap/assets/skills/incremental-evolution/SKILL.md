---
name: incremental-evolution
description: "Trigger: syncing installed harness state after repository changes. Interpret the ChangeSet, update only affected evidence and artifacts, and never re-run a full audit."
---
Keep the harness aligned with the repository through selective updates. Load `references/sync-flow.md` when a ChangeSet arrives; reuse `changes` and `checkpoint status` facts without re-reading sources.

## Activation Contract
Run when `/agent-ready sync` is dispatched or when `changes` reports paths changed since the checkpoint baseline.

## Hard Rules
- Interpret the ChangeSet first: only changed paths are update candidates.
- A sync MUST query `agent-ready changes --json` for the ChangeSet and `agent-ready checkpoint status` for the stage before deciding what to update; direct file reads never replace these facts.
- Assess whether changed evidence can affect tool needs; never skip that assessment silently.
- Update only the evidence and artifacts the ChangeSet affects; skip everything else.
- Never re-run a full audit: completed evidence is reused from the checkpoint.
- Not every dependency change requires an artifact change; decide per path.
- Every update records the ChangeSet entries that justify it.
- Manifest, lockfile, workspace, wrapper, CI, framework, build/test output, or tool-fact changes MUST trigger reassessment of tool capabilities.
- A completed reassessment MUST include reasons and either categorized recommendations or `NO_ADDITIONAL_TOOLS`.
- Irrelevant changes MUST record a reason for skipping the reassessment.
- Sync output MUST include ChangeSet/stage facts, relevant-vs-irrelevant classification, a reason-bearing reassessment or reasoned skip, and categorized recommendations or `NO_ADDITIONAL_TOOLS` when reassessment is completed.
- State persistence is model-owned: before return, the model MUST append one record to `.agent-ready/state/decisions.jsonl`. Go does not write semantic state or `decisions.jsonl`.

## Execution Steps
1. Read the ChangeSet from `agent-ready changes --json` and the current stage from `agent-ready checkpoint status`.
2. Map each changed path to the evidence and artifacts it affects.
3. Assess whether changed evidence can affect tool needs.
4. When manifest, lockfile, workspace, wrapper, CI, framework, build/test output, or tool-fact evidence changed, reassess tool capabilities with reasons and categorized recommendations or `NO_ADDITIONAL_TOOLS`.
5. When nothing relevant changed, record the reason for skipping the reassessment.
6. Decide per affected item: update it, reuse it, or leave it unchanged.
7. Apply the selective updates, record each with its justifying ChangeSet entry, and record the reassessment (or skip) decision in state.

## Mandatory Final Checklist
Before returning from sync, verify:
- [ ] ChangeSet and checkpoint stage facts are present.
- [ ] Each changed path is classified relevant or irrelevant with an explicit reason.
- [ ] Relevant changes carry a reason-bearing reassessment; irrelevant changes carry a reasoned skip.
- [ ] Completed reassessment carries categorized ecosystem, productivity, and provider recommendations or `NO_ADDITIONAL_TOOLS` with a reason.
- [ ] One model-owned `.agent-ready/state/decisions.jsonl` record was written.

## Output Contract
Return the list of updated artifacts, each tied to its ChangeSet entries; unchanged paths stay unlisted. State whether tool capabilities were reassessed (with reasons and categorized recommendations or `NO_ADDITIONAL_TOOLS`) or skipped (with the recorded reason).
