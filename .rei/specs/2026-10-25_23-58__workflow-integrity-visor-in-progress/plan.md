# Plan — Workflow integrity: visor read-only y un solo `in_progress`

## Objetivo

Dos correcciones de integridad del workflow:

1. **Visor de solo lectura.** Eliminar de `rei view` toda capacidad de aprobar
   (ruta `POST /approve`, botón, JS, `CanApprove`, CSS y tests). La aprobación
   se registra únicamente por terminal/chat (`rei status set <id> in_progress`
   ejecutado por el Leader) y, en Features, sigue pasando por la puerta
   GitHub (issue + rama). El visor conserva documentos, metadatos, validación,
   sesión, TOC, diff, SSE y los modos `--open/--stop/--enable/--disable`.
2. **Un solo `in_progress`.** `state.Transition` rechaza cualquier transición a
   `in_progress` cuando ya hay otro Work Item en ese estado. El invariante es
   duro (`--force` no lo omite), con causa y siguiente paso en el error.
   `rei doctor` y `rei check` fallan si hay más de uno; `rei items status`
   devuelve exit 1 y deja de avisar cuando hay exactamente uno.

## Archivos

| Archivo | Cambio |
|---|---|
| `internal/state/state.go` | Nuevo error centinela + helper de consulta + verificación del invariante en `Transition`. |
| `internal/state/state_test.go` | Tests de rechazo/permitido, `changes_requested`, `--force` y ausencia de escrituras parciales. |
| `internal/cli/cli.go` | `cmdItems`: exit 1 con >1 `in_progress`, sin aviso con exactamente 1. `cmdStatus`: mensaje del nuevo error con siguiente paso. |
| `internal/cli/cli_test.go` | Tests de `items status` (0/1/2 activos) y del mensaje de rechazo de `status set`. |
| `internal/doctor/doctor.go` | FAIL si hay más de un `in_progress` (lista de ids). |
| `internal/doctor/doctor_test.go` | Test de uno (OK) y de dos (fallo). |
| `internal/check/check.go` | FAIL si hay más de un `in_progress`. |
| `internal/check/check_test.go` | Tests de uno (OK) y de dos (fallo). |
| `internal/view/server.go` | Quitar ruta y handler `/approve`, `writeJSON` y el import de `state`/`encoding/json` si quedan sin uso. |
| `internal/view/render.go` | Quitar `CanApprove` del modelo y su inicialización. |
| `internal/view/assets/page.html` | Quitar el bloque `<div class="action" id="approve-area">` completo. |
| `internal/view/assets/app.js` | Quitar estado/funciones/handlers de aprobación y `setActionStatus`; actualizar el comentario de cabecera. |
| `internal/view/assets/style.css` | Quitar selectores `.action*`, `.approve-note`, `.mono` y tokens que queden sin uso (`--on-warn`, `--warn-hover`). |
| `internal/view/view.go` | Actualizar el comentario de paquete (visor de solo lectura). |
| `internal/view/render_test.go`, `internal/view/server_test.go` | Quitar aserciones de aprobación; añadir aserciones de ausencia de ruta/botón. |
| `.rei/agents/leader.md` | Caso A (sin botón), reglas duras (invariante `in_progress`) y, si aplica, Caso C con `rei status set`. |
| `.rei/docs/harness/workflow.md` | Reforzar la regla de un solo `in_progress`: quién lo aplica y qué hacer. |
| `.rei/docs/usage.md` | §4: la aprobación es por chat/terminal; el visor no aprueba. §6: el límite se aplica y se rechaza. |
| `README.md` | Visor descrito como solo lectura (y fila `rei view` en la tabla de comandos si falta), sin capacidad de aprobar. |
| `.rei/docs/project/architecture.md` | Corregir «la única mutación (Aprobar)…» y las referencias al botón/transición desde el visor. |
| `.opencode/agents/leader.md`, `.claude/agents/leader.md` | Regenerados desde el contrato con `rei init opencode` / `rei init claude` (no editar a mano). |

## Cambios

### B. Un solo `in_progress` (núcleo)

