## Exploration: Agent-Ready V1 Completion Ecosystem

### Current State

#### Verified facts

The completion document is directionally corrective, but it describes a mix of current facts, already-delivered work, genuine gaps, and future design suggestions. The repository must be evolved in place; no rewrite is warranted.

| Claim cluster | Repository evidence | Verdict |
|---|---|---|
| Preserve the Go CLI + local OpenCode harness | `cmd/agent-ready/main.go` wires `init`, `inspect`, `state`, `changes`, `checkpoint`, `validate`, `tools`, `status`, `doctor`, `update`, and `remove`. `internal/bootstrap/bootstrap.go` embeds repository-local skills/references and `.opencode/commands/agent-ready.md`. `internal/opencode/config.go` losslessly adds `./.agent-ready/skills`. | Verified and must be preserved. The stale SDD init observation describing an init-only repository is no longer current. |
| `/agent-ready` and seven-skill North Star audit exist | `internal/bootstrap/assets/commands/agent-ready.md` dispatches audit/sync/review/status. Seven skills exist under `internal/bootstrap/assets/skills/`; checkpoints, state, validation, and the skill-quality references are implemented. | Verified. Do not recreate these capabilities. |
| Tool assessment is not mandatory | `agent-ready-orchestrator/SKILL.md:17` still says “tool management is out of scope.” `references/audit-flow.md:9,34-36` reads status “when available” and has no mandatory final Tool / Capability Assessment output. | Verified gap. Tool facts are integrated, but visible assessment is optional rather than an invariant. |
| Tool recommendations are narrow | `internal/tools/recommend.go:83-99` recognizes only `dist`, `build`, `coverage` and `package-lock.json`, `bun.lock`, `go.sum`, `yarn.lock`. A lockfile alone currently emits Context7 candidate evidence. | Verified gap. This is too narrow and also too weak a basis for versioned-documentation need. |
| Recipe coverage is five tools | `internal/tools/tools_test.go:11-24` locks the catalog to `ast-grep`, `fd`, `gh`, `jq`, `rg`; recipe JSON files match. `DetectPackageManager` supports only apt, pacman, dnf, brew. | Verified gap. RTK and uv are not first-class installable tools; zypper, apk, and winget are absent. |
| Tool support levels are conflated | `tools.Recipe` combines detection/version and installation. `tools.Status` iterates only recipes and emits a flat map; it cannot honestly represent detect-only, recommend-only, install-supported, configuration-supported, side effects, or the three tool families. | Verified gap. One catalog should evolve to allow optional capabilities; avoid a duplicate parallel truth source. |
| Ecosystem detection is broad | `inventory.Facts` contains only deps, scripts, workspaces, file counts, and CI. `Inspect` parses only root `go.mod` and `package.json`; no `internal/ecosystem` package exists. | Document claim is false as a current capability; the requested gap is real. None of the V1 ecosystem matrix, package-manager conflict resolution, wrapper preference, framework signals, or mixed-ecosystem output exists. |
| `inspect --json` already exposes the proposed facts | Current `agent-ready.inspect/v1` has no ecosystems, package managers, lockfiles, frameworks, build/test tools, workspace signals, or agent assets. | Not implemented. Schema evolution and compatibility need explicit design. |
| Research is wholly missing | `targeted-research` already pins exact versions and uses repository → local docs → version metadata → official docs → specialized provider → web. | Partially false. The search ladder is good and should remain; the missing piece is an explicit External Verification Gate tied to framework-dependent artifacts and reviewer enforcement. |
| Skill creation/review protects framework accuracy | `skill-creator` forbids invented APIs generally, and `skill-reviewer` traces evidence, but neither requires `external_verified_evidence`; reviewer checks do not name framework grounding, package-manager accuracy, toolchain accuracy, or required external verification. | Verified gap. Extend existing contracts rather than add skills. |
| Sync performs tool reassessment | `incremental-evolution` maps changed paths to evidence/artifacts but never requires reassessment when manifest, lockfile, workspace, CI, or output signals change. | Verified gap. Reassessment should be relevance-triggered, not every trivial sync. |
| Existing acceptance proves audit semantics | `internal/app/acceptance_test.go` builds and runs deterministic helpers, but many C–P cases merely assert pre-seeded `expect/` or `.agent-ready/state/*.jsonl` content. No test runs an actual OpenCode `/agent-ready audit`. | Partially false. The suite is a useful regression baseline, not proof of model-driven audit behavior. |
| Current tests are healthy | `go test ./...` passed on 2026-08-09 across all packages. | Verified, but green tests do not cover the completion matrix. |
| Global OpenCode isolation and local config merge work | `internal/opencode/version_test.go` isolates HOME/XDG during preflight; `PlanConfig` uses HuJSON and adds local `skills.paths`; tool recipes execute fixed argv without shell. | Verified architecture to preserve. Global RTK/provider integration must remain a separate explicit opt-in. |

