# Tool Capability Facts Specification

## Purpose

Make tool support and recommendations independently truthful across one catalog.

## Requirements

### Requirement: Compatible categorized tools contract

`tools status --json` MUST retain `agent-ready.tools/v1`, existing fields, types, and meanings while additive support metadata remains optional. A removal, rename, type/meaning change, or newly required field MUST use a new schema version. Output MUST be byte-stable: the three families MUST order as ecosystem, productivity, provider; tools MUST order by stable identifier; capability and evidence collections MUST use documented deterministic ordering. Additive metadata — install safety level, install methods, side effects, and integration mode — MUST remain optional and MUST NOT be required to read presence or version.

#### Scenario: Existing V1 reader receives capability facts

- GIVEN a reader consumes only `schema_version`, `os`, `package_manager`, and existing tool facts
- WHEN categorized support metadata is added
- THEN the original values MUST preserve their meanings
- AND no new field MUST be required to read presence or version

#### Scenario: Safety metadata arrives

- GIVEN a reader consumes `install` capability values
- WHEN entries gain safety-level and side-effect metadata
- THEN prior values MUST keep their meanings
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

Recommendations MUST identify a capability need, candidate, observed evidence, and reason, using ecosystem facts where relevant. Go-produced output MUST NOT decide tool centrality, Tool Budget, installation, artifact suitability, or final recommendation; those semantic verdicts remain model-owned. Candidates MUST derive from the expanded deterministic signal table — per-ecosystem output-directory evidence, the full V1 lockfile set, workspace signals, and manager conflicts — not from a narrow `dist`/`build`/`coverage` or minimal-lockfile list. The RTK candidate MUST remain evidence-only; whether real commands produce large output stays model-owned.

#### Scenario: No grounded candidate

- GIVEN no documented recommendation signal is observed
- WHEN recommendations run
- THEN candidates MUST be an empty ordered collection
- AND no default tool verdict MUST be synthesized

#### Scenario: Rust output triggers an RTK candidate

- GIVEN a Rust repository with `target/` output evidence and no other signal
- WHEN recommendations run
- THEN an RTK candidate MUST appear with the observed path evidence
- AND it MUST NOT claim install, configure, or integration support

### Requirement: Catalog truth fixtures

Every changed entry MUST include behavior-driving tests for each claimed capability. Every entry claiming install support MUST be proven by tests covering detect, version, plan, execute, and verify. New system-package-manager recipes MUST fail closed on hosts without a deterministic recipe. Recipes MUST NOT execute arbitrary shell: validated executables MUST exclude shell interpreters and pipe patterns. Fixtures MUST ship in the same slice as the behavior they test.

#### Scenario: Unsupported capability is claimed

- GIVEN an entry claims install or integration support without corresponding behavior tests
- WHEN the slice is reviewed
- THEN the slice MUST be rejected

#### Scenario: Unsupported host fails closed

- GIVEN a host package manager without a deterministic recipe for a tool
- WHEN an install plan is requested
- THEN an explicit error with remediation MUST be returned
- AND nothing MUST execute

### Requirement: Install safety levels

Every catalog entry with install support MUST declare one of the five safety levels: SAFE_RECIPE, VERSION_SENSITIVE, PROJECT_WRAPPER_PREFERRED, MANUAL, GLOBAL_SIDE_EFFECT. Levels MUST be surfaced in `tools status --json` and in install plans. Reference assignments: rg, fd, jq, gh, ast-grep, uv → SAFE_RECIPE; composer, rustup → VERSION_SENSITIVE; maven, gradle → PROJECT_WRAPPER_PREFERRED; pip → runtime-coupled, preferring `python -m pip`; RTK → SAFE_RECIPE plus a separate GLOBAL_SIDE_EFFECT declaration for its OpenCode integration. GLOBAL_SIDE_EFFECT MUST NOT be bundled with binary installation; it requires the separate opt-in defined by the provider-lifecycle-truth spec.

#### Scenario: Level surfaced

- GIVEN `tools status --json` for uv
- WHEN emitted
- THEN safety level SAFE_RECIPE MUST appear

#### Scenario: RTK split declaration

