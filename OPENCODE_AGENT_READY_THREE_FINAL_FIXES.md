# Agent-Ready V1 — Correcciones finales mínimas
## Instrucciones directas para OpenCode

> **Objetivo:** aplicar únicamente tres correcciones pequeñas detectadas después de verificar los últimos cambios de `JhonMA82/agent-ready`.
>
> **No ampliar alcance.**
>
> **No añadir features nuevas.**
>
> **No tocar arquitectura, Tool Manager, providers, CLI, state, checkpoints, ownership ni skills adicionales.**

---

# 1. Alcance exacto

Modificar únicamente estos archivos:

```text
internal/bootstrap/assets/skills/agent-ready-orchestrator/references/audit-flow.md

internal/bootstrap/assets/skills/repository-analysis/SKILL.md

internal/app/driven_fixtures_test.go
```

No modificar otros archivos salvo que un test existente falle directamente por una de estas tres correcciones.

Antes de aplicar, mostrar:

```text
files planned: 3
new production packages: 0
new agents: 0
new skills: 0
new commands: 0
new providers: 0
```

Si la solución requiere cambios significativamente mayores:

```text
STOP
```

y explicar la causa.

---

# 2. Corrección 1 — `audit-flow.md`

Ruta:

```text
internal/bootstrap/assets/skills/agent-ready-orchestrator/references/audit-flow.md
```

## Problema

La tabla/definición de decisiones todavía conserva semántica antigua equivalente a:

```text
REUSE
An existing skill covers the need; create nothing.

NO_ACTION
No skill earns >= 85 on evidence; create no artifacts.
```

Estas definiciones ya contradicen el comportamiento actual de `artifact-design`.

El threshold `>=85` controla únicamente la creación de una nueva skill.

Además, `REUSE` ya no significa únicamente que exista una skill: puede tratarse de guidance, docs, pattern, canonical catalog, script u otro artefacto existente.

## Cambio requerido

Cambiar únicamente esas definiciones.

### `REUSE`

Debe expresar:

```text
REUSE

Existing guidance, artifact, skill, pattern, reference, or deterministic helper
already covers the need
AND
its current context placement is appropriate.
```

Versión corta válida:

```text
REUSE | Existing guidance/artifact covers the need and its context placement is appropriate.
```

### `NO_ACTION`

Debe expresar:

```text
NO_ACTION

No justified artifact change remains
AND
no context-placement transformation is justified
AND
all applicable assessments are complete.
```

Los assessments aplicables incluyen, cuando corresponda:

```text
Context Placement
Tool / Capability Assessment
Pattern & Exemplar Analysis
Boilerplate Assessment
```

Versión corta válida:

```text
NO_ACTION | No justified artifact or placement change remains after all applicable assessments.
```

## No hacer

No reescribir el documento.

No cambiar fases.

No cambiar nombres de stages.

No cambiar threshold/rubric.

Solo corregir las definiciones stale.

---

# 3. Corrección 2 — `repository-analysis/SKILL.md`

Ruta:

```text
internal/bootstrap/assets/skills/repository-analysis/SKILL.md
```

## Problema

`pattern_exemplar_analysis` admite:

```text
not_applicable
assessed
partial
```

pero el texto actual puede interpretarse como que el bloque solo debe persistirse cuando el análisis resulta:

```text
assessed
partial
```

Mientras tanto, los tests esperan explícitamente que un repo donde no aplica registre:

```yaml
pattern_exemplar_analysis:
  status: not_applicable
```

Esto crea una contradicción entre skill y fixture.

## Decisión

Adoptar como contrato oficial:

> Applicability must always be evaluated and persisted.

## Cambio requerido

Añadir una regla explícita equivalente a:

```text
When Pattern & Exemplar applicability has been evaluated,
repository_profile MUST include pattern_exemplar_analysis with exactly one status:

- not_applicable
- assessed
- partial
```

### `not_applicable`

Usar cuando:

```text
no meaningful repeated implementations
no example-rich boilerplate content
no repeated repository-specific implementation shape
```

Ejemplo:

```yaml
pattern_exemplar_analysis:
  status: not_applicable
```

No llenar listas innecesarias.

### `assessed`

Usar cuando:

```text
analysis completed with sufficient evidence
```

### `partial`

Usar cuando:

```text
analysis applies
but evidence remains incomplete/ambiguous
```

## Regla importante

`not_applicable` NO significa:

```text
analysis skipped
```

Significa:

```text
applicability evaluated
→ feature not relevant to this repository
```

Esto mejora trazabilidad y evita que `NO_ACTION` oculte una dimensión no evaluada.

## Output compacto

No obligar a mostrar un bloque grande al usuario cuando sea `not_applicable`.

Puede persistirse en state y mostrarse solo de forma compacta.

---

# 4. Corrección 3 — `driven_fixtures_test.go`

Ruta:

```text
internal/app/driven_fixtures_test.go
```

## Problema

El fixture TanStack actualmente contiene una prohibición equivalente a:

```go
rejectSkillPaths(t, after, []string{"dashboard", "pattern"})
```

Esto es demasiado restrictivo.

La intención correcta era evitar:

```text
generic dashboard skill
generic pattern skill
```

pero la implementación termina prohibiendo también una skill válida y específica como:

```text
add-dashboard-screen
```

aunque pudiera estar plenamente justificada por:

```text
repository-specific workflow
routing
navigation
permissions
state
validation
canonical examples
```

