# Agent-Ready

Prepare any repository so AI coding agents work better: understand it faster, use less context, and create only the artifacts that earn their place.

Agent-Ready is a lightweight Go CLI plus an OpenCode-native harness. It bootstraps a repository locally, installs a small set of high-quality reasoning skills, and turns your chosen OpenCode model into an orchestrator that audits, documents, and maintains the repository — incrementally and with minimal token spend.

> **Principle:** the quality of decisions and the context avoided matter more than the number of files, skills, or tools installed.

## Quick path

```sh
# 1. Install the binary (Linux/macOS) — see INSTALL.md for details
#    Option A: run the repository's installer:  scripts/install.sh
#    Option B: manual — download the release asset, verify checksums.txt, add to PATH
agent-ready --version

# 2. Bootstrap the repository locally
agent-ready init

# 3. Open the repo in OpenCode and run the audit
/agent-ready
```

Expected result: `agent-ready init` creates `.agent-ready/` and registers the `/agent-ready` command locally, touching nothing global. The audit proposes only evidence-backed artifacts (and `NO_ACTION` is a first-class outcome).

## What it does

| Capability | What you get |
|---|---|
| **Local bootstrap** | `init` installs the harness skills and command inside the repo; zero global OpenCode changes |
| **Cognitive audit** | `/agent-ready` analyzes the repository progressively: facts, inferences, unknowns → evidence-backed artifact decisions |
| **7 harness skills** | orchestrator, repository-analysis, targeted-research, artifact-design, skill-creator, skill-reviewer, incremental-evolution — with progressive disclosure |
| **Skill quality system** | rubric (≥85 PASS / 70–84 REVISE / <70 REJECT), canonical examples, anti-patterns |
| **Deterministic helpers** | `inspect`, `validate`, `checkpoint`, `changes`, `state`, `tools` — machine-parseable JSON facts, no semantic routing |
| **Lifecycle** | `status`, `doctor`, `update`, `remove` — ownership-tracked and safe |
| **Tool manager** | `tools status/recommend/doctor/install` with consent-gated, verified recipes |
| **Checkpoint/resume** | interrupted audits resume without repeating completed evidence |

## What it is NOT

- Not a fixed `AGENTS.md` generator
- Not a collection of templates or OpenCode agents
- Not an indiscriminate MCP installer
- Not an SDD framework or a duplicate of Gentle AI

## Documentation

- [Overview](docs/overview.md) — what the project is and the principles behind it
- [How it works](docs/how-it-works.md) — CLI vs. model responsibilities, data flow
- [Getting started](docs/getting-started.md) — install, first run, workflow
- [Usage](docs/usage.md) — command and mode reference
- [Recommended models](docs/recommended-models.md) — model guidance per phase
- [Generated structure](docs/generated-structure.md) — what the harness creates (and what it won't)
- [Productivity impact](docs/productivity.md) — token discipline and expected effects
- [Installing](INSTALL.md) — release script, manual install, checksums

## Requirements

- Git
- OpenCode 1.18.15 (pinned compatibility version)
- Linux, macOS, or Windows (WSL for the install script)

## Status

V1 released as [v0.1.0](https://github.com/JhonMA82/agent-ready/releases). CI runs go test/vet plus a real-OpenCode compatibility smoke; releases are built with GoReleaser and verified for install isolation (acceptance A/Q).
