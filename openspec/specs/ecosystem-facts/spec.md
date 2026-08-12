# Ecosystem Facts Specification

## Purpose

Expose deterministic repository evidence without turning detection into migration advice.

## Requirements

### Requirement: Compatible inspect contract

`inspect --json` MUST retain `agent-ready.inspect/v1`, all existing fields, types, and meanings while ecosystem fields are additive. Additive fields MAY remain in V1; removing, renaming, changing a field type or meaning, or making a previously optional field required MUST use a new schema version. Repeated inspection of unchanged bytes MUST produce byte-stable JSON: object keys, evidence arrays, paths, ecosystems, managers, manifests, lockfiles, workspaces, wrappers, framework signals (including version and centrality-signal evidence), output-dir signals, and build/test-tool signals MUST use documented deterministic ordering. Exact-basename rules and suffix/extension rules MUST share the same deterministic ordering.

#### Scenario: Legacy consumer reads enriched V1

- GIVEN a consumer reads only the original V1 fields
- WHEN it receives enriched inspect JSON
- THEN those fields MUST retain their prior types and meanings
- AND unknown additive fields MUST not be required for interpretation

#### Scenario: Input order varies

- GIVEN equivalent repository entries are discovered in different filesystem orders
- WHEN inspection runs
- THEN the serialized JSON MUST be byte-identical

#### Scenario: Suffix-matched facts

- GIVEN a repository mixes `*.tf` and `*.csproj` files with basename manifests
- WHEN inspection runs
- THEN suffix-matched and basename-matched signals MUST interleave in the documented order
- AND repeated runs MUST be byte-identical

### Requirement: Multi-ecosystem presence evidence

Inspection MUST retain every evidenced ecosystem concurrently and identify its manifests, lockfiles, workspace signals, project wrappers, framework signals, and build/test-tool signals. The V1 matrix MUST cover: JavaScript/TypeScript/Node, Deno, Python, Go, PHP/Composer with Laravel/Symfony signals, Rust/Cargo, Nix, JVM Maven/Gradle, .NET/NuGet, Ruby/Bundler, Elixir/Mix, Dart/Flutter, C/C++ (CMake/Conan/Meson/vcpkg), SwiftPM, and Infrastructure-as-Code (Terraform/OpenTofu, Ansible, Docker, Helm/Kustomize). Detection MUST be evidence, not a repository-class or artifact verdict; a repository MUST NOT be collapsed into a single exclusive ecosystem label.

#### Scenario: Mixed repository

- GIVEN a repository contains independently evidenced Go, JavaScript, and Python ecosystems
- WHEN inspection runs
- THEN all three MUST appear with their own source paths
- AND no ecosystem MUST be selected as the semantic primary

#### Scenario: Laravel mixed repository

- GIVEN composer.json, composer.lock, artisan, and package.json with bun.lock
- WHEN inspection runs
- THEN PHP, Composer, and Laravel signals MUST appear
- AND the JS frontend manager MUST resolve from Bun evidence
- AND no single ecosystem MUST be chosen as primary

#### Scenario: Rust and Nix repository

- GIVEN Cargo.toml, Cargo.lock, flake.nix, and flake.lock
- WHEN inspection runs
- THEN Rust and Nix MUST both appear with their own source paths

#### Scenario: Infrastructure repository

- GIVEN a repository containing only `*.tf`, Dockerfile, and Chart.yaml files
- WHEN inspection runs
- THEN IaC facts MUST be emitted
- AND no code-provider signal MUST be fabricated
- AND the audit MUST complete without error

### Requirement: Bounded heavy-tree traversal

Known dependency and output trees — `node_modules`, `vendor`, `target`, `result`, `.next`, `.nuxt`, `dist`, `build`, `coverage`, `.venv`, `venv`, `__pycache__`, `storage/logs`, `bin`, `obj`, `_build`, `deps`, `.dart_tool`, `cmake-build-*`, and `out` — MUST NOT be recursively scanned. Their existence MUST remain as path-and-kind presence evidence, and descendant contents MUST NOT affect ordinary file totals or signals.

#### Scenario: Heavy tree is present

- GIVEN a recognized heavy directory contains arbitrary descendants
- WHEN inspection runs
- THEN the directory presence MUST be emitted
- AND changing only its descendants MUST NOT change traversed-file facts

#### Scenario: Ecosystem output trees are presence-only

- GIVEN a Rust `target/`, Elixir `_build/`, or Flutter `.dart_tool/` contains arbitrary descendants
- WHEN inspection runs
- THEN each MUST be emitted as path-and-kind presence
- AND descendant changes MUST NOT alter traversed-file facts

### Requirement: Confidence and unresolved manager conflicts

Package-manager facts MUST include evidence and confidence sufficient to distinguish confirmed, inferred, and ambiguous candidates across the V1 families: npm, pnpm, yarn, bun, deno, uv, pip, poetry, pdm, pipenv, composer, cargo, go, maven, gradle, dotnet/NuGet, bundler, mix, pub, conan, and terraform/tofu. Conflicting managers MUST be retained with reasons; project wrappers MUST be reported as stronger execution evidence than global binaries. A generic manifest such as `pyproject.toml` MUST NOT alone confirm a manager; `pdm.lock` MUST confirm pdm. Managers from different ecosystems MUST NOT be conflated into cross-ecosystem conflicts. The system MUST NOT choose a manager, rewrite files, emit migration steps, or route semantic verdicts.