Esto empuja el sistema hacia `under-generation`.

## Cambio requerido

Eliminar la prohibición genérica sobre:

```text
dashboard
```

No rechazar una skill solo porque su path/nombre contenga esa palabra.

## Mantener prohibiciones útiles

Seguir rechazando skills claramente genéricas o duplicativas.

Ejemplos de targets válidos para rechazo:

```text
react
react-best-practices
tanstack
tanstack-best-practices
shadcn
shadcn-best-practices
generic-pattern
pattern-best-practices
```

Ajustar al mecanismo real del test sin introducir una lista enorme.

## Regla del fixture

El fixture debe validar:

> Pattern existence does not automatically justify a skill.

Pero también:

> Pattern existence does not automatically forbid a repository-specific skill.

El criterio correcto debe seguir siendo la rubric actual de skill quality/necessity.

---

# 5. Comportamiento esperado del fixture TanStack

Después de la corrección, estos resultados deben ser válidos:

## Caso A — no skill necesaria

```text
Pattern & Exemplar Analysis
  canonical examples identified

Artifacts
  CREATE/REUSE canonical catalog
  CREATE/REUSE dashboard pattern

Skills
  NO_ACTION
```

Válido si no hay workflow procedural suficiente.

## Caso B — skill específica justificada

```text
Patterns
  dashboard-screen pattern exists

Skill
  add-dashboard-screen

Reason
  repository-specific multi-step workflow
  navigation/routing/state/validation conventions
  canonical exemplar selection
```

También válido si supera la rubric.

## Casos inválidos

```text
react-best-practices
tanstack-guide
shadcn-best-practices
dashboard-best-practices
generic-pattern
```

---

# 6. No modificar Pattern & Exemplar architecture

No tocar:

```text
repository profile schema general
artifact graph
incremental evolution
canonical catalog format
context budget
two-example rule
design consistency analysis
legacy/deprecated detection
```

Esas partes ya están correctamente implementadas.

---

# 7. No modificar tools/providers

Este parche no tiene relación con:

```text
RTK
Context7
Serena
Semble
CodeGraph
Headroom
rg
fd
jq
ast-grep
package managers
installation recipes
```

No tocar esas áreas.

---

# 8. Tests específicos

Después del cambio ejecutar:

```bash
go test ./...
```

y los acceptance/driven tests existentes.

## Test 1 — `audit-flow`

Añadir o mantener content assertion que compruebe que:

```text
REUSE
```

ya no se define exclusivamente como:

```text
existing skill
```

y:

```text
NO_ACTION
```

ya no depende únicamente de:

```text
skill score >= 85
```

No crear un parser nuevo para Markdown.

Un assertion textual pequeño es suficiente.

## Test 2 — `not_applicable`

Fixture de repo pequeño debe persistir:

```yaml
pattern_exemplar_analysis:
  status: not_applicable
```

y pasar.

Debe quedar claro que applicability fue evaluada.

## Test 3 — TanStack skill policy

El fixture TanStack no debe fallar simplemente por encontrar una skill con:

```text
dashboard
```

en su path.

Sí debe seguir fallando ante skills genéricas prohibidas.

---

# 9. Regression manual recomendada

Después de tests, ejecutar Agent-Ready nuevamente sobre:

```text
https://github.com/arhamkhnz/tanstack-shadcn-admin-dashboard
```

Resultado esperado:

```text
Repository
  starter/template/boilerplate

Pattern & Exemplar Analysis
  assessed

Canonical examples
  evaluated

Artifact Decisions
  evidence-driven

Skills
  may be NO_ACTION
  OR
  may include a repository-specific dashboard workflow skill if justified

Tool Assessment
  unchanged

Checkpoint
  complete
```

No forzar creación de skill.

No forzar `NO_ACTION`.

La evidencia decide.

---

# 10. Acceptance criteria

El parche está terminado cuando:

- [ ] `audit-flow.md` define `REUSE` con coverage + placement.
- [ ] `audit-flow.md` define `NO_ACTION` sin depender del score de skill.
- [ ] `repository-analysis/SKILL.md` obliga a persistir `pattern_exemplar_analysis.status`.
- [ ] `not_applicable` significa “evaluado y no aplica”.
- [ ] el fixture TanStack ya no prohíbe indiscriminadamente cualquier dashboard skill.
- [ ] skills genéricas siguen siendo rechazadas.
- [ ] la rubric existente sigue controlando si una skill específica es necesaria.
- [ ] `go test ./...` pasa.
- [ ] no se añadió arquitectura.
- [ ] no se tocaron tools/providers.
- [ ] diff permanece pequeño.

---

# 11. Restricción final

No usar este cambio para:

```text
refactorizar fixtures
renombrar packages
reestructurar skills
crear nuevos schemas
crear nuevos helpers
cambiar state format
añadir más tests no relacionados
```

El objetivo es corregir únicamente tres inconsistencias pequeñas.

---

# 12. Resultado final deseado

Después de este parche, Agent-Ready debe mantener estas tres propiedades:

```text
1. NO_ACTION significa realmente que no queda ningún cambio justificado.

2. not_applicable demuestra que Pattern & Exemplar fue evaluado,
   no que fue omitido.

3. El sistema evita skills genéricas sin bloquear skills
   repository-specific legítimas.
```

Con estas tres correcciones, no ampliar más esta parte de V1 salvo que una nueva prueba real revele otro fallo concreto.
