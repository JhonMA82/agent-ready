# Progressive Disclosure

Convention that keeps skill bodies minimal and loads deeper content only when needed. Read this reference when authoring a skill body or when deciding whether content belongs inline or in a reference.

## Rule

- The `description` states the trigger so the model knows when to load the skill.
- The body states the minimal instruction set: what to do, and which reference resolves which decision.
- References load on demand: a reference is read the first time the decision it supports arises, never earlier.
- No skill dumps its full procedure inline; no skill dumps repository content into context.

## When to read a reference

| Situation | Action |
|---|---|
| A rubric decision is needed | Read skill-quality-rubric.md now, not at skill load. |
| A branching choice needs detail | Read the reference that owns that decision. |
| The body names a reference for a decision that has not arisen | Do not read it yet. |
| The flow is resolved | Stop loading; reuse checkpointed evidence instead of re-reading inputs. |

## What belongs where

| In the body | In references |
|---|---|
| Trigger, when-to-read, hard rules | Procedures, edge cases, schemas, examples |
| Decision gates the flow hits every run | Detail needed only in some runs |
| Output contract | Background the model must not hold unless deciding |

## Token discipline

- Load the smallest useful context, never whole files or whole repos.
- Stop when resolved; do not continue collecting evidence for answered questions.
- Reuse checkpointed evidence before re-reading sources.
- A skill that grows beyond the body budget signals undisclosed content: move it out, do not extend the body.

## Failure modes

- Body first, references empty: the skill dumps everything inline and defeats on-demand loading.
- Trigger missing: the model cannot know when to load the skill.
- Reference never named in the body: the model cannot know the reference exists.
