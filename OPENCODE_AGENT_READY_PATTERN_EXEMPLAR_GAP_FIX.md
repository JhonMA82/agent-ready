# Agent-Ready V1 — Parche específico: Pattern & Exemplar Preservation
## Instrucciones directas para OpenCode

> **Objetivo:** cubrir un único vacío de Agent-Ready V1: cuando un repositorio ya contiene implementaciones de ejemplo de buena calidad, Agent-Ready debe identificar cuáles son canónicas y extraer/persistir los patrones necesarios para que futuras implementaciones mantengan la misma arquitectura, estilo, diseño y UX.
>
> **No crear un subsistema nuevo. No crear una nueva skill del harness. No convertir Agent-Ready en un generador de frontend.**
>
> La solución debe reutilizar las capacidades actuales:
>
> ```text
> repository-analysis
> → artifact-design
> → existing artifact/review flow
> → incremental sync
> ```
>
> Referencia conceptual útil, NO template para copiar:
>
> ```text
> https://github.com/JhonMA82/shadcn-next-boilerplate/tree/ai-friendly
> ```
>
> Especialmente:
>
> ```text
> AGENTS.md
> docs/ai/canonical-examples.yaml
> docs/patterns/dashboard-screen.md
> ```

---

# 1. Problema concreto

Actualmente Agent-Ready puede detectar:

```text
repository understood
existing AGENTS.md
existing scripts
existing workflows
no new skill required
→ NO_ACTION
```

pero puede ignorar que el repositorio contiene múltiples implementaciones que deberían servir como **referencia persistente para trabajo futuro**.

Ejemplo:

```text
dashboard A
dashboard B
dashboard C
users screen
roles screen
```

La IA debería poder saber:

```text
qué ejemplos son canónicos
qué patrón representa cada uno
qué ejemplos evitar
qué reglas visuales/de implementación se repiten
qué debe cargar como referencia al crear algo nuevo
```

Sin esa información persistida, una nueva dashboard puede ser técnicamente correcta pero inconsistente con el resto del proyecto.

---

# 2. Principio nuevo, pequeño y general

Añadir este principio:

> **When a repository contains repeated successful implementations, Agent-Ready must evaluate whether those implementations should be indexed as canonical exemplars and whether stable recurring patterns should be persisted for future work.**

Esto aplica a cualquier stack.

Ejemplos:

```text
Frontend:
  dashboards
  forms
  tables
  detail screens

Rust/Ratatui:
  pages
  widgets
  event-handling flows

Laravel:
  controllers/actions
  jobs
  domain services

API:
  endpoints
  validation
  error handling

CLI:
  commands
  prompts
  output patterns
```

No limitarlo a UI.

---

# 3. Archivos a modificar

Modificar solo:

```text
internal/bootstrap/assets/skills/repository-analysis/SKILL.md
internal/bootstrap/assets/skills/artifact-design/SKILL.md
internal/bootstrap/assets/skills/agent-ready-orchestrator/SKILL.md
internal/bootstrap/assets/skills/incremental-evolution/SKILL.md
internal/bootstrap/content_test.go
internal/app/driven_fixtures_test.go
```

Usar el fixture TanStack/starter existente.

No crear una nueva skill.

No crear un paquete Go nuevo.

No modificar Tool Manager.

No modificar CLI.

No modificar provider logic.

---

# 4. `repository-analysis/SKILL.md`

Ruta:

```text
internal/bootstrap/assets/skills/repository-analysis/SKILL.md
```

Añadir una responsabilidad:

```text
Pattern & Exemplar Analysis
```

## Cuándo activarla

Evaluarla cuando exista evidencia de:

```text
multiple implementations of the same intent
OR
boilerplate/starter/template with example content
OR
UI-rich application with repeated screen/component composition
OR
repeated repository-specific implementation shape
```

No ejecutarla profundamente en repos donde no existan ejemplos repetidos.

Ejemplos donde NO hace falta:

```text
small script
single-screen app
library with no repeated implementation flows
empty starter
```

---

# 5. Pattern & Exemplar Analysis

El análisis debe responder:

```text
1. Are there repeated implementations of the same intent?
2. Which ones represent current architecture?
3. Which are legacy/deprecated/experimental?
4. Which examples best represent distinct use cases?
5. What stable implementation patterns repeat across them?
6. If UI-rich, what stable design/UX patterns are evidenced?
7. Is this knowledge already persisted somewhere?
8. Would indexing it reduce future exploration and inconsistency?
```

No crear artifacts todavía.

`repository-analysis` solo produce evidencia.

---

# 6. Canonical exemplar candidate

Un ejemplo puede ser candidato canónico si:

```text
current architecture
+
complete enough to learn from
+
represents a recurring intent
+
not legacy/deprecated
+
not known exception
```

