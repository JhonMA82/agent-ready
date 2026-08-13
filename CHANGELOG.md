# Changelog

All notable changes to Agent-Ready are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## Source of truth

Every entry below is cross-checked against the repository commit history; the **Covered commits** line under each release lists the commits that introduced the described behavior. Release tags follow the goreleaser `v*` filter (`goreleaser.yml` → `release`). The V1 close is additionally checkpointed by `v1.0.0-facts` and `v1.0.0-sync-contract`; the final `v1.0.0` tag marks the archived, verified release.

## [v1.1.1] - 2026-08-12

Surgical output-contract test fix per `OPENCODE_AGENT_READY_SURGICAL_OUTPUT_TEST_FIX.md`: the driven audit and fixture oracles no longer let persisted state satisfy the visible output contract, and the placement/boilerplate assertions are structural.

**Covered commits**: `b14e960`, `7c5f918`, `b6194ce`

### Fixed

- **Visible output contract enforced over the OpenCode text stream only** — the driven audit and fixture oracles no longer concatenate persisted state with visible output; `Repository`, `Context Placement`, `Artifact Decisions`, `Tool / Capability Assessment`, and `Checkpoint` must appear in the visible response, and persisted files cannot make a visible-output test pass (`b14e960`).
- **Structural Context Placement evidence** — persisted verdicts must be a `decisions.jsonl` record identifiable as the `context_placement` stage/type carrying a subject, decision/verdict, and reason/evidence; a bare `{"decision":"NO_ACTION"}` no longer satisfies the assertion (`b14e960`).
- **Structural Boilerplate Assessment** — starter/boilerplate/template repositories must persist the assessment covering extension points, editable boundaries, generated files, feature-addition workflow, and upgrade strategy; `not_found` statuses are valid, absence is not (`b14e960`).
- **NO_ACTION no longer implied by a threshold miss** — the `>=85` rubric threshold governs new-skill creation only; `NO_ACTION` requires no justified artifact change, no justified context-placement transformation, and a complete tool/capability assessment (`7c5f918`).
- **REUSE requires appropriate placement** — REUSE means existing guidance/artifact/skill covers the need AND its current context placement is appropriate (`7c5f918`).

## [v1.1.0] - 2026-08-12

Context placement refinement (placement signals, fixtures q-u) plus the final corrective output contracts per `OPENCODE_AGENT_READY_FINAL_CONTEXT_FIX.md`: mandatory Repository Profile, Context Placement and Boilerplate Assessment evidence before REUSE/NO_ACTION.

**Covered commits**: `12a1730`, `7a3f678`, `0332741`, `0cf1339`, `82cd136`, `68dad75`, `ef5c7b3`

### Added

