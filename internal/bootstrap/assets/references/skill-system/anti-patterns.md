# Skill Anti-Patterns

Catalog of content and usage patterns that fail the rubric. Read this reference when authoring or reviewing a skill, and when a score comes back REVISE or REJECT: name the anti-pattern, fix it, and re-score.

## Patterns

| Anti-pattern | Symptom | Why it fails | Fix |
|---|---|---|---|
| Context dumping | Body contains the full procedure; flow reads whole files or whole repos | Violates progressive disclosure and token discipline; bloats context, hides the trigger | Keep the body minimal; move detail to `references/`; read only what the current decision needs |
| Generic advice | Instructions apply to any codebase; no repository-specific detail | Fails repository_specificity; the model cannot execute it here | Anchor every instruction to this repository, its layout, and the installed OpenCode version (minimum compatible 1.18.15) |
| Vague trigger | `description` describes the topic, not when to load | Fails discovery_description; the model loads too late or too often | State the trigger first in one quoted line, <= 250 chars |
| Procedure-less body | Rules with no order, no gates, no output contract | Fails procedural_value; the model cannot execute | Follow the required structure: activation, hard rules, gates, steps, output contract |
| Unverifiable claims | Instructions reference files, APIs, or flows that do not exist here | Fails evidence_grounding; the model acts on invented facts | Ground every claim in a harness fact, spec requirement, or scenario |
| Unvalidated frontmatter | `name` breaks the pattern or mismatches the directory; `description` too long | Fails validation; the pinned runtime may reject the skill | Check `^[a-z0-9]+(-[a-z0-9]+)*$`, directory match, description 1-1024 chars |
| Duplicate skill | Another skill already covers the work | Fails necessity; two skills race and contradict | Reuse or extend the existing skill; only new work authorizes a new skill |
| Nice-to-have skill | No evidence-backed need, "could be useful" | Fails necessity; the audit never creates artifacts without evidence | Hold the idea until evidence exists or ASK_USER fires |
| Dual copy after extraction | Same content copied to both the skill and AGENTS | Fails progressive_disclosure and context_placement; always-on cost does not shrink | Move, do not copy; leave a short router in the source |
| Global invariant hidden in a skill | A global invariant was moved into a skill | Fails context_placement; always-on guidance becomes conditional on skill load | Keep global invariants always-on; extract only task-specific procedure |
| Bloated AGENTS.md | AGENTS.md grown with task-specific procedures | Fails context_placement; every session pays for rare procedures | Move task-specific procedures to skills; AGENTS.md stays a router |
| Reference hiding an always-on rule | A critical always-on rule sits only in a reference | Fails discoverability; the rule loads too late or never | Keep critical rules always-on; references hold detail |
| Script hiding semantic decisions | A script silently encodes decisions that need judgment | Fails context_placement; semantic choices become opaque | Scripts run deterministic steps only; semantic decisions stay visible |
| Score without justification | A verdict with no per-criterion rationale | Violates the rubric contract; the verdict cannot be reviewed | Return the full score sheet: per-criterion scores, total, verdict, and grounded justification |

## Hard rules

- Never dump repository content into context.
- Never create or accept a skill below 85 on the rubric.
- Never accept "N skills generated" as success evidence; evidence is per-skill scores with justification.
- Never let a numeric PASS override a blocking concern; record the concern in state.
