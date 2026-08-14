---
name: artifact-design
description: "Trigger: choosing what the audit will create or change. Decide CREATE, UPDATE, REUSE, REMOVE, NO_ACTION, ASK_USER, COMPACT, EXTRACT_TO_SKILL, MOVE_TO_REFERENCE, REPLACE_WITH_SCRIPT, or REUSE_EXTERNAL_SKILL from labeled evidence; no artifact spam."
---
Turn labeled evidence into a deliberate artifact decision. Load `references/artifact-decisions.md` when a decision is needed; read `../../references/skill-system/skill-quality-rubric.md` only when scoring a candidate. Before concluding REUSE on existing guidance, apply the Context Placement Gate.

## Activation Contract
Run when the loop reaches the artifact_decisions stage or when new evidence changes a prior decision.

## Hard Rules
- Decide from labeled evidence only; an evidence-backed request precedes any artifact.
- NO_ACTION is a first-class decision: the >=85 threshold governs NEW SKILL creation only. A candidate below the new-skill threshold may still justify COMPACT, EXTRACT_TO_SKILL (when extracting existing repository guidance), MOVE_TO_REFERENCE, REPLACE_WITH_SCRIPT, UPDATE, or REMOVE. NO_ACTION is valid only when no artifact or placement transformation is justified.
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

## Optional Artifact Types: Canonical Exemplar Catalog & Pattern Reference

Two optional artifact types may be proposed when repository-analysis's Pattern & Exemplar evidence supports them. They use only the existing verdicts — CREATE, UPDATE, REUSE, NO_ACTION; no new verdict exists for them.

### Canonical exemplar catalog

`CREATE` a `canonical exemplar catalog` only when: multiple useful examples exist; they represent distinct recurring intents; there is no adequate existing catalog; and future work would otherwise need repeated exploration. Recommended location `docs/ai/canonical-examples.yaml` — only if it fits the project structure; if the repository already has an equivalent location, reuse it. Never impose the path universally. The catalog must stay compact and be derived from the current repository (never copied from another project): schemaVersion; selectionRules (select the closest current example by intent; inspect no more than two examples unless more evidence is required; record intentional deviations; do not use legacy/deprecated entries as new-work references); examples with path/useFor/status; avoid entries with reasons.

### Pattern reference

`CREATE` a `pattern reference` only when: a stable non-trivial pattern repeats; the pattern is repository-specific; future work would benefit from explicit guidance; and the information is not already adequately documented. Use `docs/ai/patterns/<intent>.md` or `docs/patterns/<intent>.md` according to the existing structure. Patterns are keyed by repository intent, never by framework: `dashboard-screen.md`, `entity-management.md`, `ratatui-page.md`, `domain-service.md` are correct; `react-pattern.md`, `tailwind-pattern.md`, `rust-pattern.md` are not — and only when the evidence justifies it. Content describes only what is inferred from the repository: Use when; Canonical examples; Expected structure; Repository-specific invariants; Design/interaction expectations when applicable; Required states; Validation; Known deviations/avoidances. It must not become a framework tutorial.

### Design consistency and placement rules

- Design consistency derives from existing semantic primitives and composition patterns, never from literal duplication of classes or component markup. Correct: use existing semantic theme tokens; use the existing dashboard shell and spacing rhythm; follow the canonical responsive collapse pattern. Incorrect: copy these 43 Tailwind classes from dashboard X.
- When a catalog or pattern is created, do not copy its full content into `AGENTS.md`. At most add one compact router reference line when AGENTS.md is the right router (for example "For new UI/features, select the closest canonical example and applicable project pattern before implementation."); the detail stays on-demand.
- Creating a pattern does not justify creating a skill: a pattern records how this kind of implementation is shaped in the repo; a skill records how to execute a repeatable non-trivial workflow. `docs/ai/patterns/dashboard-screen.md` can exist without `.opencode/skills/add-dashboard-screen/`; a skill is only justified when it meets the existing rubric.

## Output Contract
Return the decision with its evidence and confidence, plus the recorded state entry; never a bare artifact count.
