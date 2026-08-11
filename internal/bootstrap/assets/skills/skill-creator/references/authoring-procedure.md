# Authoring Procedure

Procedure for authoring a candidate skill. Read this reference when authoring begins, after CREATE is decided; it resolves how to write so the candidate can pass the rubric.

## Inputs
- Approved artifact decision (CREATE + skill name), the proposal, and the evidence it cites.
- skill-quality-rubric.md: weights sum to 100; >= 85 PASS, 70-84 REVISE, < 70 REJECT.

## skill_request contract

CREATE ships a structured skill_request declaring: purpose, repository_evidence, framework_evidence, external_verified_evidence, canonical_examples, invariants, commands, and validation. Reflect every field in the artifact, or explicitly waive it with a reason. When external_verified_evidence is required and empty: never invent external framework guidance — report the missing evidence in the artifact and in state.

## Procedure
1. Verify inputs: CREATE decided and at least one FACT supports it. Missing either: hold, do not author.
2. Name the skill `^[a-z0-9]+(-[a-z0-9]+)*$`; the directory uses the same name.
3. Write the description: one quoted line, trigger first, <= 250 chars.
4. Draft the body in order: Activation Contract, Hard Rules, Decision Gates, Execution Steps, Output Contract, References.
5. Keep the body at 180-450 tokens; deeper detail goes to `references/` (progressive-disclosure.md).
6. Ground every instruction in an evidence fact or a harness fact; flag any claim you cannot trace.
7. Self-score against the rubric and write the per-criterion score sheet. Below 85: fix the weak criteria and re-score; never hand off REVISE.
8. Hand the candidate and score sheet to skill-reviewer; acceptance happens only after that gate.

## DO / DON'T
DO: state the trigger first; write imperative instructions; keep examples minimal; link local references.
DON'T: decide necessity (decided before you run); dump repository content into context; write generic advice; exceed the hard body maximum.
