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

## Gathering rules

- Reuse checkpointed evidence before re-reading sources (token-discipline.md).
- Gather only what the current stage needs; stop when the decision is supported.
- Record each fact with its source; helper output is identified by its schema name.
- Never read whole files into context; extract the fact and label it.
