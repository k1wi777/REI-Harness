# Tasks — Visor de especificaciones en el navegador

> Work Item: `2026-10-09_23-45__spec-browser-viewer` (feature).
>
> Orden de ejecución. Cada tarea indica los requisitos que implementa.
> El Implementer marca `[x]` **inmediatamente** al terminar cada tarea.

---

## Fase 1 — Fundaciones reutilizables

- [x] **T1 — Extraer `openBrowser` a `internal/browser`:** crear `internal/browser/browser.go` con `Open(url string) error` (misma lógica que `internal/update/browser.go`: `windows` → `cmd /c start`, `darwin` → `open`, resto → `xdg-open`). Añadir una costura interna para testear el comando elegido por `GOOS`. Eliminar `internal/update/browser.go` y cambiar el valor por defecto de `Options.OpenBrowser` en `internal/update/update.go` a `browser.Open`, sin tocar los tests de `update`. (R40)

- [x] **T2 — Extender `internal/config` con `viewer`:** añadir `Viewer{AutoOpenOnReady bool}` y el campo `Config.Viewer` (el valor cero da `false` cuando la clave falta, sin fallar). Implementar `SetAutoOpenOnReady(path string, enabled bool) error`: leer el JSON como `map[string]any`, fijar `viewer.autoOpenOnReady` y reescribir el mapa completo (indentado, `SetEscapeHTML(false)`, escritura atómica) preservando el resto. Tests: default `false` sin la clave; `--enable/--disable` solo cambian esa clave. (R33, R34, R35)

- [x] **T3 — Transición compartida en `internal/state`:** añadir `ErrTransitionNotAllowed` y `Transition(p, id, to, force) (from string, forced bool, err error)`, reutilizando `meta.IsValidStatus`/`meta.TransitionAllowed`/`meta.SetStatus`. Si el Work Item es la sesión activa, reescribir solo la línea `- **Estado:**` de `current.md`. Refactorizar `cmdStatus` en `internal/cli/cli.go` para usarla, conservando sus mensajes y `--force`. Tests de la transición válida, de la no permitida y de la alineación de `current.md`. (R22, R23, R24, R44)

## Fase 2 — Render

- [x] **T4 — Añadir `goldmark` a `go.mod`/`go.sum`** (`github.com/yuin/goldmark`) y verificar que no arrastra dependencias transitivas. (R40)

- [x] **T5 — Render con `goldmark` en `internal/view/render.go`:** configurar `goldmark.New()` **sin** `html.WithUnsafe()`; recorrer los documentos en el orden de `internal/show` reutilizando su resolución de rutas (`plan.md`, `requirements.md`, `design.md`, `tasks.md` en `.rei/specs/<id>/`; `impl.md`, `review.md`, `spec.md` en `.rei/progress/work-items/<id>/`); marcar los ausentes como "(no existe)". Componer la página con `html/template` y un `viewModel` (metadatos, sesión, documentos, validación, `CanApprove`). Reutilizar `validate.WorkItem` y `state.ReadSession`. Incluir `assets/page.html` y `assets/style.css` embebidos. (R10, R11, R12, R13, R14, R15, R16, R17)

## Fase 3 — Servidor y ciclo de vida

- [x] **T6 — Servidor HTTP en `internal/view/server.go`:** `net.Listen("tcp", "127.0.0.1:0")` (puerto efímero) y token aleatorio por instancia (`crypto/rand`, 32 bytes hex). Rutas: `GET /?token=` (valida el token, fija cookie `HttpOnly; SameSite=Strict; Path=/` y sirve la página), `GET /content`, `GET /events` (SSE), `POST /approve`, `GET /static/*`. El handler `/approve` llama a `state.Transition(p, id, "in_progress", false)` y responde JSON con el resultado (error → sin escribir `meta.json`), incluyendo el recordatorio de reanudar en el chat. Construir el servidor con reloj/estado inyectables para `httptest`. (R21, R22, R24, R25, R26, R27, R28)

- [x] **T7 — SSE de recarga en vivo:** en `/events`, un bucle por conexión hace polling de `mtime` de los documentos del Work Item mostrado cada ~1 s y emite `data: reload` al detectar cambios; limitar la vigilancia a esos documentos. (R19, R20)

