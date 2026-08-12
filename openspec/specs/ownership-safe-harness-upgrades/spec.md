# Ownership-Safe Harness Upgrades Specification

## Purpose

Define cross-version embedded-asset reconciliation without weakening ownership or user control.

## Requirements

### Requirement: Ownership-aware reconciliation

An upgrade MUST compare installed ownership evidence, installed bytes, and target embedded assets. It MUST update an owned asset only when its installed bytes still match the previously owned version. It MUST preserve and report user-modified or unmanaged collisions rather than overwrite or delete them. A newly embedded asset MUST be installed when its destination is absent and recorded as owned; an obsolete owned asset MAY be removed only when unchanged from its previously owned version.

#### Scenario: Unmodified asset advances

- GIVEN an older installation owns an asset whose bytes remain unchanged
- WHEN a newer harness version is applied
- THEN the asset MUST receive the newer embedded bytes
- AND ownership evidence MUST describe the installed version

#### Scenario: Modified asset is preserved

- GIVEN an owned asset differs from its previously recorded bytes
- WHEN an update changes or retires that asset
- THEN the installed bytes MUST remain unchanged
- AND the plan MUST report the ownership conflict without claiming completion for that asset

#### Scenario: New asset collides with user content

- GIVEN a newly embedded asset has an existing unmanaged destination
- WHEN the update is planned or applied
- THEN the existing bytes MUST remain unchanged
- AND the collision MUST be observable rather than silently adopted

### Requirement: Protected repository state

Reconciliation MUST NOT modify model state, checkpoints, generated artifacts, or files outside the harness ownership boundary. Planning MUST be deterministic, side-effect free, and path ordered; applying an unchanged plan MUST be idempotent.

#### Scenario: Cross-version update preserves protected data

- GIVEN a repository contains model state, checkpoints, generated artifacts, and modified owned assets
- WHEN an update plan is applied twice
- THEN protected and modified bytes MUST match their pre-update bytes
- AND the second plan MUST contain no pending writes for already reconciled assets

### Requirement: Upgrade compatibility fixtures

Every embedded-asset change MUST include behavior-driving fixtures for an older installed manifest, unchanged and modified owned files, new assets, collisions, and protected state. Embedded skill changes MUST NOT ship until these fixtures pass.

#### Scenario: Asset slice lacks upgrade evidence

- GIVEN a slice adds or changes an embedded harness asset
- WHEN its verification evidence lacks a cross-version fixture
- THEN the slice MUST be rejected
