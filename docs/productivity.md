# Productivity Impact

Agent-Ready is designed explicitly to reduce token consumption and increase agent effectiveness. This page explains the mechanisms and sets honest expectations.

## What gets better

| Mechanism | Effect |
|---|---|
| **Progressive disclosure** | Skills load only when needed; references load on demand. The orchestrator body stays lean instead of carrying every procedure inline |
| **Targeted context** | The audit starts from structured facts (`inspect`, `state`, `changes`), not from reading files; search → targeted read, never dumps |
| **Checkpoint/reuse** | Completed evidence is reused; interrupted audits resume at their stage instead of re-exploring |
| **Incremental sync** | Repository changes produce selective updates; no full re-audit |
| **Deterministic facts** | JSON facts replace re-reading and re-interpreting manifests; less hallucination, fewer wasted turns |
| **NO_ACTION** | The audit may (and often should) create nothing — the cheapest possible outcome |
| **Artifact discipline** | No skills/docs/MCPs created without evidence; less noise for every future session |

## What it costs

| Cost | When |
|---|---|
| One-time `init` | Seconds; local-only |
| First audit | Real tokens: the model explores, researches (only if needed), proposes, and you review |
| Skill authoring | Only when evidence justifies a repository-specific skill; rubric-scored and reviewed |

The harness shifts spending from *every session re-exploring* to *one deliberate audit*, then maintains itself incrementally.

## Expected effects over time

| Scenario | Before | After |
|---|---|---|
| New agent joins a repo | Full re-exploration, context dumps, invented conventions | Targeted facts + skills/docs that survived the rubric |
| Small repo with nothing special | Unnecessary skills/docs created | `NO_ACTION` recorded; zero artifacts |
| Dependency added | Guesswork or full re-read | `changes` shows exactly what moved; selective update |
| Interrupted work | Everything re-done | Checkpoint resume from the recorded stage |

## Token discipline rules the orchestrator follows

```text
smallest useful context
search → targeted read
avoid duplicate reads
load skill only when needed
load reference only when needed
stop research when resolved
reuse checkpointed evidence
stop with ASK_USER or STOP_WITH_CONCERNS after no-new-evidence iterations
```

## Measuring it

Use the deterministic helpers to observe the harness itself:

```sh
agent-ready status          # installed assets, mismatches, checkpoint state
agent-ready changes --json  # what changed since the last checkpoint
agent-ready tools recommend # whether new tooling is even justified
```

A healthy setup is one where `status` shows no drift, `changes` is small between syncs, and `doctor` exits 0.
