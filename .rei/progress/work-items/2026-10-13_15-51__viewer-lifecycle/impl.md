# Implementación — Visor: vivo tras aprobar y recomendar auto-open

## Resumen

- El servidor de fondo del visor ya **no se cierra al aprobar**: sobrevive a
  `ready → in_progress → review` y solo termina al llegar a `done` (además de
  `rei view --stop` y el idle timeout).
- El Initializer ahora **recomienda mantener activa** la apertura automática y
  el esqueleto `.rei/config.json` usa `true` por defecto.

## Archivos modificados

- `internal/view/server.go` — cierre solo en `done`; `closeOnStatusChange` →
  `closeOnDone`; comentarios actualizados.
- `internal/view/lifecycle.go` — `s.closeOnDone = true` y comentario de
  `runServe`.
- `internal/view/server_test.go` — nuevo `TestFondoSeCierraEnDone`.
- `.rei/config.json` — `viewer.autoOpenOnReady: true`.
- `.rei/agents/initializer.md` — § «Configuración del visor» reescrita.
- `.rei/specs/2026-10-13_15-51__viewer-lifecycle/plan.md` — pasos marcados.

## Cambios

1. **Cierre del servidor.** En `watchStatus` y en la rama SSE de `handleEvents`
   la condición `m.Status != meta.StatusReady` pasa a `m.Status == meta.StatusDone`.
   El flag interno se renombra a `closeOnDone` en `Server`, `Serve` y
   `runServe`. `--stop` e idle timeout intactos; reconexión SSE sin tocar.
2. **Test de ciclo de vida.** `TestFondoSeCierraEnDone` arranca un servidor real
   (`Listen` + `Serve` en goroutine) con `background`, `closeOnDone` y
   `pollInterval = 10ms`. Verifica que `in_progress` y `review` no cierran
   (`s.closed` abierto) y que `done` sí cierra y `Serve` retorna sin error.
3. **Default y recomendación.** `.rei/config.json` despliega `true`. La sección
   del Initializer conserva la pregunta, recomienda activa, usa `true` si no hay
   preferencia y respeta un `false` explícito; mismo alcance de escritura (solo
   `viewer.autoOpenOnReady`, preservando el resto).

Se mantiene el fallback compilado de `config.Load` en `false` cuando falta la
clave, según la decisión del plan (sin consumidor en runtime; coherente con
`leader.md`, que no se toca).

## Verificación

- `gofmt -l .` — sin salida.
- `go vet ./...` — OK.
- `go test ./...` — OK (incluye `TestFondoSeCierraEnDone`).
- `make build` — OK.
- `rei check --quiet` — exit 0.

## Observaciones

- Tras `done` la página pasa a «Reconectando…»: es el comportamiento esperado
  (trabajo terminado), no un fallo.
- El toolchain Go no estaba en `PATH` del entorno; la verificación se ejecutó
  con un Go 1.27.3 temporal en `/tmp` (sin cambios en el repositorio).