- **Context placement contract** — placement hierarchy (AGENTS.md → skill → reference → script), Context Placement Gate before any REUSE conclusion, placement analysis contract with qualitative cost model, usage-frequency discipline, and five new decision verbs: `COMPACT`, `EXTRACT_TO_SKILL`, `MOVE_TO_REFERENCE`, `REPLACE_WITH_SCRIPT`, `REUSE_EXTERNAL_SKILL`.
- **Extraction support** — `skill-creator` authors extractions without duplication (router left in source, placement provenance recorded); `skill-reviewer` runs placement checks (`context_savings`, `duplication_after_extraction`, `discoverability_preserved`, `always_on_guidance_not_removed`).
- **Repository kind** — `repository-analysis` classifies starter/boilerplate/template with primary/secondary/confidence and runs the boilerplate-specific audit (extension points, generated files, feature workflow, upgrade strategy).
- **Placement and tool signals** — `inspect --json` emits the `agents_md` fact; `tools recommend --json` emits `context_placement_pressure` (AGENTS.md > 300 lines, signal only), `structured_search_need` (ast-grep), and an enriched RTK candidate fired by output dirs, build/test scripts, or CI (no longer only `dist/build/coverage`).
- **Sync extension** — placement-change detection (AGENTS/skill/reference/canonical example), artifact-graph relations (`derived_from`/`routed_from`/`refresh_when`), placement provenance, and new tool-reassessment triggers.
- **Regression fixtures Q–U** — tanstack-shadcn boilerplate, long AGENTS (500+ lines), short optimal AGENTS, deterministic procedure, external canonical skill.
- **Final output contracts** — the orchestrator output contract mandates Repository Profile, Context Placement, Artifact Decisions, Tool / Capability Assessment and Checkpoint; REUSE/NO_ACTION require the Context Placement Gate, and `checkpoint --complete` is blocked when a placement verdict is missing but existing guidance informed the conclusion (`68dad75`).
- **Verdict routing** — `artifact-design` routes all 11 verdicts (CREATE/UPDATE/REUSE/REMOVE/COMPACT/EXTRACT_TO_SKILL/MOVE_TO_REFERENCE/REPLACE_WITH_SCRIPT/REUSE_EXTERNAL_SKILL/NO_ACTION/ASK_USER); the >=85 threshold applies only to new-skill creation.
- **Repository profile persistence** — `repository-analysis` persists `.agent-ready/state/repository-profile.yaml` (kind.primary, kind.confidence, topology) and runs the Boilerplate Assessment (extension points, editable boundaries, generated files, feature workflow, variants, scaffolding, upgrade strategy) for starter/boilerplate/template kinds; `monorepo` is topology, never a kind.
- **Contract tests** — content tests lock the orchestrator and artifact-design contracts; the driven audit requires observable Repository + Context Placement evidence, an explicit RTK evaluation in Productivity, and persisted profile/placement/boilerplate state; the driven tanstack cohort reuses acceptance `fixture-q` without duplicating fixtures (`68dad75`).

### Changed

- **Skill quality rubric** — added `context_placement` (weight 10); rebalanced to necessity 20, repository-specificity 15, discovery description 15, procedural value 15, progressive disclosure 10, evidence grounding 10, validation 5.
- **Driven fixture reuse** — the tanstack-starter driven cohort points at acceptance `fixture-q`; long-agents/short-optimal/deterministic coverage stays in the acceptance harness (fixtures r/s/t), so example projects are not duplicated under `driven/` (`68dad75`).

## [v1.0.0] - 2026-08-11

Covers the V1 completion change, six bounded close slices, the facts remediation checkpoint `c180e97`, and the sync-contract checkpoint `3a038f4` that bring `OPENCODE_AGENT_READY_V1_COMPLETION_2026.md` (§1–§67) to full compliance. The archived SDD report records 40/40 requirements and 83/83 scenarios passing.

**Covered commits**: `40ff819`, `8a35780`, `bba4806`, `4e6bb64`, `b65cf9b`, `304642b`, `df80798`, `dc15a5b`, `62cad46`, `d2fbbbf`, `ff57787`, `f440c56`, `29b8103`, `dc0c62a`, `2650e18`, `0e43d46`, `1b9349d`, `9912ae1`, `c180e97`, `3a038f4`

### Added

- **Ecosystem detection matrix** — repository ecosystem detection for all §7.x families, extended to the full matrix with suffix rules, the complete §8 lockfile set, output/build signals, and framework facts carrying version evidence and centrality (`bba4806`, `ff57787`).
- **Manager families** — package-manager conflict resolution and manager families; wrapper scripts (`gradlew`, `mvnw`) preferred over global tools (`4e6bb64`, `29b8103`).
- **Tool capability truth** — independent tool-support reporting; 25 detect-only entries with install never claimed; verified RTK, uv, and composer install recipes; new `tools explain` verb (`b65cf9b`, `dc0c62a`, `2650e18`).
- **Provider lifecycle truth** — serena and headroom catalog entries; Context7 gated on real need (framework facts with version, never lockfile-only); `tools doctor` health checks; opt-in global integration, default off (`0e43d46`).
- **Content contracts** — §15 Tool Budget ordering, §28 seven questions, §29 `skill_request` fields, §30 named reviewer checks, and §54 vocabulary enforced in the embedded skills (`1b9349d`).
- **Fixture matrix** — deterministic §33–§44 ecosystem tables plus NixOS Wizard and Laravel driven regression cohorts proving §52–§61 acceptance (`9912ae1`).
- **External verification gate** — reviews enforce external verification when required (`62cad46`).
- **Complete ecosystem and boilerplate facts** — PHP/Composer, Rust/Cargo, Nix, Dart, SwiftPM, container/IaC signals, empty-version framework evidence, and generated/editable file facts (`c180e97`, checkpoint `v1.0.0-facts`).

