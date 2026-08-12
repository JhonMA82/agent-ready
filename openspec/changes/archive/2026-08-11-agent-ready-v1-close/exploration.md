# Exploration: Agent-Ready V1 Close (spec §65 DoD verification + bounded remaining scope)

**Change**: `agent-ready-v1-close` | **Mode**: hybrid (OpenSpec + Engram topic `sdd/agent-ready-v1-close/explore`)
**Verified against**: HEAD `d2fbbbf` (master, 10 delivered slices `40ff819..d2fbbbf`), `go test ./...` re-run green 16/16 packages on 2026-08-10.
**Spec source**: `OPENCODE_AGENT_READY_V1_COMPLETION_2026.md` (2,990 lines, all read). Delivered subset: `openspec/changes/archive/2026-08-10-agent-ready-v1-completion-ecosystem/`.

## Executive summary

§65 has **11/35 checkboxes fully delivered, 10 partial, 14 missing**. The delivered subset covers tool-assessment mandate, JS/Python/Go manager facts, wrapper preference, External Verification Gate, honest capability truth, NO_ACTION/generic-skill discipline, global-isolation invariants, and a green suite. Everything missing clusters into six bounded work streams: ecosystem matrix completion, installer expansion, advanced-provider lifecycle truth, the fixture matrix, content contracts, and CHANGELOG. Exclusions (§64) are honored — no agents, slash commands, framework workflows, language/PM skills, MCP, TUI, daemon, database, telemetry; no rewrite. The remaining work is a natural continuation of the same architecture and is proposal-ready.

## 1. §65 Definition-of-Done compliance matrix (verified against real code)

Legend: ✅ delivered | ⚠️ partial | ❌ missing. Evidence cites current files at HEAD.

