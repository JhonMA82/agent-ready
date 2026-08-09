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
- Every decision is recorded in state (decisions.jsonl) with its evidence and confidence.
