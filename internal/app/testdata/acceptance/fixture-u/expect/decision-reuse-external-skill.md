# Decision — external canonical skill

Subject: shadcn component work guidance in AGENTS.md

Decision:
- REUSE_EXTERNAL_SKILL — the capability is already covered by the
  canonical external shadcn skill; AGENTS.md keeps only a router note
  pointing at it (refinement §41, §11 duplication order).

Evidence:
- AGENTS.md states the canonical skill covers add/update workflow,
  variant conventions, and registry sync.
- The workspace has no local shadcn skill to wrap or update.

Not created:
- No local wrapper skill is generated: a wrapper would duplicate the
  canonical skill and split its maintenance.
- No generic React skill is created for the React patterns note.

Expected effect:
  permanent_context: unchanged (router only, no duplicated content)
  duplication: avoided
  discoverability: preserved (the router names the canonical skill)
