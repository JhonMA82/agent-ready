# Changelog

All notable changes to Agent-Ready are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## Source of truth

Every entry below is cross-checked against `git log --oneline 6a07341..HEAD` on branch `feat/v1-close-changelog`; the **Covered commits** line under each release lists the exact commits the entry describes. No feature is listed that is not present in the commit history. Release tags follow the goreleaser `v*` filter (`goreleaser.yml` → `release`); the `v1.0.0` tag is deferred and is not created by this change.

## [v1.0.0] - 2026-08-11

Covers `40ff819..HEAD` (18 commits): the V1 completion change plus the six bounded close slices that bring `OPENCODE_AGENT_READY_V1_COMPLETION_2026.md` (§1–§67) to full compliance. HEAD at the time of writing is `9912ae1`.

**Covered commits**: `40ff819`, `8a35780`, `bba4806`, `4e6bb64`, `b65cf9b`, `304642b`, `df80798`, `dc15a5b`, `62cad46`, `d2fbbbf`, `ff57787`, `f440c56`, `29b8103`, `dc0c62a`, `2650e18`, `0e43d46`, `1b9349d`, `9912ae1`

### Added

- **Ecosystem detection matrix** — repository ecosystem detection for all §7.x families, extended to the full matrix with suffix rules, the complete §8 lockfile set, output/build signals, and framework facts carrying version evidence and centrality (`bba4806`, `ff57787`).
- **Manager families** — package-manager conflict resolution and manager families; wrapper scripts (`gradlew`, `mvnw`) preferred over global tools (`4e6bb64`, `29b8103`).
- **Tool capability truth** — independent tool-support reporting; 25 detect-only entries with install never claimed; verified RTK, uv, and composer install recipes; new `tools explain` verb (`b65cf9b`, `dc0c62a`, `2650e18`).
- **Provider lifecycle truth** — serena and headroom catalog entries; Context7 gated on real need (framework facts with version, never lockfile-only); `tools doctor` health checks; opt-in global integration, default off (`0e43d46`).
- **Content contracts** — §15 Tool Budget ordering, §28 seven questions, §29 `skill_request` fields, §30 named reviewer checks, and §54 vocabulary enforced in the embedded skills (`1b9349d`).
- **Fixture matrix** — deterministic §33–§44 ecosystem tables plus NixOS Wizard and Laravel driven regression cohorts proving §52–§61 acceptance (`9912ae1`).
- **External verification gate** — reviews enforce external verification when required (`62cad46`).

### Changed

- **Ownership-safe upgrades** — harness upgrades reconcile ownership and never clobber user-modified assets (`40ff819`).
- **Heavy-tree pruning** — repository traversal skips heavy trees; presence-only signals retained (`8a35780`).
- **Evidence-only recommendations** — tools recommended from ecosystem evidence; no lockfile-only or manager-conflict candidates (`304642b`).
- **Mandatory tool assessment** — audits mandate tool/capability assessment (`df80798`).
- **Relevant-sync reassessment** — tools reassessed on relevant syncs (`dc15a5b`).
- **OpenCode version decoupling** — harness no longer depends on an exact OpenCode version (`d2fbbbf`).

### Fixed

- Keyed ecosystem signal literals corrected in the detection matrix (`f440c56`).

## [v0.1.1] - 2026-08-09

Installer shipped as a release asset with a one-liner install.

**Covered commits**: `7e8f3ed`, `a2aa2ac`, `7bc55bf`

### Added

- Official project documentation: README with feature overview and direct installation explanation (`7e8f3ed`, `a2aa2ac`).
- Installer as a GitHub release asset enabling the one-liner `curl -fsSL https://github.com/JhonMA82/agent-ready/releases/latest/download/install.sh | sh` — checksum-verified (fails closed on mismatch), never requires sudo, installs to `$PREFIX/bin` (default `~/.local/bin`) (`7bc55bf`).

## [v0.1.0] - 2026-08-09

Initial public release: CLI foundation, docs, and installer.

**Covered commits**: full pre-release history to `6a07341` (34 commits); release-era commits `06731a1`..`6a07341`.

### Added

- Agent-Ready CLI foundation: `init`, `update`, `remove` lifecycle commands; deterministic fact helpers; tool catalog with `status`, `recommend`, `doctor`, and consent-gated `install`.
- Embedded agent-ready skills and acceptance fixtures shipped as owned, upgradeable assets.
- Release distribution: version metadata and `--version` flag, goreleaser pipeline, installer script with install guide and installer test harness (`06731a1`, `d0065d7`, `4a924cc`).

### Fixed

- Private-repository release downloads: authenticated asset download, POSIX fetch wrapper, API asset-download fallback, install env vars exported before verification steps, repo directory created before git init, fake opencode reporting the pinned version in the verify job (`86232c5`, `f21213a`, `5f7eead`, `9aed211`, `17822be`, `6a07341`).

---

[v1.0.0]: https://github.com/JhonMA82/agent-ready/releases/tag/v1.0.0
[v0.1.1]: https://github.com/JhonMA82/agent-ready/releases/tag/v0.1.1
[v0.1.0]: https://github.com/JhonMA82/agent-ready/releases/tag/v0.1.0
