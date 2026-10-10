# Implementation — Visor de especificaciones en el navegador

> Work Item: `2026-10-09_23-45__spec-browser-viewer` (feature).
> Implementación completa siguiendo `tasks.md` (T1–T17, R1–R44).

## Resumen

Se añade el comando `rei view [<id>] [--open|--stop|--enable|--disable]`, que
sirve en el navegador una vista enriquecida del Work Item (metadatos + documentos
renderizados + validación + sesión), con recarga en vivo por SSE y botón Aprobar.
El Markdown se renderiza con `goldmark` en Go (modo seguro), y la aprobación
reutiliza la transición validada de `rei status set` sobre `meta.json`,
alineando `current.md`. El único artefacto nuevo es un descriptor temporal de
servidor, gitignorado y con permisos `0600`.

## Archivos

**Nuevos**

- `internal/browser/browser.go` + `browser_test.go`: `Open(url)` y costura `command(goos,url)`.
- `internal/view/view.go`: `Options`/`Run` y dispatch de modos; `SIGINT` en primer plano.
- `internal/view/server.go`: HTTP 127.0.0.1:0, token por instancia, cookie
  `HttpOnly; SameSite=Strict`, rutas `/`, `/content`, `/events` (SSE), `/approve`, `/static/*`.
- `internal/view/render.go`: `goldmark` seguro, `viewModel`, `html/template`.
- `internal/view/lifecycle.go` + `process_unix.go`/`process_windows.go`:
  descriptor, spawn `--serve` con detach, `--stop`, liveness por SO, autocierre.
- `internal/view/assets/{page.html,app.js,style.css}`: UI y JS propio (SSE + Aprobar).
- `internal/view/*_test.go`, `internal/config/config_test.go`.

**Modificados**

- `internal/update/update.go`: `OpenBrowser` por defecto = `browser.Open`.
- `internal/config/config.go`: `Viewer`/`Config.Viewer` + `SetAutoOpenOnReady`.
- `internal/state/state.go`: `Transition`, `ErrTransitionNotAllowed` y alineación de `current.md`.
- `internal/cli/cli.go`: `case "view"` + `cmdView`; `cmdStatus` usa `state.Transition`.
- `internal/cli/help.go`: entrada `view`.
- `.rei/config.json`, `.gitignore`, `.rei/agents/leader.md`, `.rei/agents/initializer.md`.
- `go.mod`/`go.sum`: `github.com/yuin/goldmark v1.8.6` (sin transitivas).

**Eliminado**

- `internal/update/browser.go` (movido a `internal/browser`).

## Cambios por fase

- **F1:** `browser.Open` reutilizable; `viewer.autoOpenOnReady` con default `false`;
  `state.Transition` compartida (`ready -> in_progress`, `--force` intacto).
- **F2:** `goldmark.New()` sin `WithUnsafe` (el HTML crudo se neutraliza); orden y
  rutas de documentos iguales a `rei item show`; ausentes marcados "(no existe)".
- **F3:** servidor local con token/cookie; SSE por polling de `mtime` limitado a
  los documentos del Work Item; `/approve` llama a `state.Transition`; descriptor
  `0600` con liveness; `--open` reutiliza o relanza el binario desacoplado;
  autocierre al salir de `ready` (watcher dedicado del servidor de fondo) y por
  inactividad.
- **F4:** dispatch de `view` (modos excluyentes, uso incorrecto → 2; sin sesión → 1);
  ayuda con modos y códigos; contratos de Leader (disparo por CLI, no bloqueante)
  e Initializer (pregunta de configuración).
- **F5:** validación final.

## Verificación (V: checkpoints de T16)

- `gofmt -l internal`: sin salida.
- `make vet`: OK. `make test`: todas las suites `ok`. `make build`: OK.
- `rei check --quiet`: `exit 0`.
- Manual (proyecto temporal): `view <id>` sirve y bloquea hasta Ctrl+C (`exit 0`);
  SSE emite `data: reload` al editar un documento; `--open` arranca en segundo
  plano y repetirlo reutiliza el mismo `pid`; botón/`POST /approve` aplica
  `ready -> in_progress` y alinea `current.md`; `--stop` termina y borra el
  descriptor; `--enable`/`--disable` solo cambian `viewer.autoOpenOnReady`;
  al salir de `ready` el servidor se autocierra (~1 s); token inválido → `403`.
- Tests nuevos cubren render seguro, composición, `httptest` (token/cookie,
  `/content`, `/approve` permitido y denegado, SSE) y descriptor.

## Observaciones

- El sistema no tenía `go` instalado; se instaló `go1.27.2` en
  `~/.local/go-toolchain/go` (fuera del repo) para poder verificar.
- El autocierre por estado se reforzó con un watcher del servidor de fondo, para
  no depender de que haya un cliente SSE conectado (R30).
- No se regeneraron los agentes nativos `.opencode/`/`.claude/` (fuera del alcance del diseño).

## Rework (revisión)

Revisión `changes_requested`: `tasks.md` no estaba al día (T9–T12 sin `[x]`). Acciones aplicadas:

- **Requerido:** marcadas `[x]` T9, T10, T11 y T12 en `tasks.md` (código presente y tests verdes). T17 queda satisfecha.
- **Cosmético (opcional):** en `internal/cli/cli.go` separado el comentario doc de `cmdUpdate` del de `cmdView` (cada uno adosado a su función).

Re-verificado todo verde: `gofmt -l internal` sin salida, `make vet` OK, `make test` todas las suites `ok`, `rei check --quiet` exit 0, `rei validate <id>` OK.

