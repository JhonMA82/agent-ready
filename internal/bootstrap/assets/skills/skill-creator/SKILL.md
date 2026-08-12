---
name: skill-creator
description: "Trigger: creating a skill file. Author the approved candidate from evidence; never decide whether a skill is needed."
---
Author the approved candidate as a runtime instruction contract. Load `references/authoring-procedure.md` when authoring begins; read `../../references/skill-system/skill-quality-rubric.md` only when self-scoring.

## Activation Contract
Run when the artifact decision names CREATE for a specific skill, the proposal is approved, and evidence exists.

## Hard Rules
- Never decide necessity: CREATE was already decided; missing evidence holds the skill, it does not trigger authoring.
- Never invent files, APIs, or flows; every instruction maps to a harness fact or collected evidence.
- Author from the structured `skill_request` only: purpose, repository_evidence, framework_evidence, external_verified_evidence, canonical_examples, invariants, commands, and validation — each reflected in the artifact or explicitly waived.
- Never invent external framework guidance: when external_verified_evidence is required and empty, report the missing evidence instead of filling it.
- Extraction mode: when the approved decision is EXTRACT_TO_SKILL or MOVE_TO_REFERENCE, preserve the original semantics, never invent new conventions, use existing canonical examples, leave a short router in the original artifact when appropriate, and NEVER copy the same content into both places.
- No generic advice: anchor instructions to this repository, its layout, and the installed OpenCode version (minimum compatible 1.18.15).
- Keep the body minimal; move procedure and detail into `references/`.

## Execution Steps
1. Load the authoring procedure reference.
2. Collect the approval decision and the evidence it cites.
3. Draft the body: trigger, activation, hard rules, gates, steps, output contract.
4. Write frontmatter: name `^[a-z0-9]+(-[a-z0-9]+)*$`, directory match, description 1-1024 chars, trigger first.
5. For extraction decisions (EXTRACT_TO_SKILL, MOVE_TO_REFERENCE), follow the extraction procedure in the authoring reference: read the source section, extract only task-specific procedure, leave the router, record placement provenance.
6. Self-score against the rubric, including the context_placement criterion; below 85 is REVISE, not handoff.

## Output Contract
Return the skill file plus references and the self-score sheet with per-criterion justification; record both in state.
