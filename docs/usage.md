# Usage

Reference for the CLI commands and the `/agent-ready` modes.

## CLI commands

Run `agent-ready <command> --help` for flags. All fact helpers accept `--json`.

| Command | Purpose | Mutates? |
|---|---|---|
| `agent-ready init` | Bootstrap the harness in the current repository | Yes (repo only) |
| `agent-ready update [--dry-run]` | Refresh harness assets to the installed binary's assets; ownership-preserving | Yes |
| `agent-ready status` | Facts: manifest, installed assets, mismatches, checkpoint, config, tools | No |
| `agent-ready doctor` | Read-only checks: repo, minimum compatible OpenCode, harness, command, skills, tools | No |
| `agent-ready remove --mode harness-only\|harness+generated` | Ownership-driven removal; modified/unowned files are never deleted | Yes |
| `agent-ready --version` | Version metadata injected at build time | No |

### Fact helpers (JSON schemas `agent-ready.<name>/v1`)

| Command | Facts emitted |
|---|---|
| `agent-ready inspect --json` | OS-independent inventory: dependencies, scripts, workspaces, files, CI; `agents_md` fact (path + lines) when a root AGENTS.md exists |
| `agent-ready validate --json` | Skill frontmatter vs installed OpenCode, progressive disclosure, ownership |
| `agent-ready checkpoint save --stage S [--complete]` | Go-owned envelope (latest == history bytes) |
| `agent-ready checkpoint status` | Latest checkpoint or none |
| `agent-ready changes --json` | Diff vs last checkpoint; first run reports the full inventory as added |
| `agent-ready state --json` | Model-owned semantic state files (exists/bytes/entries) |
| `agent-ready tools status --json` | Tool presence/version facts per recipe |
| `agent-ready tools doctor` | Tool tiers: required fail-hard, recommended warn |
| `agent-ready tools recommend --json` | Candidate evidence from documented signals (§36); never verdicts. Includes `structured_search_need` (ast-grep), `context_placement_pressure` (AGENTS.md > 300 lines; signal only), and an enriched RTK candidate fired by output dirs, build/test scripts, or CI |
| `agent-ready tools install <tool> [--dry-run]` | Consent-gated install via verified recipes; checksum-safe, fail-closed |

## `/agent-ready` modes

Run inside OpenCode. The command dispatches on `$ARGUMENTS` and never overrides your model choice.

| Mode | What happens |
|---|---|
| `/agent-ready` (or `audit`) | Progressive repository analysis → context placement evaluation → evidence-backed proposals (CREATE / UPDATE / REUSE / COMPACT / EXTRACT_TO_SKILL / MOVE_TO_REFERENCE / REPLACE_WITH_SCRIPT / REUSE_EXTERNAL_SKILL / NO_ACTION / ASK_USER) → approval → apply → review → checkpoint |
| `/agent-ready sync` | Selective update after repository changes (ChangeSet → affected evidence → artifact decisions); no full re-audit |
| `/agent-ready review` | Review artifacts/proposals against the rubric and evidence |
| `/agent-ready status` | Harness state summary |

## Recommended workflow

1. `agent-ready init` once per repository.
2. `/agent-ready` for the first audit; approve only what you agree with (or decline everything — `NO_ACTION` is valid).
3. After meaningful repository changes, run `/agent-ready sync` instead of a full audit.
4. Use `agent-ready doctor` and `tools recommend` when something feels off in the tooling.

## Exit codes

| Code | Meaning |
|---|---|
| 0 | Success (including `noop`, `dry_run`, warnings) |
| 1 | Helper/required-check failure |
| 2 | Usage error |
| 3 | Refused (preflight/plan refused) |
| 4 | Commit failed |
| 5 | Recovery required (transaction journal present) |
