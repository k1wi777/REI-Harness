# Revisión — Visor: estados reales, FICHA con enlaces y cierre en done

## Resultado

**Aprobado.** La implementación cumple `requirements.md` (R1–R16), el diseño y
las 11 tareas de `tasks.md`; el diff corresponde a lo planificado.

## Verificaciones

- **Trazabilidad**: T1–T11 marcadas; el diff toca exactamente los archivos del
  diseño (`render.go`, `server.go`, `page.html`, `app.js`, `style.css`,
  `render_test.go`, `server_test.go`, `help.go`, `help_test.go`, `README`,
  `usage.md`). Sin cambios en `skeleton/`, `.rei/config.json`, `go.mod`/`go.sum`
  ni dependencias nuevas.
- **R1/R2**: `statusLabel` traduce los siete estados (`ready` = «en aprobación»,
  `done` = «terminado») y deja el desconocido sin traducir; badge y fila de la
  FICHA usan la misma función. `TestStatusLabel` cubre los 8 casos.
- **R3/R4**: `snapshot` vigila `meta.json` y `links.json` además de los
  documentos; el `reload` existente refresca badge y FICHA dentro de `#content`.
  `TestEventsEmiteReloadAlCambiarEstadoYEnlaces` lo prueba por estado y enlaces.
- **R5**: `captureState`/`restoreState` compartidos entre `reload` y `final`
  (scroll lector/índice/FICHA, diffs abiertos, documento activo y hash).
- **R6–R8**: FICHA con issue (#/URL), base (`branch@sha`) y épica
  (slug/#/URL) desde `links.Read`/`links.Problems`; ausente → «sin links.json»;
  inválido → nota con problemas y sin campos; URL solo en fichas válidas y
  validada http/https; sin `template.URL`. Cubierto por `TestFichaEnlaces`
  (válida, sin base/épica, ausente, JSON roto, campos inválidos).
- **R9–R14**: `watchStatus` es el único cierre por `done` (`markDone` con
  `sync.Once`); cada conexión SSE emite `event: final` con `status`, `label` y
  HTML renderizado; `POST /ack` autorizado, 204 y sin escrituras; cierre tras
  ack o gracia (2 s) con vaciado mínimo (200 ms); `handleEvents` ya no cierra; en
  primer plano no hay `final` ni cierre. Cubierto por `TestFinalAckCierre`,
  `TestCierrePorGraciaSinAck`, `TestPrimerPlanoNoCierraEnDone` y
  `TestAckAutorizadoSinMutaciones`; `/approve` sigue en 404.
- **R15**: `README`, `.rei/docs/usage.md` y ayuda de `view` describen etiquetas,
  FICHA y cierre; `help_test` lo verifica.
- **Checkpoints**: `V1` `go vet ./...` OK; `V2` `go test ./...` OK (además
  `go test -count=1 ./internal/view/... ./internal/cli/...` OK); `V3`
  `make build` OK; `V4` no aplica (sin código por plataforma ni release).
  `gofmt -l .` sin salida; `./bin/rei check --quiet` → 0;
  `rei init opencode/claude --check` OK; `rei doctor` sin fallos;
  `rei init status` personalizado; `rei validate <id>` OK sin `WARN` de
  presupuesto.
- **Manual** (copia desechable del repo, `view --serve`): `ready` → «en
  aprobación» con FICHA #12/`main@01234567`/épica y 2 enlaces; `/approve` → 404;
  `ready → in_progress → review` emitió `data: reload` y el badge/FICHA pasaron a
  «en curso» y «en revisión»; `view --stop` detuvo el servidor; `done` emitió
  `event: final` (`status: done`, `label: terminado`, HTML con la FICHA), el
  servidor seguía sirviendo antes del ack, `POST /ack` → 204 y cerró después.

## Observaciones

- Quedaba un servidor de fondo huérfano de la prueba manual del Implementer
  (PID 158911, sin descriptor). Se detuvo; no afecta al código.
- En primer plano el estado `done` llega por `reload` (sin `final`), conforme a
  R13 y al diseño.
- `App.js` cae a recarga completa solo si el HTML final no se puede aplicar,
  dentro de la gracia: comportamiento previsto en D5.
