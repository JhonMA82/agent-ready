# Proposal: Close Agent-Ready V1 (spec §65 DoD + six bounded streams)

## Intent

Close Agent-Ready V1 per `OPENCODE_AGENT_READY_V1_COMPLETION_2026.md`: §65 DoD **11 delivered / 10 partial / 14 missing** (matrix: `exploration.md` §1). Ship the six exploration streams as chained slices on the delivered architecture; no rewrite; §64 exclusions honored.

## Scope

### In Scope
- **Ecosystem matrix** — §7.x families (PHP/Composer/Laravel/Symfony, Rust/Cargo, Nix/flake, .NET/NuGet, Ruby/Bundler, Elixir/Mix, Dart/Flutter/pub, C/C++ CMake/Conan, SwiftPM, IaC), §8 lockfiles, §9 output/build signals, §18 framework-centrality with evidence; suffix rule-engine; manager families. Evidence-only, deterministic, no verdicts.
- **Catalog/installer truth** — §20 safety levels (SAFE_RECIPE/VERSION_SENSITIVE/PROJECT_WRAPPER_PREFERRED/MANUAL/GLOBAL_SIDE_EFFECT), §21 PMs (apt/dnf/pacman/zypper/apk/brew/winget; AUR opt-in only; no curl|sh), §45 full install rg/fd/jq/gh/ast-grep/RTK/uv + composer detect/version/explain/plan, §46 UX, §56 honesty tests; new `tools explain`.
- **Provider lifecycle truth** — §14 policies (Context7/Semble/Serena/CodeGraph/Headroom; serena+headroom entries added; Context7 real-need only, never lockfile-only), §48 metadata-first declarations, §49 doctor health, §47 RTK global opt-in (default N, never during init) + §57/§58 isolation tests.
- **Content contracts** — §15 Tool Budget rules, §28 7 questions, §29 skill_request fields, §30 named reviewer checks (framework_grounding, package_manager_accuracy, toolchain_accuracy, external_verification_when_required), §54 vocabulary; stale-phrase cleanup.
- **Fixture matrix** — §32–§44 deterministic table/cohort fixtures + driven regressions (NixOS Wizard, Laravel, monorepo, boilerplate); covers §52/§53/§55/§59/§60/§61; fixtures ship with each slice.
- **CHANGELOG.md** — v0.1.x history from git log + v1.0.0 entry; tagging deferred.

### Out of Scope (§64)
No agents; no slash commands beyond `tools explain`; no framework workflows; no language/PM skills; no mandatory MCP; no TUI/daemon/database/telemetry; no rewrite. §57/§58 isolation and delivered behavior preserved.

## Capabilities

### New Capabilities
- `provider-lifecycle-truth`: §14 policy/catalog, §48 metadata, §49 health, §47 RTK opt-in.
- `fixture-matrix`: §32–§44 deterministic tables + bounded driven regressions.

### Modified Capabilities
- `ecosystem-facts`: §7.x matrix, §8 lockfiles, §18 centrality, manager families; decision-free invariant preserved.
- `tool-capability-facts`: §9 signals, §20 levels, §21 PMs, §45 recipes/detect entries, §13 RTK first-class, §46 UX, §56 tests; deferred-installer clause replaced.
- `audit-evidence-gates`: §15 Tool Budget, §28 questions, §29 skill_request, §30 named checks, §54 vocabulary.

## Approach

Six chained slices: ecosystem → installer → providers → content → fixtures → CHANGELOG (content precedes driven fixtures so model behavior is proven against final instructions). Auto-chain, stacked-to-main (cached); ≤400 authored lines/slice, same-slice tests, clean rollback (precedent `40ff819..d2fbbbf`). Driven battery bounded (1–2 new cohorts), `-timeout 40m`. Facts decision-free; installs/providers consent-gated; Context7 slice re-runs the audit cohort.

## Affected Areas

| Area | Impact | Description |
|---|---|---|
| `internal/ecosystem/` | Modified | §7.x rules, suffix engine, §18, managers |
| `internal/tools/` | Modified | catalog, recipes, PMs, signals, doctor, explain |
| `internal/inventory/` | Modified | §23 additive facts |
| `internal/bootstrap/assets/skills/` | Modified | §15/§28/§29/§30/§54 + cleanup |
| `internal/{lifecycle,opencode}/` | Modified | §47 opt-in + isolation tests |
| `internal/app/testdata/acceptance/driven/` | New | fixture matrix + cohorts |
| `CHANGELOG.md` | New | v0.1.x + v1.0.0 |

## Risks

| Risk | Likelihood | Mitigation |
|---|---|---|
| Suffix engine breaks byte-identity tests | Med | Additive; same-slice expectation updates |
| RTK global integration touches `~/.config/opencode` | High | Opt-in default N, never in init, isolation tests |
| Context7 change alters driven outcomes | Med | Relevance-gated; re-run audit cohort |
| Provider lifecycle scope inflation | Med | Metadata-first honesty, bounded |
| Review overload | High | Auto-chain ≤400 lines/slice |
| Driven wall-clock | Med | Bound cohorts; `-timeout 40m` |

## Rollback Plan

Reverse-order slice revert; each removes only its assets/tests/fixtures. Ownership gate reruns per embedded-asset slice; never delete user-modified assets or model state.

## Dependencies

- Archived delivery (`40ff819..d2fbbbf`) stays green; each slice re-runs `go test ./...`.
- Delivery: auto-chain, stacked-to-main, 400-line review budget (cached).

## Success Criteria

- [ ] §65 DoD: all 14 missing closed; partials closed or declared with evidence.
- [ ] Fixture matrix + driven regressions green; §52–§61 acceptance proven.
- [ ] `go test ./...` green; driven battery within `-timeout 40m`.
- [ ] §57/§58 zero default global mutation proven by tests.
- [ ] CHANGELOG.md ships factual v0.1.x history + v1.0.0 entry.
