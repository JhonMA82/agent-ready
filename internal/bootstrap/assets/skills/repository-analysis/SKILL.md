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
- Classify repository kind when evidence supports it: primary kind {boilerplate, starter, template, application} with secondary kinds and a confidence score (FACT when evidenced, INFERENCE otherwise).
- When kind is boilerplate/starter/template, run the boilerplate audit: extension points; what downstream users should edit and should not edit; generated files; feature addition workflow; variants/presets; scaffolding; upgrade/update strategy; canonical customization examples. The audit never creates artifacts by itself.
- Ask the boilerplate placement questions: which instructions must be always-on; which workflows should become on-demand skills; which examples should become references.

## Execution Steps
1. Gather facts from deterministic sources: repository files and the JSON-fact helpers (inspect, state, changes, checkpoint status).
2. Reuse checkpointed evidence before re-reading sources.
3. Classify each finding per the evidence-labels discipline.
4. Build the repository profile per the `references/inventory-facts.md` contract: kind (primary/secondary/confidence), ecosystems, central frameworks, existing agent assets, context placement estimate, tool assessment.
5. Record the labeled evidence set with per-finding confidence in state.
6. Feed the Tool / Capability Assessment: include tool/capability facts (`tools status`, `tools recommend`) in the labeled evidence set; every assessment claim cites evidence and a reason.

## Output Contract
Return the evidence set: findings with FACT/INFERENCE/UNKNOWN labels, the source of each fact, confidence for every decision-relevant finding, and the tool/capability evidence behind every Tool / Capability Assessment claim.
