# Design — Visor de especificaciones en el navegador

> Work Item: `2026-10-09_23-45__spec-browser-viewer` (feature).
>
> Este documento describe **cómo** se implementa la Feature. Las referencias
> `R<n>` apuntan a `requirements.md`.

---

# 1. Estrategia

Se añade un comando `rei view` que sirve, en el navegador, una vista enriquecida
de un Work Item (metadatos + documentos + validación + sesión), renderizando el
Markdown con `goldmark` y con un JavaScript propio mínimo para la recarga en vivo
(SSE) y el botón Aprobar.

El visor **no introduce estado nuevo**: lee los mismos archivos que `rei item
show` (`meta.json`, los `.md` del Work Item y `.rei/progress/current.md`) y, al
aprobar, reutiliza la transición validada de `rei status set`. El único registro
que crea es un **descriptor temporal de servidor** (pid, puerto, token) que vive
mientras el servidor de fondo está activo y se elimina al cerrarlo.

Capas:

1. **Fundaciones reutilizables.** `internal/browser` (abrir el navegador, extraído
   de `internal/update`), `internal/config` (clave `viewer`), `internal/state`
   (transición de estado compartida + alineación de `current.md`).
2. **Paquete `internal/view`.** Render con `goldmark`, servidor HTTP local, SSE,
   handler de aprobación y gestión del servidor de fondo.
3. **CLI.** Dispatch de `view` en `internal/cli/cli.go` y entrada de ayuda en
   `internal/cli/help.go`.
4. **Arnés.** Ajustes de los Contratos de `leader.md` y `initializer.md` para el
   disparo por CLI (runtime-agnóstico) y la pregunta de configuración.

---

# 2. Archivos afectados

## Nuevos

| Archivo | Propósito |
|---------|-----------|
| `internal/browser/browser.go` | `Open(url string) error`: abrir el navegador por SO (extraído de `internal/update/browser.go`). |
| `internal/browser/browser_test.go` | Tests de `Open` (comando según `GOOS`, inyectable). |
| `internal/view/view.go` | Punto de entrada: `Options`/`Run` y dispatch de modos. |
| `internal/view/server.go` | Servidor HTTP, rutas, token, SSE, aprobación, autocierre. |
| `internal/view/render.go` | Render de documentos con `goldmark` y composición de la página. |
| `internal/view/lifecycle.go` | Descriptor temporal, spawn/`detach` del proceso de fondo, `--stop`, liveness. |
| `internal/view/process_unix.go` | Detach del proceso en Unix (`Setsid`). |
| `internal/view/process_windows.go` | Detach del proceso en Windows. |
| `internal/view/assets/page.html` | Plantilla HTML de la página. |
| `internal/view/assets/app.js` | JS propio: SSE + recarga + botón Aprobar. |
| `internal/view/assets/style.css` | Estilos. |
| `internal/view/*_test.go` | Tests del paquete (render, servidor con `httptest`, descriptor). |

## Modificados

| Archivo | Cambio |
|---------|--------|
| `internal/update/browser.go` | Eliminado; su contenido se mueve a `internal/browser`. |
| `internal/update/update.go` | `Options.OpenBrowser` sigue inyectable; su valor por defecto pasa a `browser.Open`. |
| `internal/config/config.go` | `Viewer{AutoOpenOnReady bool}` + `Config.Viewer`; `SetAutoOpenOnReady(path, bool)`. |
| `internal/state/state.go` | `Transition(p, id, to, force)` y alineación del campo Estado de `current.md`. |
| `internal/cli/cli.go` | `case "view": return cmdView(rest)`; `cmdStatus` pasa a usar la transición compartida. |
| `internal/cli/help.go` | Entrada `view` en la tabla `commands`. |
| `internal/cli/cli_test.go` | Tests del dispatch de `view`. |
| `.rei/config.json` | Se añade `"viewer": { "autoOpenOnReady": false }`. |
| `.gitignore` | Ignora el descriptor temporal del servidor (`.rei/.view-server.json`). |
| `go.mod` / `go.sum` | Dependencia `github.com/yuin/goldmark`. |
| `.rei/agents/leader.md` | Disparo del visor en la puerta de aprobación (Caso A) cuando la opción está activa. |
| `.rei/agents/initializer.md` | Pregunta y escritura de `viewer.autoOpenOnReady`. |

