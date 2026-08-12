# Audit Evidence Gates Specification

## Purpose

Make assessment, external grounding, and driven evidence mandatory while preserving model authority.

## Requirements

### Requirement: Mandatory Tool / Capability Assessment

Every successful initial audit MUST emit a categorized Tool / Capability Assessment covering ecosystem, productivity, and provider families. Each recommendation MUST include observed evidence and a reason. If none is warranted, the assessment MUST explicitly emit `NO_ADDITIONAL_TOOLS` with a reason. The model MUST retain centrality, artifact, Tool Budget, and final recommendation verdicts.

#### Scenario: Initial audit has no candidates

- GIVEN an initial audit succeeds and no candidate survives model assessment
- WHEN the final audit output is produced
- THEN it MUST contain all three categories
- AND it MUST state `NO_ADDITIONAL_TOOLS` with a reason

### Requirement: Relevant sync reassessment

A sync MUST assess whether changed evidence can affect tool needs. Manifest, lockfile, workspace, wrapper, CI, framework, build/test output, or tool-fact changes MUST trigger reassessment; irrelevant changes MUST record a reason for skipping it. A completed reassessment MUST include reasons and either categorized recommendations or `NO_ADDITIONAL_TOOLS`.

#### Scenario: Relevant and irrelevant syncs

- GIVEN one sync changes a lockfile and another changes only prose
- WHEN each sync completes
- THEN the lockfile sync MUST reassess tool capabilities with reasons
- AND the prose sync MUST record why reassessment was unnecessary

### Requirement: External Verification Gate

Before an artifact embeds central, version-sensitive framework, package-manager, or toolchain knowledge, it MUST cite current official or versioned evidence tied to the applicable version, or carry an explicit, reasoned exemption showing that no external claim is embedded. Existing repository-to-official research precedence MUST remain intact.

#### Scenario: Framework artifact lacks grounding

- GIVEN an artifact prescribes a version-sensitive framework API
- WHEN no applicable versioned evidence or valid exemption is attached
- THEN the External Verification Gate MUST fail

### Requirement: Reviewer rejection contract

Reviewers MUST reject artifacts with missing required assessment, unsupported package-manager certainty, framework/toolchain claims that fail the External Verification Gate, capability claims exceeding tested support, hidden conflicts, migration decisions presented as facts, or semantic verdicts routed through Go. `NO_ACTION` and Tool Budget remain valid model-owned outcomes.

#### Scenario: Fact output chooses a migration

- GIVEN evidence contains conflicting package managers
- WHEN an artifact silently selects one and prescribes migration
- THEN review MUST reject the artifact

### Requirement: Behavior-driving evidence per slice

Each review slice MUST stay independently reviewable under 400 changed lines and ship fixtures that execute the behavior introduced in that slice. Seeded conclusions alone MUST NOT satisfy acceptance. Existing C–P regressions and `go test ./...` MUST remain green; initial-audit and relevant-sync contracts require driven audit fixtures.

#### Scenario: Fixture only repeats an expected conclusion

- GIVEN a slice changes audit behavior but only pre-seeds expected JSONL or prose
- WHEN acceptance is evaluated
- THEN the slice MUST be rejected until a fixture drives and observes that behavior

### Requirement: V1 scope boundaries

This change MUST NOT replace the Go CLI, local harness, `/agent-ready`, seven skills, checkpoints, ownership, JSON helpers, safe recipes, or model decisions. It MUST NOT add agents, commands, generic language/package-manager skills, TUI, daemon, database, required MCP, Go verdict routing, installer expansion, provider lifecycle, or global OpenCode/RTK integration.

#### Scenario: Slice introduces excluded integration

- GIVEN a slice adds provider installation or global integration
- WHEN scope is reviewed
- THEN the slice MUST be rejected or moved to its named follow-up change
