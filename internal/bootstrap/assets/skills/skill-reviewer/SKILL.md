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
- External Verification Gate: reject framework/toolchain claims that embed version-sensitive knowledge without current official or versioned evidence tied to the applicable version, or without an explicit reasoned exemption; repository-to-official research precedence stays intact.
- Reject artifacts with missing required assessment, unsupported package-manager certainty, capability claims exceeding tested support, hidden conflicts, migration decisions presented as facts, or semantic verdicts routed through Go. NO_ACTION and Tool Budget remain valid model-owned outcomes.
- Run the named checks: `framework_grounding`, `package_manager_accuracy`, `toolchain_accuracy`, and `external_verification_when_required`. REJECT an artifact that says npm when the repository uses Bun or pnpm, says pip when it uses uv, presents `cargo test` as a validated workflow when the repository has no tests, or uses a framework API without version verification when the rule requires it.

## Execution Steps
1. Load the review procedure reference.
2. Read the candidate and every reference it names; unresolved references fail validation.
3. Verify frontmatter: name pattern, directory match, description 1-1024 chars.
4. Trace evidence grounding; apply the External Verification Gate to version-sensitive claims; scan the anti-pattern catalog.
5. Score each rubric criterion; write the score sheet.
6. Return PASS, REVISE, or REJECT; record the sheet in state.

## Output Contract
Return the verdict and the full score sheet: per-criterion scores, total, and one grounded justification per criterion.
