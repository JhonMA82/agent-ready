# Agent-Ready V1 — Instrucciones correctivas finales para OpenCode

> **Objetivo:** corregir la implementación actual de `JhonMA82/agent-ready` sin cambiar su arquitectura.
>
> La V1 ya funciona correctamente en instalación, aislamiento, `/agent-ready`, Tool Assessment, RTK, Tool Budget, providers, checkpoints y `NO_ACTION`. El problema restante es que el modelo puede concluir `REUSE`/`NO_ACTION` sin demostrar ni persistir correctamente **Context Placement** y **Repository Kind / Boilerplate Assessment**.
>
> Esta corrección debe ser pequeña y localizada.

---

# 1. Reglas de trabajo

Antes de modificar:

1. Verificar cada archivo y comportamiento en el código actual.
2. No reescribir arquitectura.
3. No añadir agentes.
4. No añadir slash commands.
5. No añadir TUI.
6. No añadir MCP propio.
7. No añadir runtime OpenCode.
8. No modificar Tool Manager salvo que un test demuestre un fallo real.
9. Preservar `NO_ACTION` como resultado válido.
10. Preservar la estrategia OpenCode-native y skills locales actuales.

La intención es:

```text
capacidades ya existentes
+
output contracts más estrictos
+
routing de decisiones completo
+
acceptance tests reales
```

---

# 2. Archivos que deben modificarse

Modificar:

```text
internal/bootstrap/assets/skills/agent-ready-orchestrator/SKILL.md
internal/bootstrap/assets/skills/artifact-design/SKILL.md
internal/bootstrap/assets/skills/artifact-design/references/artifact-decisions.md
internal/bootstrap/assets/skills/repository-analysis/SKILL.md
internal/bootstrap/assets/skills/repository-analysis/references/inventory-facts.md
internal/bootstrap/assets/skills/agent-ready-orchestrator/references/audit-flow.md
internal/bootstrap/content_test.go
internal/app/driven_audit_test.go
internal/app/driven_fixtures_test.go
```

Añadir:

```text
internal/app/testdata/acceptance/driven/tanstack-starter/
```

No modificar otros subsistemas salvo necesidad demostrada por tests.

---

# 3. `agent-ready-orchestrator/SKILL.md`

Ruta:

```text
internal/bootstrap/assets/skills/agent-ready-orchestrator/SKILL.md
```

## Problema

El orchestrator ya conoce Context Placement, Tool Assessment, `NO_ACTION` y checkpoints, pero su Output Contract no obliga a demostrar `Repository Profile` ni `Context Placement`.

## Cambio requerido

El Output Contract debe exigir:

```text
Repository Profile
Context Placement
Artifact Decisions
Tool / Capability Assessment
Checkpoint
```

Antes de devolver `NO_ACTION` deben cumplirse:

```text
repository kind classified
relevant existing guidance evaluated
context placement evaluated
no placement optimization justified
tool assessment completed
no artifact candidate justified
```

Añadir:

> Existing coverage alone is not sufficient for REUSE or NO_ACTION. Existing guidance must also pass the Context Placement Gate.

---

# 4. `artifact-design/SKILL.md`

Ruta:

```text
internal/bootstrap/assets/skills/artifact-design/SKILL.md
```

Mantener/asegurar los verdicts:

```text
CREATE
UPDATE
REUSE
REMOVE
NO_ACTION
ASK_USER
COMPACT
EXTRACT_TO_SKILL
MOVE_TO_REFERENCE
REPLACE_WITH_SCRIPT
REUSE_EXTERNAL_SKILL
```

Routing requerido:

```text
CREATE
→ proposal/review

UPDATE
→ proposal/review

REUSE
→ persist placement verdict

REMOVE
→ proposal/review

COMPACT
→ proposal/review

EXTRACT_TO_SKILL
→ skill-creator
→ skill-reviewer
→ proposal/review

MOVE_TO_REFERENCE
→ proposal/review

REPLACE_WITH_SCRIPT
→ deterministic artifact proposal
→ review

REUSE_EXTERNAL_SKILL
→ persist external coverage decision

NO_ACTION
→ only after Context Placement Gate

ASK_USER
→ stop and ask
```

La regla:

```text
if no candidate scores >= 85 → NO_ACTION
```

