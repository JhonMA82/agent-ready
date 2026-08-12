# Skill Quality Rubric

Canonical scoring instrument for candidate skills in this harness. Read this rubric when a rubric decision is needed: before creating a skill, before accepting one, or when an existing skill must be re-scored. A skill is created or accepted only after a rubric score.

## Verdict thresholds (normative)

| Total | Verdict | Meaning |
|---|---|---|
| >= 85 | PASS | Create or accept the skill as shipped. |
| 70-84 | REVISE | Rework the weak criteria, then re-score. |
| < 70 | REJECT | Do not create or accept. Record the score and its justification in state. |

Boundaries are exact: a total of 85 is PASS, 70 is REVISE, 69 is REJECT.

## Criteria and weights (normative; sum to 100)

| Weight | Criterion | Full score requires |
|---|---|---|
| 20 | necessity | The skill does work the audit cannot complete without it; it is not a duplicate of another skill and not a nice-to-have. |
| 15 | repository_specificity | Instructions are specific to this repository and its installed OpenCode version (minimum compatible 1.18.15); no generic advice that could apply to any codebase. |
| 15 | discovery_description | `description` is one quoted line, <= 250 chars, states the trigger first, and lets the model know when to load the skill. |
| 15 | procedural_value | The body gives an executable procedure: activation, hard rules, decision gates, execution steps, output contract. |
| 10 | progressive_disclosure | The body stays minimal (180-450 tokens); deeper detail lives in `references/` and loads only when needed. |
| 10 | evidence_grounding | Every instruction maps to a harness fact, spec requirement, or scenario; no invented APIs, files, or flows. |
| 10 | context_placement | Placement is deliberate: task-specific procedures live in skills/references, always-on context is minimal, no duplication after extraction, router preserved, discoverability preserved. |
| 5 | validation | Content is checkable: frontmatter `name` matches `^[a-z0-9]+(-[a-z0-9]+)*$` and its directory, `description` is 1-1024 chars, referenced files resolve. |

## Scoring procedure

1. Read the candidate skill and every reference it points to.
2. Score each criterion from its full-score requirement; award partial points only when the requirement is partially met, and say which part is missing.
3. Sum the criterion scores and apply the verdict threshold.
4. Write the score sheet: per-criterion scores, the total, the verdict, and one justification line per criterion, each grounded in the skill's content.

## Justification contract

- A score without justification is not a score. Every criterion line needs a rationale the skill itself supports.
- A bare total with no per-criterion breakdown is not a verdict; return the full score sheet.
- The rubric never substitutes for judgment: if a criterion passes numerically but the skill would mislead the model, record the concern as a blocking justification and do not PASS.
- Record the score sheet and verdict in state for every created, revised, or rejected skill.