- [x] **T8 — JS y estilos propios:** `assets/app.js` (abrir el `EventSource` de `/events`, al recibir `reload` re-fetch de `/content` y sustituir el contenido, y enviar `POST /approve` desde el botón con su feedback) y `assets/style.css`. Sin dependencias de terceros. (R18, R19, R21, R26)

- [x] **T9 — Ciclo de vida en `internal/view/lifecycle.go` (+ `process_unix.go`/`process_windows.go`):** descriptor temporal `.rei/.view-server.json` (`pid`, `port`, `token`, `workItem`, `startedAt`; permisos `0600`; escritura atómica) con liveness por SO. `--open`: reutilizar un servidor vivo o lanzar `os.Executable()` con `view <id> --serve` desacoplado (setsid/detach) y abrir el navegador con `browser.Open`. `--serve`: proceso de fondo que registra y elimina el descriptor. `--stop`: terminar el proceso y borrar el descriptor. Autocierre al salir de `ready` y por inactividad (constante, p. ej. 30 min). (R7, R8, R27, R29, R30, R31, R32, R43)

- [x] **T10 — Punto de entrada `internal/view/view.go`:** `Options`/`Run` con dispatch de modos (primer plano por defecto, `--open`, `--stop`, `--enable`, `--disable`, `--serve` interno), resolución del Work Item (id explícito o sesión activa), impresión de la URL y manejo de `SIGINT` en primer plano. Modos mutuamente excluyentes. (R1–R9)

- [x] **T11 — Tests de `internal/view`:** render de Markdown (HTML crudo neutralizado), composición de la página con documentos presentes/ausentes, servidor con `httptest` (token inválido/correcto, `/content`, `/approve` permitido y denegado, SSE), y descriptor temporal (escritura/lectura/liveness/limpieza). (R41)

## Fase 4 — CLI y arnés

- [x] **T12 — Dispatch de `view` en `internal/cli/cli.go`:** `case "view": return cmdView(rest)`; parsear `<id>` opcional y modos, resolver la sesión activa si no hay id, delegar en `view.Run` y devolver `1` si no hay sesión activa y `2` ante uso incorrecto. Añadir la entrada `view` a `internal/cli/help.go` (uso, breve y detalle de modos/códigos). Tests del dispatch y de la ayuda. (R1, R2, R3, R4, R5, R42)

- [x] **T13 — Configuración por defecto e higiene:** añadir `"viewer": { "autoOpenOnReady": false }` a `.rei/config.json` y añadir `.rei/.view-server.json` a `.gitignore`. (R33, R35, R43)

- [x] **T14 — Contrato del Leader:** en `.rei/agents/leader.md`, puerta de aprobación (Caso A), indicar que si `viewer.autoOpenOnReady` es `true` (leído de `.rei/config.json`) ejecute `rei view <work-item-id> --open` sin bloquear la conversación; si es `false`, no cambiar el flujo actual. Dejar explícito que el disparo es por CLI, sin hooks nativos. (R36, R37, R38)

- [x] **T15 — Contrato del Initializer:** en `.rei/agents/initializer.md`, añadir a la entrevista la pregunta sobre la apertura automática del visor y la escritura de `viewer.autoOpenOnReady` en `.rei/config.json`. (R39)

## Fase 5 — Validación

- [x] **T16 — Validación final** y corrección de lo que falle:
  `gofmt -l internal` (sin salida), `make vet`, `make test`, `make build`,
  `rei check --quiet` (`exit 0`).
  Verificación manual:
  - `rei view <id>` sirve la página y bloquea hasta Ctrl+C;
  - `rei view <id> --open` arranca el servidor de fondo, abre el navegador y devuelve el control; repetirlo reutiliza el servidor;
  - editar un documento recarga el navegador (SSE);
  - el botón Aprobar aplica `ready -> in_progress` y alinea `current.md`; en otro estado no aparece o no permite la transición;
  - `rei view --stop` detiene el servidor y elimina `.rei/.view-server.json`;
  - `rei view --enable`/`--disable` actualizan solo `viewer.autoOpenOnReady`;
  - `rei validate <id>` ya no reporta incoherencia tras aprobar. (R1–R44)

- [x] **T17 — Marcar como completadas (`[x]`) todas las tareas** de este archivo a medida que se terminan, dejando `tasks.md` al día. (trazabilidad)
