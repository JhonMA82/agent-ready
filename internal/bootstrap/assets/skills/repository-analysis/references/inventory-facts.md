# Inventory Facts

How to gather the deterministic evidence base. Read this reference when exploration starts or when a stage lacks facts it needs. It resolves which sources to consult and in what order.

## Sources

| Source | Provides | Used for |
|---|---|---|
| Repository files | Dependencies, scripts, configs, docs | exploration_plan |
| inspect | Dependency/script/workspace/file/CI facts (agent-ready.inspect/v1) | Plan and evidence stages |
| state | Read-only facts over decisions.jsonl, provenance.jsonl, artifact-graph.yaml, repository-profile.yaml | Resume and reuse |
| changes | Changed paths since the checkpoint baseline | Resume and selective sync |
| checkpoint status | Current stage and completeness | Every run start |
| tools status | Tool/capability facts by family (ecosystem, productivity, provider) | Tool / Capability Assessment |
| tools recommend | Candidate evidence with capability truth and reasons | Tool / Capability Assessment |

## Gathering rules

- Reuse checkpointed evidence before re-reading sources (token-discipline.md).
- Gather only what the current stage needs; stop when the decision is supported.
- Record each fact with its source; helper output is identified by its schema name.
- Never read whole files into context; extract the fact and label it.
- Feed every successful audit's Tool / Capability Assessment: all three families with evidence and reasons, or `NO_ADDITIONAL_TOOLS` with a reason.

## repository_profile contract

Build the repository profile with these fields:

```yaml
repository_profile:
  kind:
    primary: boilerplate | starter | template | application
    secondary: []
    confidence: <0.0-1.0>
  ecosystems: []
  central_frameworks: []
  existing_agent_assets:
    agents_md: null | {path, lines}
    local_skills: []
    external_skills: []
    scripts: []
  context_placement:
    always_on_estimate: <qualitative>
    task_specific_guidance_candidates: []
  tool_assessment:
    ecosystem: []
    productivity: []
    providers: []
```

Kind and confidence are classified per the repository-analysis skill (boilerplate/starter/template trigger the boilerplate audit).

## Frequency discipline

Infer usage frequency from the repository when possible — always, common, occasional, rare. When it cannot be inferred, record `unknown`: never invent frequency, and record the uncertainty instead.
