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

## Execution Steps
1. Gather facts from deterministic sources: repository files and the JSON-fact helpers (inspect, state, changes, checkpoint status).
2. Reuse checkpointed evidence before re-reading sources.
3. Classify each finding per the evidence-labels discipline.
4. Build the repository profile per the `references/inventory-facts.md` contract: kind (primary/secondary/confidence), topology (monorepo, workspace_count), ecosystems, central frameworks, existing agent assets, context placement estimate, boilerplate assessment when it applies, tool assessment.
5. Record the labeled evidence set with per-finding confidence in state.
6. Feed the Tool / Capability Assessment: include tool/capability facts (`tools status`, `tools recommend`) in the labeled evidence set; every assessment claim cites evidence and a reason.

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

The Boilerplate Assessment demonstrates that the evaluation occurred; it creates no artifacts. Before the audit completes, persist the repository profile to `.agent-ready/state/repository-profile.yaml` — kind.primary, kind.confidence, topology, and boilerplate_assessment when it applies — and reference it in decisions.jsonl; Go fact helpers only read this file. Findings carry FACT/INFERENCE/UNKNOWN labels, every decision-relevant finding has a confidence, and every Tool / Capability Assessment claim cites evidence and a reason.
