# Plan — Visor: vivo tras aprobar y recomendar auto-open

## Objetivo

1. **El servidor de fondo del visor NO debe cerrarse al aprobar.** Hoy se cierra
   en cuanto el Work Item deja de estar `ready` (R30 de la v1), así que al
   aprobar la página queda en «Reconectando…» y ya no deja ver nada. Debe seguir
   vivo mientras el Work Item avanza (`in_progress`, `review`) y cerrarse **solo**
   cuando llegue a `done`, además del `rei view --stop` explícito y el idle
   timeout ya existentes. Así se pueden ver los `impl.md`/`review.md` que se
   generan después.
2. **El Initializer debe recomendar mantener ACTIVA la apertura automática del
   visor** (hoy recomienda dejarla desactivada), arguyendo que hace más fácil
   leer las specs. Se conserva la pregunta (la decide el usuario). El valor por
   defecto del esqueleto pasa a `true`.

## Archivos

| Archivo | Cambio |
|---------|--------|
| `internal/view/server.go` | Condición de cierre del servidor de fondo (`!= ready` → `== done`). |
| `internal/view/lifecycle.go` | Nombre del flag al arrancar el servidor de fondo. |
| `internal/view/server_test.go` | Test del nuevo ciclo de vida. |
| `.rei/config.json` | Default del esqueleto: `autoOpenOnReady` → `true`. |
| `.rei/agents/initializer.md` | Recomendación del visor: activa por defecto. |

## Cambios

### 1. El servidor deja de cerrarse al aprobar

Hoy el servidor de fondo se autocierra en cuanto el Work Item deja de estar
`ready`. Hay **dos** puntos que lo hacen, ambos en `internal/view/server.go`:

- `watchStatus()`: `m.Status != meta.StatusReady` → `requestClose()`.
- `handleEvents()` (bucle SSE): `s.background && s.closeOnStatusChange` y
  `m.Status != meta.StatusReady` → `requestClose()`.

Cambiar **ambas** condiciones a `m.Status == meta.StatusDone`. Con esto el
servidor sobrevive a `ready → in_progress → review` y se cierra al alcanzar
`done`. Se mantienen intactos:

- `rei view --stop` (`stopViewer`, `internal/view/lifecycle.go`).
- El idle timeout (`monitor`, 30 min de inactividad).

Detalle:

- Renombrar el campo interno `Server.closeOnStatusChange` a `closeOnDone` (el
  nombre actual describe un comportamiento que ya no existe). Afecta a
  `Server` (declaración), `Serve` (arranque del watcher) y `handleEvents`
  (bloque SSE), y a `runServe` en `lifecycle.go` (`s.closeOnStatusChange = true`
  → `s.closeOnDone = true`).
- Actualizar los comentarios de `watchStatus`, `handleEvents`, `Serve` y
  `runServe` para describir la nueva condición terminal (`done`).

**Consecuencia esperada, no es un bug:** tras `done`, el servidor se cierra y la
página pasará a «Reconectando…»; es el comportamiento correcto porque el trabajo
terminó. No hay que «arreglarlo» ni tocar la reconexión SSE.

### 2. Default del visor a `true` y recomendación del Initializer

**(a) `.rei/config.json`** (esqueleto que despliega `rei init`): cambiar
`"viewer": { "autoOpenOnReady": false }` → `true`.

**(b) `.rei/agents/initializer.md`**, sección «Configuración del visor»:
conservar la pregunta, pero **recomendar mantenerla activa** (abrir el visor al
aprobar facilita leer las specs en el navegador en vez de hacerlo en crudo). Si
el usuario no tiene preferencia, dejar `true` (recomendado); respetar un `false`
explícito. Se sigue escribiendo `viewer.autoOpenOnReady` en `.rei/config.json`
preservando el resto del archivo (mismo alcance de escritura que hoy).

**Evaluación del `default` compilado (decisión):** el default *efectivo* es el
del esqueleto `.rei/config.json`, porque es lo que `rei init` despliega y lo que
el Leader lee textualmente. El fallback de `internal/config/config.go` cuando
**falta** la clave (valor cero `false`) **se deja como está**: no tiene
consumidor en runtime (solo los tests leen `Viewer.AutoOpenOnReady`), coincide
con el fallback conservador de `leader.md` («o la clave no existe») y preserva
R34. Cambiarlo a `true` exigiría un refactor a `*bool` sin efecto en runtime y
crearía una divergencia con `leader.md`. Se mantiene `leader.md` sin tocar.

## Restricciones

- Sin dependencias nuevas.
- No romper el resto del comportamiento del visor: selector de documento,
  aprobación en dos pasos, TOC, diff, enlaces profundos, scroll y reconexión SSE.
- `rei view --stop` y el idle timeout siguen siendo condiciones de cierre válidas.
- No tocar `leader.md` (fuera del alcance de este Work Item).

## Pasos

- [x] 1. `internal/view/server.go`: cambiar las dos condiciones de cierre
  (`watchStatus` y rama SSE de `handleEvents`) a `m.Status == meta.StatusDone`.
- [x] 2. `internal/view/server.go` + `internal/view/lifecycle.go`: renombrar
  `closeOnStatusChange` → `closeOnDone` y actualizar los comentarios.
- [x] 3. `internal/view/server_test.go`: añadir un test del ciclo de vida con
  servidor de fondo (`background` + `closeOnDone`, `pollInterval` corto):
  `ready → in_progress` NO cierra; `→ done` SÍ cierra.
- [x] 4. `.rei/config.json`: `viewer.autoOpenOnReady` a `true`.
- [x] 5. `.rei/agents/initializer.md`: reescribir «Configuración del visor»
  para recomendar activa y usar `true` por defecto.
- [x] 6. Verificación: `gofmt -l .`, `go vet ./...`, `go test ./...`, `make build`.
