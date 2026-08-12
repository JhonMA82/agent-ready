# Delta for Audit Evidence Gates

## MODIFIED Requirements

### Requirement: Mandatory Tool / Capability Assessment

Every successful initial audit MUST emit a categorized Tool / Capability Assessment covering ecosystem, productivity, and provider families. Each recommendation MUST include observed evidence and a reason. If none is warranted, the assessment MUST explicitly emit `NO_ADDITIONAL_TOOLS` with a reason. The model MUST retain centrality, artifact, Tool Budget, and final recommendation verdicts. The assessment MUST read tool status and recommendation facts, evaluate the three families, apply the Tool Budget minimal-set ordering, and render the result to the user; installation MUST never be automatic during audit.
(Previously: no explicit read-status, apply-budget, and render steps in the assessment contract.)

#### Scenario: Initial audit has no candidates

- GIVEN an initial audit succeeds and no candidate survives model assessment
- WHEN the final audit output is produced
- THEN it MUST contain all three categories
- AND it MUST state `NO_ADDITIONAL_TOOLS` with a reason

#### Scenario: Tool Budget applied

- GIVEN a small repository with baseline productivity tools
- WHEN the assessment is produced
- THEN the recommended set MUST follow the minimal-set ordering
- AND each recommendation MUST carry a reason

### Requirement: Reviewer rejection contract

Reviewers MUST reject artifacts with missing required assessment, unsupported package-manager certainty, framework/toolchain claims that fail the External Verification Gate, capability claims exceeding tested support, hidden conflicts, migration decisions presented as facts, or semantic verdicts routed through Go. `NO_ACTION` and Tool Budget remain valid model-owned outcomes. Reviewers MUST apply the named checks `framework_grounding`, `package_manager_accuracy`, `toolchain_accuracy`, and `external_verification_when_required`. Reviewers MUST reject when an artifact says npm but the repository uses Bun or pnpm; says pip but the repository uses uv; presents `cargo test` as a validated workflow when the repository has no tests; or uses a framework API without version verification when the rule requires it.
(Previously: rejection contract without named checks.)

#### Scenario: Fact output chooses a migration

- GIVEN evidence contains conflicting package managers
- WHEN an artifact silently selects one and prescribes migration
- THEN review MUST reject the artifact

#### Scenario: Named check fails

- GIVEN an artifact prescribes pip while repository evidence shows uv
- WHEN the review runs the named checks
- THEN `package_manager_accuracy` MUST fail
- AND the artifact MUST be rejected

### Requirement: V1 scope boundaries

This change MUST NOT replace the Go CLI, local harness, `/agent-ready`, seven skills, checkpoints, ownership, JSON helpers, safe recipes, or model decisions. It MUST NOT add agents, slash commands beyond `tools explain`, framework workflows, generic language/package-manager skills, mandatory MCP, TUI, daemon, database, telemetry, Go verdict routing, or a rewrite. Installer expansion, provider lifecycle, RTK global-integration opt-in, and the fixture matrix are in scope under consent-gated, metadata-first honesty.
(Previously: installer expansion, provider lifecycle, and global OpenCode/RTK integration were outright exclusions.)

#### Scenario: Slice introduces an excluded component

- GIVEN a slice adds an agent, a generic language skill, or mandatory MCP
- WHEN scope is reviewed
- THEN the slice MUST be rejected

#### Scenario: Consent-gated installer slice is in scope

- GIVEN a slice adds a verified recipe with the consent UX
- WHEN scope is reviewed
- THEN the slice MUST be within scope

## ADDED Requirements

### Requirement: Tool Budget rules

Orchestrator skill assets MUST state the explicit Tool Budget minimal-set ordering: rg + fd + jq; then + ast-grep when structural search is needed; then + Semble OR Serena when semantic retrieval or navigation is justified; CodeGraph only when graph capability adds clear value; Headroom only when context compression remains a measured problem. The Tool Budget MUST recommend the minimal set covering observed needs and MUST remain a model-owned outcome.

#### Scenario: Small repo minimal set

- GIVEN a small repository
- WHEN the Tool Budget is applied
- THEN the rg + fd + jq set MUST suffice
- AND heavier tools MUST carry explicit justification

#### Scenario: Ordering respected

- GIVEN evidence of a structural-search need
- WHEN the Tool Budget is applied
- THEN ast-grep MUST precede any semantic provider in the ordering

### Requirement: Artifact-design seven questions

Before deciding on a skill, artifact design MUST answer the seven questions: repository-specific; repeatable; non-trivial; contains project-specific decisions or invariants; AGENTS/docs solve it more cheaply; a deterministic script solves it better; framework-specific guidance requires external verification. The decision output MUST be one of CREATE, UPDATE, REUSE, NO_ACTION, or ASK_USER.

#### Scenario: All seven answered

- GIVEN a proposed skill
- WHEN artifact design runs
- THEN all seven questions MUST be answered with evidence

#### Scenario: Script cheaper

- GIVEN a deterministic script would fully solve the need
- WHEN artifact design runs
- THEN the skill MUST NOT be created
- AND NO_ACTION or REUSE MUST be chosen

### Requirement: Skill-creator request contract

`skill-creator` MUST receive a structured `skill_request` declaring purpose, repository_evidence, framework_evidence, external_verified_evidence, canonical_examples, invariants, commands, and validation. When external_verified_evidence is required and empty, the creator MUST NOT invent external framework guidance and MUST report the missing evidence instead.

#### Scenario: Complete request

- GIVEN a skill_request with all fields populated
- WHEN the creator runs
- THEN each field MUST be reflected in the artifact or explicitly waived

#### Scenario: Empty required evidence

- GIVEN external_verified_evidence required and empty
- WHEN the creator runs
- THEN no external framework guidance MUST be invented
- AND the missing evidence MUST be reported

### Requirement: Research-quality vocabulary

Targeted-research output MUST name `external_verified_evidence` non-empty for artifacts embedding central framework guidance, or MUST carry the explicit exemption `external_verification_not_required`. The skill reviewer MUST verify the vocabulary on every central-framework artifact.

#### Scenario: Evidence named

- GIVEN an artifact embeds central framework guidance
- WHEN research output is produced
- THEN external_verified_evidence MUST be non-empty

#### Scenario: Exemption named

- GIVEN research is not required
- WHEN output is produced
- THEN `external_verification_not_required` MUST be stated

### Requirement: Stale-phrase cleanup

Skill assets MUST NOT contain the phrase `tool management is out of scope` or its equivalents. The governing rule MUST read: tool installation is never automatic during audit, but tool/capability assessment is mandatory.

#### Scenario: Assets are clean

- GIVEN the skill assets of the delivered slices
- WHEN scanned
- THEN the stale phrase MUST NOT appear anywhere
