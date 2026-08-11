# How It Works

Agent-Ready splits responsibilities cleanly: the Go CLI produces deterministic facts and safe operations; the model interprets meaning and makes decisions.

```
user opens OpenCode
    → chooses a model normally
    → runs /agent-ready
    → OpenCode discovers the local harness skills
    → the model orchestrates analysis, research, and generation
    → calls the CLI for facts when it needs them
```

## Responsibility split

| Layer | Owns | Never does |
|---|---|---|
| **Go CLI** | Bootstrap, detection, inventory, manifests, Git, hashes, checkpoints, validation, tool detection/recipes, safe config merges, ownership, remove | Semantic routing (`if path contains "src/api" → create skill`), recommendations-as-verdicts, artifact decisions |
| **Model (OpenCode)** | What to explore, what context is needed, which skill to load, when to research, what an artifact should be, when to ask the user, when to return `NO_ACTION` | Mutating global OpenCode, guessing tool recipes, inventing conventions |

The rule: **Go delivers facts; the model interprets the meaning.**

## The audit loop

The orchestrator skill teaches the model to keep asking:

```
What do I know? What don't I know? What evidence do I need?
Which capability helps? Do I need to research? Do I need to ask?
Do I have enough confidence? Does an artifact add value?
Is creating nothing better? Should I review again?
```

It is a map, not a script. Stages may be skipped on resume, revisited on new evidence, and stopped by the first applicable stop condition (`ASK_USER` or `STOP_WITH_CONCERNS`).

## Data flow

1. `agent-ready init` materializes `.agent-ready/` assets and the `/agent-ready` command into the repository.
2. `/agent-ready audit` starts: the orchestrator reads deterministic facts (`inspect`, `state`, `changes`, `checkpoint status`, `tools status`).
3. `repository-analysis` explores with FACT / INFERENCE / UNKNOWN labels.
4. `targeted-research` closes gaps only when necessary (repo → local docs → version metadata → official docs → specialized providers → web last).
5. `artifact-design` decides: CREATE / UPDATE / REUSE / REMOVE / NO_ACTION / ASK_USER — each with evidence and alternatives.
6. Proposals are presented for approval; the model applies approved artifacts and records decisions.
7. `skill-reviewer` gates any new skill against the rubric (≥85 PASS).
8. `checkpoint save --stage S` records progress; a later run resumes without repeating completed evidence.

## Checkpoints and ownership

- **Go owns** checkpoint envelopes (`.agent-ready/checkpoints/`): stage, hashes, resume facts.
- **The model owns** semantic state (`.agent-ready/state/decisions.jsonl`, `provenance.jsonl`, etc.).
- State and checkpoints are **outside** the ownership manifest, so model-written state never makes `init` refuse on rerun.

## Skill quality system

Every harness skill is validated against a rubric (necessity 25, repository-specificity 20, discovery description 15, procedural value 15, progressive disclosure 10, evidence grounding 10, validation 5):

| Score | Verdict |
|---|---|
| ≥ 85 | PASS |
| 70–84 | REVISE |
| < 70 | REJECT |

Scores force justification; canonical examples (`excellent-simple`, `excellent-complex`, `bad-generic`, `unnecessary-skill`) calibrate what good looks like. The frontmatter of every skill is machine-validated against the installed OpenCode rules (minimum compatible 1.18.15).
