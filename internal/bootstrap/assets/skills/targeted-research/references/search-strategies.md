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
- **Exact version**: search against the version actually in the repository — the installed OpenCode version (minimum compatible 1.18.15), the dependency version, or the version the question names. Never accept advice for an unspecified or different version.
- **Tool facts first**: when a decision involves tool capabilities, run `agent-ready tools status --json` (and `tools recommend --json` for candidate evidence) when available and treat the output as FACT. When the command is absent, reason from repository signals and note candidates without blocking — never invent tool state.
- **Provenance**: every answer records its source and version; a source-less answer is UNKNOWN, not FACT.
- **External Verification Gate**: before an answer embeds central, version-sensitive framework, package-manager, or toolchain knowledge, it MUST cite current official or versioned evidence tied to the applicable version, or carry an explicit reasoned exemption showing that no external claim is embedded. Existing repository-to-official research precedence stays intact: repository and local sources still come first, and official or versioned sources are consulted for the exact version in use.
- **Vocabulary**: output embedding central framework guidance MUST state `external_verified_evidence` non-empty, or carry the explicit exemption `external_verification_not_required`; the skill reviewer verifies this on every central-framework artifact.
- **Stopping condition**: stop at the first ladder step that answers the question with a source. Escalating the ladder is justified only when the current step fails.
- **No tool, no block**: absence of a search provider never blocks the audit; derive from local sources or stop with ASK_USER. Tool installation is never automatic during audit; tool/capability assessment is mandatory.
