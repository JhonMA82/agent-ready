# Design: Close Agent-Ready V1 (six bounded streams)

## Technical Approach

Incremental close on the delivered architecture (verified at HEAD `d2fbbbf`, 16/16 pkg green). Six auto-chained slices, stacked-to-main, ≤400 authored lines/slice, tests+fixtures shipped per slice, `go test ./...` green per slice, clean per-slice rollback (precedent `40ff819..d2fbbbf`). Invariants preserved: Go emits facts only (model owns centrality/Tool Budget/final verdicts), all JSON schemas stay `*/v1` with additive optional fields only, installs/providers consent-gated, §64 exclusions honored (no agents/commands beyond `tools explain`). Content slice precedes driven fixtures so model behavior is proven against final instructions.

## Architecture Decisions

| # | Decision | Choice | Alternatives rejected | Rationale |
|---|---|---|---|---|
| D1 | Suffix engine | Extend `rule{id,names}` with `suffixes []string`; match `strings.HasSuffix(path.Base)`; shared (ID,Path) sort | Glob engine | Additive, byte-stable, covers `*.tf/*.sln/*.csproj/*.gemspec/*.xcodeproj`; no ordering complexity |
| D2 | Framework facts | New `FrameworkFact{Name,Version,Evidence,CentralitySignals}`; per-family deterministic version parsers (Cargo.toml/composer.json/package.json regex/JSON); bounded centrality import scan (≤20 sorted files); `Detect` gains `root` param | Basename-only signals | Spec §18 requires version+centrality evidence; version never inferred when absent |
| D3 | Context7 gating | Candidate fires **only** on framework-facts-with-version evidence; lockfile-only and manager-conflict candidates removed | Keep lockfile signal, let model filter | §8/§14.1 prohibit lockfile-alone; model owns "artifact teaches API"/"docs insufficient" |
| D4 | Provider signals | CodeGraph only on workspace/cross-package topology (drop `deps>40` threshold-as-verdict); Semble on documented scale band (counts = observed evidence, never verdict); Serena on LSP-language source surface + module boundaries; Headroom only when RTK evidence + output-pressure signals coexist | Threshold verdicts | §14.2–14.5; candidates evidence-only (§62) |
| D5 | Provider metadata-first | §48 six fields declared per provider (install method, project-init, integration mode, side effects, uninstall, health); implement only deterministic lifecycle (detect/version/doctor); **no provider install recipes in V1** | Full lifecycle subsystem | Bounds scope; `install:supported ⟺ verified` honesty invariant |
| D6 | RTK global opt-in | Separate post-install `Enable global integration? [y/N]` default N, never during init; on Y with no verified integration recipe → explicit remediation; harness never writes global config directly | Auto-configure on Y | §13/§47; §57/§58 zero-default-mutation tests |
| D7 | `tools explain` | Extend `cli.Helper` with optional `Use` (positional arg, `cobra.ExactArgs(1)`); unknown id → error naming id | Bespoke cobra command | Keeps helper exit contract (0/1); no execution |
| D8 | Detect-only entries | 25 ecosystem-toolchain entries (npm…tofu) with executables/versionArgs, install unsupported or MANUAL | Install claims | §45 capability truth without recipes |
| D9 | Fixture split | Driven: NixOS Wizard + Laravel only; monorepo + boilerplate = deterministic oracles | 4 driven cohorts | Spec fixture-matrix MAY clause; bounds battery to `-timeout 40m` |
| D10 | PM detection | Order apt, pacman, dnf, brew, zypper, apk, winget; AUR never auto-selected (opt-in remediation); Nix environment-only | Universal AUR/Nix installers | §21 |
| D11 | Heavy trees | `heavyTrees` += dist, build, coverage, result, .next, .nuxt, venv, __pycache__, storage/logs, _build, deps, .dart_tool, out + prefix match `cmake-build-`; presence-only | Traverse-and-prune | §9; same-slice byte-expectation updates |

## Slice Sequence (auto-chain, stacked-to-main, ≤400 lines/slice)

| Slice | Scope | Test strategy | Rollback boundary |
|---|---|---|---|
| 1 Ecosystem matrix | §7.x rules+suffix engine, §8 23 lockfiles, §9 output signals (inventory Presence), §18 framework facts, manager families (pdm.lock→pdm confirmed), heavyTrees | Table tests in `internal/ecosystem`/`inventory`; update byte-identity expectations same slice (`TestDetectMixedRepositoryFacts`, `TestInspectDeterministic`); new `TestSuffixRules`, `TestFrameworkVersionEvidence`, `TestOutputSignals`, `TestFullLockfileCoverage`, `TestHeavyTreePresenceOnly` | `internal/ecosystem/*.go`, `internal/inventory/inventory.go` + tests |
| 2 Catalog/installer | §20 safety levels (status+plan), §21 PMs + interpreter denylist, §45 recipes rtk/uv + composer detect/version/explain/plan + 25 detect-only entries, §46 UX render, `tools explain`, §56 tests | §56 5-behavior tests per new `install:true` entry (PATH-faking pattern); fail-closed host tests; UX golden; explain known/unknown; `TestStatusFamiliesOrderedAndStable` updated | `internal/tools/*`, `internal/cli/cli.go`, `cmd/agent-ready/main.go` + recipes JSON + tests |
| 3 Provider lifecycle | serena/headroom entries, D3 Context7 gating, D4 signals, §48 metadata, §49 doctor health, §47 opt-in, §57/§58 isolation | Recommend tables (lockfile-only→no Context7; ratatui→candidate; small repo→absent; monorepo→3 candidates); doctor broken-version/missing-index; opt-in default-N leaves `~/.config/opencode` byte-identical; init silent; **re-run driven audit cohort** (`TestDrivenAudit`) | `internal/tools/recommend.go|catalog.go|doctor.go`, isolation tests |
| 4 Content contracts | §15 Tool Budget (audit-flow.md+SKILL.md), §28 7 questions, §29 skill_request, §30 named checks, §54 vocabulary, stale-phrase cleanup (`search-strategies.md:24`) | Extend `content_test.go` locks per contract; re-run `TestEmbeddedAssetUpgradeGate` (assets changed → manifest hashes regenerate); **re-run driven review cohort** (`TestDrivenReview`) | `internal/bootstrap/assets/skills/**` + content/upgrade tests |
| 5 Fixture matrix | Deterministic tables §33–§42 + §52/§59/§60/§61 + monorepo/boilerplate oracles (§43/§44/§55); driven NixOS Wizard + Laravel cohorts (unseeded structural oracles, reuse `TestDrivenAudit` driver) | New `internal/ecosystem/fixtures_test.go` tables; `driven_fixtures_test.go` oracles (Rust/Cargo/Nix/flake/Ratatui 0.29/stale-pnpm/generic-skill-rejected; PHP≠Node, Bun frontend); battery `-timeout 40m` | `internal/app/testdata/acceptance/driven/{nixos-wizard,laravel}/`, new test files |
| 6 CHANGELOG | `CHANGELOG.md`: v0.1.0@`6a07341`, v0.1.1@`7bc55bf`, v1.0.0 covering `40ff819..d2fbbbf` | No Go tests; verify cross-checks git log vs goreleaser filters; tagging deferred | `CHANGELOG.md` |

