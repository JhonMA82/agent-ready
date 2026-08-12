# Agent-Ready V1 — Context Placement & Token Optimization Refinement
## Instrucciones correctivas para OpenCode

> **Objetivo:** corregir la V1 existente de Agent-Ready después de la prueba real sobre `arhamkhnz/tanstack-shadcn-admin-dashboard`, sin rehacer la arquitectura que ya funciona.
>
> **Preservar:** CLI Go global, `agent-ready init`, `/agent-ready`, skills internas locales, checkpoints, Tool Budget, `NO_ACTION`, aislamiento por proyecto y modelo seleccionado por el usuario en OpenCode.
>
> **Añadir:** optimización explícita de ubicación de contexto, evaluación obligatoria de RTK y clasificación más precisa de repositorios tipo starter/boilerplate/template.
>
> **No introducir:** más agentes, más slash commands, TUI, runtime propio de OpenCode, MCP obligatorio, skills genéricas por framework o routing semántico rígido en Go.

---

# 0. Hallazgo principal

La prueba sobre `tanstack-shadcn-admin-dashboard` produjo correctamente `NO_ACTION` porque la mayor parte del guidance ya estaba cubierta, shadcn ya estaba delegado a una skill canónica, los scripts deterministas resolvían generación y no había evidencia suficiente para nuevas skills.

Ese comportamiento debe preservarse.

Sin embargo, apareció una limitación conceptual:

> Agent-Ready evalúa si una necesidad está cubierta, pero todavía no evalúa suficientemente bien **si esa información está ubicada en el nivel correcto de contexto**.

Ejemplo observado:

```text
Routing/screen creation
→ AGENTS.md
→ 13-step procedure
→ loaded every session
→ REUSE
```

Eso puede ser funcionalmente correcto y, al mismo tiempo, subóptimo para tokens.

---

# 1. Nuevo principio V1: Context Placement Optimization

Añadir como principio central:

> **Agent-Ready no solo debe decidir si la información existe. Debe decidir si está almacenada en el lugar donde produce el menor coste de contexto sin perder discoverability ni precisión.**

El problema ya no es solo `missing guidance`. También puede ser:

```text
correct guidance
+
wrong placement
```

---

# 2. Jerarquía de ubicación de contexto

Usar esta jerarquía conceptual:

```text
ALWAYS-ON CONTEXT
    ↓
AGENTS.md

TASK-SPECIFIC PROCEDURAL CONTEXT
    ↓
Skill

DETAILED REFERENCE KNOWLEDGE
    ↓
docs / references

DETERMINISTIC OPERATION
    ↓
script / CLI helper
```

No convertirla en routing hardcoded de Go. El modelo debe aplicar esta jerarquía mediante `artifact-design`.

---

# 3. Regla para AGENTS.md

`AGENTS.md` debe contener únicamente información que:

- aplica a una gran parte de las tareas;
- el agente necesita conocer casi siempre;
- define invariantes globales;
- define comandos esenciales;
- sirve como router hacia contexto más específico.

Ejemplos válidos:

```text
package manager
core commands
workspace structure
critical architectural constraints
forbidden operations
where to find docs
which skills to use
```

Ejemplos que deben revisarse críticamente:

```text
12–20 step feature procedure
full migration workflow
complete release procedure
framework tutorial
long list of edge cases
large canonical examples
```

---

# 4. Regla para skills

Una skill debe preferirse cuando la información es:

```text
task-specific
+
procedural
+
repeatable
+
repository-specific
+
non-trivial
```

Ejemplos válidos:

```text
add-dashboard-screen
create-database-migration
publish-workspace-package
add-domain-module
```

No crear:

```text
react-best-practices
typescript
tanstack-basics
rust-guide
laravel-best-practices
```

---

# 5. Regla para references/docs

Usar references/docs cuando:

- la información es detallada;
- no debe cargarse siempre;
- no constituye por sí sola un procedimiento;
- sirve de apoyo a una skill o AGENTS.md;
- tiene valor humano además del agente.

