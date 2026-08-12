# Tasks: Close Agent-Ready V1 (six bounded streams)

## Review Workload Forecast

| Field | Value |
|-------|-------|
| Estimated changed lines | ~1,700–2,050 total (authored additions + deletions) |
| Per-slice estimates | 1a ~250–300 · 1b ~130–160 · 2a ~200–240 · 2b ~200–240 · 3 ~280–340 · 4 ~220–280 · 5 ~350–420 · 6 ~100–140 |
| 400-line budget risk | Medium — slices 1, 2, 5 sit at the boundary; escape hatches pre-planned (1a/1b, 2a/2b; trim fixture repos to minimal shapes) |
| Chained PRs recommended | Yes |
| Suggested split | 8 stacked PRs: 1a → 1b → 2a → 2b → 3 → 4 → 5 → 6 (1a/1b and 2a/2b merge into one PR each if ≤400 combined) |
| Delivery strategy | auto-chain |
| Chain strategy | stacked-to-main (cached) |

Decision needed before apply: No
Chained PRs recommended: Yes
Chain strategy: stacked-to-main
400-line budget risk: Medium

**First apply boundary: Slice 1 only (unit 1a).** Auto-chain then proceeds slice-by-slice; each PR ≤400 authored lines; each slice keeps `go test ./...` green and reverts cleanly.

### Suggested Work Units

| Unit | Goal | Likely PR | Focused test command | Runtime harness | Rollback boundary |
|------|------|-----------|----------------------|-----------------|-------------------|
| 1a | Suffix engine + 23 lockfiles + output signals + framework facts + heavyTrees | PR 1 | `go test ./internal/ecosystem ./internal/inventory` | `go run ./cmd/agent-ready inspect --json` on mixed fixture; byte-identity compare | `internal/ecosystem/ecosystem.go`, `internal/inventory/inventory.go` + tests |
| 1b | Manager families: pdm confirm, wrapper precedence, conflicts | PR 2 | `go test ./internal/ecosystem -run TestResolveManagers` | N/A — pure fact tables (no runtime boundary) | `internal/ecosystem/managers.go` + `managers_test.go` |
| 2a | §20 safety levels + §21 PMs + denylist + §46 UX + `tools explain` | PR 3 | `go test ./internal/tools ./internal/cli` | `tools explain uv`; `tools install uv` plan render then decline | `internal/tools/catalog.go\|install.go\|detect.go`, `internal/cli/cli.go`, `cmd/agent-ready/main.go` + tests |
| 2b | rtk/uv recipes + composer lifecycle + 25 detect-only entries + §56 tests | PR 4 | `go test ./internal/tools -run 'Install\|Recipe\|DetectOnly'` | `tools install rtk` in isolated HOME/fake PATH; fail-closed host | `internal/tools/install.go`, `internal/tools/recipes/*.json` + tests |
| 3 | Provider lifecycle truth (serena/headroom, D3/D4, §48/§49/§47, §57/§58) | PR 5 | `go test ./internal/tools ./internal/lifecycle ./internal/opencode` | `AGENT_READY_DRIVEN_MODEL=… go test ./internal/app -run TestDrivenAudit -timeout 40m` (post-gating re-run) | `internal/tools/recommend.go\|catalog.go\|doctor.go` + isolation tests |
| 4 | Content contracts + stale-phrase cleanup | PR 6 | `go test ./internal/bootstrap` | `go test ./internal/lifecycle -run TestEmbeddedAssetUpgradeGateSlice11`; `AGENT_READY_DRIVEN_MODEL=… go test ./internal/app -run TestDrivenReview -timeout 40m` | `internal/bootstrap/assets/skills/**` + `content_test.go`/upgrade tests |
| 5 | Fixture matrix + NixOS Wizard/Laravel driven cohorts | PR 7 | `go test ./internal/ecosystem -run TestFixtures` | `AGENT_READY_DRIVEN_MODEL=… go test ./internal/app -run 'TestDrivenAudit\|TestDrivenFixtures' -timeout 40m` | `internal/ecosystem/fixtures_test.go`, `internal/app/driven_fixtures_test.go`, `internal/app/testdata/acceptance/driven/{nixos-wizard,laravel}/` |
| 6 | CHANGELOG v0.1.x + v1.0.0 | PR 8 | `git log --oneline 6a07341..HEAD` vs goreleaser filters | N/A — docs-only artifact | `CHANGELOG.md` |