Two additional implementation constraints were discovered:

1. `inventory.collectFiles` excludes only `.git`; it currently traverses dependency/output trees. Ecosystem detection must record directories such as `node_modules`, `target`, `vendor`, `.venv`, `bin`, and `obj` without recursively paying their file/context cost.
2. `lifecycle.UpdatePlan` only reconciles paths already listed in the installed manifest. `bootstrap.canonicalManifest` does not list the manifest itself, and newly embedded assets are skipped. Cross-version tests are absent. Any completion change that modifies or adds embedded harness assets needs a narrow ownership-preserving migration fix, otherwise installed manifests can remain stale.

#### Overlap with `sdd/agent-ready-northstar-audit/proposal`

The North Star proposal is implemented and archived, not an unimplemented competing design. Its seven skills, JSON helpers, checkpoint/state split, NO_ACTION discipline, C–P baseline, and local isolation are prerequisites for this change. This change should be specified as modifications to that delivered audit capability, not as a second audit system.

The corrective document overlaps the archived proposal in repository analysis, targeted research, artifact design, review, sync, and acceptance. It resolves two former deferrals: Tool Manager integration is now present but incomplete, and tool assessment must become mandatory. It must not reopen the proposal’s intentionally deferred `apply` helper, `.opencode/skills/` generation, or deep MERGE/DEPRECATE/REMOVE workflow unless separately justified.

### Affected Areas

- `internal/inventory/inventory.go` — existing public fact entry point; currently only Go/npm dependency facts, scripts, workspaces, counts, and CI.
- `internal/ecosystem/` (new only if a focused package keeps tables/resolution clearer) — deterministic ecosystem, manifest, lockfile, manager-conflict, wrapper, framework, and mixed-repository facts; no artifact verdicts.
- `internal/tools/catalog.go` — evolve the single catalog from recipe-only entries to honest optional detection, recommendation, install, configuration, integration, and side-effect capabilities.
- `internal/tools/detect.go` — categorized status facts for ecosystem tools, productivity tools, and providers.
- `internal/tools/recommend.go` — consume richer repository facts; emit evidence candidates and reasons, never install verdicts.
- `internal/tools/install.go` and `internal/tools/recipes/` — later safe-recipe expansion; preserve fixed argv, explicit consent, post-install verification, and no automatic global integration.
- `internal/bootstrap/assets/skills/agent-ready-orchestrator/{SKILL.md,references/audit-flow.md}` — mandatory visible assessment and five-area audit summary.
- `internal/bootstrap/assets/skills/{repository-analysis,targeted-research,artifact-design,skill-creator,skill-reviewer,incremental-evolution}/` — ecosystem reconnaissance, External Verification Gate, package-manager/toolchain grounding, and relevant-sync reassessment.
- `internal/bootstrap/bootstrap.go`, `internal/lifecycle/update.go` — cross-version ownership-safe delivery of changed/new embedded assets.
- `internal/app/acceptance_test.go` and `internal/app/testdata/acceptance/` — retain C–P and add behavior-driving ecosystem cohorts rather than seeded conclusions.
- `docs/` — update only after behavior and fixtures are proven.

### Acceptance-Fixture Implications

The completion matrix should be organized by root-cause behavior, not one bespoke test per document section:

1. **Deterministic detector table tests**: JS managers (npm/pnpm/Bun/Deno), Python managers (uv/pip/Poetry/PDM/Pipenv/Conda candidate), PHP/Composer, Rust+Nix, Go, JVM wrappers, .NET/NuGet, Ruby/Bundler, Elixir/Mix, Dart/Flutter, C/C++, SwiftPM, and basic IaC. Assert evidence, mixed ecosystems, stable ordering, ignored heavy trees, and no migration decisions.
2. **Resolution/conflict tests**: lockfile and package metadata precedence, `pnpm-lock.yaml` + `bun.lock` conflict, `pyproject.toml` not implying uv, project wrappers preferred, and multiple ecosystems retained concurrently.
3. **Tool contract tests**: detect/version/recommend/install/configure support are independently truthful; categorized status; every `install: true` entry proves plan, execute, and verify; detect-only tools never claim installation.
4. **Harness content-contract tests**: every initial audit prints Tool / Capability Assessment (including NO_ADDITIONAL_TOOLS); relevant syncs reassess; framework-dependent skills require versioned evidence; reviewer rejects package-manager or framework drift; Tool Budget and NO_ACTION remain first-class.
5. **Driven/manual regressions**: run actual `/agent-ready audit` against clean representative repositories, including `km-clay/nixos-wizard`, Laravel, small-repository provider minimalism, monorepo provider evaluation, and boilerplate. Semantic verdicts should allow alternatives but require explicit reasons. These cannot be replaced by pre-seeded JSONL.

