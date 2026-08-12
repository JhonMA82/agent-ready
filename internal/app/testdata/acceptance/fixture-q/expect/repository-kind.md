# Repository kind

repository_kind:
  primary: boilerplate
  secondary:
    - application
  confidence: 0.9

Evidence:
- extension points: add route, add feature, add shadcn component, generate preset
- generated files: route manifests under generated/ (regenerated via
  npm run generate-routes)
- scaffolding scripts: scripts/generate-routes
- template markers: components.json (shadcn config), src/routes examples,
  "What downstream users should edit" section in AGENTS.md
- dependency signature: @tanstack/react-router + @tanstack/react-start +
  react + shadcn-adjacent packages (class-variance-authority,
  tailwind-merge)

Classification: starter/boilerplate is supported by the extension model
(extension points + generated files + scaffolding), not by file counts.
