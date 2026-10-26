# Implementación — Workflow integrity: visor read-only y un solo in_progress

## Resumen

Ejecutado `plan.md` sin desviaciones ni contradicciones. El visor queda en solo
lectura (sin `/approve`, botón, JS, `CanApprove` ni CSS) y el invariante de un
único `in_progress` se aplica en `state.Transition`, `doctor`, `check` e
`items status`, con tests y documentación actualizada.

## Archivos modificados

- `internal/state/state.go`, `state_test.go`
- `internal/cli/cli.go`, `cli_test.go`
- `internal/doctor/doctor.go`, `doctor_test.go`
- `internal/check/check.go`, `check_test.go`
- `internal/view/server.go`, `render.go`, `view.go`, `server_test.go`, `render_test.go`
- `internal/view/assets/page.html`, `app.js`, `style.css`
- `.rei/agents/leader.md`, `.rei/docs/harness/workflow.md`, `.rei/docs/usage.md`,
  `.rei/docs/project/architecture.md`, `README.md`
- `.opencode/agents/leader.md`, `.claude/agents/leader.md` (regenerados)
- `.rei/specs/.../plan.md` (pasos marcados)

## Cambios

- **Invariante**: `ErrAnotherInProgress` y `state.InProgressItems`; `Transition`
  rechaza `-> in_progress` con otro activo antes de validar o escribir, también
  con `force`. Sin escrituras parciales en `meta.json` ni `current.md`.
- **CLI**: `status set` traduce el error con causa (id activo) y siguiente paso
  (cerrar/bloquear/revisar), exit 1; `items status` sale 1 con >1 activos, sin
  `AVISO` con uno y conserva `<< ACTIVO`.
- **doctor/check**: FAIL con >1 `in_progress` listando ids; con 0/1 no avisan
  (check emite un `ok` que `--quiet` oculta).
- **Visor**: eliminadas ruta/handler `/approve`, `writeJSON`, `CanApprove`, el
  bloque de aprobación de la página, el JS de aprobación y los selectores/tokens
  CSS asociados. Se mantienen documentos, validación, sesión, TOC, diff, SSE,
  `setNotice`, enlaces profundos, cierre en `done` y `--open/--stop/--enable/--disable`.
- **Tests**: rechazo/permitido/`changes_requested`/`force` y ausencia de cambios;
  `items status` 0/1/2 y `status set` rechazado; `doctor` y `check` con 1/2;
  `POST /approve?token=...` responde 404 y no muta `meta.json`; render sin
  `id="approve"` ni `/approve`, conservando documentos, validación, sesión, TOC y
  colofón.
- **Documentación**: Leader (visor solo lectura, `rei status set` y regla dura),
  workflow (invariante y quién lo aplica), usage (§4 y §6), README (fila
  `rei view`), arquitectura (sin mutación desde el visor, CLI único emisor).

## Verificación

- `gofmt -l .` sin salida; `go vet ./...` OK.
- `go test ./...` OK.
- `make build` OK.
- `./bin/rei check --quiet` = 0.
- `./bin/rei init opencode --check` = 0; `./bin/rei init claude --check` = 0.
- `./bin/rei doctor` = 0; `./bin/rei init status` = 0.
- `./bin/rei validate 2026-10-25_23-58__workflow-integrity-visor-in-progress` = OK.
- Búsqueda: no quedan referencias a aprobación en `internal/` salvo las
  aserciones de ausencia de los tests.

## Observaciones

- `skeleton/` no se ha tocado; no se añadieron dependencias.
- No hay código específico por plataforma, por lo que V4 no aplica.
- No se hizo commit, push, PR ni merge.