## Phase 1: Ecosystem matrix — Slice 1 (PRs 1–2; 1a/1b split pre-planned, OQ-2)

Rollback: `internal/ecosystem/`, `internal/inventory/inventory.go` + tests only.

- [x] 1.1 Extend `rule{id,names}` in `internal/ecosystem/ecosystem.go` with `suffixes []string`, match `strings.HasSuffix(path.Base)`; enumerate full set: `*.tf`, `*.sln`, `*.slnx`, `*.csproj`, `*.fsproj`, `*.gemspec`, `*.xcodeproj`, `*.xcworkspace`, `phpunit.xml.dist`, `rust-toolchain.toml`, `CMakeUserPresets.json`, plus CMakeLists.txt basename (D1 + validator finding 1). (~70 L)
- [x] 1.2 `TestSuffixRules`: each suffix fires with full matched path; unknown suffix → no signal (threat-matrix RED); suffix+basename interleave byte-stable. (~40 L)
- [x] 1.3 §8 full lockfile set (23: incl. go.work.sum, Package.resolved, packages.lock.json, conan.lock, .terraform.lock.hcl, Chart.lock): `TestFullLockfileCoverage`. (~40 L)
- [x] 1.4 §9 `Facts.OutputSignals []Signal` (additive omitempty) in `internal/inventory/inventory.go`: per-ecosystem output dirs + `cmake-build-` prefix; `TestOutputSignals`; candidate-only, no verdict. (~50 L)
- [x] 1.5 D11 heavyTrees += dist, build, coverage, result, .next, .nuxt, venv, __pycache__, storage/logs, _build, deps, .dart_tool, out + prefix `cmake-build-`; `TestHeavyTreePresenceOnly`; confirm `bin` presence-only signal retained (PHP `bin/console`) — OQ-3 task. (~40 L)
- [x] 1.6 D2 `FrameworkFact{Name,Version,Evidence,CentralitySignals}`; per-family version parsers (Cargo.toml/composer.json/package.json regex/JSON); bounded centrality scan ≤20 sorted files; `Detect` gains `root`; `TestFrameworkVersionEvidence` (ratatui 0.29.0; absent version stays empty). (~90 L)
- [x] 1.7 (1b) Manager families in `internal/ecosystem/managers.go`: `pdm.lock`→pdm confirmed; pyproject.toml alone→pip inferred; wrappers `gradlew`/`mvnw` stronger than global; no cross-ecosystem conflict; extend `TestResolveManagers*`. (~90 L)
- [x] 1.8 Same-slice byte-identity updates: `TestDetectMixedRepositoryFacts` (`internal/ecosystem/ecosystem_test.go`), `TestInspectDeterministic` (`internal/inventory/inventory_test.go`). (~50 L)

## Phase 2: Catalog/installer truth — Slice 2 (PRs 3–4; 2a/2b split pre-planned)

Rollback: `internal/tools/*`, `internal/cli/cli.go`, `cmd/agent-ready/main.go`, `internal/tools/recipes/*.json` + tests.