- GIVEN the rtk entry
- WHEN status and plan are emitted
- THEN binary-install safety and the separate global-side-effect declaration MUST appear as distinct metadata

### Requirement: System package-manager coverage

The CLI MUST detect apt, dnf, pacman, zypper, apk, brew, and winget. Recipes MUST be extended per tool where a deterministic recipe exists. AUR (yay/paru) MUST be opt-in only; the CLI MUST NOT select an AUR recipe automatically. Nix MAY be detected as an environment but MUST NOT be used as an automatic universal installer. The CLI MUST NOT execute `curl | sh`; recipe validation MUST reject shell interpreters and pipe patterns.

#### Scenario: New managers detected

- GIVEN zypper, apk, or winget hosts
- WHEN package-manager detection runs
- THEN the matching manager MUST be returned

#### Scenario: AUR requires opt-in

- GIVEN a host where only AUR helpers are available
- WHEN an install plan is requested
- THEN no AUR recipe MUST be selected automatically
- AND an opt-in-only remediation MUST be reported

#### Scenario: Shell pipe rejected

- GIVEN a recipe proposal using sh/bash with a pipe
- WHEN recipe validation runs
- THEN the recipe MUST be rejected

### Requirement: Full install coverage V1

`tools install` MUST provide verified recipes with post-install verification for rg, fd, jq, gh, ast-grep, RTK, and uv. Composer MUST support detect, version, explain, and plan on all V1 platforms, executing only where a deterministic recipe exists. Detect/recommend catalog entries MUST exist for npm, pnpm, yarn, bun, deno, uv, pip, poetry, pdm, pipenv, composer, cargo, rustup, go, mvn, gradle, dotnet, bundle, mix, dart, flutter, cmake, conan, nix, terraform, and tofu, without implying install support.

#### Scenario: Composer lifecycle

- GIVEN any V1 platform
- WHEN composer detect, version, explain, and plan are requested
- THEN they MUST succeed
- AND execution MUST happen only where a deterministic recipe exists

#### Scenario: Detect-only entries

- GIVEN status for npm, cargo, or terraform
- WHEN emitted
- THEN detect and version MAY be supported
- AND install MUST remain unsupported or declared MANUAL

### Requirement: RTK first-class support

RTK MUST be a first-class productivity entry supporting detect, version, recommend, install plan, install, verify, and OpenCode-integration detection. RTK binary installation MUST be separate from RTK OpenCode global integration; the separate opt-in prompt is defined by the provider-lifecycle-truth spec and MUST NOT appear during `agent-ready init`.

#### Scenario: Full RTK lifecycle

- GIVEN a supported package manager and explicit consent
- WHEN install rtk runs
- THEN the binary installs AND post-install verification runs

#### Scenario: Init stays silent

- GIVEN `agent-ready init`
- WHEN it completes
- THEN no RTK global-integration prompt MUST appear

### Requirement: Install UX contract

`tools install` MUST render before any execution: tool, kind, repository-required evidence, plan (platform, method, executable, args), and the three Changes lines — installs user-level/global executable; does NOT modify OpenCode; does NOT modify project dependencies — followed by `Proceed? [y/N]`. Consent MUST NOT default to yes; empty or unreadable input MUST decline.

#### Scenario: Complete plan rendered

- GIVEN `tools install uv`
- WHEN the plan is shown
- THEN tool, kind, evidence, plan fields, and all three Changes lines MUST appear before the prompt

#### Scenario: Empty input declines

- GIVEN empty input at the consent prompt
- WHEN confirmation is read
- THEN installation MUST NOT run

### Requirement: Tools explain verb

`agent-ready tools explain <tool>` MUST render the catalog entry's declared capabilities, safety level, install methods, side effects, and integration metadata without executing or installing anything. An unknown tool MUST produce an explicit not-in-catalog error.

#### Scenario: Explain a known tool

- GIVEN `tools explain uv`
- WHEN run
- THEN capability, safety-level, and method facts MUST render

#### Scenario: Explain an unknown tool

- GIVEN an id outside the catalog
- WHEN run
- THEN an explicit error MUST name the unknown id
- AND nothing MUST execute
