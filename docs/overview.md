# Overview

Agent-Ready makes a repository *agent-ready*: a place where an AI coding agent can start working with accurate context, low token spend, and evidence-backed conventions — without a human pre-writing exhaustive documentation.

## The problem

Agents waste context and make worse decisions when a repository is unfamiliar:

- Repetitive exploration of the same codebase, session after session.
- Full-context dumps instead of targeted reads.
- Hallucinated conventions that don't match the repository.
- Artifact sprawl: skills, docs, and MCPs created because they *could* help, not because evidence shows they will.

## The approach

A small Go CLI provides deterministic facts and safe operations. A set of high-quality OpenCode skills teaches your model how to reason about a repository: what to explore, what to ignore, when to ask, and — critically — when to create **nothing**.

The model you already use in OpenCode does the thinking; Agent-Ready provides the discipline.

## Principles

| Principle | Meaning |
|---|---|
| **Smallest useful context** | Search → targeted read; never dump repositories into the conversation |
| **Evidence over assumption** | Artifacts are created only with evidence; findings are labeled FACT / INFERENCE / UNKNOWN |
| **NO_ACTION is first class** | Not creating something is a valid, recorded outcome |
| **Progressive disclosure** | Skills stay lean; details load only when needed |
| **Deterministic facts** | The CLI reports what *is* (JSON), the model decides what it *means* |
| **Incremental, not from scratch** | Sync and resume reuse prior evidence; no full re-audits |
| **Local and isolated** | Everything happens inside the repository; global OpenCode state is never touched |

## What it is not

- A fixed `AGENTS.md` generator — the audit decides, and may decide nothing is needed.
- A template collection — no boilerplate libraries are installed.
- An MCP installer — tools are recommended by evidence and installed only with consent.
- An SDD framework — Agent-Ready prepares repositories; it does not run development processes.
- A duplicate of Gentle AI — no runtime, session management, or agent layers are added.

## The North Star

> Enter a repository — existing project, boilerplate, monorepo, or small app — understand it progressively, decide what it truly needs for agents to work better, create only justified artifacts, and keep that preparation synchronized as the project evolves.

## V1 scope (released)

- Deterministic CLI: `init`, `update`, `status`, `doctor`, `remove`, `tools`, and JSON-fact helpers.
- Cognitive layer: 7 harness skills, skill quality system, `/agent-ready` modes (audit, sync, review, status).
- Distribution: GoReleaser releases, POSIX installer with checksum verification, CI-verified install isolation.

Out of scope for V1: TUI, managed sessions, global OpenCode mutation, package-manager publishing, artifact signing, and the deep skill lifecycle approval flows (MERGE/DEPRECATE/REMOVE).