- En `internal/state`:
  - `var ErrAnotherInProgress = errors.New("ya existe otro Work Item en in_progress")`.
  - Helper exportado (p. ej. `func InProgressItems(p *paths.Project) ([]string, error)`)
    que filtre `ListWorkItems` por `meta.StatusInProgress`; lo reutilizan
    `doctor`, `check` y `cli` para no duplicar el criterio.
- `Transition(p, id, to, force)`:
  - Tras el early-return `from == to` y **antes** de validar la transición o de
    escribir, si `to == meta.StatusInProgress` consulta el helper; si hay algún
    id distinto de `id`, devuelve
    `fmt.Errorf("%w: %s", ErrAnotherInProgress, otroID)` sin tocar `meta.json`
    ni `current.md` (ni con `force == true`).
  - El error debe ser identificable con `errors.Is` y contener causa (id activo)
    y siguiente paso a nivel de CLI.
- `cmdStatus` (`internal/cli/cli.go`): comprobar `errors.Is(err, state.ErrAnotherInProgress)`
  **antes** del caso genérico y emitir:
  `ERROR: no se puede pasar <id> a in_progress: <activo> ya está en curso; ciérralo (review/done), bloquéalo o revísalo antes de continuar.` exit 1.
  `--force` sigue funcionando solo para transiciones no estándar, nunca para
  este invariante.
- `doctor`: contador de ids `in_progress`; si `> 1`, `fail("Hay %d Work Items en in_progress: %s. Solo puede haber uno.", n, ids)`.
  Con 0 o 1 no se añade aviso.
- `check.Run`: nuevo bloque tras la estructura (p. ej. «2b.») con
  `state.InProgressItems`; `> 1` → `fail(...)`; `<= 1` → `ok("Un solo Work Item en in_progress.")`
  (el `ok` desaparece con `--quiet`; el `fail` no). No tocar `RunDeclared` (`rei test`).
- `cmdItems`: `active > 1` → mensaje `ERROR: hay %d Work Items en in_progress (%s); solo puede haber uno.` a stderr y `return 1`;
  `active == 1` → se elimina el `AVISO` actual; se conserva el marcador `<< ACTIVO`.

### A. Visor de solo lectura

- `internal/view/server.go`: eliminar `mux.HandleFunc("/approve", ...)`,
  `handleApprove`, `writeJSON` y los imports que queden sin uso (`encoding/json`,
  `state`; `meta` se conserva por `watchStatus`).
- `internal/view/render.go`: quitar el campo `CanApprove` y su asignación en
  `buildModel`; el resto del modelo no cambia.
- `internal/view/assets/page.html`: quitar el bloque `approve-area` (incluida la
  sección «Aprobación»); conservar «Ficha» y el colofón.
- `internal/view/assets/app.js`: quitar `approveState`, `resetApprove`,
  `onApproveClick`, `doApprove`, `setActionStatus`, los dos `closest("#approve*")`
  y actualizar la cabecera (solo SSE, índice, enlaces profundos). `setNotice` se
  mantiene (lo usa el SSE).
- `internal/view/assets/style.css`: quitar `.action`, `.approve-note`,
  `.approve-note .mono, .mono`, `.action button*`, `.action.confirming*`,
  `.action-status*` y la regla de `.action button` en el media query; retirar
  `--on-warn` y `--warn-hover` si quedan sin uso (`--warn` se conserva por
  `status-ready` y `issue-WARN`).
- `internal/view/view.go`: comentario de paquete sin «botón Aprobar».
- Tests de `internal/view`:
  - `server_test.go`: borrar `TestApprovePermitido` y `TestApproveDenegado`;
    añadir una aserción de que `POST /approve?token=...` responde `404`
    (ruta inexistente) y que no cambia `meta.json`.
  - `render_test.go`: quitar `CanApprove` y las aserciones de `id="approve"`;
    verificar que la página/fragmento siguen incluyendo documentos, validación,
    sesión, TOC y colofón, y que **no** contienen `id="approve"` ni `/approve`.
  - `TestFondoSeCierraEnDone` se mantiene (cambia el estado por `meta.SetStatus`);
    solo ajustar comentarios que digan «aprobar».

### Documentación

- `.rei/agents/leader.md`: Caso A paso 7 sin «botón Aprobar» (visor opcional de
  solo lectura, no sustituye la aprobación del chat); en Reglas duras, que solo
  puede haber un `in_progress` y que el CLI lo rechaza (incluido `--force`);
  Caso C puede indicar `rei status set <id> in_progress` como forma canónica.