Ejemplos:

```text
architecture-map
canonical-examples
edge-cases
known-pitfalls
extension-points
dependency-map
```

---

# 6. Regla para scripts

Preferir script/CLI cuando:

- el procedimiento es determinista;
- los pasos no requieren juicio semántico;
- puede validarse automáticamente;
- repetirlo mediante instrucciones sería más caro o frágil.

Ejemplo:

```text
generate routes
validate presets
check architecture constraints
collect manifests
compute ChangeSet
```

No crear una skill para explicar 10 comandos deterministas si un script seguro puede ejecutarlos.

---

# 7. Nuevas decisiones de artifact-design

Extender las decisiones actuales:

```text
CREATE
UPDATE
REUSE
REMOVE
NO_ACTION
ASK_USER
```

con:

```text
COMPACT
EXTRACT_TO_SKILL
MOVE_TO_REFERENCE
REPLACE_WITH_SCRIPT
REUSE_EXTERNAL_SKILL
```

Interpretación:

- `COMPACT`: mantener el mismo artefacto, pero reducir contexto permanente.
- `EXTRACT_TO_SKILL`: mover procedimiento task-specific desde AGENTS/docs a una skill.
- `MOVE_TO_REFERENCE`: mover detalle excesivo a reference/docs y dejar un router pequeño.
- `REPLACE_WITH_SCRIPT`: sustituir instrucciones deterministas largas por helper validable.
- `REUSE_EXTERNAL_SKILL`: reutilizar una skill canónica existente sin wrapper local innecesario.

---

# 8. Context Placement Analysis contract

Cada candidato importante debe poder producir:

```yaml
context_placement:
  subject: dashboard-screen-creation

  current:
    location: AGENTS.md
    loaded: every_session
    size_estimate: 13-step-procedure

  applicability:
    always_needed: false
    task_specific: true
    procedural: true
    repository_specific: true

  alternatives:
    - keep_in_agents
    - extract_to_skill
    - move_detail_to_reference

  recommended:
    action: extract_to_skill

  expected_effect:
    permanent_context: decrease
    on_demand_context: increase_only_when_relevant
    discoverability: preserved

  evidence:
    - AGENTS.md
    - existing route examples
    - screen conventions
```

No es obligatorio mostrar el YAML literal en pantalla, pero el estado debe poder registrar la decisión.

---

# 9. Context Cost model

No intentar calcular tokens con falsa precisión.

Usar clases cualitativas:

```text
VERY_LOW
LOW
MEDIUM
HIGH
VERY_HIGH
```

Evaluar:

```text
always_loaded_cost
on_demand_cost
duplication
frequency_of_use
discoverability
maintenance_cost
```

Ejemplo:

```yaml
context_cost:
  current:
    always_loaded_cost: HIGH
    frequency_of_use: LOW

  proposed:
    always_loaded_cost: LOW
    on_demand_cost: MEDIUM

  result:
    optimize: true
```

---

# 10. Frecuencia de uso

Cuando pueda inferirse del repo:

```text
always
common
occasional
rare
unknown
```

No inventar frecuencia. Si es `unknown`, registrar incertidumbre. No mover contenido solo por longitud.

---

# 11. Duplicación

Antes de crear una skill:

```text
check existing AGENTS
check local skills
check canonical external skills
check docs
check scripts
```

Si la información ya existe, preferir:

```text
REUSE
COMPACT
EXTRACT_TO_SKILL
MOVE_TO_REFERENCE
REPLACE_WITH_SCRIPT
REUSE_EXTERNAL_SKILL
```

antes de crear duplicación.

---

# 12. Caso de referencia: tanstack-shadcn-admin-dashboard

Usar este repo como regression fixture manual/automatizable.

El resultado NO debe forzarse a crear una skill.

Pero el audit debe evaluar explícitamente el workflow de creación de pantallas con estas alternativas:

```text
REUSE existing AGENTS procedure
EXTRACT_TO_SKILL add-dashboard-screen
COMPACT AGENTS + reference
```

La decisión debe considerar:

- procedimiento de 13 pasos;
- carga permanente de AGENTS.md;
- especificidad del workflow;
- repetibilidad;
- discoverability;
- si ya existe una skill externa equivalente;
- beneficio real en tokens.

`REUSE` es válido solo si explica por qué mantenerlo always-on es mejor.

`EXTRACT_TO_SKILL` es válido si demuestra ahorro de contexto sin perder guidance.

---

# 13. Skill candidata de referencia

Si la evidencia lo justifica, una skill posible:

```text
.opencode/skills/add-dashboard-screen/
```

Debe contener solo guidance específico del repo:

```text
route placement
feature/module location
sidebar registration
SSR/loading/error state expectations
theme token usage
responsive conventions
canonical screen examples
validation commands
```

No enseñar React/TanStack/shadcn genéricos.

Si shadcn ya tiene skill canónica:

```text
delegate shadcn-specific work
```

en vez de duplicarla.

---

# 14. RTK debe evaluarse siempre

RTK debe formar parte de la matriz obligatoria del audit.

No significa recomendarlo siempre.

Estados válidos:

```text
RECOMMENDED
CONSIDER
NOT_JUSTIFIED
ALREADY_AVAILABLE
```

Nunca omitirlo silenciosamente.

---

# 15. RTK decision model

No basar el veredicto únicamente en:

```text
dist/
build/
coverage/
```

Evaluar señales:

```text
build command output volume
test output volume
git diff/status/log usage
package manager output
cargo/nix output
CI/log inspection
frequency of shell-heavy workflows
```

No ejecutar comandos costosos solo para medir sin consentimiento. Puede usar evidencia observada durante el audit.

---

# 16. RTK output example

```yaml
tool:
  id: rtk
  status: consider

  evidence:
    - npm build and check are common project commands
    - no measured high-volume output yet

  reason:
    token reduction may help, but benefit has not been demonstrated

  next:
    evaluate after observing actual command output
```

O:

```yaml
tool:
  id: rtk
  status: not_justified
  reason:
    current repository workflows produce small outputs and rg/fd/jq cover navigation efficiently
```

---

# 17. Tool assessment completeness

Todo audit debe mostrar explícitamente:

```text
Ecosystem
Productivity
Providers
```

Productivity debe evaluar al menos:

```text
rg
fd
jq
RTK
gh when GitHub remote exists
ast-grep when structured search may help
```

Providers:

```text
Context7
Semble
Serena
CodeGraph
Headroom
```

No tienen que recomendarse.

---

# 18. Provider decision completeness

Evitar:

```text
provider candidates: none
```

sin indicar qué se evaluó.

La salida debe poder mostrar:

```text
Context7   NOT_JUSTIFIED
Serena     NOT_JUSTIFIED
CodeGraph  NOT_JUSTIFIED
Semble     NOT_JUSTIFIED
Headroom   NOT_JUSTIFIED
```

con razones compactas.

Esto permite verificar que Tool Budget realmente ocurrió.

---

# 19. Repository kind debe ser visible

La prueba del dashboard debe identificar claramente que el repo es starter/template/boilerplate si la evidencia lo soporta.

Agregar en repository profile:

```yaml
repository_kind:
  primary: boilerplate
  secondary:
    - application
  confidence: 0.9
```

o equivalente.

---

# 20. Boilerplate-specific audit

Cuando `repository_kind` sea:

```text
boilerplate
starter
template
```

evaluar:

```text
extension points
what downstream users should edit
what they should not edit
generated files
feature addition workflow
variants/presets
scaffolding
upgrade/update strategy
canonical customization examples
```

No significa crear artifacts automáticamente.

---

# 21. Boilerplate artifact candidates

Candidatos permitidos:

```text
PROJECT.template.md
docs/ai/extension-points.md
docs/ai/boilerplate-variants.md
skill: add-feature
skill: add-screen
skill: create-variant
```

solo con evidencia.

---

# 22. Boilerplate context-placement concern

En boilerplates es especialmente probable encontrar un `AGENTS.md` grande.

Agent-Ready debe preguntar:

```text
Which instructions must be always-on?
Which workflows should become on-demand skills?
Which examples should become references?
```

Esto debe ser parte del audit.

---

# 23. Repository profile improvements

Añadir o confirmar:

```yaml
repository_profile:
  kind:
    primary:
    secondary: []

  ecosystems: []
  central_frameworks: []

  existing_agent_assets:
    agents_md:
    local_skills:
    external_skills:
    scripts:

  context_placement:
    always_on_estimate:
    task_specific_guidance_candidates: []

  tool_assessment:
    ecosystem: []
    productivity: []
    providers: []
```

---

# 24. Orchestrator changes

Actualizar `agent-ready-orchestrator/SKILL.md`.

Agregar:

> Coverage is not sufficient to conclude REUSE. Evaluate context placement before final artifact decisions.

El orquestador debe preguntar:

```text
Is this guidance always applicable?
Is it task-specific?
Is it procedural?
Is it too detailed for always-on context?
Would extraction reduce token cost?
Would extraction harm discoverability?
```

---

# 25. artifact-design changes

Agregar una sección:

```text
Context Placement Gate
```

Antes de concluir `REUSE` sobre guidance existente.

Flow conceptual:

```text
Need already covered?
    ↓ yes
Is placement optimal?
    ↓
yes → REUSE
no
 ↓
COMPACT / EXTRACT_TO_SKILL /
MOVE_TO_REFERENCE / REPLACE_WITH_SCRIPT
```

---

# 26. skill-creator changes

Cuando recibe una extracción desde AGENTS/docs:

- preservar semántica original;
- no inventar nuevas convenciones;
- usar canonical examples existentes;
- dejar un router corto en el artefacto original cuando corresponda;
- no copiar el mismo contenido en ambos lugares.

---

# 27. skill-reviewer changes

Añadir checks:

```text
context_savings
duplication_after_extraction
discoverability_preserved
always_on_guidance_not_removed_if_global
```

Una extracción debe rechazarse si:

- mueve una invariante global a una skill;
- hace más difícil descubrir una regla crítica;
- duplica contenido en lugar de moverlo;
- no produce ahorro real.

---

# 28. General artifact reviewer

Añadir `placement_review`.

Debe validar:

- AGENTS no creció innecesariamente;
- procedures largos tienen justificación;
- references no contienen reglas que deban ser always-on;
- scripts no esconden decisiones semánticas;
- skills no duplican external skills.

---

# 29. `NO_ACTION` sigue siendo primera clase

No convertir esta corrección en una máquina de extracción.

Resultado correcto puede seguir siendo `NO_ACTION` si:

- coverage es completa;
- ubicación es óptima;
- Tool Budget no identifica mejoras;
- no hay research necesario.

La mejora es que `NO_ACTION` debe estar respaldado por:

```text
coverage check
+
placement check
+
tool assessment
```

---

# 30. New audit stage terminology

Internamente registrar:

```text
exploration_plan
repository_analysis
context_placement
artifact_decisions
tool_assessment
approval
apply
review
checkpoint
```

No es obligatorio mostrar cada una como fila si perjudica UX.

---

# 31. Suggested audit output

Ejemplo:

