# Audit Flow

The stages of a full `/agent-ready audit` run. Read this reference when an audit starts or when the loop must decide the next stage. It resolves how exploration, evidence, decisions, and review connect.

## Stage machine

| Stage | Work | Produces |
|---|---|---|
| exploration_plan | Gather facts (inspect, state, changes, checkpoint status; `tools status --json` when available); label findings FACT / INFERENCE / UNKNOWN | Exploration plan with labeled findings |
| targeted_context | Load only what the plan requires; never whole files | Smallest useful context |
| evidence | Collect missing facts; reuse checkpointed evidence | Labeled evidence set |
| context_placement | Evaluate whether existing guidance sits at the cheapest context level (AGENTS.md / skill / reference / script hierarchy) before any REUSE conclusion | Placement verdict per existing candidate |
| artifact_decisions | Choose CREATE / UPDATE / REUSE / REMOVE / NO_ACTION / ASK_USER with confidence and rationale | Decision list |
| proposal | Evidence-backed proposal for each CREATE or UPDATE | Proposal |
| approval | User approval for each proposal; ASK_USER when context is missing | Approved proposal |
| apply | Model writes approved artifacts and state files directly (no apply helper) | Created artifacts and state records |
| review | skill-reviewer gate on every candidate, scored against `../../references/skill-system/skill-quality-rubric.md` | PASS / REVISE / REJECT |
| checkpoint | `checkpoint save --stage S`; `--complete` when the run finishes | Checkpoint envelope |

The stage machine registers these names internally: exploration_plan, repository_analysis, context_placement, artifact_decisions, tool_assessment, approval, apply, review, checkpoint. Not every name needs a user-facing row: sub-stages (targeted_context, evidence, proposal) may be merged or reported inline when a full list would hurt the run output.

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

Every successful initial audit MUST end with a categorized Tool / Capability Assessment covering three families — ecosystem (gh, go, node), productivity (rg, fd, jq, RTK, gh when a GitHub remote exists, ast-grep when structured search may help), and provider (Context7, Semble, Serena, CodeGraph, Headroom). Productivity MUST name RTK individually with one explicit status — RECOMMENDED, CONSIDER, NOT_JUSTIFIED, or ALREADY_AVAILABLE — and its verdict rests on observed signals (command output volume, shell-heavy workflows), never only on dist/build/coverage presence. Each provider is explicitly evaluated; NOT_JUSTIFIED is a valid outcome with a compact reason. Tool facts come from `agent-ready tools status --json`; `agent-ready tools recommend --json` offers candidate evidence the model weighs against the tool budget. The final output MUST name all three families verbatim — ecosystem, productivity, provider — and either list recommendations with observed evidence and a reason, or state `NO_ADDITIONAL_TOOLS` with a reason; never omit or abbreviate a family, and record the assessment in the output and in the recorded decisions. The model keeps Tool Budget and final recommendation verdicts. An absent tool never blocks a stage; reason from repository signals and degrade gracefully.

## Tool Budget

The assessment applies the Tool Budget minimal-set ordering: for a small repository, rg + fd + jq suffice; add ast-grep only when structural search is needed — ast-grep precedes any semantic provider; add Semble OR Serena when semantic retrieval or navigation is justified; add CodeGraph only when the graph capability adds clear value; add Headroom only when context compression remains a measured problem. Recommend the minimal set that covers observed needs; every heavier tool carries explicit justification. The Tool Budget and final recommendations remain model-owned outcomes; installation is never automatic during audit, and tool/capability assessment is mandatory.

## Audit output template

Shape of a complete audit output (model-owned verdicts; the point is showing the analysis):

```text
Audit outcome: <CREATE | NO_ACTION | COMPACT | EXTRACT_TO_SKILL | MOVE_TO_REFERENCE | REPLACE_WITH_SCRIPT | REUSE_EXTERNAL_SKILL>

Repository
  <kind, ecosystems, central frameworks, file count>

Context Placement
  <per existing guidance: location, REVIEWED, task-specific?, decision and reason, alternative considered>

Artifacts
  <decision list>

Tools
  Productivity
    rg          <status>
    fd          <status>
    jq          <status>
    RTK         <RECOMMENDED | CONSIDER | NOT_JUSTIFIED | ALREADY_AVAILABLE>
    ast-grep    <status>
    gh          <status when GitHub remote exists>
  Providers
    Context7    <status or NOT_JUSTIFIED with compact reason>
    Semble      <status>
    Serena      <status>
    CodeGraph   <status>
    Headroom    <status>

Checkpoint
  saved
```

Compact provider output is allowed: `Providers: none justified. Evaluated: Context7, Semble, Serena, CodeGraph, Headroom.` RTK always appears individually in productivity.

## Token optimization objective

Agent-Ready optimizes both content and its distribution between always-on and on-demand context. Qualitative metrics: always_on_context_reduced, duplicate_guidance_avoided, targeted_skill_loads, references_loaded_on_demand, unnecessary_artifacts_avoided, tool_output_reduced.

## Token estimates

Never claim exact token savings unless measured. Use qualitative classes (VERY_LOW, LOW, MEDIUM, HIGH, VERY_HIGH), for example `expected permanent context reduction: MEDIUM`.

## Checkpoint gate

`checkpoint --complete` MUST NOT be emitted when:

```text
relevant existing guidance was used to justify
REUSE or NO_ACTION

AND

no Context Placement verdict was recorded.
```

If repository kind is:

```text
starter
boilerplate
template
```

then the Repository output MUST include Boilerplate Assessment.
