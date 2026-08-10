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
