# Fixture Matrix Specification

## Purpose

Make the V1 fixture matrix deterministic and provable: tables for ecosystem correctness, bounded driven regressions with structural oracles, and acceptance proof for no-spam, minimalism, conflict, wrapper, and mixed-ecosystem behavior.

## Requirements

### Requirement: Deterministic ecosystem correctness tables

Go table fixtures MUST exist for the mandatory ecosystems: Laravel (§33), Python uv (§34), Python pip (§35), JavaScript managers including npm, pnpm workspace, Bun, and Deno (§36), JVM wrapper (§37), .NET/NuGet (§38), Ruby/Bundler (§39), Elixir/Mix (§40), Flutter/pub (§41), and C/C++ CMake plus Conan (§42). Each fixture MUST assert §52 correctness: PHP is not Node; Python is not always uv; package.json is not always npm; and Cargo.lock, flake.lock, composer.lock, Gemfile.lock, pubspec.lock, and mix.lock are recognized.

#### Scenario: Laravel table

- GIVEN a composer.json, artisan, package.json, and bun.lock fixture
- WHEN detection runs
- THEN PHP, Composer, and Laravel facts MUST appear
- AND the JS manager MUST resolve to Bun

#### Scenario: uv and pip tables

- GIVEN the uv fixture (pyproject.toml plus uv.lock)
- WHEN detection runs
- THEN uv MUST be confirmed and pip MUST NOT be primary
- AND GIVEN the pip fixture (pyproject.toml plus requirements.txt, no higher-confidence lockfile)
- THEN pip-compatible MUST be inferred and uv MUST NOT be forced

### Requirement: Deterministic conflict and wrapper tables

Table fixtures MUST prove §59: `pnpm-lock.yaml` plus `bun.lock` without clear evidence resolve to a retained conflict with both candidates and no arbitrary choice. Table fixtures MUST prove §60: `./gradlew` and `./mvnw` are preferred over global Gradle/Maven and no global install is recommended.

#### Scenario: Conflict retained

- GIVEN the §59 fixture
- WHEN manager facts are emitted
- THEN both managers and a conflict reason MUST appear
- AND no selection MUST be made

#### Scenario: Wrapper preferred

- GIVEN a Gradle fixture with gradlew
- WHEN facts are emitted
- THEN the wrapper MUST appear as stronger evidence than a global binary
- AND no global install MUST be recommended

### Requirement: Mixed-ecosystem facts matrix

Fixtures MUST prove §61: Rust plus Cargo plus Nix (nixos-wizard shape), PHP plus Composer plus JS frontend (Laravel shape), and Dart plus native subprojects (Flutter shape) are all supported in one repository and never classified with a single exclusive label.

#### Scenario: Rust plus Nix facts

- GIVEN the mixed fixture
- WHEN detection runs
- THEN Rust and Nix MUST both appear with their own paths

#### Scenario: No exclusive label

- GIVEN any mixed fixture
- WHEN inspect JSON is emitted
- THEN no single ecosystem field MUST select a primary

### Requirement: Monorepo provider assessment fixture

A monorepo fixture with multiple workspaces/packages MUST exercise provider evaluation (§43): candidates for semantic_retrieval, symbol_navigation, and dependency_graph MUST be evaluated with evidence; there MUST be no obligation to recommend all of Semble, Serena, and CodeGraph; the model-owned Tool Budget MUST be applied to the minimal set.

#### Scenario: Monorepo candidates

- GIVEN the monorepo fixture
- WHEN recommendations run
- THEN provider candidates MUST appear with workspace evidence
- AND the set MUST NOT be forced to include all three providers

### Requirement: Boilerplate assessment fixture

A boilerplate/starter fixture MUST expose detection of extension points, generated-vs-editable files, variants, scaffolding, downstream customization, and upgrade workflow (§44). A project-specific skill MAY be proposed only with evidence; generic language or scaffold skills MUST NOT be generated.

#### Scenario: Extension points detected

- GIVEN a boilerplate fixture
- WHEN inspection runs
- THEN generated-vs-editable markers and extension points MUST appear as facts

#### Scenario: No generic scaffold skill

- GIVEN the same fixture
- WHEN the audit completes
- THEN no generic scaffold or language skill MUST be proposed without evidence

### Requirement: No-artifact-spam acceptance

Adding ecosystems MUST NOT add skills (§53). Ecosystem detection alone — Rust detected, Laravel detected, uv detected — MUST NOT suffice for a skill; only a specific, repeatable, non-trivial workflow justifies one. Driven cohorts MUST assert `NO_ADDITIONAL_TOOLS` or a rejected generic skill where applicable.

#### Scenario: Rust alone is not enough

- GIVEN a Rust fixture with no specific workflow
- WHEN the audit completes
- THEN no Rust skill MUST be proposed
- AND any NO_ACTION decision MUST carry a reason

### Requirement: Provider minimalism acceptance

Small-repo fixtures MUST accept `CodeGraph NOT_JUSTIFIED`, `Serena NOT_JUSTIFIED`, and `Semble NOT_JUSTIFIED` as valid outcomes (§55). Complex fixtures MUST seriously evaluate at least one specialized provider. The outcome `install everything` MUST NOT be accepted anywhere.

#### Scenario: Small repo verdicts

- GIVEN a small fixture
- WHEN provider evaluation completes
- THEN NOT_JUSTIFIED outcomes with reasons MUST be valid
- AND no install-everything outcome MUST appear

### Requirement: Driven regression cohorts

Driven regression cohorts (model plus OpenCode) MUST cover at least the NixOS Wizard-shaped fixture (§32) as flagship — Rust detected, Cargo.lock detected, Nix detected, flake.lock detected, Ratatui centrality recognized, stale pnpm guidance detected, tool assessment shown, external verification considered or used for a Ratatui artifact, generic Rust skill rejected — and a Laravel fixture (§33). Monorepo and boilerplate MAY be covered by deterministic oracles instead of driven runs. Driven additions MUST be bounded (1–2 new cohorts) and the battery MUST stay within the documented `-timeout 40m` guardrail. Cohorts MUST be unseeded: assertions MUST check structural properties via oracles, not pre-seeded expected conclusions.

#### Scenario: NixOS Wizard regression

- GIVEN the fixture repository
- WHEN the driven audit runs
- THEN the §32 structural checklist MUST pass via oracle

#### Scenario: Unseeded cohort

- GIVEN a driven cohort
- WHEN its oracle runs
- THEN no pre-seeded JSONL or prose conclusion MUST be compared
- AND structural assertions MUST be the only oracle input
