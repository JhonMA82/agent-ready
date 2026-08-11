---
name: opencode-usage
description: "Trigger: running the installed OpenCode CLI (minimum compatible 1.18.15) for this repository. Every option cites official or local versioned evidence."
---
Run OpenCode commands for this repository with verified options only.

## Activation Contract
Run when a decision requires invoking the OpenCode runtime installed in this repository.

## Hard Rules
- Never use an option that is not documented for the installed OpenCode version.
- Record the source and version of every option with the decision.
- Stop when the command contract is verified locally; do not guess flags.

## Evidence
- Local versioned evidence: `opencode --version` reports the installed version and `opencode run --help` documents `--dir`, `--model`, `--format json`, and `--command` for the runtime installed in this repository (repository-to-official precedence: local sources come first).

## Execution Steps
1. Confirm the installed runtime: `opencode --version` and record the reported version.
2. Run `opencode run --dir <repo> --model <model> --format json --command <mode>`.
3. Record the source and version of every option with the decision.

## Output Contract
Return the executed command, the verified options with their versioned sources, and the recorded decision.
