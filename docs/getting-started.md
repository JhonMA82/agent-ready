# Getting Started

Install the binary, bootstrap a repository, and run your first audit.

## 1. Install

Choose one of the options below (details in [INSTALL.md](../INSTALL.md)).

```sh
# Option A — one-liner: fetches the installer from the latest release
curl -fsSL https://github.com/JhonMA82/agent-ready/releases/latest/download/install.sh | sh

# Option B — installer script from a checkout: verifies the sha256 checksum,
# installs to $PREFIX/bin (default ~/.local/bin).
VERSION=latest PREFIX=$HOME/.local ./scripts/install.sh

# Option C — manual: download the asset for your platform from the release
# page, verify checksums.txt, extract, and add the binary to PATH.
```

The installer verifies the sha256 checksum before installing anything; a mismatch aborts with nothing installed. Windows: download the zip and add the folder to PATH.

Verify:

```sh
agent-ready --version   # e.g. agent-ready 0.1.0 (commit)
```

## 2. Bootstrap a repository

```sh
cd /path/to/your/repo
agent-ready init
```

Expected result:

```text
Outcome: changed
- create .agent-ready/manifest.json
- create .agent-ready/skills/...
- create .opencode/commands/agent-ready.md
- create opencode.json          (only if the repo had none)
Next: /agent-ready
```

`init` touches only this repository: no global OpenCode files, no other projects. Re-running it is a no-op; `agent-ready status` shows the current state; `agent-ready doctor` checks the setup.

## 3. Run the audit

Open the repository in OpenCode (1.18.15) and run:

```
/agent-ready
```

The default mode is `audit`. The orchestrator will explore progressively, propose only evidence-backed artifacts, and ask before creating anything. On a small repository the honest outcome is often `NO_ACTION` — that is a success, not a failure.

## 4. Keep it maintained

```text
/agent-ready status    → current harness state
/agent-ready sync      → selective updates when the repo changes
/agent-ready review    → review proposed artifacts
agent-ready update     → refresh harness assets to the installed binary
agent-ready remove     → remove the harness (two modes, ownership-safe)
```

## Requirements

- Git
- OpenCode 1.18.15 (pinned compatibility version)
- A model of your choice — see [Recommended models](recommended-models.md)

## Troubleshooting

| Symptom | Check |
|---|---|
| `init` refuses: "OpenCode 1.18.15 required" | `opencode --version`; install the pinned version |
| `init` refuses on a conflict | an unowned or modified target exists; resolve or remove it, then retry |
| `/agent-ready` not found | verify `.opencode/commands/agent-ready.md` exists and OpenCode is 1.18.15 |
| `doctor` exits 1 | a required check failed (`git`/`opencode`/initialized harness); read the named check |
