# Proposal: Complete Agent-Ready V1 Ecosystem Intelligence

## Intent

Close verified ecosystem, tool-truth, and audit-contract gaps incrementally without rewriting V1 or batching all 67 corrective sections.

## Scope

### In Scope
- First, reconcile embedded assets across versions while preserving ownership and modifications.
- Emit stable ecosystem, manifest, lockfile, workspace, wrapper, framework-signal, and package-manager conflict facts; prune heavy trees but retain presence evidence.
- Evolve the single tool catalog to report ecosystem, productivity, and provider tools with independent detect/version/recommend/install/configure/integration/side-effect capabilities.
- Require categorized Tool / Capability Assessment for initial audits and relevant syncs, including reasons and `NO_ADDITIONAL_TOOLS`.
- Gate artifacts embedding central versioned framework knowledge, with reviewer checks and behavior-driving fixtures per slice.
- User-directed amendment (2026-08-10): decouple the harness from the exact OpenCode version pin; host runtime updates must never block init, audit, sync, or review. Minimum-compatible-version floor with recorded version facts and non-blocking drift warnings.

### Out of Scope
- Replacing the Go CLI, local harness, `/agent-ready`, seven skills, checkpoints, ownership, JSON helpers, safe recipes, or model decisions.
- New agents, commands, generic language/package-manager skills, TUI, daemon, database, required MCP, or Go verdict routing.
- Safe recipe/system-package-manager expansion: follow-up `agent-ready-v1-safe-installer-expansion`.
- Provider lifecycle and global OpenCode/RTK integration: follow-up `agent-ready-v1-provider-global-integrations`.

## Capabilities

### New Capabilities
- `ownership-safe-harness-upgrades`: Cross-version reconciliation preserving user changes.
- `ecosystem-facts`: Multi-ecosystem and manager-conflict evidence.
- `tool-capability-facts`: Categorized support and recommendation evidence.
- `audit-evidence-gates`: Mandatory assessment, verification, and grounded review.

### Modified Capabilities
None; no baseline OpenSpec capability specs exist.

## Approach

Auto-chain upgrade safety, ecosystem/schema facts, tool truth, audit contracts, then driven regressions. Every slice that changes embedded assets declares its exact asset delta and passes the same parameterized cross-version ownership gate in that slice. Go emits evidence and conflicts; the model retains centrality, artifact, Tool Budget, and recommendation verdicts.

## Affected Areas

| Area | Impact | Description |
|---|---|---|
| `internal/{bootstrap,lifecycle,inventory,ecosystem,tools}` | Modified/New | Upgrade safety and deterministic facts |
| `internal/bootstrap/assets/skills/` | Modified | Assessment and verification contracts |
| `internal/app/testdata/acceptance/driven/` | New | Unseeded behavior-driving audit, sync, and review cohorts nested under the existing acceptance fixture root; existing C–P fixture paths remain unchanged |

## Risks

| Risk | Likelihood | Mitigation |
|---|---|---|
| Schema or upgrade regression | High | Compatibility and cross-version fixtures first |
| False certainty or expensive scans | Medium | Evidence/conflict output, stable pruning, model verdicts |
| Review overload | High | Auto-chain under the 400-line budget |

## Rollback Plan

Revert slices in reverse order. Compatibility and ownership fixtures gate boundaries; never remove user-modified assets or model state.

## Dependencies

- Ownership-safe upgrade slice must pass before embedded skill changes ship; each later embedded-asset slice must rerun the same gate against its declared asset delta.

## Success Criteria

- [ ] Every embedded-asset slice carries same-slice cross-version evidence that advances unchanged assets, preserves modified assets and collisions, installs new assets, and protects model state.
- [ ] Fixtures distinguish npm/pnpm/Bun/Deno, uv/pip, mixed ecosystems, wrappers, and unresolved pnpm/Bun conflicts without choosing.
- [ ] Tool facts never claim install/configuration support without corresponding tested capability.
- [ ] Every successful audit exposes categorized assessment; relevant syncs reassess.
- [ ] Framework-dependent fixtures require versioned evidence or an explicit exemption; reviewer checks enforce it.
- [ ] C–P regressions and `go test ./...` remain green.