#### Scenario: pnpm and Bun conflict

- GIVEN both `pnpm-lock.yaml` and `bun.lock` provide current evidence
- WHEN manager facts are emitted
- THEN both candidates and a conflict reason MUST appear
- AND no preferred manager or migration decision MUST appear

#### Scenario: pdm lockfile confirms its family

- GIVEN pdm.lock and pyproject.toml
- WHEN manager facts are emitted
- THEN pdm MUST be confirmed with pdm.lock evidence
- AND no uv or poetry implication MAY appear
- AND GIVEN pyproject.toml alone
- THEN pip-compatible MUST be inferred, not confirmed

#### Scenario: Families resolve independently

- GIVEN composer.lock and Cargo.lock in one repository
- WHEN manager facts are emitted
- THEN composer and cargo MUST both appear confirmed within their ecosystems
- AND no cross-ecosystem conflict reason MUST be invented

### Requirement: Suffix-based detection rules

The detection rule engine MUST match repository files by suffix or extension in addition to exact basenames, covering at least `*.tf`, `*.sln`, `*.slnx`, `*.csproj`, `*.fsproj`, `*.gemspec`, `*.xcodeproj`, `*.xcworkspace`, and variant files such as `phpunit.xml.dist`, `rust-toolchain.toml`, and `CMakeUserPresets.json`. Suffix rules MUST be additive: existing exact-basename rules and their emitted facts MUST remain unchanged. Suffix-matched signals MUST carry the full matched path and MUST follow the documented deterministic ordering.

#### Scenario: Terraform files

- GIVEN `*.tf` files exist at any depth
- WHEN inspection runs
- THEN terraform/IaC signals MUST be emitted with each matched path

#### Scenario: Suffix and basename coexist

- GIVEN CMakeLists.txt, CMakePresets.json, and `*.csproj` files
- WHEN inspection runs
- THEN all matched rules MUST fire with deterministic, byte-stable output

#### Scenario: Unknown suffix

- GIVEN files with an unrecognized suffix
- WHEN inspection runs
- THEN no ecosystem signal MUST be fabricated for them

### Requirement: Full lockfile coverage V1

Inspection MUST recognize the full V1 lockfile set as versioned-dependency evidence: package-lock.json, npm-shrinkwrap.json, pnpm-lock.yaml, yarn.lock, bun.lock, deno.lock, uv.lock, poetry.lock, Pipfile.lock, pdm.lock, composer.lock, Cargo.lock, flake.lock, go.sum, go.work.sum, Gemfile.lock, mix.lock, pubspec.lock, Package.resolved, packages.lock.json, conan.lock, .terraform.lock.hcl, and Chart.lock. Lockfile presence MUST be emitted as versioned-dependency evidence only; provider-recommendation gating is defined by the provider-lifecycle-truth spec.

#### Scenario: Rust plus Nix lockfiles

- GIVEN Cargo.lock and flake.lock
- WHEN inspection runs
- THEN both MUST be listed as lockfiles with their paths

#### Scenario: Container and IaC lockfiles

- GIVEN packages.lock.json, .terraform.lock.hcl, and Chart.lock
- WHEN inspection runs
- THEN all three MUST be recognized with their ecosystem association

### Requirement: Per-ecosystem output signals

Inspection MUST emit per-ecosystem output-directory presence signals beyond `dist`, `build`, and `coverage`: `target`, `result`, `.next`, `.nuxt`, `node_modules`, `.venv`, `venv`, `__pycache__`, `vendor`, `storage/logs`, `bin`, `obj`, `_build`, `deps`, `.dart_tool`, `build` (Flutter), `cmake-build-*`, and `out`. Output presence MUST be a candidate signal only: Go MUST NOT decide that a tool is needed because output exists; that verdict remains model-owned.

#### Scenario: Rust output signal

- GIVEN a Rust repository with `target/`
- WHEN inspection runs
- THEN the output signal MUST appear associated with the Rust ecosystem

#### Scenario: Node and Next outputs

- GIVEN `.next/` and `node_modules`
- WHEN inspection runs
- THEN both MUST appear as presence signals
- AND node_modules MUST NOT be traversed

#### Scenario: Elixir outputs

- GIVEN `_build/` and `deps/`
- WHEN inspection runs
- THEN both MUST be emitted as presence signals

### Requirement: Framework centrality with evidence

Framework facts MUST carry: name; version parsed deterministically from manifest evidence when present; evidence paths; and bounded `centrality_signals` (which files reference the framework). Version MUST NOT be inferred when absent. Go MUST NOT emit a centrality verdict — central, supporting, incidental, and unknown remain model-owned.

#### Scenario: Versioned framework

- GIVEN Cargo.toml declares ratatui 0.29.0 and UI files import it
- WHEN framework facts are emitted
- THEN name, version, evidence path, and import-referencing centrality signals MUST appear

#### Scenario: Version absent

- GIVEN a manifest references a framework without a version
- WHEN framework facts are emitted
- THEN version MUST be empty
- AND no version MAY be inferred
