---
name: artifact-design
description: "Trigger: choosing what the audit will create or change. Decide CREATE, UPDATE, REUSE, REMOVE, NO_ACTION, ASK_USER, COMPACT, EXTRACT_TO_SKILL, MOVE_TO_REFERENCE, REPLACE_WITH_SCRIPT, or REUSE_EXTERNAL_SKILL from labeled evidence; no artifact spam."
---
Turn labeled evidence into a deliberate artifact decision. Load `references/artifact-decisions.md` when a decision is needed; read `../../references/skill-system/skill-quality-rubric.md` only when scoring a candidate. Before concluding REUSE on existing guidance, apply the Context Placement Gate.

## Activation Contract
Run when the loop reaches the artifact_decisions stage or when new evidence changes a prior decision.

## Hard Rules
- Decide from labeled evidence only; an evidence-backed request precedes any artifact.
- NO_ACTION is a first-class decision: nothing scoring below threshold means nothing is created.
- No artifact spam: prefer REUSE over CREATE, keep the artifact set minimal, and record avoided artifacts.
- Never propose an artifact on UNKNOWN-only evidence; gather more or stop with ASK_USER.
- Surface conflicts, never hide them: conflicting package-manager or ecosystem evidence is named with the decision, and migration is proposed, never stated as fact.
- No unsupported certainty: package-manager certainty and capability claims never exceed the tested support in `tools status --json`.
- Placement is part of the decision: coverage without optimal placement does not conclude REUSE.
- Never claim exact token savings; use qualitative classes (VERY_LOW to VERY_HIGH).
- Do not move content just because it is long; if frequency is unknown, record the uncertainty.
- Record every decision with its evidence and confidence in state (decisions.jsonl).

## Seven Questions

Before any decision, answer all seven with evidence: is it repository-specific; is it repeatable; is it non-trivial; does it contain project-specific decisions or invariants; do AGENTS/docs solve it more cheaply; does a deterministic script solve it better; does framework-specific guidance require external verification? A script that fully solves the need means the skill is not created.

## Execution Steps
1. Collect the labeled evidence set for the current decision point.
2. Choose the decision output — CREATE, UPDATE, REUSE, NO_ACTION, or ASK_USER (REMOVE is sync-scope only), plus the placement verbs COMPACT, EXTRACT_TO_SKILL, MOVE_TO_REFERENCE, REPLACE_WITH_SCRIPT, or REUSE_EXTERNAL_SKILL — per `references/artifact-decisions.md`.
3. Record the decision, its evidence, and confidence in state.
4. Route the outcome per Verdict Routing: CREATE/UPDATE to proposal, REUSE to a persisted placement verdict, NO_ACTION to stop only after the Context Placement Gate, ASK_USER to the user.

## Verdict Routing

```text
CREATE → proposal/review
UPDATE → proposal/review
REUSE → persist placement verdict
REMOVE → proposal/review
COMPACT → proposal/review
EXTRACT_TO_SKILL → skill-creator → skill-reviewer → proposal/review
MOVE_TO_REFERENCE → proposal/review
REPLACE_WITH_SCRIPT → deterministic artifact proposal → review
REUSE_EXTERNAL_SKILL → persist external coverage decision
NO_ACTION → only after Context Placement Gate
ASK_USER → stop and ask
```

The rubric threshold >= 85 only controls the creation of new skills. It never blocks the placement transformations COMPACT, EXTRACT_TO_SKILL, MOVE_TO_REFERENCE, or REPLACE_WITH_SCRIPT: those route through their own transformations above, not through skill creation.

## Output Contract
Return the decision with its evidence and confidence, plus the recorded state entry; never a bare artifact count.