Every implementation slice should ship its fixtures with the behavior it introduces. A final fixture-only tail would repeat the current “tests taught to agree” risk.

### Approaches

1. **Incremental completion on the existing architecture (recommended)** — keep `inventory.Inspect`, the existing tools package, seven skills, ownership manifest, and slash command; evolve contracts in bounded root-cause slices.
   - Pros: Preserves working behavior; removes duplicated interpretation; allows compatibility tests and reviewable rollback; aligns with Go-facts/model-decisions.
   - Cons: Requires explicit schema migration and many coordinated fixtures; cross-version ownership must be corrected first.
   - Effort: High. `auto-chain` is required; the 400-line budget risk is High.

2. **Implement the entire 67-section document as one completion batch** — add every detector, installer, provider integration, content rule, and fixture together.
   - Pros: One nominal V1 completion milestone.
   - Cons: Unreviewable, conflates detection with installation/integration, encourages speculative provider APIs, and cannot stay within the 400-line review budget.
   - Effort: Very High; not recommended.

3. **Content-only correction** — update orchestrator/research/reviewer wording while leaving deterministic facts narrow.
   - Pros: Small and fast.
   - Cons: Produces mandatory prose without reliable evidence, cannot satisfy ecosystem correctness, and risks hallucinated tool/framework conclusions.
   - Effort: Low; insufficient.

### Recommendation

Proceed incrementally and treat the completion document as an umbrella roadmap. For the named change, propose these ordered boundaries:

1. **Upgrade safety prerequisite**: prove and fix cross-version manifest reconciliation without weakening ownership or touching model state/checkpoints.
2. **Ecosystem facts**: bounded deterministic detection and manager conflict facts, exposed through an evolved inspect schema; no semantic repository class or artifact decision in Go.
3. **Tool truth model**: evolve the existing catalog so detection, recommendation, installation, configuration, and side effects are independent; categorized status and richer evidence-only recommendations consume ecosystem facts.
4. **Audit intelligence contracts**: mandatory visible assessment, External Verification Gate, reviewer grounding checks, and relevant-sync reassessment by modifying the existing seven-skill suite.
5. **Acceptance cohorts**: ship detector/resolver/tool/content tests with each slice, then driven/manual mixed-ecosystem regressions.

Keep safe installation expansion (RTK, uv, conditional Composer, extra system package managers) as a separately reviewable chain after the truth model is stable. Keep advanced provider install/init/uninstall/health and RTK global OpenCode integration in a further explicit opt-in change; detection and recommendation candidates may land earlier. These boundaries make incomplete support honest while preserving the completion roadmap.

The proposal should explicitly modify the delivered North Star audit and Tool Manager capabilities. It should not add agents, slash commands, language skills, package-manager skills, a TUI, daemon, database, required MCP, or Go semantic routing.

### Risks

- `agent-ready.inspect/v1`, `agent-ready.tools/v1`, and recommendation consumers need an explicit compatibility/versioning policy.
- Broad recursive scans can become slow and context-expensive unless output/dependency trees are pruned while preserving presence facts.
- Package-manager and framework detection can become a hardcoded workflow engine; keep outputs as evidence with confidence/conflict, not decisions.
- Existing installed repositories may receive stale or incomplete harness updates until cross-version manifest reconciliation is covered.
- Current green acceptance tests overstate semantic coverage because many conclusions are seeded.
- Provider and framework APIs in the August 2026 document are not repository facts; implementation must verify current official/versioned documentation before encoding recipes or guidance.
- Global side effects are a high-severity boundary: binary installation and OpenCode integration require separate consent and isolation tests.
- The full roadmap is far above 400 changed lines; `auto-chain` must split autonomous, test-backed work units with clean rollback.

### Ready for Proposal

Yes, with the recommended boundary. The proposal should define this change as the ecosystem-facts, tool-truth, and audit-contract completion of the existing architecture, include the ownership migration prerequisite and acceptance strategy, and explicitly defer provider integration/global side effects to named follow-up changes. It must separate verified current facts from corrective recommendations and must not claim the entire completion document is already validated truth.
