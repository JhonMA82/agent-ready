---
name: repository-analysis
description: "Trigger: analyzing repository evidence in an audit. Inspect repository facts, label findings FACT/INFERENCE/UNKNOWN, and build the evidence base for decisions."
---
Produce the labeled evidence base the audit decides on. Load `references/evidence-labels.md` when classifying findings; load `references/inventory-facts.md` when gathering facts.

## Activation Contract
Run when the orchestrator starts exploration, when a new finding needs classification, or when a decision lacks evidence.

## Hard Rules
- Label every finding FACT, INFERENCE, or UNKNOWN; never present an inference as a fact.
- Never propose an artifact on UNKNOWN-only evidence; gather the fact or stop with ASK_USER.
- Never dump repository content into context; record evidence as labeled facts.
- Classify repository kind when evidence supports it: primary kind from {application, library, cli, starter, boilerplate, template, infrastructure, mixed} with secondary kinds and a confidence score (FACT when evidenced, INFERENCE otherwise). `monorepo` is a topology fact, never a kind: express it as `topology.monorepo: true`.
- When kind.primary is boilerplate/starter/template, run the Boilerplate Assessment: extension points; editable boundaries (what downstream users should edit and should not edit); generated files; feature addition workflow; variants/presets; scaffolding; upgrade/update strategy; canonical customization examples. The assessment never creates artifacts by itself; it only demonstrates that the evaluation happened.
- Ask the boilerplate placement questions: which instructions must be always-on; which workflows should become on-demand skills; which examples should become references.
- When evidence shows repeated implementations of the same intent, example-rich boilerplate/starter/template content, a UI-rich application with repeated screen/component composition, or a repeated repository-specific implementation shape, run the Pattern & Exemplar Analysis below. Do not run it deeply in repositories without repeated examples: small script, single-screen app, library with no repeated implementation flows, empty starter.

## Execution Steps
1. Gather facts from deterministic sources: repository files and the JSON-fact helpers (inspect, state, changes, checkpoint status).
2. Reuse checkpointed evidence before re-reading sources.
3. Classify each finding per the evidence-labels discipline.
4. Build the repository profile per the `references/inventory-facts.md` contract: kind (primary/secondary/confidence), topology (monorepo, workspace_count), ecosystems, central frameworks, existing agent assets, context placement estimate, boilerplate assessment when it applies, tool assessment.
5. Record the labeled evidence set with per-finding confidence in state.
6. Feed the Tool / Capability Assessment: include tool/capability facts (`tools status`, `tools recommend`) in the labeled evidence set; every assessment claim cites evidence and a reason.

## Pattern & Exemplar Analysis

Evaluate this analysis when there is evidence of: multiple implementations of the same intent; boilerplate/starter/template with example content; a UI-rich application with repeated screen/component composition; or a repeated repository-specific implementation shape. Skip the deep analysis for small scripts, single-screen apps, libraries with no repeated implementation flows, and empty starters — but always evaluate applicability and persist its verdict.

Applicability must always be evaluated and persisted: whenever Pattern & Exemplar applicability has been evaluated, `repository_profile` MUST include `pattern_exemplar_analysis` with exactly one status:

- `not_applicable` — no meaningful repeated implementations, no example-rich boilerplate content, and no repeated repository-specific implementation shape. This means applicability was evaluated and the feature is not relevant to this repository — never that the analysis was skipped. Persist it compactly (`status: not_applicable`, no unnecessary lists); no large user-facing block is required.
- `assessed` — analysis completed with sufficient evidence.
- `partial` — analysis applies but evidence remains incomplete/ambiguous.

The analysis answers:
1. Are there repeated implementations of the same intent?
2. Which ones represent current architecture?
3. Which are legacy/deprecated/experimental?
4. Which examples best represent distinct use cases?
5. What stable implementation patterns repeat across them?
6. If UI-rich, what stable design/UX patterns are evidenced?
7. Is this knowledge already persisted somewhere?
8. Would indexing it reduce future exploration and inconsistency?

It produces evidence only — no artifacts. artifact-design decides CREATE/UPDATE/REUSE/NO_ACTION from this evidence.

### Canonical exemplar candidates