---

# 3. Componentes

## 3.1 `internal/browser`

```go
// Open intenta abrir url en el navegador por defecto, según el SO.
// Es best-effort: devuelve el error para que el llamante lo ignore o lo reporte.
func Open(url string) error
```

Implementación idéntica a la actual `openBrowser` de `internal/update`. Para poder
testear sin abrir navegadores, se añade una costura interna
(`openCommand func(goos, url string) *exec.Cmd`) o una variable de paquete
sobrescribible; los tests verifican el comando elegido por `GOOS`
(`windows` → `cmd /c start`, `darwin` → `open`, resto → `xdg-open`).
`internal/update` deja de definir `openBrowser` y usa `browser.Open` como valor
por defecto de su costura `Options.OpenBrowser` (los tests de `update` siguen
inyectando su propia función). (`R40`)

## 3.2 `internal/config`

```go
type Viewer struct {
    AutoOpenOnReady bool `json:"autoOpenOnReady"`
}

type Config struct {
    Checks []Check `json:"checks"`
    Viewer Viewer  `json:"viewer"`
}

// SetAutoOpenOnReady escribe viewer.autoOpenOnReady preservando el resto del
// archivo, de forma atómica e indentada.
func SetAutoOpenOnReady(path string, enabled bool) error
```

- `Load` no cambia su contrato: si la clave `viewer` falta, `Config.Viewer` queda
  en su valor cero (`false`) y no falla (`R34`).
- `SetAutoOpenOnReady` lee el JSON existente como `map[string]any`, escribe la
  clave anidada `viewer.autoOpenOnReady` y vuelca el mapa completo con
  `SetIndent("", "  ")` y `SetEscapeHTML(false)`, de forma atómica (temp + rename,
  como `meta.writeFileAtomic`). Así preserva `checks` y cualquier otra clave
  (`R35`).
- El default `false` también se materializa en el esqueleto: `.rei/config.json`
  pasa a `{ "checks": [], "viewer": { "autoOpenOnReady": false } }` (`R33`).

## 3.3 `internal/state` — transición y alineación

`rei status set` y el botón Aprobar deben compartir exactamente la misma lógica.
Se extrae a `internal/state` (que ya lee `meta.json` y `current.md`):

```go
// ErrTransitionNotAllowed indica que la transición no está permitida por workflow.
var ErrTransitionNotAllowed = errors.New("transición no permitida")

// Transition valida y aplica una transición de estado sobre meta.json. Si el
// Work Item es la sesión activa, alinea el campo Estado de current.md.
// Devuelve el estado previo y si fue necesario forzar la transición.
func Transition(p *paths.Project, id, to string, force bool) (from string, forced bool, err error)
```

- Reutiliza `meta.IsValidStatus`, `meta.TransitionAllowed` y `meta.SetStatus`
  (exactamente el código de `rei status set`). (`R22`)
- Si `!meta.TransitionAllowed(from, to) && !force` devuelve
  `ErrTransitionNotAllowed` sin escribir. (`R24`)
- Tras `meta.SetStatus`, si `state.ReadSession(p).WorkItem == id`, reescribe la
  línea `- **Estado:** ...` de `current.md` con el nuevo estado. La edición es
  quirúrgica (reemplaza solo esa línea) y conserva el resto del archivo. (`R23`)
