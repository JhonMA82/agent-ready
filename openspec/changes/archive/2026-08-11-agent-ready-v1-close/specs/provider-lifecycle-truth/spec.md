# Provider Lifecycle Truth Specification

## Purpose

Declare advanced-provider policy, lifecycle metadata, and health truth so recommendation stays evidence-only, installation stays consent-gated, and global side effects stay opt-in.

## Requirements

### Requirement: Provider policy catalog

The tools catalog MUST contain entries for Context7, Semble, Serena, CodeGraph, and Headroom, each declaring its capability honestly: Context7 `versioned_documentation`; Semble `semantic_retrieval`; Serena `symbol_intelligence`, `semantic_navigation`, `symbolic_editing`; CodeGraph `dependency_graph`, `call_graph`, `blast_radius`, `architecture_query`; Headroom `general_context_compression`. Recommendation candidates MUST remain evidence-only. Policy guidance on when to recommend MAY live in orchestrator skill assets, not in Go.

#### Scenario: All five providers present

- GIVEN `tools status --json`
- WHEN run
- THEN all five provider entries MUST appear with their declared capabilities

#### Scenario: Policy content not hardcoded

- GIVEN the catalog is inspected
- THEN Go output MUST contain no recommendation verdicts or thresholds for any provider

### Requirement: Context7 real-need gating

The Context7 candidate MUST NOT be generated from lockfile presence alone. It MUST be generated only when observed evidence shows a central, version-sensitive framework or dependency relevant to a proposed artifact, or that repository/local documentation is insufficient. When Context7 is unavailable, the fallback MUST be official documentation via web fetch.

#### Scenario: Lockfile-only repository

- GIVEN a repository with only lockfiles and no central framework evidence
- WHEN recommendations run
- THEN no Context7 candidate MUST appear

#### Scenario: Central version-sensitive framework

- GIVEN ratatui 0.29.0 is central and a proposed artifact teaches its API
- WHEN recommendations run
- THEN a Context7 candidate MUST appear with capability, signal, and observed evidence

### Requirement: Conditional provider evaluation

Provider candidates MUST be evaluated conditionally on observed signals: Semble when textual retrieval would reduce candidates (medium/large repositories); Serena when a supported language server and cross-file symbol-navigation needs exist; CodeGraph when multi-workspace or cross-package topology evidence exists — never from a files>N or deps>N threshold as a verdict; Headroom only when context-compression pressure is measured and RTK is insufficient. Each provider MAY legitimately resolve to NOT_JUSTIFIED on small repositories.

#### Scenario: Small repository

- GIVEN a small single-crate repository
- WHEN recommendations run
- THEN CodeGraph, Serena, and Semble candidates MAY be absent or carry not-justified evidence
- AND no default verdict MUST be synthesized

#### Scenario: Complex monorepo

- GIVEN workspace and cross-package signals
- WHEN recommendations run
- THEN candidates for semantic_retrieval, symbol_navigation, and dependency_graph MUST appear with observed evidence
- AND there MUST be no obligation to recommend all three providers

### Requirement: Provider metadata declarations

Each provider MUST declare, metadata-first: install method, project-init requirement, OpenCode integration mode, global/local side effects, uninstall, and health check. Only deterministic lifecycle MUST be implemented; anything else MUST be declared unsupported, and install support requires a verified recipe. Project-local configuration MUST be preferred when supported. Providers MUST NOT be installed or configured during audit without explicit approval.

#### Scenario: Unverified install declared unsupported

- GIVEN a provider has no verified recipe
- WHEN status is emitted
- THEN install, configure, and integration MUST be unsupported
- AND the declared install method MUST reflect reality

#### Scenario: Audit never installs

- GIVEN an audit runs with provider evidence
- WHEN it completes
- THEN no provider MUST be installed or configured

### Requirement: Provider health checks

`tools doctor` MUST validate for each provider: executable exists; version parses; project index/config exists when required; OpenCode integration is detectable when enabled; and provider health when inexpensive. A provider MUST NOT be reported healthy merely because a binary exists; each failed check MUST carry a reason.

#### Scenario: Broken version

- GIVEN a provider binary exists but version fails to parse
- WHEN doctor runs
- THEN the provider MUST be reported unhealthy with the failing check and reason

#### Scenario: Missing project index

- GIVEN a provider requires a project index and it is absent
- WHEN doctor runs
- THEN the provider MUST be reported unhealthy with a path-referencing reason

### Requirement: RTK global integration opt-in

After a successful RTK binary install, the CLI MUST ask a separate question — `Enable global integration? [y/N]` — describing that the integration modifies global OpenCode configuration. The default MUST be N. The question MUST NOT appear during `agent-ready init`. RTK MUST remain usable explicitly without the integration (`rtk cargo test`, `rtk git status`, `rtk read`).

#### Scenario: Separate opt-in prompt

- GIVEN an rtk binary install succeeds
- WHEN the flow continues
- THEN the global-integration prompt MUST appear with default N
- AND declining MUST leave global configuration untouched

#### Scenario: Init is silent

- GIVEN `agent-ready init` runs
- WHEN it completes
- THEN no integration prompt MUST be shown

### Requirement: OpenCode isolation invariants

`agent-ready init`, `/agent-ready`, `agent-ready tools install <normal tool>`, and `agent-ready remove` MUST NOT modify `~/.config/opencode` unless the user separately approves a global integration. The local `.agent-ready` integration via `skills.paths` MUST remain lossless. Tests MUST prove zero default global mutation.

#### Scenario: Default install is isolated

- GIVEN a normal tool install with consent and no approved global integration
- WHEN it completes
- THEN `~/.config/opencode` content MUST be unchanged

#### Scenario: Lossless local merge

- GIVEN an existing local OpenCode config
- WHEN init merges `.agent-ready/skills`
- THEN the resulting config MUST retain all prior keys and values
