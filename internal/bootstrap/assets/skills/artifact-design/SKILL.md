---
name: artifact-design
description: "Trigger: choosing what the audit will create or change. Decide CREATE, UPDATE, REUSE, REMOVE, NO_ACTION, or ASK_USER from labeled evidence; never generate artifact spam."
---
Turn labeled evidence into a deliberate artifact decision. Load `references/artifact-decisions.md` when a decision is needed; read `../../references/skill-system/skill-quality-rubric.md` only when scoring a candidate.

## Activation Contract
Run when the loop reaches the artifact_decisions stage or when new evidence changes a prior decision.

## Hard Rules
- Decide from labeled evidence only; an evidence-backed request precedes any artifact.
- NO_ACTION is a first-class decision: nothing scoring below threshold means nothing is created.
- No artifact spam: prefer REUSE over CREATE, keep the artifact set minimal, and record avoided artifacts.
- Never propose an artifact on UNKNOWN-only evidence; gather more or stop with ASK_USER.
- Surface conflicts, never hide them: conflicting package-manager or ecosystem evidence is named with the decision, and migration is proposed, never stated as fact.
- No unsupported certainty: package-manager certainty and capability claims never exceed the tested support in `tools status --json`.
- Record every decision with its evidence and confidence in state (decisions.jsonl).

## Seven Questions

Before any decision, answer all seven with evidence: is it repository-specific; is it repeatable; is it non-trivial; does it contain project-specific decisions or invariants; do AGENTS/docs solve it more cheaply; does a deterministic script solve it better; does framework-specific guidance require external verification? A script that fully solves the need means the skill is not created.

## Execution Steps
1. Collect the labeled evidence set for the current decision point.
2. Choose the decision output — CREATE, UPDATE, REUSE, NO_ACTION, or ASK_USER (REMOVE is sync-scope only) — per `references/artifact-decisions.md`.
3. Record the decision, its evidence, and confidence in state.
4. Route the outcome: CREATE/UPDATE to proposal, NO_ACTION to stop, ASK_USER to the user.

## Output Contract
Return the decision with its evidence and confidence, plus the recorded state entry; never a bare artifact count.
