# Agent-Ready V1 — Parche específico de Output Contract y Acceptance Tests
## Instrucciones para OpenCode

> **Objetivo:** corregir únicamente el problema restante detectado en la salida real de `/agent-ready` sin cambiar arquitectura, Tool Manager, providers, CLI, state model ni diseño general de V1.
>
> **Problema concreto:** los contratos de las skills ya conocen `Repository Profile`, `Context Placement` y `Boilerplate Assessment`, pero los acceptance tests actuales pueden dar PASS aunque la salida visible omita esas secciones porque mezclan texto mostrado al usuario con contenido persistido en state.
>
> **Alcance:** parche pequeño sobre 5 archivos.
>
> **Regla principal:** NO hacer refactors generales ni introducir nuevas abstracciones.

---

# 1. Archivos a modificar

Modificar únicamente:

```text
internal/app/driven_audit_test.go
internal/app/driven_fixtures_test.go
internal/bootstrap/assets/skills/agent-ready-orchestrator/SKILL.md
internal/bootstrap/assets/skills/artifact-design/SKILL.md
internal/bootstrap/assets/skills/agent-ready-orchestrator/references/audit-flow.md
```

No modificar otros archivos salvo que un test existente falle por una dependencia directa y el cambio sea mínimo.

---

# 2. No tocar

No modificar:

```text
internal/tools/
internal/ecosystem/
internal/state/
internal/checkpoint/
internal/ownership/
internal/bootstrap/assets/skills/skill-creator/
internal/bootstrap/assets/skills/skill-reviewer/
internal/bootstrap/assets/skills/repository-analysis/
.opencode/
cmd/
CLI architecture
provider catalog
RTK logic
installation recipes
```

No añadir:

```text
nuevos agents
nuevas skills
nuevos slash commands
MCP
TUI
daemon
database
new registry
new orchestration layer
```

---

# 3. `internal/app/driven_audit_test.go`

## Problema

Actualmente el test combina:

```text
visible OpenCode output
+
persisted files/state
```

y busca términos como:

```text
repository
context placement
checkpoint
RTK
```

sobre ese corpus combinado.

Esto permite falsos positivos: la respuesta visible puede omitir `Repository` y `Context Placement`, pero el test pasa porque esas palabras existen en state.

## Cambio requerido

Separar:

```go
visibleOutput
persistedState
```

Conceptualmente:

```go
visible := events.text.String()

assertVisibleAuditOutput(t, visible)
assertPersistedAuditState(t, repo)
```

No es obligatorio usar exactamente esos nombres.

### `assertVisibleAuditOutput`

Debe validar SOLO texto visible enviado por OpenCode.

Exigir:

```text
Repository
Context Placement
Artifact / Artifact Decisions
Tool / Capability Assessment
Checkpoint
```

Debe comprobar además:

```text
ecosystem
productivity
provider
RTK
```

No exigir verdict específico para RTK.

### `assertPersistedAuditState`

Validar por separado:

```text
repository profile
decisions/provenance
checkpoint
context placement decision when applicable
boilerplate assessment when applicable
```

Los archivos persistidos NO pueden hacer pasar el test de output visible.

---

# 4. Endurecer Context Placement assertion

## Problema

No aceptar que palabras genéricas como:

```text
reuse
no_action
placement
verdict
```

sean evidencia suficiente.

Un simple:

```json
{"decision":"NO_ACTION"}
```

NO demuestra Context Placement.

## Cambio requerido

Validar estructuralmente usando el state actual, preferentemente `decisions.jsonl`.

Debe existir un registro identificable como:

```text
stage == context_placement
```

o:

```text
type == context_placement
```

o equivalente real ya usado por el proyecto.

Debe contener como mínimo:

```text
subject
decision/verdict
reason or evidence
```

Ejemplo válido:

```json
{
  "stage": "context_placement",
  "subject": "screen-creation",
  "decision": "REUSE",
  "reason": "..."
}
```