No elegir simplemente:

```text
largest file
newest file
first matching file
```

Registrar razón.

Ejemplo:

```yaml
canonical_candidate:
  id: finance-dashboard
  intent:
    - KPI hierarchy
    - chart-heavy dashboard
  path: src/...
  status: current
  evidence:
    - uses current layout primitives
    - uses semantic theme tokens
    - follows current route structure
```

---

# 7. Anti-examples / exclusions

También detectar cuando sea importante:

```text
legacy
deprecated
migration-only
experimental
generated
protected primitive
known exception
```

Ejemplo:

```yaml
avoid:
  - path: src/.../legacy
    reason: obsolete implementation pattern
```

La IA futura no debe copiar accidentalmente código viejo solo porque existe.

---

# 8. UI-rich repositories: Design Consistency

Cuando el repositorio sea claramente UI-rich, evaluar además evidencia sobre:

```text
layout/composition
spacing/density
typography hierarchy
semantic theme/token usage
responsive behavior
loading/empty/error states
interaction behavior
accessibility patterns
component composition
chart/table/form conventions when present
```

No inventar un design system.

Solo registrar patrones repetidos y demostrables en el repo.

Si una dimensión no tiene evidencia suficiente:

```text
UNKNOWN
```

No completar con conocimiento genérico de Tailwind, shadcn, React, etc.

---

# 9. Output de repository-analysis

Extender el repository profile con un bloque opcional:

```yaml
pattern_exemplar_analysis:

  status: <not_applicable | assessed | partial>

  repeated_intents: []

  canonical_candidates: []

  avoid_examples: []

  stable_patterns: []

  design_consistency:
    applicable: <true | false>
    observed:
      layout: []
      spacing_density: []
      typography: []
      theme_tokens: []
      responsive: []
      states: []
      interactions: []
      accessibility: []

  existing_persistence:
    canonical_catalog: <path | null>
    pattern_docs: []

  persistence_gap:
    exists: <true | false>
    reason:
```

No es necesario poblar listas vacías enormes en la salida visible.

---

# 10. Context budget durante el análisis

No leer todas las implementaciones completas.

Usar progressive exploration:

```text
find candidate cluster
→ inspect structure/search results
→ choose representative candidates
→ read smallest useful portions
→ compare
→ expand only if evidence is insufficient
```

Por defecto:

```text
1 primary candidate
+
maximum 1 secondary comparison
```

Solo leer más si hay una razón explícita.

---

# 11. `artifact-design/SKILL.md`

Ruta:

```text
internal/bootstrap/assets/skills/artifact-design/SKILL.md
```

Añadir dos tipos de artefacto posibles, NO obligatorios:

```text
canonical exemplar catalog
pattern reference
```

No crear un nuevo verdict.

Usar:

```text
CREATE
UPDATE
REUSE
NO_ACTION
```

---

# 12. Cuándo crear un canonical exemplar catalog

`CREATE` solo si:

```text
multiple useful examples exist
AND
they represent distinct recurring intents
AND
there is no adequate existing catalog
AND
future work would otherwise need repeated exploration
```

Ubicación recomendada:

```text
docs/ai/canonical-examples.yaml
```

solo si encaja con la estructura del proyecto.

Si el repo ya tiene una ubicación equivalente, reutilizarla.

No imponer el path universalmente.

---

# 13. Contenido mínimo del canonical catalog

Debe ser compacto.

Ejemplo:

```yaml
schemaVersion: 1

selectionRules:
  - Select the closest current example by intent.
  - Inspect no more than two examples unless more evidence is required.
  - Record intentional deviations.
  - Do not use legacy/deprecated entries as new-work references.

examples:

  operational-dashboard:
    path: src/...
    useFor:
      - operational metrics
      - status composition
    status: current

  entity-management:
    path: src/...
    useFor:
      - tabular entity management
      - filtering and row actions
    status: current

avoid:
  - path: src/.../legacy
    reason: obsolete architecture
```

No copiar el catálogo de otro proyecto.

Derivarlo del repo actual.

---

# 14. Cuándo crear un pattern reference

Crear un pattern reference solo si:

```text
a stable non-trivial pattern repeats
AND
the pattern is repository-specific
AND
future work would benefit from explicit guidance
AND
the information is not already adequately documented
```

Posible ubicación:

```text
docs/ai/patterns/<intent>.md
```

o:

```text
docs/patterns/<intent>.md
```

según estructura existente.

No generar patrones por framework:

```text
react-pattern.md
tailwind-pattern.md
rust-pattern.md
```

Generar por intención del proyecto:

```text
dashboard-screen.md
entity-management.md
ratatui-page.md
domain-service.md
```

solo cuando la evidencia lo justifique.

---

# 15. Contenido mínimo de un pattern reference

