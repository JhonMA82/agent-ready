---
name: targeted-research
description: "Trigger: closing an evidence gap in an audit. Research a concrete question about the repository, its dependencies, or the installed OpenCode runtime; record source and version; stop when answered."
---
Close evidence gaps with the smallest useful search. Load `references/search-strategies.md` before choosing a search order; label findings per repository-analysis `references/evidence-labels.md`.

## Activation Contract
Run when a decision lacks a FACT and repository evidence alone cannot close the gap.

## Hard Rules
- Research a concrete question that names the decision it supports, never a topic.
- Pin the exact version before consulting any documentation.
- Search repository first, then local docs, then version metadata, then official docs, then a specialized provider; use the broader web only if necessary.
- Record source and version with every answer; an answer without provenance is not evidence.
- External Verification Gate: before an answer embeds central, version-sensitive framework, package-manager, or toolchain knowledge, attach current official or versioned evidence tied to the applicable version, or an explicit reasoned exemption showing that no external claim is embedded; repository-to-official research precedence stays intact.
- Name the evidence or the exemption: output that embeds central framework guidance MUST state `external_verified_evidence` non-empty, or carry the explicit exemption `external_verification_not_required`.
- Stop when the question is answered; keep searching only if the answer still lacks a source.
- No search tool available never blocks: derive from local sources or stop with ASK_USER.

## Execution Steps
1. State the concrete question and the decision it supports.
2. Run the search ladder in `references/search-strategies.md` order.
3. Fix the exact version: the installed OpenCode version (minimum compatible 1.18.15) or the version the question names.
4. Record each answer with its source and version; classify FACT or INFERENCE; attach gate evidence or a reasoned exemption for version-sensitive answers.
5. Stop at the first source that answers; report the stopping condition.

## Output Contract
Return the answer with source provenance and version, or an UNKNOWN naming the evidence that would resolve it; record both in state.