- `.rei/docs/harness/workflow.md`: en «Reglas generales», precisar que el
  invariante lo aplican `rei status set`, `rei check`, `rei doctor` e
  `rei items status`, y el siguiente paso cuando se rechaza.
- `.rei/docs/usage.md`: §4 aclarar que la aprobación se da en el chat/terminal y
  que el visor es de solo lectura (no hay botón de aprobación); §6 indicar que
  el límite de un `in_progress` se rechaza de forma dura.
- `README.md`: describir `rei view` como visor de solo lectura y añadir la fila
  `rei view [<id>]` a la tabla de comandos si no está.
- `.rei/docs/project/architecture.md`: línea del principio «El visor observa y
  reutiliza» (sin mutación), flujo de `rei view` (sin botón Aprobar), restricción
  «El visor no muta documentos» (ninguna transición desde el visor) y decisión
  sobre transiciones centralizadas (CLI como único emisor).
- Regenerar adaptadores con `rei init opencode` y `rei init claude`; verificar con
  `--check`. No tocar `skeleton/`.

## Restricciones

- Sin dependencias nuevas (solo biblioteca estándar; `goldmark` sigue siendo la
  única externa).
- No modificar `skeleton/` ni los documentos históricos de Work Items cerrados.
- No romper el visor de lectura: documentos, metadatos, validación, sesión, TOC,
  diff, SSE y `--open/--stop/--enable/--disable` deben seguir funcionando.
- El invariante no puede omitirse con `--force` ni dejar escrituras parciales
  (`meta.json` y `current.md` intactos al rechazar).
- No cambiar el alcance de `meta.json`.
- V4 solo si aparece código específico por plataforma (no se espera aquí).

## Pasos

- [x] 1. `internal/state/state.go`: añadir `ErrAnotherInProgress` y el helper de
  Work Items `in_progress`; aplicar el invariante en `Transition` antes de
  cualquier escritura y con independencia de `force`.
- [x] 2. `internal/state/state_test.go`: tests de rechazo con otro activo
  (incluido `force=true`), transición permitida sin otro activo,
  `changes_requested -> in_progress` permitido y rechazado, y verificación de que
  `meta.json`/`current.md` no cambian al rechazar.
- [x] 3. `internal/cli/cli.go`: `cmdItems` (exit 1 con >1, sin aviso con 1) y
  `cmdStatus` (mensaje específico con causa y siguiente paso para
  `ErrAnotherInProgress`).
- [x] 4. `internal/cli/cli_test.go`: tests de `items status` con 0, 1 y 2 activos
  (incluido que con 1 no se imprime aviso) y de `status set` rechazado.
- [x] 5. `internal/doctor/doctor.go` + `doctor_test.go`: FAIL con dos
  `in_progress` (ids en el mensaje) y OK con uno.
- [x] 6. `internal/check/check.go` + `check_test.go`: FAIL con dos `in_progress`
  (también con `--quiet`) y OK con uno.
- [x] 7. `internal/view/server.go` y `render.go`: quitar `/approve`,
  `handleApprove`, `writeJSON`, imports sin uso y `CanApprove`.
- [x] 8. `internal/view/assets/{page.html,app.js,style.css}`: quitar botón, JS de
  aprobación y CSS asociado; actualizar comentarios.
- [x] 9. `internal/view/view.go`: comentario de paquete de solo lectura.
- [x] 10. `internal/view/{server_test.go,render_test.go}`: eliminar tests de
  aprobación, añadir aserciones de ausencia de ruta/botón y mantener cobertura de
  documentos/validación/sesión/TOC/SSE.
- [x] 11. Documentación: `leader.md`, `workflow.md`, `usage.md`, `README.md`,
  `architecture.md`; regenerar adaptadores y comprobar con `--check`.
- [x] 12. Verificación final: `go test ./...`, `go vet ./...`, `make build`,
  `rei check`, `rei doctor`, `rei init status`, `rei validate`, `rei init opencode --check`,
  `rei init claude --check`; comprobar por búsqueda que no quedan referencias a
  `/approve` ni al botón en `internal/`.