solo debe aplicar a **creación de nuevas skills**.

No debe impedir:

```text
COMPACT
EXTRACT_TO_SKILL
MOVE_TO_REFERENCE
REPLACE_WITH_SCRIPT
```

---

# 5. `artifact-decisions.md`

Ruta:

```text
internal/bootstrap/assets/skills/artifact-design/references/artifact-decisions.md
```

Cambiar `REUSE` a:

```text
Use when existing guidance/artifact/skill already covers the need
AND its current context placement is appropriate.
```

Cambiar `NO_ACTION` para exigir:

```text
no justified CREATE/UPDATE/REMOVE
AND
no justified placement transformation
AND
relevant existing guidance passed Context Placement Gate
AND
tool assessment is complete
```

Context Placement Gate:

```text
Is this guidance always applicable?
Is it task-specific?
Is it procedural?
Is it too detailed for always-on context?
Would extraction reduce permanent context?
Would extraction reduce discoverability or remove global invariants?
```

---

# 6. `repository-analysis/SKILL.md`

Ruta:

```text
internal/bootstrap/assets/skills/repository-analysis/SKILL.md
```

El Output Contract debe emitir:

```yaml
repository_profile:

  kind:
    primary:
    secondary: []
    confidence:

  topology:
    monorepo:
    workspace_count:

  ecosystems: []

  central_frameworks: []

  existing_agent_assets:
    agents_md:
    local_skills:
    external_skills:
    scripts:

  context_placement:
    always_on:
    task_specific_candidates:

  tool_assessment:
    ecosystem: []
    productivity: []
    providers: []
```

Kinds permitidos:

```text
application
library
cli
starter
boilerplate
template
infrastructure
mixed
```

No usar `monorepo` como identidad excluyente. Usar:

```yaml
topology:
  monorepo: true
```

---

# 7. Boilerplate Assessment

Si:

```text
kind.primary ∈ {boilerplate, starter, template}
```

evaluar:

```text
extension points
editable boundaries
generated files
feature addition workflow
variants/presets
scaffolding
upgrade strategy
canonical customization examples
```

Contrato:

```yaml
boilerplate_assessment:

  extension_points: []
  editable_boundaries: []
  generated_files: []

  feature_addition_workflow:
    status:
    evidence: []

  variants:
    status:
    evidence: []

  scaffolding:
    status:
    evidence: []

  upgrade_strategy:
    status:
    evidence: []

  canonical_customization_examples: []
```

No significa crear artifacts.

Solo demostrar que la evaluación ocurrió.

---

# 8. `inventory-facts.md`

Ruta:

```text
internal/bootstrap/assets/skills/repository-analysis/references/inventory-facts.md
```

Actualizar el contrato para soportar:

```text
repository kind
topology
boilerplate assessment
context placement candidates
existing agent assets
```

Go puede detectar hechos; el modelo clasifica semánticamente.

---

# 9. `audit-flow.md`

Ruta:

```text
internal/bootstrap/assets/skills/agent-ready-orchestrator/references/audit-flow.md
```

No reescribirlo.

Añadir:

```text
checkpoint --complete MUST NOT be emitted when:

relevant existing guidance was used to justify
REUSE or NO_ACTION

AND

no Context Placement verdict was recorded.
```

Y:

```text
If repository kind is:

starter
boilerplate
template

then Repository output MUST include
Boilerplate Assessment.
```

---

# 10. Context Placement Decision

Cuando guidance existente sea relevante para una decisión, persistir equivalente a:

```yaml
context_placement:

  subject: screen-creation

  current:
    location: AGENTS.md
    loaded: every_session

  properties:
    always_applicable: false
    task_specific: true
    procedural: true
    repository_specific: true

  alternatives:
    - reuse
    - extract_to_skill
    - move_to_reference

  decision: reuse

  reason:
    "Screen creation is one of the dominant workflows and the existing section also carries routing invariants."

  expected_effect:
    permanent_context: unchanged
    discoverability: preserved
```

O:

```yaml
decision: extract_to_skill
```

cuando corresponda.

No exigir YAML literal si el state actual usa JSONL.

---

# 11. Context Cost

No calcular tokens con falsa precisión.

Usar:

```text
VERY_LOW
LOW
MEDIUM
HIGH
VERY_HIGH
```

Dimensiones:

```text
always_loaded_cost
on_demand_cost
frequency_of_use
duplication
discoverability
maintenance_cost
```

---

# 12. `internal/bootstrap/content_test.go`

Añadir assertions para:

```text
COMPACT
EXTRACT_TO_SKILL
MOVE_TO_REFERENCE
REPLACE_WITH_SCRIPT
REUSE_EXTERNAL_SKILL
Context Placement
Repository Profile
Boilerplate Assessment
RTK
```

El test debe fallar si:

```text
agent-ready-orchestrator/SKILL.md
```

no contiene en su Output Contract:

```text
Repository
Context Placement
Tool / Capability Assessment
Checkpoint
```

También debe fallar si `artifact-design/SKILL.md` no contiene routing explícito para:

```text
EXTRACT_TO_SKILL
COMPACT
MOVE_TO_REFERENCE
REPLACE_WITH_SCRIPT
```

---

# 13. `internal/app/driven_audit_test.go`

Ampliar `assertAuditStructure()`.

No exigir outcome específico.

Sí exigir evidencia observable de:

```text
Repository
Context Placement
Artifact Decisions
Tool / Capability Assessment
Checkpoint
```

Tool Assessment debe contener:

```text
ecosystem
productivity
provider
```

y Productivity debe evaluar explícitamente:

```text
RTK
```

aunque resulte `NOT_JUSTIFIED`.

No exigir skill creada.

No exigir `NO_ACTION`.

---

# 14. State assertions

Después de audit completo:

```text
repository-profile.yaml
```

debe contener:

```text
kind.primary
kind.confidence
```

Cuando aplique:

```text
boilerplate_assessment
```

debe quedar persistido o trazable en state/decisions/provenance.

Cuando `REUSE` o `NO_ACTION` dependa de guidance existente:

```text
Context Placement verdict
```

debe quedar registrado.

---

# 15. `internal/app/driven_fixtures_test.go`

Añadir fixture/cohort:

```text
tanstack-starter
```

No depender del repo remoto durante tests.

---

# 16. Fixture `tanstack-starter`

Ruta:

```text
internal/app/testdata/acceptance/driven/tanstack-starter/
```

Contenido mínimo:

```text
package.json
package-lock.json
AGENTS.md
src/
  routes/
  features/
scripts/
  generate-routes.*
```

`AGENTS.md` debe contener:

- stack global;
- reglas always-on;
- procedimiento largo de screen creation;
- referencia a una skill canónica externa de shadcn;
- comandos deterministas.

`package.json` debe contener:

```text
generate:routes
generate:presets
```

o equivalentes.

---

# 17. Oracle TanStack

Debe comprobar:

## Repository classification

```text
starter/template/boilerplate recognized
```

## Existing guidance

```text
screen workflow detected
```

## Context Placement

Evaluar:

```text
REUSE
vs
EXTRACT_TO_SKILL
vs
MOVE_TO_REFERENCE
```

No imponer cuál gana.

## Skills genéricas

Rechazar:

```text
generic React skill
generic TanStack skill
generic shadcn skill
```

## External skill

Permitir:

```text
REUSE_EXTERNAL_SKILL
```

si el fixture expone canonical shadcn skill.

## Deterministic scripts

Reconocer route/preset generation como scripts, no skills.

## Tools

Evaluar:

```text
RTK
Context7
Semble
Serena
CodeGraph
Headroom
```

sin obligación de recomendar ninguno.

## NO_ACTION

Debe seguir siendo válido solo si:

```text
placement evaluated
+
boilerplate assessment complete
+
tool assessment complete
```

---

# 18. Fixture `long-agents`

Si no existe uno equivalente, añadir un fixture con:

```text
AGENTS.md > 400 lines
```

que contenga:

```text
global invariants
migration workflow
release workflow
large examples
```

Esperar según evidencia:

```text
COMPACT
EXTRACT_TO_SKILL
MOVE_TO_REFERENCE
```

No aceptar `REUSE everything` únicamente porque la información exista.

---

# 19. Fixture `short-optimal-agents`

AGENTS corto y bien ubicado.

Debe permitir:

```text
REUSE
NO_ACTION
```