## Data Flow

```
WalkDir (heavy-tree skip → Presence) → ecosystem.Detect(root, paths) → Facts
  (manifests/lockfiles/output_signals/framework_facts/managers/conflicts)
  → recommend: signal table → Candidate[] (evidence-only) → model verdict
install: Plan (fail-closed) → §46 render (kind/evidence/plan/3 Changes) → [y/N]
  → exec fixed argv → verify → rtk: [y/N] global opt-in (never during init)
```

## Interfaces / Contracts

```go
type rule struct { id string; names, suffixes []string }
type FrameworkFact struct {
  Name string `json:"name"`; Version string `json:"version,omitempty"`
  Evidence []string `json:"evidence"`; CentralitySignals []Signal `json:"centrality_signals"`
}
// Facts gains (additive, omitempty): OutputSignals []Signal; FrameworkFacts []FrameworkFact
type Entry struct { ID string; Family Family; Executables, VersionArgs []string
  Install map[string]RecipeOp; Capabilities Capabilities
  SafetyLevel SafetyLevel `json:"safety_level,omitempty"`; Methods []string
  SideEffects, IntegrationMode string } // surfaced in status --json, optional
type ExplainFacts struct { SchemaVersion string; ID string; Kind string
  Capabilities Capabilities; SafetyLevel SafetyLevel; Methods []string
  SideEffects, Integration string }          // agent-ready.explain/v1
// InstallPlan += Kind, Evidence, Method, Level, SideEffects
// ValidateRecipe += interpreter denylist (sh/bash/zsh/fish/dash/ksh/pwsh/cmd)
// cli.Helper += Use string (positional arg → Options.Tool)
```

## Testing Strategy

Per-slice: focused table tests (go-testing decision gates) + full `go test ./...`; driven cohorts (NixOS Wizard, Laravel, re-runs of audit/review cohorts) skip without `AGENT_READY_DRIVEN_MODEL`/opencode/auth and run in verify with `-timeout 40m`. Byte-stability: update expectations in the same slice as the additive rule. Isolation: fake-PATH + isolated HOME/XDG patterns (existing `TestPreflightIsolatesGlobalTrees` precedent).

## Threat Matrix

| Boundary | Cases | Applicability | Design response | Planned RED tests |
|---|---|---|---|---|
| Executable classification | suffix-matched files, fixture docs, recipe executables | **Applicable** — suffix engine + recipe validation classify files/executables | Suffix signals are data, never executed; recipes = fixed argv, no shell | Unknown suffix → no signal; `sh`/`bash` executable rejected; pipe args rejected |
| Git repository selection | `git -C`, relative/absolute roots | **Applicable** — new driven cohorts reuse `gitRoot` | Same gitRoot helper; temp repos | Reuse `TestDrivenAuditGitSelectors`; new cohorts inherit |
| Commit state | staged, empty index | **N/A** — no index/worktree manipulation in Go surface | — | — |
| Push state | tracking, refspec | **N/A** — no push automation | — | — |
| PR commands | `--head`, composed | **N/A** — PR tooling outside Go surface | — | — |

## Migration / Rollout

No data migration: all schemas remain `*/v1` (additive optional fields only). Embedded-asset changes (slice 4) regenerate ownership manifest hashes → same-slice upgrade-gate re-run; `agent-ready update` re-syncs owned assets; user-modified assets/model state never deleted. Delivery: auto-chain, stacked-to-main, 400-line budget.

## Failure Behavior / Rollback

Fail-closed: no PM → explicit error + remediation, nothing executes; consent never defaults yes (empty/unreadable declines); doctor unhealthy carries failing check + reason; unknown tool id named. Rollback: reverse-order slice revert; each slice removes only its assets/tests/fixtures (table above); never touches `~/.config/opencode` or `.agent-ready/state`.

## Open Questions

- [ ] RTK/uv/composer recipe executables must be verified deterministic per PM during implementation (brew/winget/apt candidates); unresolvable PMs fail closed.
- [ ] Slice 1 rule-table volume may approach 400 lines; if breached, split manager-family expansion into slice 1b (auto-chain tolerates).
- [ ] `bin` as heavy tree (PHP `bin/console`): presence-only signal retained — confirm no signal loss.