An example may be a canonical exemplar candidate when it is: current architecture; complete enough to learn from; represents a recurring intent; not legacy/deprecated; not a known exception. Never pick simply the largest file, the newest file, or the first matching file. Record the reason for every candidate:

```yaml
canonical_candidate:
  id: finance-dashboard
  intent:
    - KPI hierarchy
    - chart-heavy dashboard
  path: src/...
  status: current
  evidence:
    - uses current layout primitives
    - uses semantic theme tokens
    - follows current route structure
```

### Anti-examples / exclusions

Also detect, when important: legacy, deprecated, migration-only, experimental, generated, protected primitive, known exception.

```yaml
avoid:
  - path: src/.../legacy
    reason: obsolete implementation pattern
```

Future work must not copy old code accidentally just because it exists.

### UI-rich repositories: design consistency

When the repository is clearly UI-rich, also evaluate evidence on: layout/composition; spacing/density; typography hierarchy; semantic theme/token usage; responsive behavior; loading/empty/error states; interaction behavior; accessibility patterns; component composition; chart/table/form conventions when present. Do not invent a design system; record only repeated, demonstrable repository patterns. When a dimension lacks sufficient evidence, record `UNKNOWN`; never fill it with generic Tailwind, shadcn, or React knowledge.

### Context budget

Do not read complete implementations. Use progressive exploration: find candidate cluster → inspect structure/search results → choose representative candidates → read smallest useful portions → compare → expand only if evidence is insufficient. By default: 1 primary candidate + maximum 1 secondary comparison — no more than two examples total unless there is an explicit reason to read more.

### No aggressive auto-learning

Do not promote every new file to canonical. Promoting a new example requires evidence: current architecture; complete implementation; distinct or better representative intent; not experimental; not legacy. When in doubt: `NO_CHANGE` or `ASK_USER`.

## Output Contract
Return the labeled evidence set and the repository profile:

```yaml
repository_profile:
  kind:
    primary: <application | library | cli | starter | boilerplate | template | infrastructure | mixed>
    secondary: []
    confidence: <0.0-1.0>
  topology:
    monorepo: <true | false>
    workspace_count: <int>
  ecosystems: []
  central_frameworks: []
  existing_agent_assets:
    agents_md: <null | {path, lines}>
    local_skills: []
    external_skills: []
    scripts: []
  context_placement:
    always_on: []
    task_specific_candidates: []
  tool_assessment:
    ecosystem: []
    productivity: []
    providers: []
```

When kind.primary is starter/boilerplate/template, the profile MUST also carry the Boilerplate Assessment:

```yaml
boilerplate_assessment:
  extension_points: []
  editable_boundaries: []
  generated_files: []
  feature_addition_workflow:
    status: <assessed | partial | not_found>
    evidence: []
  variants:
    status: <assessed | partial | not_found>
    evidence: []
  scaffolding:
    status: <assessed | partial | not_found>
    evidence: []
  upgrade_strategy:
    status: <assessed | partial | not_found>
    evidence: []
  canonical_customization_examples: []
```

Whenever Pattern & Exemplar applicability has been evaluated, the profile MUST carry this block with exactly one status (`not_applicable` | `assessed` | `partial`):

```yaml
pattern_exemplar_analysis:
  status: <not_applicable | assessed | partial>
  repeated_intents: []
  canonical_candidates: []
  avoid_examples: []
  stable_patterns: []
  design_consistency:
    applicable: <true | false>
    observed:
      layout: []
      spacing_density: []
      typography: []
      theme_tokens: []
      responsive: []
      states: []
      interactions: []
      accessibility: []
  existing_persistence:
    canonical_catalog: <path | null>
    pattern_docs: []
  persistence_gap:
    exists: <true | false>
    reason:
```

The Boilerplate Assessment demonstrates that the evaluation occurred; it creates no artifacts. Before the audit completes, persist the repository profile to `.agent-ready/state/repository-profile.yaml` — kind.primary, kind.confidence, topology, boilerplate_assessment when it applies, and pattern_exemplar_analysis with its status whenever applicability was evaluated — and reference it in decisions.jsonl; Go fact helpers only read this file. Findings carry FACT/INFERENCE/UNKNOWN labels, every decision-relevant finding has a confidence, and every Tool / Capability Assessment claim cites evidence and a reason.