```text
Audit outcome: NO_ACTION

Repository
  TanStack Start / React / shadcn starter
  npm
  367 files

Context Placement
  AGENTS.md
    screen creation: REVIEWED
    13-step procedure is task-specific

  Decision
    REUSE
    Reason: current project expects screen creation to be one of the dominant workflows
    and the instruction also carries global routing invariants.

  Alternative considered
    EXTRACT_TO_SKILL add-dashboard-screen
    Rejected because expected savings are small relative to discoverability cost.

Artifacts
  NO_ACTION

Tools
  Productivity
    rg          AVAILABLE
    fd          AVAILABLE
    jq          AVAILABLE
    ast-grep    AVAILABLE
    RTK         NOT_JUSTIFIED

  Providers
    Context7    NOT_JUSTIFIED
    Semble      NOT_JUSTIFIED
    Serena      NOT_JUSTIFIED
    CodeGraph   NOT_JUSTIFIED
    Headroom    NOT_JUSTIFIED

Checkpoint
  saved
```

O puede decidir `EXTRACT_TO_SKILL`. Lo importante es demostrar el análisis.

---

# 32. Token optimization objective

Añadir explícitamente a la definición de Agent-Ready:

> **Agent-Ready optimiza tanto el contenido como su distribución entre always-on y on-demand context.**

Métricas cualitativas:

```text
always_on_context_reduced
duplicate_guidance_avoided
targeted_skill_loads
references_loaded_on_demand
unnecessary_artifacts_avoided
tool_output_reduced
```

---

# 33. Context-placement provenance

Si mueve contenido:

```yaml
placement_change:
  from:
    path: AGENTS.md
    section: "Adding a screen"

  to:
    path: .opencode/skills/add-dashboard-screen/SKILL.md

  reason:
    task_specific_procedure

  preserved_router:
    path: AGENTS.md
    text: "Use the add-dashboard-screen skill for new dashboard screens."

  source_hash: ...
```

Esto facilita `sync`.

---

# 34. Incremental sync interaction

`/agent-ready sync` debe detectar:

```text
AGENTS content changed
skill changed
reference changed
canonical example changed
```

Si una skill fue extraída desde AGENTS:

```text
update source dependency graph
```

No volver a duplicar automáticamente.

---

# 35. Artifact graph extension

Permitir relaciones:

```yaml
artifact:
  path: .opencode/skills/add-dashboard-screen/SKILL.md

  derived_from:
    - AGENTS.md#adding-screen
    - src/routes/**
    - src/features/**

  routed_from:
    - AGENTS.md

  refresh_when:
    - source_section_changed
    - route_structure_changed
    - canonical_example_changed
```

---

# 36. Tool reassessment on sync

No evaluar providers desde cero en cada sync.

Reevaluar cuando:

```text
repo complexity materially changes
workspace count changes
new framework added
new language ecosystem added
tool output problem observed
new provider already installed
```

Esto ahorra tokens.

---

# 37. Regression fixture: tanstack-shadcn-admin-dashboard

Añadir test/eval específico.

Debe comprobar:

```text
repository kind recognized as starter/boilerplate/template when evidence supports it
npm correctly detected
shadcn external skill reuse recognized
generate-routes scripts treated as deterministic
generic React/TanStack skills rejected
screen-creation context placement evaluated
RTK explicitly evaluated
advanced providers explicitly evaluated
NO_ACTION remains possible
```

No obligar `EXTRACT_TO_SKILL`.

---

# 38. Regression fixture: long AGENTS

Crear fixture sintético con `AGENTS.md` 500+ líneas:

- 50 líneas globales;
- 150 líneas migration workflow;
- 100 líneas release workflow;
- 200 líneas examples.

Resultado esperado:

```text
COMPACT AGENTS
EXTRACT migration skill
EXTRACT release skill
MOVE examples to references
```

No aceptar `REUSE` automático solo porque toda la información ya existe.

---

# 39. Regression fixture: short optimal AGENTS

Fixture `AGENTS.md` ~60 líneas con guidance global bien colocada.

Resultado:

```text
REUSE
NO_ACTION
```

No extraer por dogma.

---

# 40. Regression fixture: deterministic procedure

AGENTS contiene un procedimiento largo, pero repo ya tiene un script determinista equivalente.

Resultado esperado:

```text
REPLACE_WITH_SCRIPT / COMPACT
```

no skill redundante.

---