| # | §65 checkbox | Verdict | Evidence (code path) |
|---|---|---|---|
| 1 | Tool assessment en todo audit | ✅ | `internal/bootstrap/assets/skills/agent-ready-orchestrator/SKILL.md:17,30,33` + `references/audit-flow.md:36` mandate three-family assessment w/ reasons or `NO_ADDITIONAL_TOOLS`; driven proof `internal/app/driven_audit_test.go:TestDrivenAudit` (verify-report PASS). §51 acceptance covered. |
| 2 | Ecosystem Detector soporta matriz V1 | ⚠️ | `internal/ecosystem/ecosystem.go:42-91` covers only 3 manifest ecosystems (go, javascript, python), 5 frameworks (angular/django/nextjs), 5 build / 3 test tools. PHP, Rust, Nix, .NET, Ruby, Elixir, Dart, C/C++ (beyond CMakeLists), Swift, IaC absent. |
| 3 | Cargo.lock y flake.lock | ❌ | No Cargo.toml/Cargo.lock/flake.nix/flake.lock rules in `internal/ecosystem/{ecosystem.go,managers.go}`. |
| 4 | Composer/composer.lock | ❌ | No composer.json/composer.lock rules; `composer` absent from `internal/tools/catalog.go` support table (only 11 entries: ast-grep, codegraph, context7, fd, gh, go, jq, node, rg, rtk, semble). |
| 5 | uv/pip/Poetry/Pipenv/PDM distinguidos | ⚠️ | `managers.go:48-66`: uv.lock, poetry.lock, Pipfile.lock, requirements.txt confirmed/inferred; **pdm.lock missing** (pdm only as ambiguous pyproject candidate, `managers.go:80`). pyproject≠uv covered by `TestResolveManagersFamilyEvidencePrecedence`. |
| 6 | npm/pnpm/Yarn/Bun/Deno distinguidos | ✅ | `managers.go:48-66` + conflicts `managerConflicts`; `TestResolveManagersPnpmBunConflict` covers §59. |
| 7 | Maven/Gradle wrapper distinguidos | ✅ | `managers.go:41-44` tier-3 wrappers; `TestResolveManagersWrapperPrecedence` covers §60. Note: pom.xml/build.gradle are NOT manifests — only wrapper files. |
| 8 | .NET/NuGet detectado | ❌ | No *.sln/*.csproj/packages.lock.json/NuGet.Config rules. |
| 9 | Bundler detectado | ❌ | No Gemfile/Gemfile.lock/*.gemspec/.ruby-version rules. |
| 10 | Mix detectado | ❌ | No mix.exs/mix.lock rules. |
| 11 | Dart/Flutter pub detectado | ❌ | No pubspec.yaml/pubspec.lock rules. |
| 12 | CMake/Conan detectado | ⚠️ | CMakeLists.txt only (`ecosystem.go:80` build rule). No CMakePresets/CMakeUserPresets, conanfile.py/txt, conan.lock, vcpkg.json, meson. |
| 13 | SwiftPM detectado | ❌ | No Package.swift/Package.resolved rules. |
| 14 | Infra repos no rompen audit | ⚠️ | No IaC rules at all (*.tf, Dockerfile, compose.yaml, Chart.yaml, ansible.cfg). Detection is additive/empty-safe (`TestRecommendEmptyRepo` green), so infra repos don't break by construction — but nothing proves or reports them. |
| 15 | External Verification Gate | ✅ | `skills/targeted-research/{SKILL.md,references/search-strategies.md}`, `skills/skill-reviewer/{SKILL.md,references/review-procedure.md}`; driven proof `TestDrivenReview/ungrounded` PASS. |
| 16 | Context7 por necesidad real | ⚠️ | `internal/tools/recommend.go:75-77` fires Context7 on ANY lockfile — exactly the too-weak signal §8/§14.1 prohibit ("lockfile alone insufficient"). Candidate exists; signal needs framework/dependency relevance. |
| 17 | RTK tool de primera clase | ❌ | `catalog.go:105`: rtk = provider, detect/version/install **unsupported**; only a recommend candidate (`recommend.go:56-58`). No §13 detect/version/install/verify/OpenCode-integration-detection, no §47 UX. |
| 18 | ast-grep por búsqueda estructural | ⚠️ | ast-grep catalog entry + install recipe (brew only) + `TestInstallSupportBackedByRecipe`; but **no structural-search recommend signal** (§11 example: large Rust modules). Detect/version/install yes; evaluation-by-signal no. |
| 19 | Semble/Serena/CodeGraph/Headroom condicionales | ⚠️ | CodeGraph candidate on workspaces/deps>40 (`recommend.go:62-68`), Context7 on lockfiles; semble present all-unsupported (`catalog.go:106`); **serena and headroom absent from catalog**. No §14 policy content. |
| 20 | Tool Budget evita proliferación | ⚠️ | Referenced as model-owned outcome (`audit-flow.md:36`, `review-procedure.md` rejection contract) but **no explicit budget rules/minimal-set ordering** (§15 example absent from assets). |
| 21 | Install support honesto | ✅ | Seven-state capability truth (`catalog.go:52-69`); `install:supported ⟺ verified recipe` (`TestInstallSupportBackedByRecipe`, `TestPlanSelectionAndFailClosed`). §20 five safety levels (SAFE_RECIPE etc.) NOT modeled — additive gap, not dishonesty. |
| 22 | Recipes sin shell arbitrario | ✅ | `ValidateRecipe` (`catalog.go:185-203`) rejects shell metachars; fixed argv `exec.Command` (`install.go:89`); `TestValidateRecipeRejectsShellMeta`. |
| 23 | Integración global opt-in separado | ✅ | No global integration exists; `ConfirmConsent` default N (`install.go:116-125`); `internal/opencode/version.go` probe isolates HOME/XDG; `TestPreflightIsolatesGlobalTrees`. §47 UX itself pending (see stream 3). |
| 24 | NixOS Wizard regression | ❌ | No Rust/Nix/Ratatui fixture. Driven audit fixture is Go+JS only (`internal/app/testdata/acceptance/driven/audit/{go.mod,main.go,package.json}`). |
| 25 | Laravel fixture | ❌ | No PHP fixture anywhere. |
| 26 | Python uv fixture | ❌ | uv detection unit-tested (`TestResolveManagersConfidenceLevels`); no uv repo fixture/driven run. |
| 27 | Python pip fixture | ❌ | pip inference unit-tested; no pip fixture. |
| 28 | JS manager fixtures | ⚠️ | Unit-level manager resolution/conflicts only (`managers_test.go`); no npm/pnpm-workspace/Bun/Deno repo fixtures. |
| 29 | JVM/.NET/Ruby/Elixir/Flutter/C++ fixtures | ❌ | Only JVM wrapper unit test; none of the others exist. |
| 30 | Monorepo provider assessment | ⚠️ | Workspace signals (turbo/nx/lerna/pnpm-workspace/go.work) + CodeGraph candidate exist; no monorepo fixture, no driven assessment. |
| 31 | Boilerplate assessment | ❌ | No boilerplate detection (extension points, generated-vs-editable, variants) anywhere. |
| 32 | NO_ACTION primera clase | ✅ | `artifact-design/references/artifact-decisions.md` NO_ACTION row + no-spam rules; `TestRecommendNoSynthesizedDefault`; driven outputs include `NO_ADDITIONAL_TOOLS`. |
| 33 | Sin skills genéricas | ✅ | skill-creator "No generic advice" + anti-patterns + fixture-g generic-react rejection + driven review rejection contract. |
| 34 | Cero contaminación global | ✅ | Local-only lossless config merge (`internal/opencode/config.go`), §58 shape preserved, probe isolation, no global writes. |
| 35 | `go test ./...` + acceptance | ✅ | Re-run 2026-08-10: 16/16 packages ok, `internal/app` 9.5s (acceptance C–P + driven cohorts as documented in archive verify-report). |

**Totals: 11 ✅ / 10 ⚠️ / 14 ❌.**

Also verified (outside §65): §23 inspect schema delivered except `agent_assets` field (`internal/inventory/inventory.go` Facts embeds ecosystem.Facts; no agent_assets); §24 status categories delivered (families), but `required_by_repo`/`integration_status` per-entry fields absent (`detect.go:22-43`); §49 doctor covers executable/version/project-state/OpenCode-integration checks but **no provider health** (`internal/lifecycle/doctor.go`, `internal/tools/doctor.go`); residual stale phrase "Tool Manager is out of scope" still in `skills/targeted-research/references/search-strategies.md:24` (contradicts §26, delivered).

## 2. Bounded remaining scope (six work streams, verified code paths)

### Stream 1 — Ecosystem matrix completion (facts only, no verdicts)
- **§7.x**: extend `internal/ecosystem/ecosystem.go` rules: PHP (composer.json, composer.lock, artisan, phpunit.xml*, pint.json, symfony.lock, bin/console), Rust (Cargo.toml, Cargo.lock, rust-toolchain*), Nix (flake.nix, flake.lock, default.nix, shell.nix), .NET (*.sln/slnx/csproj/fsproj, Directory.Build.*, Directory.Packages.props, NuGet.Config, packages.lock.json, global.json), Ruby (Gemfile, Gemfile.lock, *.gemspec, .ruby-version), Elixir (mix.exs, mix.lock), Dart (pubspec.yaml, pubspec.lock, analysis_options.yaml), C/C++ (CMakePresets.json, CMakeUserPresets.json, conanfile.py/txt, conan.lock, vcpkg.json, meson.build), Swift (Package.swift, Package.resolved), IaC (*.tf, .terraform.lock.hcl, Dockerfile, compose.yaml, docker-compose.yml, Chart.yaml, Chart.lock, values.yaml, kustomization.yaml, ansible.cfg, requirements.yml/yaml), Go go.work.sum, Python pdm.lock.
- **Rule-engine**: current rules match basename only; suffix/extension matching needed for *.tf, *.sln, *.csproj, *.xcodeproj — small additive change to the `rule` struct, keep determinism and sorted output (existing `TestDetectMixedRepositoryFacts` byte assertions must stay green).
- **§9 output signals**: extend `hasOutputDirs` (`internal/tools/recommend.go:144-151`, currently dist/build/coverage only) per-ecosystem: target/, result, .next/, .nuxt/, node_modules presence, .venv/, venv/, __pycache__/, vendor/, storage/logs/, bin/, obj/, _build/, deps/, .dart_tool/, cmake-build-*/, out/. RTK candidate stays evidence-only ("actual commands produce large output" is model-owned, §9).
- **§18 framework centrality**: extend `frameworkRules` to carry version evidence (deterministic: parse version from manifest) and `centrality_signals` (bounded import evidence, e.g. which files reference the framework); model decides central/supporting/incidental/unknown. No Go verdict.
- **Managers**: extend `internal/ecosystem/managers.go` conflict/ambiguity families (composer, cargo, bundler, mix, pub, conan, terraform) without adding selection tokens (decision-free invariant, §62).