Debe describir únicamente lo inferido del repo:

```text
Use when
Canonical examples
Expected structure
Repository-specific invariants
Design/interaction expectations when applicable
Required states
Validation
Known deviations/avoidances
```

No debe convertirse en tutorial del framework.

---

# 16. Design consistency no significa copiar estilos literalmente

Regla:

> Derive from existing semantic primitives and composition patterns, not from literal duplication of classes or component markup.

Correcto:

```text
Use existing semantic theme tokens.
Use the existing dashboard shell and spacing rhythm.
Follow the canonical responsive collapse pattern.
```

Incorrecto:

```text
copy these 43 Tailwind classes from dashboard X
```

---

# 17. AGENTS.md como router

Si se crea un canonical catalog o pattern reference:

No copiar el contenido completo a `AGENTS.md`.

Como máximo añadir una referencia compacta si `AGENTS.md` es el router adecuado:

```text
For new UI/features, select the closest canonical example and applicable project pattern before implementation.
```

El detalle permanece on-demand.

---

# 18. Skill vs pattern

No crear una skill solo porque se creó un pattern.

```text
pattern
→ how this kind of implementation is shaped in the repo

skill
→ how to execute a repeatable non-trivial workflow
```

Ejemplo:

```text
docs/ai/patterns/dashboard-screen.md
```

puede existir sin:

```text
.opencode/skills/add-dashboard-screen/
```

La skill solo se justifica si cumple la rubric existente.

---

# 19. `agent-ready-orchestrator/SKILL.md`

Ruta:

```text
internal/bootstrap/assets/skills/agent-ready-orchestrator/SKILL.md
```

No agregar otra fase grande.

Añadir únicamente:

> When repeated implementations or example-rich boilerplates are detected, require repository-analysis to evaluate Pattern & Exemplar coverage before final artifact decisions.

Añadir a la checklist previa a `NO_ACTION`:

```text
pattern/exemplar coverage evaluated when applicable
```

Así un repo con muchos ejemplos no termina `NO_ACTION` solo porque ya tenga AGENTS.md.

---

# 20. Output visible

No añadir una sección obligatoria grande.

Cuando aplique, resumir dentro de Repository/Artifact Decisions:

```text
Patterns & Exemplars
  repeated dashboard patterns detected
  canonical examples are not indexed
  CREATE canonical-examples
  CREATE dashboard-screen pattern
```

O:

```text
Patterns & Exemplars
  already covered by docs/ai/canonical-examples.yaml
  REUSE
```

---

# 21. `incremental-evolution/SKILL.md`

Ruta:

```text
internal/bootstrap/assets/skills/incremental-evolution/SKILL.md
```

Añadir solo esta regla:

> If a changed path belongs to a recorded canonical exemplar, removes one, changes its status, or introduces a clearly new repeated implementation intent, reassess only the affected canonical catalog/pattern entries.

No re-auditar todo el repo.

Ejemplo:

```text
new dashboard added
→ compare with existing dashboard intents
→ if same intent: NO_CHANGE
→ if new canonical intent: UPDATE catalog
```

---

# 22. No hacer auto-learning agresivo

No convertir cualquier archivo nuevo en canonical.

Para promover un nuevo ejemplo debe haber evidencia:

```text
current architecture
complete implementation
distinct or better representative intent
not experimental
not legacy
```

Si hay duda:

```text
NO_CHANGE
```

o:

```text
ASK_USER
```

---

# 23. `content_test.go`

Ruta:

```text
internal/bootstrap/content_test.go
```

Añadir assertions mínimas para:

```text
Pattern & Exemplar
canonical exemplar
legacy/deprecated
no more than two examples
```

y comprobar que el orchestrator incluye:

```text
pattern/exemplar coverage evaluated when applicable
```

No crear tests de texto gigantes.

---

# 24. `driven_fixtures_test.go`

Ruta:

```text
internal/app/driven_fixtures_test.go
```

Reutilizar el fixture TanStack/starter existente.

No crear otro fixture grande.

Añadir al oracle:

```text
pattern/exemplar analysis occurred
```

Debe permitir:

```text
CREATE canonical catalog/pattern
```

o:

```text
REUSE existing catalog/pattern
```

según el fixture.

---

# 25. Ajuste mínimo al fixture TanStack

Añadir solo lo necesario para tener:

```text
2-3 dashboard/example implementations
1 legacy/deprecated example
shared theme/layout primitives
```

No copiar el repo real completo.

Debe permitir comprobar:

```text
canonical candidates identified
legacy example excluded
design consistency evidence extracted
maximum-two-example context discipline
```

---

# 26. Acceptance — missing catalog

Fixture contiene:

```text
dashboard/default
dashboard/finance
dashboard/operations
dashboard/legacy
```

