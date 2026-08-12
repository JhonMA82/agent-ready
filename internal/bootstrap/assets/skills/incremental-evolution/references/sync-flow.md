# Sync Flow

How a ChangeSet becomes selective updates. Read this reference when `changes` reports a diff or when sync must decide what to touch. It resolves how to interpret the ChangeSet and when tool capabilities must be reassessed.

## ChangeSet interpretation (normative)

| ChangeSet entry | Meaning | Sync response |
|---|---|---|
| `added` | Path exists now, not at baseline | Evaluate evidence; update only if it affects a shipped artifact |
| `changed` | Path content differs from baseline | Update the evidence and artifacts that cite this path |
| `removed` | Path gone since baseline | Re-check dependents; removal decisions stay within sync scope (no lifecycle approval flows) |
| `first_run` | No baseline existed | Treat as initial inventory, not a change signal; no updates from it alone |

## Tool-need reassessment (normative)

- The ChangeSet and stage always come from the fact helpers: run `agent-ready changes --json` and `agent-ready checkpoint status`; direct file reads never replace them.
- A sync MUST assess whether changed evidence can affect tool needs; never skip the assessment silently.
- Manifest, lockfile, workspace, wrapper, CI, framework, build/test output, or tool-fact changes MUST trigger reassessment of tool capabilities.
- Reassess tool needs when the repository changes materially: repo complexity materially changes, workspace count changes, new framework added, new language ecosystem added, tool output problem observed, or a new provider is already installed.
- A completed reassessment MUST include reasons and either categorized recommendations (ecosystem, productivity, provider) or `NO_ADDITIONAL_TOOLS`.
- Irrelevant changes (prose, docs, formatting) MUST record a reason for skipping the reassessment.

## Placement-change detection

Detect placement changes in the ChangeSet: AGENTS content changed, skill changed, reference changed, canonical example changed. When a skill was extracted from AGENTS, update the source dependency graph (derived_from / routed_from / refresh_when) instead of re-reading sources; never re-duplicate automatically — an extraction's source change updates the graph, it does not copy content back.

## Artifact graph relations and placement provenance

Artifact relations are recorded on the artifact graph:

```yaml
artifact:
  path: .opencode/skills/add-dashboard-screen/SKILL.md
  derived_from:
    - AGENTS.md#adding-screen
    - src/routes/**
  routed_from:
    - AGENTS.md
  refresh_when:
    - source_section_changed
    - route_structure_changed
    - canonical_example_changed
```

Placement moves record provenance:

```yaml
placement_change:
  from:
    path: AGENTS.md
    section: "Adding a screen"
  to:
    path: .opencode/skills/add-dashboard-screen/SKILL.md
  reason: task_specific_procedure
  preserved_router:
    path: AGENTS.md
    text: "Use the add-dashboard-screen skill for new dashboard screens."
  source_hash: <hash>
```

Sync reads this provenance to decide whether a source change refreshes the extracted artifact or only the graph.

## Rules

- Selective updates only: a changed path updates the evidence that cites it and nothing else; artifacts whose evidence is unchanged are not touched.
- No full re-audit: sync never re-runs completed stages; it reuses checkpointed evidence (R14) and reads `changes` hashes, never model state.
- Not every dependency requires change: a dependency bump updates the evidence that cites it, not every artifact in the harness.
- Reassessment reads `tools status` and `tools recommend` facts and never re-runs the initial audit; skip decisions are recorded in state.
- Unchanged paths never appear in sync output; they are already aligned with the checkpoint baseline.
- Nothing changed and no pending plan: return NO_ACTION with zero artifacts.

## Output templates (shape guidance)

These examples show the required shape, not prescribed semantic conclusions. Model-owned verdicts and artifact choices remain free.
Every classification and reassessment record MUST carry an explicit reason grounded in the observed ChangeSet.

### Relevant lockfile change

```text
ChangeSet: changed package-lock.json
Stage: baseline
Classification: relevant — reason: the new lockfile is versioned-dependency evidence and may change capability needs.
Reassessment: completed — reason: lockfile evidence can affect tool selection.
Recommendations: ecosystem=<model-owned>; productivity=<model-owned>; provider=<model-owned>
```

```jsonl
{"type":"sync","path":"package-lock.json","relevance":"relevant","reason":"lockfile adds versioned-dependency evidence that may affect capability needs","reassessment":"completed because lockfile evidence can affect tool selection","recommendations":{"ecosystem":"<model-owned>","productivity":"<model-owned>","provider":"<model-owned>"}}
```

### Irrelevant prose change

```text
ChangeSet: changed README.md
Stage: baseline
Classification: irrelevant — reason: prose-only content does not alter repository tool or framework evidence.
Reassessment: skipped — reason: no tool-capability facts changed.
```

```jsonl
{"type":"sync","path":"README.md","relevance":"irrelevant","reason":"prose-only content does not alter repository tool or framework evidence","reassessment":"skipped","skip_reason":"no tool-capability facts changed"}
```

Before return, append one model-owned JSONL state record to `.agent-ready/state/decisions.jsonl` for the sync. The Go layer observes facts and does not write semantic state.
