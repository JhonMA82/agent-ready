# Search Strategies

How to close an evidence gap with the smallest useful search. Read this reference when a decision lacks a fact and you must choose where to look. It resolves the search order, version pinning, and the stopping condition.

## Search ladder (normative order)

| Step | Source | Use when |
|---|---|---|
| 1 | Repository itself: files, docs, configs, history | The answer may already exist in the repo; always check first |
| 2 | Local documentation: repo docs, `.agent-ready/references/` | Repo facts need a local interpretation rule |
| 3 | Version metadata: pinned versions, dependency files, tool versions | The answer depends on the exact version in use |
| 4 | Official documentation for the exact version | The version is known and the behavior is in question |
| 5 | A specialized provider for the tool or language | Official docs do not cover the case |
| 6 | Broader web only if necessary | All prior steps failed; last resort, never a default |

## Rules

- **Concrete question**: state what decision the research supports before searching; a question without a decision is a topic, not research.
- **Exact version**: search against the version actually in the repository — pinned OpenCode 1.18.15, the dependency version, or the version the question names. Never accept advice for an unspecified or different version.
- **Provenance**: every answer records its source and version; a source-less answer is UNKNOWN, not FACT.
- **Stopping condition**: stop at the first ladder step that answers the question with a source. Escalating the ladder is justified only when the current step fails.
- **No tool, no block**: absence of a search provider never blocks the audit; derive from local sources or stop with ASK_USER (tool candidates may be noted; Tool Manager is out of scope).
