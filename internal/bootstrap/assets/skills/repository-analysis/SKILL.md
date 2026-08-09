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

## Execution Steps
1. Gather facts from deterministic sources: repository files and the JSON-fact helpers (inspect, state, changes, checkpoint status).
2. Reuse checkpointed evidence before re-reading sources.
3. Classify each finding per the evidence-labels discipline.
4. Record the labeled evidence set with per-finding confidence in state.

## Output Contract
Return the evidence set: findings with FACT/INFERENCE/UNKNOWN labels, the source of each fact, and confidence for every decision-relevant finding.
