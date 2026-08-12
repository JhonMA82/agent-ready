# Context Placement — screen creation

Subject: dashboard-screen-creation

Current:
  location: AGENTS.md
  loaded: every_session
  size_estimate: 13-step-procedure

Applicability:
  always_needed: false
  task_specific: true
  procedural: true
  repository_specific: true

Alternatives considered:
  - REUSE existing AGENTS procedure
  - EXTRACT_TO_SKILL add-dashboard-screen
  - COMPACT AGENTS + reference

Decision:
  REUSE
  Reason: screen creation is one of the dominant workflows in this
  boilerplate and the 13-step procedure also carries global routing
  invariants (route placement, sidebar registration, generated-file rule)
  that belong in always-on context; keep it loaded every session.

Alternative rejected:
  EXTRACT_TO_SKILL add-dashboard-screen
  Rejected because the expected permanent-context savings are small
  relative to the discoverability cost of hiding a dominant workflow
  behind an on-demand skill load.

Alternative considered:
  COMPACT AGENTS + reference
  Not applied: the procedure already routes to canonical examples and
  validation commands instead of embedding their content.

Expected effect:
  permanent_context: unchanged
  on_demand_context: unchanged
  discoverability: preserved