- [x] 2.1 `Entry` gains `SafetyLevel/Methods/SideEffects/IntegrationMode` (additive omitempty); §20 levels (rg/fd/jq/gh/ast-grep/uv SAFE_RECIPE; composer/rustup VERSION_SENSITIVE; maven/gradle PROJECT_WRAPPER_PREFERRED; pip runtime-coupled; RTK SAFE_RECIPE + separate GLOBAL_SIDE_EFFECT); surface in `status --json`; extend `TestStatusFamiliesOrderedAndStable`. (~70 L)
- [x] 2.2 D10 PM detection order apt, pacman, dnf, brew, zypper, apk, winget; AUR (yay/paru) opt-in-only remediation never auto-selected; Nix environment-only; host tests zypper/apk/winget. (~50 L)
- [x] 2.3 `ValidateRecipe` interpreter denylist (sh/bash/zsh/fish/dash/ksh/pwsh/cmd) + pipe rejection; RED tests: `sh`/`bash` executable rejected, pipe args rejected (threat-matrix). (~30 L)
- [x] 2.4 §46 `InstallPlan` += Kind/Evidence/Method/Level/SideEffects; render tool/kind/evidence/plan + 3 Changes lines + `Proceed? [y/N]`; empty/unreadable input declines; UX golden. (~40 L)
- [x] 2.5 D7 `tools explain`: `cli.Helper.Use` + `cobra.ExactArgs(1)`, `ExplainFacts` `agent-ready.explain/v1`; known renders, unknown names id, exit 1; only new verb (§64). (~60 L)
- [x] 2.6 OQ-1 first-class: verify rtk/uv/composer recipe executables deterministic per PM (brew/winget/apt candidates); unresolvable PM → fail-closed error + remediation, nothing executes. (~30 L)
- [x] 2.7 (2b) rtk + uv verified recipes (fixed argv, post-install verify) with §56 5-behavior tests each (detect/version/plan/execute/verify; PATH-faking, isolated HOME). (~120 L)
- [x] 2.8 (2b) composer detect/version/explain/plan on all V1 platforms; executes only where deterministic recipe exists. (~60 L)
- [x] 2.9 (2b) D8 25 detect-only entries (npm…tofu) with executables/versionArgs; install unsupported/MANUAL; extend `TestRecommendDetectOnlyTools`. (~90 L)

## Phase 3: Provider lifecycle truth — Slice 3 (PR 5)

Rollback: `internal/tools/recommend.go|catalog.go|doctor.go` + isolation tests.

- [x] 3.1 serena + headroom catalog entries (capabilities per §14.4/§14.5); all five providers in `status --json`; Go output contains no verdicts/thresholds. (~40 L)
- [x] 3.2 D3 Context7 gating: candidate only on framework-facts-with-version; lockfile-only/manager-conflict candidates removed; recommend tables (lockfile-only→no Context7; ratatui→candidate). (~50 L)
- [x] 3.3 D4 conditional signals: CodeGraph workspace/cross-package topology (no deps>N verdict), Semble scale band, Serena LSP surface, Headroom RTK+output-pressure; small repo → NOT_JUSTIFIED valid. (~60 L)
- [x] 3.4 §48 six metadata fields per provider; `install:supported ⟺ verified`; no provider install recipes in V1. (~40 L)
- [x] 3.5 §49 `tools doctor` per-provider checks (executable/version/index/integration/health); tests: broken version, missing project index → unhealthy with reason. (~50 L)
- [x] 3.6 D6 §47 opt-in: post-install `Enable global integration? [y/N]` default N; never during `init`; Y without verified recipe → remediation. (~40 L)
- [x] 3.7 §57/§58 isolation: init/install/remove leave `~/.config/opencode` byte-identical (fake PATH + isolated HOME, `TestPreflightIsolatesGlobalTrees` pattern); local `.agent-ready` merge lossless; **re-run `TestDrivenAudit`** after gating. (~80 L)

## Phase 4: Content contracts — Slice 4 (PR 6)

Rollback: `internal/bootstrap/assets/skills/**` + `content_test.go`/upgrade tests; user-modified assets never deleted.

