# External skill reuse — shadcn

Decision:
  REUSE_EXTERNAL_SKILL

shadcn-specific component work (add/update component, registry sync,
variant conventions) is delegated to the canonical external shadcn skill.
No local wrapper skill is generated: a wrapper would duplicate the
canonical skill and split its maintenance.

Not created:
- No generic React skill (react-best-practices is a forbidden pattern).
- No TanStack basics skill (tanstack-basics is a forbidden pattern).
- No shadcn wrapper skill.

The repository keeps only a router note ("keep components.json in sync
when adding shadcn components") plus the canonical external reference.
