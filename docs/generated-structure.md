# Generated Structure

What `agent-ready init` creates in a repository — and what it deliberately will not create.

## After `init`

```text
project/
├── .agent-ready/
│   ├── manifest.json              # ownership manifest (schema agent-ready.manifest/v1)
│   ├── skills/                    # internal harness skills (7)
│   │   ├── agent-ready-orchestrator/
│   │   ├── repository-analysis/
│   │   ├── targeted-research/
│   │   ├── artifact-design/
│   │   ├── skill-creator/
│   │   ├── skill-reviewer/
│   │   └── incremental-evolution/
│   ├── references/
│   │   └── skill-system/          # rubric, authoring guide, anti-patterns, canonical examples
│   ├── state/                     # model-owned semantic state (created by init, NOT manifest-owned)
│   └── checkpoints/               # Go-owned resume envelopes (NOT manifest-owned)
├── .opencode/
│   └── commands/
│       └── agent-ready.md         # the /agent-ready command (no model/agent keys)
└── opencode.json                  # created only if missing; adds skills.paths to existing config
```

All files are byte-tracked in the ownership manifest, so `doctor`/`status` detect drift, `update` reconciles, and `remove` never deletes user-modified content.

## Runtime files (not in the manifest)

| Path | Owner | Purpose |
|---|---|---|
| `.agent-ready/state/` | model | decisions.jsonl, provenance.jsonl, artifact-graph.yaml, repository-profile.yaml |
| `.agent-ready/checkpoints/` | Go | latest.json + history/<id>.json envelopes |

These are created by `init` but intentionally **outside** the ownership manifest, so model-written state never causes a later `init` rerun to refuse.

## What the audit MAY create (only with evidence + approval)

```text
AGENTS.md                      # short router, only if justified
docs/ai/…                      # context-map, invariants, workflows, etc. — per decision
.opencode/skills/<name>/       # project-specific skills (authored via skill-creator,
                               # gated by skill-reviewer, rubric ≥ 85)
scripts/agent/…                # deterministic helpers when a skill is not the right tool
PROJECT.template.md            # boilerplate repositories only
```

There is **no mandatory set**. The audit decides per artifact and records the decision with evidence, alternatives, and confidence.

## What the harness will NOT create

- A skill for every dependency or framework.
- Copies of official documentation.
- Generic language/framework skills.
- Anything without repository evidence.
- Global OpenCode skills, agents, or commands (ever).

## Config integration

`opencode.json` / `opencode.jsonc` is merged losslessly: comments, trailing commas, newline style, and unrelated keys are preserved. The only change is the addition of:

```jsonc
{ "skills": { "paths": ["./.agent-ready/skills"] } }
```

If both `opencode.json` and `opencode.jsonc` exist with conflicting `skills` definitions, `init` refuses rather than guessing.
