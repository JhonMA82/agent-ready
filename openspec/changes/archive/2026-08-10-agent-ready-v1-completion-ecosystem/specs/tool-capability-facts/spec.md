# Tool Capability Facts Specification

## Purpose

Make tool support and recommendations independently truthful across one catalog.

## Requirements

### Requirement: Compatible categorized tools contract

`tools status --json` MUST retain `agent-ready.tools/v1`, existing fields, types, and meanings while additive support metadata remains optional. A removal, rename, type/meaning change, or newly required field MUST use a new schema version. Output MUST be byte-stable: the three families MUST order as ecosystem, productivity, provider; tools MUST order by stable identifier; capability and evidence collections MUST use documented deterministic ordering.

#### Scenario: Existing V1 reader receives capability facts

- GIVEN a reader consumes only `schema_version`, `os`, `package_manager`, and existing tool facts
- WHEN categorized support metadata is added
- THEN the original values MUST preserve their meanings
- AND no new field MUST be required to read presence or version

### Requirement: Independent capability truth

Each catalog entry MUST independently state support for detect, version, recommend, install, configure, integration, and side effects. Unsupported and unknown MUST remain distinguishable from supported. Presence or recommendation MUST NOT imply installation, configuration, or integration support. `install: supported` MUST have tested plan, execution, post-install verification, and explicit-consent behavior.

#### Scenario: Detect-only ecosystem tool

- GIVEN a tool has tested detection but no safe recipe
- WHEN status and recommendations are emitted
- THEN detect support MAY be true while install and configure support MUST be false
- AND the tool MUST NOT become installable by implication

#### Scenario: Provider candidate has no lifecycle support

- GIVEN repository evidence makes a provider relevant
- WHEN recommendation facts are emitted
- THEN the provider MAY appear with evidence and reasons
- AND install, configure, integration, and side-effect support MUST remain unsupported

### Requirement: Evidence-only recommendations

Recommendations MUST identify a capability need, candidate, observed evidence, and reason, using ecosystem facts where relevant. Go-produced output MUST NOT decide tool centrality, Tool Budget, installation, artifact suitability, or final recommendation; those semantic verdicts remain model-owned.

#### Scenario: No grounded candidate

- GIVEN no documented recommendation signal is observed
- WHEN recommendations run
- THEN candidates MUST be an empty ordered collection
- AND no default tool verdict MUST be synthesized

### Requirement: Catalog truth fixtures

Every changed entry MUST include behavior-driving tests for each claimed capability. This change MUST NOT add safe recipes, system-package-manager coverage, provider lifecycle, or global OpenCode/RTK integration.

#### Scenario: Unsupported capability is claimed

- GIVEN an entry claims install or integration support without corresponding behavior tests
- WHEN the slice is reviewed
- THEN the slice MUST be rejected
