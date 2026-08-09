# Installing Agent-Ready

Agent-Ready is a single Go binary that bootstraps and maintains a repository
as agent-ready. Installation never modifies global OpenCode configuration
(`~/.config/opencode` stays untouched), and `agent-ready init` only touches
the repository it runs in (spec acceptance A and Q).

## Requirements

- Linux, macOS, or Windows (WSL for the script on Windows)
- Git
- OpenCode 1.18.15 (pinned compatibility version)

## Option 1: One-liner (Linux/macOS)

```sh
curl -fsSL https://github.com/JhonMA82/agent-ready/releases/latest/download/install.sh | sh
```

Fetches the installer from the latest release and installs to `~/.local/bin`
with default settings. The installer verifies the sha256 checksum before
installing anything; a mismatch aborts with nothing installed. Review first
with `... | less` if you prefer. While the repository is private, downloads
require a token — see the authenticated variant in the README or use
Option 2 from a checkout.

## Option 2: Installer script from a checkout

```sh
# Install the latest release into ~/.local/bin
VERSION=latest PREFIX=$HOME/.local ./scripts/install.sh

# Or a specific release
VERSION=v0.1.1 PREFIX=$HOME/.local ./scripts/install.sh

# The script verifies the sha256 checksum before installing anything;
# a mismatch aborts with nothing installed.
```

The script downloads the asset and its `checksums.txt` from the release,
verifies the hash, and installs to `$PREFIX/bin/agent-ready`. It never
requires sudo; for a system-wide install, run with `PREFIX=/usr/local` under
your own privilege mechanism.

## Option 3: Manual install

1. Open the release page and download the asset for your platform:
   `agent-ready_<version>_<os>_<arch>.tar.gz` (or `.zip` on Windows).
2. Verify the checksum:

   ```sh
   sha256sum -c checksums.txt   # or: shasum -a 256 -c checksums.txt
   ```

3. Extract and install:

   ```sh
   tar -xzf agent-ready_<version>_linux_amd64.tar.gz
   install -m 0755 agent-ready "$HOME/.local/bin/agent-ready"
   export PATH="$HOME/.local/bin:$PATH"
   ```

## Windows

Download the `.zip` asset, extract it, and add the folder to `PATH`. There is
no installer for Windows in V1.

## First run

```sh
agent-ready init          # bootstraps the repository harness locally
agent-ready --version     # confirms the installed version
```

Then open the repository in OpenCode and run `/agent-ready`.

## Updating

`agent-ready update` refreshes the repository's harness assets to the ones
shipped in the installed binary (spec §32). It never regenerates `AGENTS.md`,
`docs/ai`, or project skills. For a newer binary version, reinstall via the
release script.

## Uninstalling

`agent-ready remove --mode harness-only` removes the harness assets from the
current repository; `--mode harness+generated` also removes generated state
and the installed config entry when it is byte-identical to the installed
one. Manual OpenCode configuration and global tools are never touched.

## Guarantees

- No global OpenCode mutation at install or init time.
- Checksum-verified installs (fail closed on mismatch).
- Repository-local bootstrap with ownership tracking for safe removal.
