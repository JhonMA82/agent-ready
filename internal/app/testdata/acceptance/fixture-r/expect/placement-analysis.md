# Placement analysis — long AGENTS.md

AGENTS.md is 500 lines (50 global / 150 migration / 100 release / 200
examples and edge cases) and loads every session. The audit splits it by
placement, not by topic alone:

Outcomes:
- COMPACT AGENTS — keep only the always-on global guidance: package
  manager, core commands, workspace structure, critical constraints,
  forbidden operations, docs index, skills router.
- EXTRACT migration workflow to a skill — task-specific, procedural,
  repeatable, repository-specific; 150 lines never needed by most tasks.
- EXTRACT release workflow to a skill — task-specific, procedural,
  repeatable, repository-specific; 100 lines needed only during releases.
- MOVE examples to references — 200 lines of edge cases belong in
  docs/edge-cases.md and load on demand, not every session.

Router preservation (placement_change contract):
- AGENTS.md keeps a short router in place of the removed sections:
  "Use the db-migration skill for schema changes; use the
  deploy-runbook skill for releases; see docs/edge-cases.md for edge
  cases." The router text preserves discoverability of the moved content.

Not accepted:
- REUSE without analysis is not accepted just because all information
  already exists (refinement §38): placement, not coverage, decides.

Expected effect:
  permanent_context: decrease
  on_demand_context: increase_only_when_relevant
  discoverability: preserved
