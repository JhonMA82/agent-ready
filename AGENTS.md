# Agent-Ready — Project Rules

## Delivery gate: version-and-release is MANDATORY

Before executing any delivery action — commit or push of finished work, PR, tag, or GitHub release — including requests phrased as "sube los cambios", "subir a git", "tarea terminada", "versionar", "bump de version", "release", "changelog" or "tag": load the `version-and-release` skill FIRST (via the skill tool; global skill at `~/.config/opencode/skills/version-and-release/SKILL.md`) and follow its full workflow. A push-only interpretation is never enough in this repo.

Delivery convention:

- `chore(release): vX.Y.Z` commit with the CHANGELOG.md entry (Keep a Changelog + "Covered commits").
- Annotated tag `vX.Y.Z`.
- GitHub release from the changelog section.
- Version is injected by goreleaser from the tag; there is no version file to bump (changelog + tag only).

If the skill cannot be loaded, STOP and ask the user instead of pushing without it. When in doubt about the dimension, ask.

## Conventions

- Root spec docs (`OPENCODE_AGENT_READY_*.md`) are tracked and sequential: each new spec supersedes the previous one (replace, don't accumulate).
- Technical artifacts (embedded skills, tests, docs) are written in English.