- `cmdStatus` en `internal/cli/cli.go` se refactoriza para llamar a
  `state.Transition`, conservando sus mensajes y su flag `--force`. El PDF del
  comportamiento del CLI no cambia salvo que ahora también alinea `current.md`
  (lo que ya pedía el Contrato del Leader `rei validate` + corrección manual).
  (`R22`, `R44`)
- El handler `/approve` del visor llama a `state.Transition(p, id,
  meta.StatusInProgress, false)`. (`R22`)

## 3.4 `internal/view` — render

```go
// viewModel es el conjunto de datos que la plantilla HTML necesita.
type viewModel struct {
    Meta       *meta.Meta
    Session    state.Session
    Docs       []docView // nombre, ruta, ¿existe?, HTML renderizado
    Validation []validate.Issue
    CanApprove bool // status == ready
}

// renderMarkdown convierte Markdown a HTML con goldmark en modo seguro.
func renderMarkdown(src []byte) (template.HTML, error)
```

- El render usa `goldmark.New()` **sin** `html.WithUnsafe()`, de modo que el HTML
  incrustado en los documentos no se interpreta (seguro por defecto). (`R17`)
- Los documentos se recorren en el orden de `internal/show` (`plan.md`,
  `requirements.md`, `design.md`, `tasks.md`, `impl.md`, `review.md`,
  `spec.md`), reutilizando la resolución de rutas de `internal/show` (`docPath`):
  planificación en `.rei/specs/<id>/`, progreso en
  `.rei/progress/work-items/<id>/`. Un documento ausente se muestra como
  "(no existe)". (`R11`, `R12`, `R16`)
- La validación reutiliza `validate.WorkItem`; la sesión, `state.ReadSession`.
  (`R13`, `R14`)
- La página se compone con `html/template` (escape automático) y sirve el HTML ya
  renderizado; el JS no renderiza Markdown. (`R15`, `R18`)

## 3.5 `internal/view` — servidor HTTP

```go
type Server struct {
    p      *paths.Project
    id     string
    token  string
    srv    *http.Server
    // reloj, timeout de inactividad y manejador de cierre inyectables para tests
}
```

Rutas (todas exigen autenticación salvo la carga inicial con token válido):

| Ruta | Método | Descripción |
|------|--------|-------------|
| `/?token=<t>` | GET | Página completa. Valida el token; si es correcto fija una cookie de sesión y sirve la página. |
| `/content` | GET | Fragmento HTML dinámico (metadatos + documentos + validación) para la recarga. |
| `/events` | GET | SSE; emite `data: reload` cuando cambia algún documento del Work Item. |
| `/approve` | POST | Aplica la transición `ready -> in_progress`; responde JSON. |
| `/static/*` | GET | `app.js` y `style.css` embebidos con `go:embed`. |

- **Bind y puerto:** `net.Listen("tcp", "127.0.0.1:0")`; el puerto lo asigna el
  SO y se reporta en la URL. (`R27`)
- **Token:** 32 bytes aleatorios en hex generados por instancia
  (`crypto/rand`); viajan en la URL y se exigen para cargar el visor. Tras la
  validación inicial se fija una cookie `HttpOnly; SameSite=Strict; Path=/`, con
  la que se autentican las peticiones posteriores (SSE, `/content`, `/approve`),
  de modo que el JS no necesita transportar el token. (`R28`)
- **SSE:** un bucle por conexión hace polling de `mtime` de los documentos del
  Work Item cada ~1 s; al detectar cambios envía `data: reload`. También
  comprueba el estado de `meta.json`: si deja de ser `ready`, dispara el cierre
  (`R30`). El polling solo cubre los documentos del Work Item mostrado (`R20`).
- **Inactividad:** un temporizador (constante, p. ej. 30 min) se reinicia con
  cada petición; al expirar sin peticiones, el servidor se cierra (`R31`).
- **Aprobar:** `POST /approve` devuelve `409`/JSON de error si
  `state.Transition` falla (`R24`); si tiene éxito, responde
  `{"ok":true,"status":"in_progress"}` e incluye un mensaje indicando que hay que
  reanudar en el chat (`R26`). No crea ningún archivo de aprobación (`R25`).