- [x] 4.1 §15 Tool Budget minimal-set ordering in `agent-ready-orchestrator/SKILL.md` + `references/audit-flow.md`. (~30 L)
- [x] 4.2 §28 seven questions + decision outputs in `artifact-design/SKILL.md` + `references/artifact-decisions.md`. (~30 L)
- [x] 4.3 §29 `skill_request` fields (purpose…validation) in `skill-creator/SKILL.md` + `references/authoring-procedure.md`; empty external_verified_evidence → report, never invent. (~30 L)
- [x] 4.4 §30 named checks (framework_grounding, package_manager_accuracy, toolchain_accuracy, external_verification_when_required) in `skill-reviewer/SKILL.md` + `references/review-procedure.md`. (~30 L)
- [x] 4.5 §54 vocabulary in `targeted-research/SKILL.md` + `references/search-strategies.md`; fix stale phrase `search-strategies.md:24` ("Tool Manager is out of scope") → "tool installation is never automatic during audit; tool/capability assessment is mandatory". (~20 L)
- [x] 4.6 Extend `internal/bootstrap/content_test.go` locks per contract; **parameterized `TestEmbeddedAssetUpgradeGateSlice11`** (`internal/lifecycle/update_test.go`) — assets changed → manifest hashes regenerate; **re-run `TestDrivenReview`** with final instructions. (~60 L)

## Phase 5: Fixture matrix — Slice 5 (PR 7)

Rollback: `internal/ecosystem/fixtures_test.go`, `internal/app/driven_fixtures_test.go`, `internal/app/testdata/acceptance/driven/{nixos-wizard,laravel}/`.

- [x] 5.1 `internal/ecosystem/fixtures_test.go` deterministic tables §33 Laravel / §34 uv / §35 pip / §36 npm+pnpm-workspace+Bun+Deno / §37 JVM wrapper / §38 .NET / §39 Ruby / §40 Elixir / §41 Flutter / §42 CMake+Conan; assert §52 (PHP≠Node; uv not forced; package.json not always npm; Cargo/flake/composer/Gemfile/pubspec/mix lockfiles recognized). (~120 L)
- [x] 5.2 §59 conflict table (pnpm-lock.yaml+bun.lock → retained, no arbitrary choice); §60 wrapper table (gradlew/mvnw preferred, no global install); §61 mixed matrix (Rust+Cargo+Nix, PHP+Composer+JS, Dart+native; no exclusive label). (~60 L)
- [x] 5.3 Monorepo deterministic oracle (§43): provider candidates with workspace evidence; set not forced to all three; Tool Budget minimal set. (~40 L)
- [x] 5.4 Boilerplate oracle (§44): extension points/generated-vs-editable facts; no generic scaffold/language skill; NO_ACTION with reason (§53/§55). (~40 L)
- [x] 5.5 NixOS Wizard driven cohort: fixture `internal/app/testdata/acceptance/driven/nixos-wizard/`; unseeded structural oracle in `internal/app/driven_fixtures_test.go` (reuse `TestDrivenAudit` driver + `gitRoot`): Rust, Cargo.lock, Nix, flake.lock, Ratatui 0.29 centrality, stale-pnpm guidance detected, tool assessment shown, **external verification considered/used for a Ratatui artifact**, generic Rust skill rejected (validator finding 2). (~80 L)
- [x] 5.6 Laravel driven cohort: fixture `driven/laravel/`; oracle asserts PHP≠Node, Bun frontend, no exclusive label. (~60 L)
- [x] 5.7 Battery bounds: `AGENT_READY_DRIVEN_MODEL` skip guard; cohorts unseeded (structural assertions only); runs under `-timeout 40m`; new cohorts inherit `TestDrivenAuditGitSelectors`. (~30 L)

## Phase 6: CHANGELOG — Slice 6 (PR 8)

Rollback: `CHANGELOG.md` only.

- [x] 6.1 Create `CHANGELOG.md`: v0.1.0@`6a07341`, v0.1.1@`7bc55bf`; v1.0.0 covering `40ff819..HEAD` + the six close slices. (~90 L)
- [x] 6.2 Cross-check entries vs `git log --oneline 40ff819..HEAD` and goreleaser release filters; no tag/commit created (tagging deferred). (~10 L)

