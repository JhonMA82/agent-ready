---
name: skill-reviewer
description: "Trigger: accepting a candidate skill. Score it against the rubric and gate acceptance; never accept without a verdict."
---
Run the mandatory gate before any candidate skill is accepted. Load `references/review-procedure.md` when a review begins; read `../../references/skill-system/skill-quality-rubric.md` when scoring.

## Activation Contract
Run before a candidate skill is created or accepted, and after every REVISE rework.

## Hard Rules
- Never accept a skill without a rubric score, verdict, and per-criterion justification.
- Never PASS instructions that are not evidence-backed; record blocking concerns in state.
- Never accept below 85: 70-84 is REVISE, < 70 is REJECT.

## Execution Steps
1. Load the review procedure reference.
2. Read the candidate and every reference it names; unresolved references fail validation.
3. Verify frontmatter: name pattern, directory match, description 1-1024 chars.
4. Trace evidence grounding; scan the anti-pattern catalog.
5. Score each rubric criterion; write the score sheet.
6. Return PASS, REVISE, or REJECT; record the sheet in state.

## Output Contract
Return the verdict and the full score sheet: per-criterion scores, total, and one grounded justification per criterion.
