---
name: dependency-bump-triage
description: "Trigger: a PR bumps a dependency in go.mod. Verify the bump builds and passes focused tests before merge."
---
Gate dependency bumps. Read `../../references/skill-system/skill-quality-rubric.md` when self-scoring before handoff.

## Activation Contract
Run when a PR changes go.mod or go.sum and the change is a version bump.

## Hard Rules
- Never approve a bump without `go build ./...` and `go test ./...` passing on the bump commit.
- Never rely on CI alone; run the focused commands on the exact commit.

## Execution Steps
1. Read the diff: module, old and new versions, indirect changes.
2. Run `go build ./...`; a failure blocks approval.
3. Run `go test ./...`; record failing tests with output.
4. Check the upstream changelog between versions for breaking changes.
5. Return the verdict with the commands run and their exit codes.

## Output Contract
One-line verdict (approve / revert / hold) plus the commands run and exit codes.
