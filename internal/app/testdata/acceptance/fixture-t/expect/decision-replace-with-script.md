# Decision — deterministic procedure

Subject: generate-routes / validate-presets procedures in AGENTS.md

Decision:
- REPLACE_WITH_SCRIPT — both procedures are deterministic: fixed steps,
  no semantic judgment, mechanically checkable results (refinement §6).
- COMPACT the AGENTS.md section to the router: "Run
  `npm run generate-routes` to regenerate routes; run
  `npm run validate-presets` to validate presets."

Evidence:
- The equivalent deterministic scripts already exist in the repository
  (scripts/generate-routes, scripts/validate-presets).
- The procedures and the scripts describe the same steps; keeping both
  in sync is pure duplication (refinement §11 duplication order).

Not created:
- No skill explaining the deterministic commands: a safe, validatable
  script replaces instruction-based repetition (refinement §6, §40).

Expected effect:
  permanent_context: decrease
  duplication: removed
  discoverability: preserved (router keeps the commands visible)
