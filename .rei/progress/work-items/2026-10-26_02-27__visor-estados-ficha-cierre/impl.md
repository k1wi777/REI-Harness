# Implementación — Visor: estados reales, FICHA con enlaces y cierre en done

## Resumen

Se ejecutaron T1–T11 según `design.md`: etiquetas corregidas, instantánea SSE
ampliada a `meta.json`/`links.json`, FICHA de enlaces con `internal/links`,
evento `final` con confirmación `POST /ack` y cierre centralizado en
`watchStatus`. Sin contradicciones con el diseño.

## Archivos modificados

- Código: `internal/view/render.go`, `internal/view/server.go`,
  `internal/view/assets/{page.html,app.js,style.css}` y `internal/cli/help.go`.
- Tests: `internal/view/render_test.go`, `internal/view/server_test.go` y
  `internal/cli/help_test.go`.
- Docs: `README.md`, `.rei/docs/usage.md`.
- Planificación: `tasks.md` (T1–T11 marcadas).

## Cambios

- **T1** `statusLabel`: `ready` → «en aprobación», `done` → «terminado»; badge y
  FICHA siguen usando la misma función. Test de tabla con los 7 estados y un
  valor desconocido.
- **T2** `Server.snapshot` añade `meta.json` y `links.json`; se mantienen el
  sondeo por `mtime` y el evento `reload` existente.
- **T3/T4** `buildModel` carga `links.json` con `links.Read` y `links.Problems`.
  Ficha válida: issue (#/URL), base (`branch@sha`) y épica (slug/#/URL); base o
  épica nulas → «sin registrar»/«sin épica». Ausente → «sin links.json»;
  ilegible o inválido → «links.json inválido» con problemas y sin campos. Las
  URL solo se enlazan en fichas válidas; sin `template.URL`.
- **T5/T7** `app.js`: `captureState`/`restoreState` compartidos por `reload` y
  `final` (scroll lector/índice/FICHA, diffs abiertos y hash). El evento `final`
  se registra con `addEventListener`, aplica el HTML, cierra el `EventSource` y
  envía `POST /ack`; se inhiben reconexión y «Reconectando…».
- **T6/T7** `watchStatus` marca `done` una vez (`sync.Once`) y es el único que
  cierra: espera ack o gracia (2 s) y respeta un vaciado mínimo (200 ms). Cada
  conexión SSE emite `event: final` con `status`, `label` y HTML de `#content`
  renderizado en done. `handleEvents` ya no cierra. `/ack` valida token/cookie,
  responde 204 y no muta nada. En primer plano no se emite `final` ni se cierra.
- **T8/T9** Tests de etiquetas y FICHA (válida completa, sin base/épica,
  ausente, JSON roto, campos inválidos), `reload` por estado y enlaces, secuencia
  final → ack → cierre, gracia sin ack, primer plano y `/ack` sin mutaciones.
- **T10** Ayuda de `view`, `README` y `usage.md` describen etiquetas, FICHA de
  enlaces y cierre tras «terminado».

## Verificación

- `gofmt -l .` → sin salida; `go vet ./...` → OK.
- `go test ./...` → ok en todos los paquetes.
- `make build` → OK; `./bin/rei check --quiet` → 0.
- `./bin/rei init opencode --check` y `claude --check` → OK; `./bin/rei doctor`
  → «Sin fallos»; `init status` → personalizado; `./bin/rei validate <id>` → OK.
- Prueba manual con el binario y `view --serve` (entorno sin GUI): `/content`
  mostró badge y FICHA «en curso», «sin links.json», `/approve` → 404 y `/ack`
  sin token → 403. Al pasar a `review`, el SSE emitió `data: reload`, la FICHA
  pasó a «en revisión» y el servidor siguió vivo; `view --stop` lo detuvo.
- El flujo `done` → `final` → ack → cierre y la gracia se validan en
  `TestFinalAckCierre` y `TestCierrePorGraciaSinAck` (no se cambió este Work Item
  a `done`).
- `git diff --name-only` no toca `skeleton/`, `.rei/config.json` ni
  `go.mod`/`go.sum`.

## Observaciones

- `internal/links`, `internal/state`, `internal/meta` y el flujo de aprobación
  quedan intactos; el visor sigue siendo de solo lectura.
- Sin dependencias nuevas y sin commit/push.
