# Revisión — Workflow integrity: visor read-only y un solo `in_progress`

## Resultado

**Aprobado (`done`).** La planificación se cumple sin desviaciones: el visor es
de solo lectura y el invariante de un único `in_progress` se aplica en
`state.Transition`, `doctor`, `check` e `items status`, con tests y
documentación coherente.

## Verificaciones

- **Trazabilidad (12/12 pasos).** El diff coincide con `plan.md`: `state`,
  `cli`, `doctor`, `check`, `view` (Go, HTML, JS, CSS), tests y documentación.
  Sin cambios fuera de alcance (no se tocó `skeleton/`).
- **Visor de solo lectura.** No quedan en `internal/` (fuera de aserciones de
  ausencia en tests) `handleApprove`, `/approve`, `CanApprove`, `writeJSON` ni
  selectores/tokens de aprobación; `app.js` conserva `setNotice`, SSE, índice y
  enlaces profundos. En vivo (`rei view`): `GET /` 200, `GET /content` 200,
  `/static/*` 200 y **`POST /approve?token=…` → 404** con `meta.json` intacto
  (`review`). `TestFondoSeCierraEnDone` mantiene el cierre en `done`.
  `--enable/--disable/--stop` OK (0) y modos mixtos 2.
- **Invariante en proyecto temporal.** A `ready→in_progress` permitido; B
  rechazado con `ERROR: … ya está en curso; ciérralo (review/done), bloquéalo o
  revísalo…` y exit 1, **también con `--force`**, sin cambiar su `meta.json`.
  Con dos `in_progress` simulados: `items status` exit 1 (mensaje a stderr),
  `doctor` y `check` (`--quiet` incluido) FAIL con los ids. Con uno: los tres OK
  y sin `AVISO`; con cero: OK. Proyecto recién inicializado: `check`/`doctor`/
  `items status` = 0.
- **Tests.** Cubren rechazo/permitido, `--force`, `changes_requested`,
  ausencia de escrituras en `meta.json`/`current.md`, `items status` 0/1/2,
  `doctor` 1/2, `check` 1/2 y `/approve` 404.
- **Documentación.** `leader.md` (visor solo lectura, `rei status set`, regla
  dura), `workflow.md`, `usage.md` §4/§6, `README.md` (fila `rei view`) y
  `architecture.md` coherentes; adaptadores nativos regenerados y `--check` OK.
  `skeleton/` intacto.
- **Checkpoints.** V1 `go vet ./...` OK; V2 `go test -count=1 ./...` OK; V3
  `make build` OK; V4 no aplica (sin código por plataforma). `gofmt -l internal
  cmd` sin salida. Sin dependencias nuevas (`go.mod`/`go.sum` sin cambios).
- **CLI.** `rei check --quiet` 0; `init opencode/claude --check` 0; `doctor` 0;
  `init status` 0; `validate <id>` OK.

## Observaciones

- `DESIGN.md` (raíz, fuera del alcance del plan) aún describe el botón primario
  «Aprobar» y los tokens asociados; conviene limpiarlo en un Work Item de
  documentación de diseño. No afecta al criterio de aceptación de esta Task.
- Efecto lateral preexistente detectado durante la revisión: `rei view
  --enable/--disable` reescribe `.rei/config.json` con otro formato; se restauró
  y no forma parte de este cambio.

## Acciones requeridas

Ninguna.
