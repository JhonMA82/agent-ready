# Skill Authoring Guide

Procedures for authoring a harness skill. Read this guide when authoring, revising, or extending a skill in this repository. The skill-quality-rubric.md defines acceptance; this guide defines how to write so a score is achievable.

## A skill is a runtime contract

A skill is an instruction contract for the model, not human documentation. Every line must be executable advice: what to load, when, what to decide, what to produce. History, motivation, and tutorial background do not belong in a skill.

## Required structure

Order sections as follows; drop a section only when it is truly irrelevant:

1. Frontmatter — discovery metadata.
2. Activation Contract — exact situations that load the skill.
3. Hard Rules — constraints the model MUST NOT violate.
4. Decision Gates — short tables for branching choices.
5. Execution Steps — ordered operational workflow.
6. Output Contract — required final format.
7. References — local files only; supporting detail lives there, not in the body.

## Frontmatter rules

- `name` matches `^[a-z0-9]+(-[a-z0-9]+)*$` and equals the containing directory name.
- `description` is one physical line, quoted, YAML-safe, <= 250 chars, with trigger words first.
- No `Keywords` section; discovery uses the description.

## Body budget

- Target 180-450 tokens; recommended maximum 700; hard maximum 1000.
- Move examples, schemas, and edge cases into `references/`; never exceed the hard maximum by padding the body.

## Progressive disclosure

- The body states the trigger and the minimal instruction set; it never dumps the full procedure inline.
- Deeper content loads only when needed: a reference is read the first time the decision it supports arises, not earlier.
- The body names each reference with the decision it resolves, so the model knows what to load when.

## Writing rules

DO: write imperative instructions ("Load X", "Check Y", "Return Z"); lead with the activation trigger and hard constraints; use compact tables for decision gates; keep examples minimal and executable; link local files for detail.

DON'T: explain history or motivation; duplicate long docs inside the body; add advice the model cannot execute; use external URLs as primary references; bury critical rules below examples.

## Verification before handoff

Before a candidate skill is submitted: confirm frontmatter validity (name pattern, directory match, description length), confirm every reference path resolves, and self-score the skill against skill-quality-rubric.md. A skill below 85 is returned as REVISE with the failing criteria named.
