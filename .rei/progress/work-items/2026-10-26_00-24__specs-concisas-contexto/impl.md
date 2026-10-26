# Implementación — Specs concisas, issue body sin design y contexto bajo control

## Resumen

Implementados los 9 pasos de `plan.md`: presupuesto de planificación, criterio de
tipo (Task para docs o cambios pequeños sin cambio de comportamiento), reglas de
contexto en roles/`AGENTS.md`/`progress.md`, `WARN` blando de presupuesto en
`rei validate` con tests, e issue body sin `design.md` completo (comentario
idempotente opcional). Tras la revisión se corrigió la incoherencia del criterio
de tipo entre `specs.md`, `task.md`, `workflow.md` y `leader.md`.

## Archivos modificados

- Docs: `.rei/docs/harness/{specs,task,workflow,progress}.md`.
- GitHub: `integrations/github.md` y `github/{templates,core,backlog}.md`.
- Roles: `.rei/agents/{leader,spec_author,implementer,reviewer}.md`.
- Raíz: `AGENTS.md` §9.
- Código/tests: `internal/validate/validate.go`, `validate_test.go`.
- Generados: `.opencode/agents/*` y `.claude/agents/*` (desde el Contrato;
  `--check` en verde).

## Cambios

- `specs.md`: «Presupuesto de planificación» (requisitos 10–20/máx. **30**; design
  150–300/máx. 400; tasks 8–12/máx. 15), regla dividir o justificar, concisión y
  criterio de tipo: Task = documentación/contratos o cambio pequeño **sin cambio de
  comportamiento ni de arquitectura**; comportamiento o arquitectura ⇒ Feature.
- `task.md`: misma conjunción («siempre que no cambie el comportamiento ni la
  arquitectura»), `plan.md` 150–200 líneas (máx. 200), sin requirements/design.
- `workflow.md`: retirados «cambios menores de comportamiento» y «pequeños ajustes
  de interfaz» de Task; pasan a Feature como «modificaciones del comportamiento,
  por menores que sean» y «pequeños ajustes de interfaz que cambian lo que el
  usuario ve o recibe». «Corrección de errores localizados» se recalifica a
  erratas/ajustes de texto sin cambio observable. Desempate: duda ⇒ Feature.
- Roles: Leader define `task` = docs/contratos o cambio pequeño sin cambio de
  comportamiento ni arquitectura, y `feature` incluye cualquier cambio de
  comportamiento (aunque el diff sea pequeño); Spec Author respeta el presupuesto;
  Implementer/Reviewer abren solo el Work Item y el Reviewer verifica el
  presupuesto (`WARN` no es rechazo automático). Todos: no `history.md`, no
  escanear `.rei/specs/**`.
- `validate.go`: constantes de presupuesto, `budgetApplicable` (ready,
  in_progress, review, changes_requested), conteo de requisitos con `regexp`,
  checkboxes de tareas y líneas de design/plan; `WARN` con presupuesto y «Justifica
  o divide», nunca `FAIL`.
- Tests: límite exacto sin WARN; 31/401/16 con 3 WARN; plan de 201 líneas con
  WARN; `done` fuera de presupuesto sin WARN.
- GitHub: cuerpo de issue y bloque de spec sin design; design solo como comentario
  idempotente `<!-- rei:comment:design:<id> -->` si `.rei` no se versiona o el
  usuario pide trazabilidad; marcador añadido a la tabla.
- Contexto: `AGENTS.md` §9 y `progress.md` con no-history, no-escanear specs y
  `rei items status`/`rei item show`.

## Verificación

- `gofmt -l internal cmd`: sin salida. `go vet ./...`: OK. `go test ./...`: OK.
  `make build`: OK.
- `./bin/rei check --quiet`: exit 0 (V1).
- `./bin/rei init opencode --check` y `rei init claude --check`: exit 0 (tras
  regenerar los adaptadores desde `leader.md`).
- `./bin/rei doctor`: exit 0; `./bin/rei init status`: exit 0.
- `./bin/rei validate 2026-10-26_00-24__specs-concisas-contexto`: OK, sin WARN de
  presupuesto (plan de 118 líneas).
- Tests de presupuesto PASS: `TestPlanningBudgetWithinLimit`,
  `TestPlanningBudgetExceededWarns`, `TestTaskPlanBudgetWarns`,
  `TestPlanningBudgetDoneNoWarn`.
- V4 no aplica: sin código específico de plataforma.

## Observaciones

- `rei init` se ejecutó solo para regenerar los adaptadores derivados del Contrato;
  no se tocaron flujos, comandos, labels ni marcadores existentes.
- Sin dependencias nuevas; no se tocó `skeleton/`.