# 41. Regression fixture: external canonical skill

Repo contiene guidance para una capability ya cubierta por skill canónica externa.

Resultado:

```text
REUSE_EXTERNAL_SKILL
```

No generar wrapper sin necesidad.

---

# 42. Skill quality rubric update

Añadir dimensión:

```yaml
context_placement:
  weight: 10
```

Ejemplo rebalanceado:

```yaml
necessity: 20
repository_specificity: 15
discovery_description: 15
procedural_value: 15
progressive_disclosure: 10
evidence_grounding: 10
context_placement: 10
validation: 5
```

---

# 43. Reviewer score

Una skill puede ser técnicamente excelente pero debe rechazarse si:

```text
same content remains fully in AGENTS
```

porque no aporta progressive disclosure real.

---

# 44. Boilerplate assessment output

Cuando aplica:

```text
Repository Kind
  boilerplate

Extension Model
  add route
  add feature
  add shadcn component
  generate preset

Agent Context
  always-on
    core stack
    folder boundaries

  on-demand candidates
    add dashboard screen
    create preset

  deterministic
    generate routes
```

El output puede resumirse en pantalla, pero debe quedar registrado.

---

# 45. No tool recommendation inflation

La evaluación explícita NO significa mencionar 10 tools con párrafos largos.

Salida compacta permitida:

```text
Providers: none justified.
Evaluated: Context7, Semble, Serena, CodeGraph, Headroom.
```

RTK sí debe aparecer individualmente en productivity por ser baseline de optimización de tokens.

---

# 46. `agent-ready tools recommend --json`

Añadir o confirmar señales para:

```text
context_placement_pressure
high_output_pressure
structured_search_need
semantic_retrieval_need
symbol_navigation_need
graph_need
versioned_docs_need
```

Go solo reporta señales. El modelo emite decisión.

---

# 47. Avoid fake token estimates

No decir:

```text
this saves 1,284 tokens
```

salvo medición real.

Usar:

```text
expected permanent context reduction: MEDIUM
```

---

# 48. Definition of Done

Esta corrección queda completa cuando:

- [ ] `artifact-design` evalúa placement antes de `REUSE`.
- [ ] existen decisiones `COMPACT`, `EXTRACT_TO_SKILL`, `MOVE_TO_REFERENCE`, `REPLACE_WITH_SCRIPT`, `REUSE_EXTERNAL_SKILL`.
- [ ] AGENTS always-on vs skill on-demand queda modelado.
- [ ] `skill-creator` soporta extracción sin duplicación.
- [ ] `skill-reviewer` valida context savings.
- [ ] reviewer general valida placement.
- [ ] `repository-profile` identifica starter/boilerplate/template.
- [ ] boilerplate audit evalúa extension points.
- [ ] RTK aparece explícitamente en todo Tool Assessment.
- [ ] RTK no depende únicamente de `dist/build/coverage`.
- [ ] advanced providers siguen bajo Tool Budget.
- [ ] `NO_ACTION` sigue siendo válido.
- [ ] tanstack-shadcn regression pasa.
- [ ] long-AGENTS fixture optimiza placement.
- [ ] short-AGENTS fixture no sobreoptimiza.
- [ ] deterministic-workflow fixture usa script.
- [ ] external-skill fixture reutiliza canonical skill.
- [ ] `go test ./...` y evals pasan.

---

# 49. Regla final

La nueva pregunta que Agent-Ready debe responder es:

> **¿La información correcta existe y está en el lugar correcto para que el agente pague su coste de contexto solo cuando realmente la necesita?**

La mejor salida puede ser:

```text
CREATE
NO_ACTION
COMPACT
EXTRACT_TO_SKILL
MOVE_TO_REFERENCE
REPLACE_WITH_SCRIPT
REUSE_EXTERNAL_SKILL
```

El objetivo sigue siendo:

> **menos contexto permanente, menos duplicación, mejores skills, mejores decisiones y cero artifact spam.**
