# Recommended Models

Agent-Ready is model-agnostic by design: the `/agent-ready` command never sets a provider or model, so whatever you select in OpenCode is what runs the harness. What matters is matching the model to the phase.

## Guidance by phase

| Phase | Workload | Model tier | Why |
|---|---|---|---|
| **Audit orchestration** | Progressive exploration, evidence labeling, artifact decisions, proposal drafting | Strong reasoning (frontier or large reasoning model) | The quality of the audit is the quality of the reasoning: what to explore, what to ignore, when to ask |
| **Skill authoring/review** | Writing lean skills, rubric scoring, canonical-example discipline | Strong reasoning + instruction following | The skill quality system depends on consistent rubric interpretation |
| **Sync / incremental** | ChangeSet interpretation, selective updates | Strong reasoning, but only for changed evidence | Cheaper than audit; context is small by design |
| **Fact gathering** | `inspect`, `validate`, `checkpoint`, `changes`, `state`, `tools` | Any model (or no model at all) | These are deterministic CLI facts; the model only reads them |
| **Mechanical edits** | Applying approved artifacts | Small/cheap model is fine | The decision was already made; execution is routine |

## Practical notes

- **Start with your best reasoning model** for the first `/agent-ready` audit. Later syncs need less.
- The harness **reduces** model burden: progressive disclosure means skills and references load only when needed, and checkpointed evidence is reused instead of re-explored.
- Token-heavy repositories benefit most: the audit never dumps the repository into context; it reads targeted facts first.
- DeepSeek, Claude, GPT-class, and other OpenCode-supported providers all work; there is no provider-specific integration.

## What the harness assumes about the model

The model must be able to:

- Follow the orchestrator's adaptive loop (map, not script).
- Label findings FACT / INFERENCE / UNKNOWN honestly.
- Create nothing when evidence does not justify it (`NO_ACTION`).
- Apply the rubric thresholds exactly (≥85 PASS, 70–84 REVISE, <70 REJECT).

If a model cannot hold that discipline, the harness still produces deterministic facts — the audit quality, not the tooling, degrades.
