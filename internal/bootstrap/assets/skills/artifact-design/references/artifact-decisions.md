# Artifact Decisions

The eleven decision verbs of the audit loop. Read this reference when a decision point needs an artifact verdict. It resolves what each verb means, what evidence it requires, and when it applies.

## Decision table (normative)

| Decision | Meaning | Requires |
|---|---|---|
| CREATE | Author a new skill with skill-creator | Labeled evidence + rubric score >= 85 |
| UPDATE | Revise an existing skill for changed evidence | Changed evidence; no full re-author |
| REUSE | An existing skill already covers the need | Evidence the coverage exists AND placement is optimal |
| REMOVE | Drop a shipped skill (sync scope only) | Evidence the skill no longer applies; approval flows (MERGE/DEPRECATE/REMOVE) are out of scope |
| NO_ACTION | No evidence-supported artifact exists | Scored candidates, all below threshold |
| ASK_USER | Stop and request missing context | Named unknown only the user can resolve |
| COMPACT | Keep the same artifact, reduce permanent context | Existing artifact carries more always-on weight than its frequency justifies |
| EXTRACT_TO_SKILL | Move a task-specific procedure from AGENTS/docs into a skill | Task-specific + procedural + repeatable + repository-specific guidance; router left in the source |
| MOVE_TO_REFERENCE | Move excessive detail to reference/docs and leave a short router | Detail that should not load every session; not a standalone procedure |
| REPLACE_WITH_SCRIPT | Replace long deterministic instructions with a validatable helper | Steps need no semantic judgment; the script can be validated automatically |
| REUSE_EXTERNAL_SKILL | Reuse a canonical external skill without a local wrapper | Canonical coverage exists; a wrapper adds nothing |

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

## Context Placement Analysis contract

Every important existing-guidance candidate can produce this contract. It may be recorded in state without showing literal YAML on screen:

```yaml
context_placement:
  subject: <guidance subject>

  current:
    location: <AGENTS.md | skill | reference | script>
    loaded: <every_session | on_demand>
    size_estimate: <e.g. 13-step-procedure>

  applicability:
    always_needed: false
    task_specific: true
    procedural: true
    repository_specific: true

  alternatives:
    - keep_in_agents
    - extract_to_skill
    - move_detail_to_reference

  recommended:
    action: <keep_in_agents | extract_to_skill | move_detail_to_reference | replace_with_script>

  expected_effect:
    permanent_context: <decrease | unchanged | increase>
    on_demand_context: <increase_only_when_relevant | ...>
    discoverability: <preserved | improved | reduced>

  evidence:
    - <source paths>
```

## Context cost model

Never compute tokens with fake precision; use qualitative classes: VERY_LOW, LOW, MEDIUM, HIGH, VERY_HIGH over the dimensions always_loaded_cost, on_demand_cost, duplication, frequency_of_use, discoverability, maintenance_cost.

```yaml
context_cost:
  current:
    always_loaded_cost: HIGH
    frequency_of_use: LOW
  proposed:
    always_loaded_cost: LOW
    on_demand_cost: MEDIUM
  result:
    optimize: true
```

## Usage frequency

Use the vocabulary always, common, occasional, rare, unknown — infer it from the repository when possible. Never invent frequency: if it is `unknown`, record the uncertainty. Do not move content just because it is long.

## Duplication check order

Before creating a skill, check in order: existing AGENTS, local skills, canonical external skills, docs, scripts. If the information already exists, prefer REUSE, COMPACT, EXTRACT_TO_SKILL, MOVE_TO_REFERENCE, REPLACE_WITH_SCRIPT, REUSE_EXTERNAL_SKILL over duplication.

## Context Placement Gate

Before concluding REUSE on existing guidance:

```text
Need already covered?
    ↓ yes
Is placement optimal?
    ↓
yes → REUSE
no
 ↓
COMPACT / EXTRACT_TO_SKILL / MOVE_TO_REFERENCE / REPLACE_WITH_SCRIPT
```

REUSE stays valid only when keeping the guidance always-on is the better choice; otherwise the placement verbs apply.

## Placement questions (gate)

Answer these six with evidence before concluding REUSE on existing guidance:
1. Is this guidance always applicable?
2. Is it task-specific?
3. Is it procedural?
4. Is it too detailed for always-on context?
5. Would extraction reduce token cost?
6. Would extraction harm discoverability?

## Seven questions (decision gate)

Before deciding on a skill, answer all seven with evidence:
1. Is the need repository-specific?
2. Is it repeatable?
3. Is it non-trivial?
4. Does it contain project-specific decisions or invariants?
5. Do AGENTS/docs solve it more cheaply?
6. Does a deterministic script solve it better? A script that fully solves the need MUST NOT be created as a skill: choose NO_ACTION or REUSE.
7. Does framework-specific guidance require external verification?

The decision output is one of CREATE, UPDATE, REUSE, NO_ACTION, or ASK_USER (REMOVE is sync-scope only), extended by the placement verbs COMPACT, EXTRACT_TO_SKILL, MOVE_TO_REFERENCE, REPLACE_WITH_SCRIPT, REUSE_EXTERNAL_SKILL. All seven questions are answered with evidence on every decision point; the placement questions above gate every REUSE conclusion.
