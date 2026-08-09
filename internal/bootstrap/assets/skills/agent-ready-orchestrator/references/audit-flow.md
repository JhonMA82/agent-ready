# Audit Flow

The stages of a full `/agent-ready audit` run. Read this reference when an audit starts or when the loop must decide the next stage. It resolves how exploration, evidence, decisions, and review connect.

## Stage machine

| Stage | Work | Produces |
|---|---|---|
| exploration_plan | Gather facts (inspect, state, changes, checkpoint status; `tools status --json` when available); label findings FACT / INFERENCE / UNKNOWN | Exploration plan with labeled findings |
| targeted_context | Load only what the plan requires; never whole files | Smallest useful context |
| evidence | Collect missing facts; reuse checkpointed evidence | Labeled evidence set |
| artifact_decisions | Choose CREATE / UPDATE / REUSE / REMOVE / NO_ACTION / ASK_USER with confidence and rationale | Decision list |
| proposal | Evidence-backed proposal for each CREATE or UPDATE | Proposal |
| approval | User approval for each proposal; ASK_USER when context is missing | Approved proposal |
| apply | Model writes approved artifacts and state files directly (no apply helper) | Created artifacts and state records |
| review | skill-reviewer gate on every candidate, scored against `../../references/skill-system/skill-quality-rubric.md` | PASS / REVISE / REJECT |
| checkpoint | `checkpoint save --stage S`; `--complete` when the run finishes | Checkpoint envelope |

## Decisions

| Decision | Means |
|---|---|
| CREATE | Rubric-scored candidate with evidence; author with skill-creator |
| UPDATE | Revise an existing skill; changed evidence only |
| REUSE | An existing skill covers the need; create nothing |
| REMOVE | Sync-only scope; lifecycle approval flows (MERGE/DEPRECATE/REMOVE) are out of scope |
| NO_ACTION | No skill earns >= 85 on evidence; create no artifacts |
| ASK_USER | Stop and request missing context |

## Ordering

Stages are a map, not a script: resume skips completed stages, new evidence revisits earlier stages, and the first applicable stop condition ends the run.

## Tool capabilities

Tool facts come from `agent-ready tools status --json` when the command is available; `tools recommend --json` offers candidate evidence the model weighs against the tool budget. Without the Tool Manager, reason from repository signals and degrade gracefully — missing tool facts never block a stage.