No sobreoptimizar.

---

# 20. Fixture deterministic workflow

AGENTS describe varios pasos manuales, pero el repo ya tiene un script equivalente.

Esperar:

```text
REPLACE_WITH_SCRIPT
```

o:

```text
COMPACT + reuse script
```

No crear skill redundante.

---

# 21. Regresión manual final

Repetir sobre:

```text
arhamkhnz/tanstack-shadcn-admin-dashboard
```

La salida puede continuar siendo:

```text
NO_ACTION
```

pero debe demostrar:

```text
Repository Kind
  starter/template/boilerplate

Boilerplate Assessment
  extension points
  screen workflow
  generated routes/presets
  customization boundaries

Context Placement
  screen creation evaluated
  REUSE or EXTRACT decision
  alternative considered

Tools
  RTK evaluated
  providers evaluated

Checkpoint
```

---

# 22. Ejemplo válido

```text
Outcome: NO_ACTION

Repository
  kind: starter / template
  stack: TanStack Start, React, shadcn
  package manager: npm

Boilerplate Assessment
  screen extension workflow: covered
  route generation: deterministic
  preset generation: deterministic
  shadcn customization: external canonical skill
  upgrade strategy: documented/partial

Context Placement
  screen creation:
    current: AGENTS.md
    decision: REUSE
    alternative: EXTRACT_TO_SKILL
    reason:
      workflow is common and contains global routing invariants;
      extraction savings are not sufficient.

Artifacts
  NO_ACTION

Tools
  RTK        NOT_JUSTIFIED
  Context7   NOT_JUSTIFIED
  Semble     NOT_JUSTIFIED
  Serena     NOT_JUSTIFIED
  CodeGraph  NOT_JUSTIFIED
  Headroom   NOT_JUSTIFIED

Checkpoint
  complete
```

También es válido `EXTRACT_TO_SKILL` si la evidencia lo justifica.

---

# 23. No modificar ahora

No tocar salvo bug comprobado:

```text
internal/tools/recommend.go
internal/tools/recipes/
skill-creator base
skill-reviewer rubric base
RTK detection
provider catalog
OpenCode command integration
checkpoint engine
ownership
Go CLI architecture
```

---

# 24. No añadir

No resolver esto agregando:

```text
nuevo agent
nuevo MCP
nuevo slash command
nuevo registry
nuevo daemon
más providers
hardcoded TanStack rules
hardcoded boilerplate paths
```

---

# 25. Definition of Done

La corrección queda completa cuando:

- [ ] orchestrator Output Contract exige Repository Profile.
- [ ] orchestrator Output Contract exige Context Placement.
- [ ] `NO_ACTION` requiere placement gate.
- [ ] artifact-design enruta los 11 verdicts.
- [ ] threshold >=85 solo controla creación de skills.
- [ ] `REUSE` requiere placement válido.
- [ ] `NO_ACTION` incluye ausencia de placement optimization.
- [ ] repository-analysis persiste kind/confidence.
- [ ] starter/boilerplate/template activa Boilerplate Assessment.
- [ ] topology monorepo no reemplaza kind.
- [ ] audit-flow bloquea checkpoint sin placement cuando aplica.
- [ ] content tests cubren los nuevos contracts.
- [ ] driven audit exige Repository + Context Placement.
- [ ] RTK sigue evaluándose.
- [ ] fixture TanStack pasa.
- [ ] long AGENTS fixture detecta optimization.
- [ ] short optimal AGENTS permite NO_ACTION.
- [ ] deterministic workflow evita skill redundante.
- [ ] `go test ./...` pasa.
- [ ] regresión manual del dashboard produce output explicable.

---

# 26. Regla final

No intentar hacer que Agent-Ready “genere más”.

La corrección debe hacer que justifique mejor:

```text
CREATE
REUSE
NO_ACTION
COMPACT
EXTRACT_TO_SKILL
MOVE_TO_REFERENCE
REPLACE_WITH_SCRIPT
REUSE_EXTERNAL_SKILL
```

La pregunta final del audit debe ser:

> **¿La información necesaria existe, está en el lugar correcto y cuesta el mínimo contexto razonable para este tipo de repositorio?**

Si la respuesta es sí, `NO_ACTION` es una salida excelente.