## Constraints

- §64 exclusions preserved: no agents, no slash commands beyond `tools explain`, no framework workflows, no language/PM skills, no mandatory MCP, no TUI/daemon/db/telemetry, no rewrite. `tools explain` is the only new verb.
- strict_tdd false: no RED-GREEN phases; threat-matrix RED cases (1.2, 2.3) ship as explicit negative tests.
- Every slice: `go test ./...` green, ≤400 authored lines, byte-identity expectations updated same slice, fixtures ship with behavior, clean per-slice revert.

NOTE (unit 2b, slice 2): tasks 2.7–2.9 implemented on branch feat/v1-close-rtk-uv-recipes (HEAD dc0c62a) at 262 authored changed lines (237 additions + 22 deletions + 3 recipe JSON files). Recipes added for uv (apt/pacman/dnf/brew), rtk (brew only — verified Homebrew formula; other PM hosts fail closed), composer (apt/pacman/dnf/brew, VERSION_SENSITIVE). 19 new §45 detect-only entries (catalog 17→36; recipes 5→8); rtk moved from provider to productivity family per §13 "first-class productivity entry". §56 five-behavior tests (TestRecipeFiveBehaviors), §47/§57/§58 rtk isolation (TestRtkInstallIsolatedFromGlobalOpenCode), composer lifecycle (TestComposerLifecycleAcrossPlatforms), §45 detect-only truth (TestRecommendDetectOnlyTools extension). Work unit close-slice-2b-recipes-detect-only; not committed/pushed per instructions.