- El servidor se construye de forma que sus handlers y el cierre sean testeables
  con `httptest` y un reloj/estado inyectables (`R41`).

## 3.6 `internal/view` — ciclo de vida del servidor de fondo

Modos:

- **Primer plano** (`rei view [<id>]`): arranca el servidor en el proceso actual,
  imprime la URL y bloquea hasta Ctrl+C (se maneja `SIGINT` para cierre
  *graceful*). No abre el navegador ni registra descriptor. (`R6`)
- **`--open`**: 
  1. Si existe un descriptor vivo (`internal/view/lifecycle.go`), reutiliza ese
     servidor, abre el navegador con su URL y termina. (`R29`)
  2. Si no, lanza el mismo binario en segundo plano
     (`os.Executable()` + subcomando interno `view <id> --serve`, `detach`
     según SO), espera a que el descriptor aparezca (poll con timeout), abre el
     navegador con la URL y termina. (`R7`)
- **`--serve`** (interno, oculto): el proceso de fondo. Escribe el descriptor,
  sirve y, al cerrarse, elimina el descriptor. (`R32`, `R43`)
- **`--stop`**: lee el descriptor, termina el proceso (señal) y elimina el
  descriptor. Si no hay descriptor vivo, informa y sale `0`. (`R8`, `R32`)
- **`--enable` / `--disable`**: `config.SetAutoOpenOnReady(path, bool)`. (`R9`)

Descriptor temporal: `.rei/.view-server.json` (añadido a `.gitignore`), con
`pid`, `port`, `token`, `workItem` y `startedAt`, escrito con permisos `0600` de
forma atómica. La liveness del pid se comprueba con una costura por SO
(`syscall.Kill(pid, 0)` en Unix; `os.FindProcess` en Windows); un descriptor
cuyo pid ya no vive se considera obsoleto y se elimina. Al ser un artefacto
puramente operativo (no decide el workflow), no contradice "sin estado paralelo"
(`R25`).

## 3.7 CLI

- `internal/cli/cli.go`: `case "view": return cmdView(rest)`. `cmdView` parsea el
  `<id>` opcional y los modos, resuelve el Work Item (`project()` + sesión activa
  si no hay id) y delega en `view.Run`. Argumentos inválidos → uso por `stderr` y
  `2`. (`R1`–`R5`)
- `internal/cli/help.go`: nueva entrada `view` con
  `usage: "view [<id>] [--open|--stop|--enable|--disable]"`, descripción y
  detalle de modos y códigos de salida. (`R42`)

## 3.8 Contratos del arnés

- **`leader.md` (Caso A / puerta de aprobación):** si
  `viewer.autoOpenOnReady` es `true` (leído de `.rei/config.json`), el Leader
  ejecuta `rei view <work-item-id> --open` —no bloqueante— para que el usuario
  revise en el navegador; la aprobación sigue siendo la del chat (o el botón). Si
  la opción es `false`, no ejecuta nada y el flujo es idéntico al actual.
  (`R36`, `R37`, `R38`)
- **`initializer.md`:** en la entrevista de personalización se pregunta si se
  desea activar la apertura automática del visor y se escribe
  `viewer.autoOpenOnReady` en `.rei/config.json` (ya está dentro de su alcance de
  escritura). (`R39`)

El disparo es una instrucción del Contrato ejecutada por el Leader mediante el
CLI; no se usa ningún hook nativo del runtime. (`R38`)

---

# 4. Decisiones de diseño

**D1 — Token por instancia, no cookie de autenticación inventada.**
"Token de un uso en la URL" se interpreta como un token **efímero por instancia de
servidor**: se genera al arrancar, viaja en la URL y se descarta con el servidor.
Es la lectura coherente con la reutilización exigida (un token estrictamente de un
solo request sería incompatible con reutilizar un servidor ya abierto). La
seguridad se completa atando `127.0.0.1` + token, suficiente para un servicio
local frente a *drive-by* de otras páginas o procesos locales.

