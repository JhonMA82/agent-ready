# Review Procedure

Procedure for the mandatory review gate. Read this reference when a candidate skill must be accepted; it resolves how to verify and score so acceptance is never a numeric shortcut.

## Inputs
- Candidate skill (SKILL.md + references), typically under `.opencode/skills/`; its self-score sheet, and the evidence it claims.
- skill-quality-rubric.md: >= 85 PASS, 70-84 REVISE, < 70 REJECT; 85 is PASS, 70 is REVISE, 69 is REJECT.
- anti-patterns.md: the failure catalog to scan for.

## Procedure
1. Read the candidate and every reference it names; a missing reference is a validation failure.
2. Verify frontmatter: `name` matches `^[a-z0-9]+(-[a-z0-9]+)*$` and the directory; `description` 1-1024 chars, trigger first.
3. Verify evidence grounding: trace each instruction to a FACT or harness fact; before scoring, run `agent-ready state --json` and `agent-ready inspect --json` (both are available in every initialized repository; skipping them is a blocking concern); flag invented files, APIs, or flows.
4. Apply the External Verification Gate to every version-sensitive framework, package-manager, or toolchain claim: it MUST cite current official or versioned evidence tied to the applicable version, or carry an explicit reasoned exemption showing that no external claim is embedded. repository-to-official research precedence stays intact; a claim that fails the gate is a blocking concern.
5. Scan the anti-patterns: context dumping, generic advice, vague trigger, procedure-less body, duplicate skill, nice-to-have.
6. Score each criterion against its full-score requirement; partial points name the missing part.
7. Apply thresholds: >= 85 PASS, 70-84 REVISE, < 70 REJECT.
8. Write the score sheet: per-criterion scores, total, verdict, one grounded justification per criterion.
9. Record the sheet in state; REVISE names the failing criteria; REJECT records the score and justification.

## Rejection contract

REJECT an artifact when it has missing required assessment, unsupported package-manager certainty, framework/toolchain claims that fail the External Verification Gate, capability claims exceeding tested support, hidden conflicts, migration decisions presented as facts, or semantic verdicts routed through Go. NO_ACTION and Tool Budget remain valid model-owned outcomes and are never grounds for rejection.

## Blocking concerns
- A numeric PASS never overrides a blocking concern: if the skill would mislead the model, record the concern and do not PASS.
- "N skills generated" is never evidence; per-skill scores with justification are.