### Changed

- **Ownership-safe upgrades** — harness upgrades reconcile ownership and never clobber user-modified assets (`40ff819`).
- **Heavy-tree pruning** — repository traversal skips heavy trees; presence-only signals retained (`8a35780`).
- **Evidence-only recommendations** — tools recommended from ecosystem evidence; no lockfile-only or manager-conflict candidates (`304642b`).
- **Mandatory tool assessment** — audits mandate tool/capability assessment (`df80798`).
- **Relevant-sync reassessment** — tools reassessed on relevant syncs (`dc15a5b`).
- **OpenCode version decoupling** — harness no longer depends on an exact OpenCode version (`d2fbbbf`).
- **Explicit sync handoff** — sync loads its specialized flow, records reason-bearing relevance/reassessment decisions, and persists model-owned state before returning (`3a038f4`, checkpoint `v1.0.0-sync-contract`).

### Fixed

- Keyed ecosystem signal literals corrected in the detection matrix (`f440c56`).

## [v0.1.1] - 2026-08-09

Installer shipped as a release asset with a one-liner install.

**Covered commits**: `7e8f3ed`, `a2aa2ac`, `7bc55bf`

### Added

- Official project documentation: README with feature overview and direct installation explanation (`7e8f3ed`, `a2aa2ac`).
- Installer as a GitHub release asset enabling the one-liner `curl -fsSL https://github.com/JhonMA82/agent-ready/releases/latest/download/install.sh | sh` — checksum-verified (fails closed on mismatch), never requires sudo, installs to `$PREFIX/bin` (default `~/.local/bin`) (`7bc55bf`).

## [v0.1.0] - 2026-08-09

Initial public release: CLI foundation, docs, and installer.

**Covered commits**: full pre-release history to `6a07341` (34 commits); release-era commits `06731a1`..`6a07341`.

### Added

- Agent-Ready CLI foundation: `init`, `update`, `remove` lifecycle commands; deterministic fact helpers; tool catalog with `status`, `recommend`, `doctor`, and consent-gated `install`.
- Embedded agent-ready skills and acceptance fixtures shipped as owned, upgradeable assets.
- Release distribution: version metadata and `--version` flag, goreleaser pipeline, installer script with install guide and installer test harness (`06731a1`, `d0065d7`, `4a924cc`).

### Fixed

- Private-repository release downloads: authenticated asset download, POSIX fetch wrapper, API asset-download fallback, install env vars exported before verification steps, repo directory created before git init, fake opencode reporting the pinned version in the verify job (`86232c5`, `f21213a`, `5f7eead`, `9aed211`, `17822be`, `6a07341`).

---

[v1.1.1]: https://github.com/JhonMA82/agent-ready/releases/tag/v1.1.1
[v1.1.0]: https://github.com/JhonMA82/agent-ready/releases/tag/v1.1.0
[v1.0.0]: https://github.com/JhonMA82/agent-ready/releases/tag/v1.0.0
[v0.1.1]: https://github.com/JhonMA82/agent-ready/releases/tag/v0.1.1
[v0.1.0]: https://github.com/JhonMA82/agent-ready/releases/tag/v0.1.0