**D2 — JS servidor-renderizado.**
El Markdown se convierte a HTML en Go (`goldmark`) y la página se compone con
`html/template`; el JS solo abre el SSE, sustituye el contenido y envía el POST de
aprobación. Mantener el render en Go hace que la superficie de JS sea mínima, sin
terceros (`R18`), y evita depender de un parser Markdown en el navegador.

**D3 — Reutilizar `meta`/`state` como única verdad.**
El botón Aprobar no crea estado: llama a la misma transición validada que
`rei status set` y alinea `current.md` (`R22`–`R25`). La única duplicación
tolerada es un descriptor operativo del servidor, fuera del modelo de aprobación.

**D4 — Descriptor temporal en `.rei/`, gitignorado.**
Se prefiere `.rei/.view-server.json` (frente al directorio temporal del SO) porque
`--stop`, la reutilización y la liveness deben resolverse desde el propio
proyecto; se ignora en git y se elimina al cerrar, por lo que no es estado durable
(`R43`).

**D5 — Proceso de fondo por re-ejecución del propio binario.**
`--open` relanza `os.Executable()` con `view <id> --serve` y lo desacopla del
terminal (setsid/detach), evitando bloquear la conversación (`R7`) y sin
reimplementar el servidor en un segundo camino de código.

**D6 — `--open` no decide por config.**
`--open` siempre abre el visor cuando se invoca; la decisión de invocarlo según
`viewer.autoOpenOnReady` es del Leader (`R36`, `R37`). Así `rei view --open` sigue
siendo útil manualmente aunque la opción esté desactivada.

**D7 — Cierre por estado además de por inactividad.**
El bucle SSE ya recorre `meta.json` para el refresco; se aprovecha para detectar
la salida de `ready` y cerrar (`R30`). Combinado con el timeout de inactividad
(`R31`) y `--stop` (`R32`), el servidor no queda huérfano.

**D8 — Fase 2 fuera de v1.**
Marcar secciones y enviar feedback (que mapea al Caso C del Spec Author) queda
explícitamente fuera de esta Feature; el diseño deja el handler `/approve` como
único punto de escritura, extensible en el futuro sin rehacer el modelo.

---

# 5. Alternativas descartadas

**A1 — Renderizar Markdown en el navegador con una librería JS.**
Descartada: introduce JS de terceros, prohibido (`R18`), y duplica la lógica de
render. `goldmark` en Go es testeable y seguro por defecto (`R17`).

**A2 — Servir el archivo `.md` en crudo.**
Descartada: es justo la fricción que la Feature busca eliminar (`R15`).

**A3 — Archivo de estado de aprobación propio del visor.**
Descartada: crearía un modelo de aprobación paralelo y podría desincronizarse;
`meta.json`/`current.md` ya codifican la aprobación (`R25`).

**A4 — Servidor en primer plano para el disparo del Leader.**
Descartada: bloquearía el proceso y la conversación; `--open` con proceso
desacoplado resuelve R7.

**A5 — Hook nativo del runtime (p. ej. OpenCode) como disparo.**
Descartada: rompe la agnosticidad. El disparo va por CLI desde el Contrato del
Leader (`R38`).

**A6 — Token de un solo request con handshake por cookie y emisión de tokens vía
endpoint administrativo.**
Descartada: complica el ciclo de vida (minting, secreto administrativo) sin
aportar seguridad relevante en localhost y dificulta la reutilización del
servidor. Se adopta D1.

**A7 — Descriptor en el directorio temporal del SO.**
Descartada: `--stop` y la reutilización desde el proyecto son menos fiables y el
archivo puede limpiarse por el SO mientras el servidor sigue vivo. Se adopta D4.
