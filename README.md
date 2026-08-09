# Agent-Ready

Prepare any repository so AI coding agents work better: understand it faster, use less context, and create only the artifacts that earn their place.

Agent-Ready is a lightweight Go CLI plus an OpenCode-native harness. It bootstraps a repository locally, installs a small set of high-quality reasoning skills, and turns your chosen OpenCode model into an orchestrator that audits, documents, and maintains the repository — incrementally and with minimal token spend.

> **Principle:** the quality of decisions and the context avoided matter more than the number of files, skills, or tools installed.

## Quick path

```sh
# 1. Install the binary — see the Installation section below
agent-ready --version

# 2. Bootstrap the repository locally
agent-ready init

# 3. Open the repo in OpenCode and run the audit
/agent-ready
```

Expected result: `agent-ready init` creates `.agent-ready/` and registers the `/agent-ready` command locally, touching nothing global. The audit proposes only evidence-backed artifacts (and `NO_ACTION` is a first-class outcome).

## Installation

**Requirements:** Git · OpenCode 1.18.15 (pinned) · Linux, macOS, or Windows (WSL for the script).

### Option A — installer script (Linux/macOS, recommended)

The installer lives in this repository at `scripts/install.sh`. It downloads the release asset for your platform, verifies the sha256 checksum (fails closed on mismatch), and installs to `$PREFIX/bin` (default `~/.local/bin`). It never requires sudo.

```sh
# From a checkout of this repository:
VERSION=latest PREFIX=$HOME/.local ./scripts/install.sh

# Or pin an exact release:
VERSION=v0.1.0 PREFIX=$HOME/.local ./scripts/install.sh

# For a system-wide install, choose your own privilege mechanism:
sudo env VERSION=latest PREFIX=/usr/local ./scripts/install.sh
```

After installing, add the prefix to your PATH if it isn't there:

```sh
export PATH="$HOME/.local/bin:$PATH"   # add to your shell profile
```

### Option B — manual install from a release

1. Open the [releases page](https://github.com/JhonMA82/agent-ready/releases) and download the asset for your platform, e.g. `agent-ready_0.1.0_linux_amd64.tar.gz` (or `.zip` on Windows), plus `checksums.txt`.
2. Verify the checksum before installing:

   ```sh
   sha256sum -c checksums.txt     # macOS: shasum -a 256 -c checksums.txt
   ```

3. Extract and install:

   ```sh
   tar -xzf agent-ready_0.1.0_linux_amd64.tar.gz
   install -m 0755 agent-ready "$HOME/.local/bin/agent-ready"
   export PATH="$HOME/.local/bin:$PATH"
   ```

### Windows

Download the `.zip` asset, extract it, and add the folder to `PATH`. There is no Windows installer in V1.

### Verify the installation

```sh
agent-ready --version   # e.g. agent-ready 0.1.0 (commit)
```

Full details, uninstall, and update semantics: [INSTALL.md](INSTALL.md).

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
