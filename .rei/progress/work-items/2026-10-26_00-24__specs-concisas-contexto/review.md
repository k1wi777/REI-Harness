# Revisión — Specs concisas, issue body sin design y contexto bajo control

## Resultado

**done.** Las correcciones resuelven el bloqueo anterior (criterio de tipo
contradictorio) sin regresiones en el resto del alcance.

## Correcciones verificadas

- `specs.md`: máximo de requisitos único **30** (sin «25–30»); criterio de tipo
  «documentación/contratos o cambio pequeño y localizado, **sin cambio de
  comportamiento ni de arquitectura**; cualquier cambio de comportamiento o
  arquitectura es Feature, aunque el diff sea pequeño».
- `task.md`: misma conjunción («siempre que no cambie el comportamiento ni la
  arquitectura») y remisión a Feature.
- `workflow.md`: ejemplos de Task solo sin cambio observable; «modificaciones del
  comportamiento existente, por menores que sean» y «pequeños ajustes de interfaz
  que cambian lo que el usuario ve o recibe» pasan a Feature; «cambios menores de
  comportamiento» y «pequeños ajustes de interfaz» ya no aparecen como Task.
- `leader.md`: `task` exige «sin cambio de comportamiento ni de arquitectura»;
  `feature` incluye cualquier cambio de comportamiento.
- Sin coincidencias de «25–30» ni de ejemplos contradictorios en docs/roles.

## Verificaciones ejecutadas

- `V1`: `./bin/rei check --quiet` → exit 0.
- `gofmt -l internal cmd` → sin salida; `go vet ./...` → OK.
- `V2`: `go test ./...` → OK; tests de presupuesto (`WithinLimit`,
  `ExceededWarns`, `TaskPlanBudgetWarns`, `DoneNoWarn`) PASS.
- `V3`: `make build` → OK.
- `V4`: no aplica (sin código específico de plataforma).
- `rei init opencode --check` y `rei init claude --check` → todos `[OK]`; los
  adaptadores reflejan el `leader.md` corregido.
- `rei doctor` → exit 0; `rei init status` → exit 0, documentación personalizada.
- `rei validate 2026-10-26_00-24__specs-concisas-contexto` → OK, sin WARN de
  presupuesto (`plan.md` de 118 líneas).
- Reproducción CLI (revisión previa, código sin cambios): `plan.md` de 200 líneas
  → OK/exit 0; 201 → `WARN` con presupuesto y «Justifica o divide», exit 0;
  `done` con 201 → sin WARN.
- `plan.md`: 118 líneas, 9/9 pasos; `impl.md` actualizado.

## Sin regresiones

- Presupuesto: `validate.go` mantiene 30/400/15/200, `budgetApplicable` (`ready`,
  `in_progress`, `review`, `changes_requested`), mensajes con presupuesto y
  siempre `LevelWarn`.
- Issue body: `templates.md` §1/§2 sin design; §7 con comentario opcional
  idempotente `<!-- rei:comment:design:<id> -->`; `core.md`/`backlog.md`/
  `github.md` coherentes; comandos, labels, marcadores existentes y flujo intactos.
- Contexto: `AGENTS.md` §9 y `progress.md` con no-history/no-scan; los cuatro
  roles lo incorporan; el Reviewer verifica el presupuesto.
- Sin cambios en `skeleton/`, dependencias, specs históricas ni `history.md`.

## Observaciones

- Ninguna bloqueante. El artefacto `plan.md` accidental en la raíz del repo se
  eliminó durante esta revisión; no formaba parte del cambio.