No crear formato nuevo si ya existe uno compatible.

Usar `encoding/json`; evitar parsers/abstracciones innecesarias.

---

# 5. Endurecer Boilerplate Assessment assertion

Si `repository-profile` clasifica:

```text
starter
boilerplate
template
```

entonces debe existir evidencia estructurada de:

```text
boilerplate_assessment
```

Debe demostrar evaluación de al menos:

```text
extension_points
editable_boundaries
generated_files
feature_addition_workflow
upgrade_strategy
```

No exigir findings positivos.

Válido:

```yaml
upgrade_strategy:
  status: not_found
```

El objetivo es demostrar que fue evaluado.

---

# 6. `internal/app/driven_fixtures_test.go`

Separar también:

```text
visible output assertions
```

de:

```text
state persistence assertions
```

No concatenar ambos para satisfacer un mismo oracle.

Para el fixture TanStack/starter exigir visible:

```text
Repository
Context Placement
Tool / Capability Assessment
Checkpoint
```

Y persistido:

```text
repository kind
context placement verdict
boilerplate assessment
checkpoint
```

No obligar:

```text
skill created
```

No obligar:

```text
NO_ACTION
```

---

# 7. `agent-ready-orchestrator/SKILL.md`

Ruta:

```text
internal/bootstrap/assets/skills/agent-ready-orchestrator/SKILL.md
```

Buscar wording equivalente a:

```text
when nothing scores above threshold,
return NO_ACTION
```

Cambiar por:

```text
When no new-skill candidate clears its threshold
AND no other artifact change or context-placement transformation is justified,
return NO_ACTION.
```

Motivo:

El threshold `>=85` pertenece a creación de una nueva skill.

No debe bloquear:

```text
COMPACT
EXTRACT_TO_SKILL
MOVE_TO_REFERENCE
REPLACE_WITH_SCRIPT
```

No modificar el resto salvo consistencia mínima.

---

# 8. `artifact-design/SKILL.md`

Ruta:

```text
internal/bootstrap/assets/skills/artifact-design/SKILL.md
```

Buscar cualquier frase donde:

```text
score < 85
```

implique automáticamente:

```text
NO_ACTION
```

Reemplazar por:

```text
The >=85 threshold governs NEW SKILL creation only.

A candidate below the new-skill threshold may still justify:
- COMPACT
- EXTRACT_TO_SKILL when extracting existing repository guidance
- MOVE_TO_REFERENCE
- REPLACE_WITH_SCRIPT
- UPDATE
- REMOVE

NO_ACTION is valid only when no artifact or placement transformation is justified.
```

No modificar rubric.

No modificar `skill-reviewer`.

---

# 9. `audit-flow.md`

Ruta:

```text
internal/bootstrap/assets/skills/agent-ready-orchestrator/references/audit-flow.md
```

Cambiar únicamente semántica obsoleta.

## `REUSE`

Debe ser:

```text
REUSE =
existing guidance/artifact/skill covers the need
AND its current context placement is appropriate.
```

## `NO_ACTION`

Debe ser:

```text
NO_ACTION =
no justified artifact change remains
AND
no context-placement transformation is justified
AND
tool/capability assessment is complete.
```

No reescribir el flujo completo.

---

# 10. Test de falso positivo visible

Crear/adaptar un test donde la respuesta visible contiene:

```text
Outcome: NO_ACTION

Artifact Decisions:
  REUSE existing AGENTS guidance

Tool Assessment:
  ...
Checkpoint:
  complete
```

pero omite:

```text
Repository
Context Placement
```

Resultado requerido después del fix:

```text
FAIL
```

Aunque state contenga esos datos.

---

# 11. Test de Context Placement persistido

Caso insuficiente:

```json
{"decision":"NO_ACTION"}
```

Debe:

```text
FAIL
```

Caso válido:

```json
{
  "stage":"context_placement",
  "subject":"screen-creation",
  "decision":"REUSE",
  "reason":"..."
}
```