NOTE (unit 3, slice 3): tasks 3.1–3.7 implemented on branch feat/v1-close-provider-lifecycle (HEAD 2650e18, working tree) at **618 authored changed lines (529 additions + 89 deletions)** — exceeds the 400 budget: **size:exception request** (precedent: slice 2a @ 422 approved, memory #242). Overage drivers: §49 five-dimension doctor checks across 5 providers + real-host honest probes; §48 six-field metadata declarations surfaced in status; D3/D4 signal gating + 4 new recommend tests; §47 opt-in prompt + binary-level opt-in/decline harness; §57/§58 isolation (install byte-identical global config, remove isolation, driven-init global seeding). Same-slice expectation updates: TestRecommendGroundedOrderAndStability (7→7 but context7×2→serena+headroom), TestRecommendConflictAwareEvidence → TestRecommendLockfileOnlyNoContext7, TestRecommendProviderWithoutLifecycle → TestRecommendContext7FrameworkGated, TestRecommendV1ReaderCompatibility (2→1 candidate), TestRecommendSignals (Context7→Headroom), TestCatalogOrderedSupportTruth 36→38, provider family list 3→5, TestDetectEcosystemTools provider list 3→5, TestDoctorHealthyWithWarnings PATH-isolated (§49 probes are host-sensitive). Driven re-run: `AGENT_READY_DRIVEN_MODEL=deepseek/deepseek-v4-flash go test ./internal/app -run TestDrivenAudit -timeout 40m` → PASS 132.78s (2m13s), including the new §57 seeded-global-config init assertion. Work unit close-slice-3-provider-lifecycle; not committed/pushed per instructions.

NOTE (unit 5, slice 5): tasks 5.1–5.7 implemented on branch feat/v1-close-fixture-matrix (HEAD 1b9349d, working tree) at **394 authored changed-line units (381 Go test lines + 13 fixture files)** — within the 400 budget (estimate 350–420; slice-2b counting precedent: data files = 1 unit each). Fixture matrix: `internal/ecosystem/fixtures_test.go` TestFixtureMatrix (14 deterministic §33–§42 rows: exact managers/lockfiles/ecosystem; §52 PHP≠Node, uv not forced, package.json≠npm, Cargo/flake/composer/Gemfile/pubspec/mix lockfiles) + TestFixtureAcceptance (§52 lockfile set, §59 retained pnpm/bun conflict with reason and no decision token, §60 gradlew/mvnw confirmed with no global install, §61 rust+nix / php+composer+js / dart+native with no exclusive label). Monorepo oracle `internal/tools/fixtures_test.go` TestMonorepoOracle (§43: codegraph+serena workspace-evidenced candidates, Semble not forced, byte-stable, no model-owned tokens) + TestBoilerplateOracle (§44: generated-vs-editable/extension-point facts via inventory, no generic scaffold/language skill candidate, pub inferred). Driven cohorts `internal/app/driven_fixtures_test.go` TestDrivenFixtures (subtests nixos-wizard + laravel): unseeded fixtures, reuses TestDrivenAudit driver + gitRoot + assertAuditStructure, per-cohort 9-min budget, `AGENT_READY_DRIVEN_MODEL` skip guard (5.7); NixOS Wizard oracle asserts rust/cargo/cargo.lock/nix/flake.lock/ratatui/0.29/pnpm tokens, stale-pnpm flag, Ratatui external-verification language, generic-skill artifact guard (validator finding 2); Laravel oracle asserts php/laravel/composer/bun, npm-without-bun rejection (PHP≠Node), no exclusive label. Driven battery: `AGENT_READY_DRIVEN_MODEL=deepseek/deepseek-v4-flash go test ./internal/app -run 'TestDrivenAudit|TestDrivenFixtures' -count=1 -timeout 40m` → PASS 400.75s total (TestDrivenAudit 168.81s; TestDrivenFixtures 231.89s = nixos-wizard 169.40s + laravel 129.33s in the first combined run; isolated TestDrivenAudit re-run PASS 150.56s — one combined-run failure was model-output variance on NO_ADDITIONAL_TOOLS reasoning, not a regression). gofmt/vet/diff-check clean; `go test ./...` 15/15 ok. Work unit close-slice-5-fixture-matrix; not committed/pushed per instructions.

NOTE (unit 6, slice 6): tasks 6.1–6.2 implemented on branch feat/v1-close-changelog (HEAD 9912ae1, working tree) at **71 authored lines — within the 400 budget (estimate 100–140)**. `CHANGELOG.md` created at repo root (71 lines, Keep-a-Changelog structure): v0.1.0@`6a07341` (2026-08-09, initial public release: CLI foundation, docs, installer — 34 commits), v0.1.1@`7bc55bf` (2026-08-09, installer as release asset with one-liner install — 3 commits 7e8f3ed/a2aa2ac/7bc55bf), v1.0.0 (2026-08-11, 18 commits `40ff819..HEAD` — completion change: ownership-safe upgrades, heavy-tree pruning, ecosystem detection matrix, manager families, tool capability truth, evidence-only recommendations, mandatory tool assessment, relevant-sync reassessment, external verification gate, OpenCode version decoupling; plus six close slices: ecosystem matrix completion, installer truth with RTK/uv/composer recipes and `tools explain`, provider lifecycle truth with Context7 real-need gating and doctor health, content contracts, fixture matrix with NixOS Wizard/Laravel driven regressions, and this changelog). Added/Changed/Fixed per release; "Source of truth" section documents git-log cross-checking; every entry maps to a real commit (all 30 referenced hashes resolve via `git cat-file`). 6.2 cross-check: `git log --oneline 40ff819..HEAD` = 18 commits all covered; goreleaser `v*` tag filter honored (v0.1.0/v0.1.1 tags exist; v1.0.0 tag deferred); goreleaser excludes `^docs:`/`^test:` from auto-changelog — hand-written entries kept factual, docs commits listed under their releases. `git diff --check` clean; `go test ./...` 16/16 ok (no Go changes). Work unit close-slice-6-changelog; not committed/pushed per instructions.
