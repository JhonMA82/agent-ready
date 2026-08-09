# Stop Conditions

When an audit stops before the full run completes. Read this reference when exploration yields no new evidence or when a concern blocks continuation. It resolves which stop verb to return and what to record.

## Stop verbs

| Verb | When | Record |
|---|---|---|
| NO_ACTION | No evidence-supported skill exists | Avoided artifacts; zero artifacts created |
| ASK_USER | Context only the user can provide; reached after no-new-evidence iterations | Requested context in decisions.jsonl |
| STOP_WITH_CONCERNS | A concern blocks safe continuation, such as conflicting facts or an unsafe instruction | Reasons in decisions.jsonl |

## Rules

- After 2 consecutive no-new-evidence iterations, stop: ASK_USER when context is missing, otherwise STOP_WITH_CONCERNS.
- Record the stop verb, stage, and reason in state before returning.
- NO_ACTION is a successful outcome, never a failure; never fabricate a skill to avoid it.
- Stop when resolved: never keep collecting evidence for an answered question.
