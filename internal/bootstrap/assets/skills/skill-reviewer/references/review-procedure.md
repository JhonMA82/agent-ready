# Review Procedure

Procedure for the mandatory review gate. Read this reference when a candidate skill must be accepted; it resolves how to verify and score so acceptance is never a numeric shortcut.

## Inputs
- Candidate skill (SKILL.md + references), its self-score sheet, and the evidence it claims.
- skill-quality-rubric.md: >= 85 PASS, 70-84 REVISE, < 70 REJECT; 85 is PASS, 70 is REVISE, 69 is REJECT.
- anti-patterns.md: the failure catalog to scan for.

## Procedure
1. Read the candidate and every reference it names; a missing reference is a validation failure.
2. Verify frontmatter: `name` matches `^[a-z0-9]+(-[a-z0-9]+)*$` and the directory; `description` 1-1024 chars, trigger first.
3. Verify evidence grounding: trace each instruction to a FACT or harness fact; flag invented files, APIs, or flows.
4. Scan the anti-patterns: context dumping, generic advice, vague trigger, procedure-less body, duplicate skill, nice-to-have.
5. Score each criterion against its full-score requirement; partial points name the missing part.
6. Apply thresholds: >= 85 PASS, 70-84 REVISE, < 70 REJECT.
7. Write the score sheet: per-criterion scores, total, verdict, one grounded justification per criterion.
8. Record the sheet in state; REVISE names the failing criteria; REJECT records the score and justification.

## Blocking concerns
- A numeric PASS never overrides a blocking concern: if the skill would mislead the model, record the concern and do not PASS.
- "N skills generated" is never evidence; per-skill scores with justification are.
