# Ecosystem Facts Specification

## Purpose

Expose deterministic repository evidence without turning detection into migration advice.

## Requirements

### Requirement: Compatible inspect contract

`inspect --json` MUST retain `agent-ready.inspect/v1`, all existing fields, types, and meanings while ecosystem fields are additive. Additive fields MAY remain in V1; removing, renaming, changing a field type or meaning, or making a previously optional field required MUST use a new schema version. Repeated inspection of unchanged bytes MUST produce byte-stable JSON: object keys, evidence arrays, paths, ecosystems, managers, manifests, lockfiles, workspaces, wrappers, framework signals, and build/test-tool signals MUST use documented deterministic ordering.

#### Scenario: Legacy consumer reads enriched V1

- GIVEN a consumer reads only the original V1 fields
- WHEN it receives enriched inspect JSON
- THEN those fields MUST retain their prior types and meanings
- AND unknown additive fields MUST not be required for interpretation

#### Scenario: Input order varies

- GIVEN equivalent repository entries are discovered in different filesystem orders
- WHEN inspection runs
- THEN the serialized JSON MUST be byte-identical

### Requirement: Multi-ecosystem presence evidence

Inspection MUST retain every evidenced ecosystem concurrently and identify its manifests, lockfiles, workspace signals, project wrappers, framework signals, and build/test-tool signals. Detection MUST be evidence, not a repository-class or artifact verdict.

#### Scenario: Mixed repository

- GIVEN a repository contains independently evidenced Go, JavaScript, and Python ecosystems
- WHEN inspection runs
- THEN all three MUST appear with their own source paths
- AND no ecosystem MUST be selected as the semantic primary

### Requirement: Bounded heavy-tree traversal

Known dependency and output trees, including `node_modules`, `vendor`, `target`, `.venv`, `bin`, and `obj`, MUST NOT be recursively scanned. Their existence MUST remain as path-and-kind presence evidence, and descendant contents MUST NOT affect ordinary file totals or signals.

#### Scenario: Heavy tree is present

- GIVEN a recognized heavy directory contains arbitrary descendants
- WHEN inspection runs
- THEN the directory presence MUST be emitted
- AND changing only its descendants MUST NOT change traversed-file facts

### Requirement: Confidence and unresolved manager conflicts

Package-manager facts MUST include evidence and confidence sufficient to distinguish confirmed, inferred, and ambiguous candidates. Conflicting managers MUST be retained with reasons; project wrappers MUST be reported as stronger execution evidence than global binaries. A generic manifest such as `pyproject.toml` MUST NOT alone confirm a manager. The system MUST NOT choose a manager, rewrite files, emit migration steps, or route semantic verdicts.

#### Scenario: pnpm and Bun conflict

- GIVEN both `pnpm-lock.yaml` and `bun.lock` provide current evidence
- WHEN manager facts are emitted
- THEN both candidates and a conflict reason MUST appear
- AND no preferred manager or migration decision MUST appear