### Stream 2 — Installer expansion
- **§13 RTK first-class**: catalog rtk → detect/version supported (executables+version args), install recipes where verified (brew/winget + standalone binary w/ checksum — no curl|sh, §21), post-install verify (existing `Install` path), **separate** OpenCode-global-integration opt-in (§47) never during init.
- **§20 safety levels**: additive `level` metadata (SAFE_RECIPE / VERSION_SENSITIVE / PROJECT_WRAPPER_PREFERRED / MANUAL / GLOBAL_SIDE_EFFECT) on catalog entries/recipes; surfaced in plan/status. uv=SAFE_RECIPE, composer=VERSION_SENSITIVE, maven/gradle=PROJECT_WRAPPER_PREFERRED, rustup=VERSION_SENSITIVE, RTK=SAFE_RECIPE+GLOBAL_SIDE_EFFECT.
- **§21 system PMs**: `DetectPackageManager` (`detect.go:49-56`, currently apt/pacman/dnf/brew) += zypper, apk, winget; recipes extended per tool where deterministic. AUR **opt-in only** (no automatic yay/paru recipe). No curl|sh: existing shell-meta rejection covers args; consider explicit executable denylist (sh/bash/curl-pipe) in `ValidateRecipe`.
- **§45 coverage**: full install for rg/fd/jq/gh/ast-grep (already, 5 recipes) + RTK + uv; composer detect+version+explain+plan on all platforms (execute only where a deterministic recipe exists). New `tools explain` verb (does not exist today — verified no "explain" symbol in cmd/ or internal/tools/). Detect/recommend catalog entries for npm, pnpm, yarn, bun, deno, uv, pip, poetry, pdm, pipenv, composer, cargo, rustup, go (exists), mvn, gradle, dotnet, bundle, mix, dart, flutter, cmake, conan, nix, terraform, tofu — the `Entry{Executables, VersionArgs}` mechanism supports them without install claims.
- **§46 UX**: extend `ConfirmConsent`/plan rendering with platform, method, executable, args, and the three "Changes" lines (installs user-level/global executable; does NOT modify OpenCode; does NOT modify project deps).
- **§56 honesty tests**: extend the existing pattern (`TestPlanSelectionAndFailClosed`, `TestInstallExecutesRecipeAndVerifies`, `TestInstallVerifyFailureFailsClosed`, `TestInstallSupportBackedByRecipe`, `TestConfirmConsentNeverDefaultsYes`) to every new `install:true` entry: detect/version/plan/execute/verify.

