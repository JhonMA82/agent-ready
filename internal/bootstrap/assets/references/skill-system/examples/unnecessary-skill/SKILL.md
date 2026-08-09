---
name: change-reviewer
description: "Trigger: reviewing a proposed change. Evaluate changes against review best practices before acceptance."
---
Evaluate proposed changes before acceptance. Read `../../references/skill-system/skill-quality-rubric.md` when scoring.

## Activation Contract
Run when a change is proposed for acceptance.

## Hard Rules
- Never accept a change without review.
- Record findings for every reviewed change.

## Execution Steps
1. Read the proposed change and its evidence.
2. Check correctness, tests, and evidence grounding.
3. Score the change against the rubric.
4. Return the verdict with justification.

## Output Contract
Verdict with per-criterion justification; findings recorded in state.
