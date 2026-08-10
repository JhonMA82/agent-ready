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

## Tool / Capability Assessment

Every successful initial audit MUST end with a categorized Tool / Capability Assessment covering three families — ecosystem (gh, go, node), productivity (ast-grep, fd, jq, rg), and provider (codegraph, context7). Tool facts come from `agent-ready tools status --json`; `agent-ready tools recommend --json` offers candidate evidence the model weighs against the tool budget. The final output MUST name all three families verbatim — ecosystem, productivity, provider — and either list recommendations with observed evidence and a reason, or state `NO_ADDITIONAL_TOOLS` with a reason; never omit or abbreviate a family, and record the assessment in the output and in the recorded decisions. The model keeps Tool Budget and final recommendation verdicts. An absent tool never blocks a stage; reason from repository signals and degrade gracefully.
