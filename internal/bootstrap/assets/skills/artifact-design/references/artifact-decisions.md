# Artifact Decisions

The six decision verbs of the audit loop. Read this reference when a decision point needs an artifact verdict. It resolves what each verb means, what evidence it requires, and when it applies.

## Decision table (normative)

| Decision | Meaning | Requires |
|---|---|---|
| CREATE | Author a new skill with skill-creator | Labeled evidence + rubric score >= 85 |
| UPDATE | Revise an existing skill for changed evidence | Changed evidence; no full re-author |
| REUSE | An existing skill already covers the need | Evidence the coverage exists |
| REMOVE | Drop a shipped skill (sync scope only) | Evidence the skill no longer applies; approval flows (MERGE/DEPRECATE/REMOVE) are out of scope |
| NO_ACTION | No evidence-supported artifact exists | Scored candidates, all below threshold |
| ASK_USER | Stop and request missing context | Named unknown only the user can resolve |

## Rules

- Evidence first: the labeled evidence set precedes the decision; a decision without evidence is not recorded as one.
- No artifact spam: prefer REUSE, keep the set minimal, and record avoided artifacts; more artifacts is not more success.
- UNKNOWN-only evidence never decides CREATE; it decides gather-more or ASK_USER.
- NO_ACTION is success: a repo that needs no new skills produces no artifacts.
- "N skills generated" is never accepted as evidence of progress (R11).
- Hidden conflicts fail review: conflicting package-manager or ecosystem evidence is surfaced with the decision, never silently resolved.
- Migration is a proposal, not a fact: a migration decision requires evidence and approval; presenting it as a fact fails review.
- Certainty is bounded by tested support: package-manager certainty and capability claims never exceed `tools status --json` support states.
- Every decision is recorded in state (decisions.jsonl) with its evidence and confidence.

## Seven questions (decision gate)

Before deciding on a skill, answer all seven with evidence:
1. Is the need repository-specific?
2. Is it repeatable?
3. Is it non-trivial?
4. Does it contain project-specific decisions or invariants?
5. Do AGENTS/docs solve it more cheaply?
6. Does a deterministic script solve it better? A script that fully solves the need MUST NOT be created as a skill: choose NO_ACTION or REUSE.
7. Does framework-specific guidance require external verification?

The decision output is one of CREATE, UPDATE, REUSE, NO_ACTION, or ASK_USER (REMOVE is sync-scope only). All seven questions are answered with evidence on every decision point.