sin canonical catalog.

Esperado:

```text
Pattern & Exemplar Analysis: assessed

canonical candidates:
  current dashboards

avoid:
  legacy

Artifact Decision:
  CREATE canonical exemplar catalog
```

Un pattern reference solo se crea si también existe evidencia suficiente de patrón estable no documentado.

---

# 27. Acceptance — already prepared repo

Mismo fixture, pero ya contiene:

```text
docs/ai/canonical-examples.yaml
docs/ai/patterns/dashboard-screen.md
```

y son actuales/correctos.

Esperado:

```text
REUSE
```

No duplicados.

`NO_ACTION` puede ser correcto.

---

# 28. Acceptance — no repeated examples

Repo pequeño sin repeated implementations.

Esperado:

```text
pattern_exemplar_analysis:
  status: not_applicable
```

No crear nada.

---

# 29. Acceptance — design evidence

En fixture UI-rich, detectar evidencia real de al menos algunas dimensiones:

```text
semantic tokens
layout composition
responsive behavior
states
```

No es obligatorio llenar todas.

No aceptar afirmaciones genéricas sin evidence/paths.

---

# 30. Canonical context budget

Cuando futuras instructions/artifacts usen el catálogo:

```text
select closest canonical example
+
at most one secondary example
```

por defecto.

No ordenar leer todo el código de todos los ejemplos.

El catálogo sí puede leerse completo porque debe ser pequeño.

---

# 31. Resultado razonable para `tanstack-shadcn-admin-dashboard`

```text
Repository
  starter/template
  TanStack Start + React + shadcn

Patterns & Exemplars
  multiple dashboard and management-screen implementations detected
  current examples are not semantically indexed
  legacy/example exclusions identified

Artifact Decisions
  CREATE docs/ai/canonical-examples.yaml
    reason: reduce repeated exploration and preserve implementation/design consistency

  CREATE docs/ai/patterns/dashboard-screen.md
    only if stable project-specific dashboard composition is evidenced

  NO_ACTION generic React/TanStack/shadcn skills

Context strategy
  choose closest canonical example
  maximum one secondary comparison

Tools
  unchanged

Checkpoint
  complete
```

No obligar exactamente ese resultado.

La evidencia decide.

---

# 32. Qué tomar de `shadcn-next-boilerplate:ai-friendly`

Tomar únicamente:

```text
AGENTS as compact router
canonical example catalog
patterns by repository intent
exclude legacy examples
select closest exemplar
maximum two exemplars
record intentional deviations
```

No copiar automáticamente:

```text
templates/
create-dashboard generator
create-crud generator
create-feature generator
Next.js-specific paths
shadcn-specific rules
```

Eso pertenece a ese proyecto.

---

# 33. No añadir

Prohibido crear:

```text
new harness skill: design-pattern-analysis
new Go package: design/
new database
new indexer
new embeddings
new vector store
new provider
new MCP
new command
new agent
new template library
```

No hace falta.

---

# 34. No cambiar Tool Manager

Este cambio no está relacionado con:

```text
RTK
Context7
Serena
Semble
CodeGraph
Headroom
package managers
install recipes
```

No modificar esas áreas.

---

# 35. Definition of Done

- [ ] repository-analysis evalúa Pattern & Exemplar cuando aplica.
- [ ] no lo ejecuta profundamente cuando no aplica.
- [ ] identifica canonical candidates con evidencia.
- [ ] identifica legacy/deprecated/avoid cuando aplica.
- [ ] UI-rich repos evalúan design consistency desde evidencia.
- [ ] artifact-design puede proponer canonical catalog.
- [ ] artifact-design puede proponer pattern reference.
- [ ] no se crean skills genéricas.
- [ ] AGENTS permanece compacto/router.
- [ ] canonical context budget limita a máximo dos ejemplos por defecto.
- [ ] orchestrator no permite `NO_ACTION` sin evaluar esta dimensión cuando aplica.
- [ ] incremental sync reevalúa solo entries afectadas.
- [ ] fixture TanStack cubre el comportamiento.
- [ ] repo ya preparado produce REUSE/NO_ACTION.
- [ ] repo sin ejemplos repetidos produce not_applicable.
- [ ] no se añadió ningún subsistema nuevo.
- [ ] `go test ./...` pasa.
- [ ] diff permanece pequeño y localizado.

---

# 36. Regla final

Agent-Ready debe preparar el repositorio para que el código futuro no sea solo correcto.

Debe conseguir que:

> **una nueva implementación parezca pertenecer al mismo proyecto.**

La forma mínima de lograrlo es:

```text
existing good implementations
→ canonical exemplars
→ stable project patterns
→ compact persistent guidance
→ load only the closest examples when needed
```

No hace falta más arquitectura para V1.