### Stream 3 — Advanced provider lifecycle truth
- **§14 policy + catalog**: add serena and headroom entries (absent today); provider entries declare capability truth honestly (context7/semble/serena/codegraph/headroom); recommendation candidates stay evidence-only (§62); policy content (when to recommend per §14.1-14.5) lives in orchestrator assets, not Go.
- **§48 provider metadata**: declare per provider: install method, project init requirement, OpenCode integration mode, global/local side effects, uninstall, health check — metadata-first; implement only what is deterministic, declare the rest unsupported (honesty invariant, `install:supported ⟺ verified`).
- **§49 provider health**: `tools doctor`/`doctor` gain provider checks: executable exists, version parses, project index/config exists if required, OpenCode integration detectable if enabled, provider health if inexpensive.
- **§47 RTK global opt-in UX**: after binary install, separate `Enable global integration? [y/N]` default N; never during `agent-ready init`; isolation tests proving ~/.config/opencode untouched unless approved (§57) and lossless local merge preserved (§58).

### Stream 4 — Fixture matrix (deterministic vs driven)
- **Deterministic table/cohort fixtures (no model)**: per-ecosystem detector tables (§33 Laravel, §34 uv, §35 pip, §36 npm/pnpm-workspace/Bun/Deno, §37 JVM wrapper, §38 .NET, §39 Ruby, §40 Elixir, §41 Flutter, §42 CMake/Conan, §43 monorepo facts, §44 boilerplate detection markers) as Go table tests in `internal/ecosystem/`, `internal/inventory/`, `internal/tools/` — same shape as existing `managers_test.go`. Covers §52 (PHP≠Node, Cargo.lock/flake.lock/composer.lock/Gemfile.lock/pubspec.lock/mix.lock recognized), §55-facts, §59 (exists), §60 (exists), §61 (mixed facts exist, extend matrix).
- **Driven regressions (model + OpenCode, bounded)**: §32 NixOS Wizard (`km-clay/nixos-wizard`-shaped fixture: Rust+Cargo+Nix+Ratatui, §66 outcome) as the flagship; Laravel and Monorepo/Boilerplate assessments as needed for §55 (provider minimalism small repo) and §43/§44. Driven battery stays within the documented `-timeout 40m` guardrail (verify-report SUGGESTION #1).
- **Acceptance §53 no-artifact-spam / §55 minimalism**: extend the C–P pattern and driven cohorts; generic-skill rejection already proven (fixture-g).

### Stream 5 — Content contracts
- **§15 Tool Budget explicit**: add a Tool Budget rule block (minimal-set ordering: rg+fd+jq → +ast-grep → +Semble OR Serena → CodeGraph only when graph value → Headroom only when measured pressure) to `audit-flow.md`/orchestrator SKILL.md.
- **§28 7 questions**: add the 7-question checklist (repository-specific / repeatable / non-trivial / project invariants / AGENTS-docs cheaper / deterministic script better / external verification required) to `artifact-design/references/artifact-decisions.md`.
- **§29 skill_request**: add the structured `skill_request` contract (purpose, repository_evidence, framework_evidence, external_verified_evidence, canonical_examples, invariants, commands, validation) to `skill-creator/SKILL.md` + `references/authoring-procedure.md`; gate: never invent external guidance when external_verified_evidence required and empty.
- **§30 named reviewer checks**: add `framework_grounding`, `package_manager_accuracy`, `toolchain_accuracy`, `external_verification_when_required` as named checks (rubric criteria or reviewer-procedure steps) — behavior partially exists via the rejection contract + External Verification Gate; names and rubric mapping absent.
- **§54 research quality**: name `external_verified_evidence` (or `external_verification_not_required` exemption) in `targeted-research` output contract; reviewer verifies.
- **Cleanup**: remove residual "Tool Manager is out of scope" (`search-strategies.md:24`).

### Stream 6 — CHANGELOG.md
- None exists (verified: no CHANGELOG.md at repo root; goreleaser.yml has a `changelog:` filter block using git log, no file). Tags `v0.1.0` (at 6a07341) and `v0.1.1` (at 7bc55bf) exist; the 10 delivered slices (`40ff819..d2fbbbf`) are untagged. Write CHANGELOG.md with v0.1.x history (reconstructed from git log) + v1.0.0 entry. Release tagging deferred (per session instruction "out of scope unless trivial").

## 3. Approaches

1. **Incremental close on the delivered architecture (recommended)** — six ordered streams as chained slices: facts → installer → providers → content → fixtures → CHANGELOG.
   - Pros: each slice is reviewable ≤400 lines, extends existing tested contracts (capability truth, decision-free facts, driven oracles), honors §64 exclusions, preserves the archived verify evidence.
   - Cons: many coordinated rules/tests; rule-engine extension must keep existing byte assertions green; driven regressions are wall-clock heavy.
   - Effort: High (auto-chain required; 400-line budget risk High).

2. **Full-spec one-shot completion** — implement §7–§61 in one batch.
   - Pros: single nominal milestone.
   - Cons: unreviewable (far beyond 400 lines), conflates facts/installers/providers/fixtures, risks provider-API speculation — the same failure the prior exploration rejected. Effort: Very High; not recommended.

3. **Facts + content only, defer installers/providers** — complete streams 1, 4 (deterministic), 5, 6.
   - Pros: smallest safe batch.
   - Cons: §13/§20/§21/§45/§47/§48/§49 stay open; RTK stays non-first-class; DoD keeps ~14 unchecked items. Effort: Medium; insufficient for "close V1" unless the user re-scopes.

## 4. Recommendation

Proceed with Approach 1, proposed as one change `agent-ready-v1-close` with chained slices in dependency order:

1. **Ecosystem matrix** (§7.x/§8/§9/§18 facts + rule-engine suffix support + manager families) — foundation for everything else.
2. **Catalog/installer truth** (§20 levels, §21 PMs, §45 recipes for RTK/uv + composer detect/version/explain/plan, §46 UX, §56 tests).
3. **Provider lifecycle truth** (§14 entries incl. serena/headroom, §48 metadata, §49 health, §47 RTK opt-in with §57/§58 isolation tests).
4. **Content contracts** (§15/§28/§29/§30/§54 + stale-phrase cleanup) — must precede driven fixtures so model behavior is proven against final instructions.
5. **Fixture matrix** (deterministic tables first, then NixOS Wizard + Laravel + monorepo/boilerplate driven regressions).
6. **CHANGELOG.md** (v0.1.x history + v1.0.0 entry).

Each slice ships its fixtures/tests with the behavior (no final fixture-only tail — per prior exploration's "tests taught to agree" warning). Delivery: `auto-chain`, stacked-to-main, ≤400 authored lines/slice. Exclusions honored throughout.

## 5. Risks

- **Rule-engine extension** (basename → suffix matching) can break existing byte-identity assertions (`TestDetectMixedRepositoryFacts`, `TestInspectDeterministic`); keep additive and update expectations in the same slice.
- **Installer execution tests**: recipes execute real package managers; keep the PATH-faking pattern of existing install tests; new PMs (apk/zypper/winget) are metadata + fail-closed tests on non-supported hosts.
- **RTK global integration is a high-severity boundary**: any implementation must stay opt-in (default N), never during init, with isolation tests proving zero `~/.config/opencode` mutation by default.
- **Content-contract changes shift reviewer behavior**: the §30 rubric/check changes need re-proof via content tests (`TestResearchDesignEvolutionContent` pattern) and the driven review cohort in the content slice.
- **Driven regression wall-clock**: NixOS Wizard + Laravel audits add minutes to the battery; keep the 40m timeout documented and bound driven additions (max 1-2 new driven cohorts).
- **Scope inflation in providers**: §48 lifecycle could grow into a subsystem; bound it by metadata-first honesty — implement only deterministic lifecycle, declare the rest unsupported.
- **Context7 signal is currently wrong per spec** (§16 partial): lockfile-only firing must become framework/dependency-relevance-gated; changing it will alter driven outcomes — re-run the audit cohort in that slice.
- **CHANGELOG drift** vs goreleaser git-log changelog: keep the file manual and factual; tagging v1.0.0 stays a delivery decision.
- **400-line budget risk: High** → auto-chain with clean rollback per slice (precedent: prior 10-slice chain).

## 6. Ready for Proposal

**Yes.** The proposal should: (a) define the six streams above with the slice order, (b) declare every §65 checkbox status from the matrix in Section 1, (c) keep facts decision-free and install/providers consent-gated, (d) ship fixtures with each slice, (e) exclude §64 items and the installer/provider deferred boundaries only where metadata-first honesty applies, (f) forecast chained PRs (High budget risk) and the driven-battery timeout guardrail. No rewrite; no new agents/commands beyond the single `tools explain` verb; no mandatory MCP.