Debe:

```text
PASS
```

---

# 12. Test de Boilerplate Assessment

Perfil:

```yaml
kind:
  primary: starter
```

sin assessment:

```text
FAIL
```

Con:

```yaml
boilerplate_assessment:
  extension_points:
    status: found
  editable_boundaries:
    status: found
  generated_files:
    status: found
  feature_addition_workflow:
    status: covered
  upgrade_strategy:
    status: not_found
```

Debe:

```text
PASS
```

---

# 13. Mantener audit compacto

NO convertir la salida en reporte gigante.

Formato suficiente:

```text
Repository
  starter/template, TanStack Start + React + shadcn

Context Placement
  screen creation → REUSE
  considered EXTRACT_TO_SKILL; not justified

Artifacts
  NO_ACTION

Tools
  productivity: rg/fd/jq/ast-grep available; RTK NOT_JUSTIFIED
  providers: none justified

Checkpoint
  complete
```

Los detalles permanecen en state.

---

# 14. Resultado esperado en el dashboard TanStack

Después del fix:

```text
Outcome: NO_ACTION
```

puede seguir siendo correcto.

Pero la salida visible debe demostrar:

```text
Repository
  starter/template/boilerplate

Context Placement
  screen creation evaluated
  current location identified
  REUSE or EXTRACT_TO_SKILL decision
  alternative considered

Artifact Decisions
  valid verdict

Tool / Capability Assessment
  ecosystem
  productivity
  provider
  RTK explicitly evaluated

Checkpoint
  complete
```

State debe contener:

```text
repository kind
boilerplate assessment
context placement verdict
checkpoint
```

---

# 15. No modificar ahora

No tocar salvo bug comprobado:

```text
internal/tools/recommend.go
internal/tools/recipes/
skill-creator
skill-reviewer
RTK detection
provider catalog
OpenCode command integration
checkpoint engine
ownership
Go CLI architecture
```

---

# 16. Definition of Done

El parche está completo cuando:

- [ ] `driven_audit_test.go` separa output visible de state.
- [ ] persisted state no puede hacer pasar el test visible.
- [ ] Context Placement assertion es estructurada.
- [ ] `NO_ACTION` por sí solo no satisface Context Placement.
- [ ] Boilerplate Assessment assertion es estructurada.
- [ ] starter/boilerplate/template exige assessment persistido.
- [ ] `driven_fixtures_test.go` separa visible/state assertions.
- [ ] orchestrator ya no asocia threshold automáticamente a `NO_ACTION`.
- [ ] artifact-design aclara que >=85 solo aplica a nuevas skills.
- [ ] audit-flow define correctamente `REUSE`.
- [ ] audit-flow define correctamente `NO_ACTION`.
- [ ] no se modificó Tool Manager.
- [ ] no se añadió ninguna nueva arquitectura.
- [ ] `go test ./...` pasa.
- [ ] fixture TanStack pasa.
- [ ] prueba manual del dashboard muestra `Repository` + `Context Placement`.
- [ ] diff final permanece pequeño y localizado.

---

# 17. Restricción de tamaño

Antes de aplicar, mostrar plan:

```text
production files planned: 3
test files planned: 2
new production packages: 0
new agents: 0
new skills: 0
new commands: 0
new providers: 0
```

Si la solución requiere cambios significativos fuera de estos 5 archivos:

```text
STOP
```

y explicar la razón antes de continuar.

No aprovechar esta tarea para refactorizar código no relacionado.

---

# 18. Regla final

Este parche NO busca que Agent-Ready genere más artifacts.

Busca que:

```text
NO_ACTION
```

sea confiable.

Un `NO_ACTION` correcto debe demostrar:

```text
repository understood
+
context placement evaluated
+
tools evaluated
+
state persisted
```

y los tests deben validar cada capa por separado, sin falsos positivos por concatenación de output visible y state interno.
