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
6. Apply the placement checks to extraction candidates; a failing check is a blocking concern.
7. Score each criterion against its full-score requirement, including context_placement; partial points name the missing part.
8. Apply thresholds: >= 85 PASS, 70-84 REVISE, < 70 REJECT.
9. Write the score sheet: per-criterion scores, total, verdict, one grounded justification per criterion.
10. Record the sheet in state; REVISE names the failing criteria; REJECT records the score and justification.

## Rejection contract

REJECT an artifact when it has missing required assessment, unsupported package-manager certainty, framework/toolchain claims that fail the External Verification Gate, capability claims exceeding tested support, hidden conflicts, migration decisions presented as facts, or semantic verdicts routed through Go. REJECT an extraction that moves a global invariant into a skill, makes a critical rule harder to discover, duplicates content instead of moving it, or produces no real context saving. A technically excellent skill is REJECTED if the same content remains fully in AGENTS: it adds no progressive disclosure. NO_ACTION and Tool Budget remain valid model-owned outcomes and are never grounds for rejection.

## Placement checks

Run all four placement checks on every extraction candidate:
- `context_savings`: the move reduces permanent context without losing guidance.
- `duplication_after_extraction`: the content must not remain duplicated in the source.
- `discoverability_preserved`: critical rules stay discoverable after the move.
- `always_on_guidance_not_removed`: a global invariant must not be moved into a skill.

A failing placement check is a blocking concern: REJECT the extraction.

## Named checks

Run all four named checks on every candidate:
- `framework_grounding`: framework guidance traces to repository or verified external evidence for the version in use.
- `package_manager_accuracy`: the named package manager matches repository evidence (npm vs Bun or pnpm; pip vs uv).
- `toolchain_accuracy`: claimed validated workflows exist (no `cargo test` as validated when the repository has no tests).
- `external_verification_when_required`: version-sensitive claims carry the evidence the rule requires, and the `external_verified_evidence` / `external_verification_not_required` vocabulary is verified on every central-framework artifact.

A failing named check is a blocking concern: REJECT the artifact.

## Blocking concerns
- A numeric PASS never overrides a blocking concern: if the skill would mislead the model, record the concern and do not PASS.
- "N skills generated" is never evidence; per-skill scores with justification are.
