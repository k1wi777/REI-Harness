# Diseño — P2: UX del CLI, validación offline y documentación

> Decisiones con peso para implementar la Feature. No repite requisitos ni
> describe arquitectura general. Presupuesto: 150–300 líneas (objetivo).

## Estrategia

P2 es una suma de cambios locales y acotados que comparten dos piezas nuevas:
un lector único de `links.json` y un contrato más estricto de `session start`.
El resto son ajustes de salida, validación y documentación. No se añaden
dependencias, no se cambia el workflow ni el formato de `meta.json`, y no se
toca `skeleton/` (el cambio vive en el harness/proyecto vivo).

El orden de implementación es: pieza compartida de `links.json` → consumidores
(`show`, `items status`, `validate`) → CLI de GitHub → transición `done` →
sesión → repositorio canónico → documentación y limpieza.

## 1. Paquete `internal/links`

Se crea un paquete hoja, `internal/links`, que solo usa la biblioteca estándar.
No importa `meta`, `paths`, `github` ni `validate`, por lo que no introduce
ciclos y puede ser consumido por presentación (`show`, `cli`) y dominio
(`validate`).

Modelo y API:

- `Links`: `provider`, `repo`, `issue`, `url`, `base` opcional y `epic`
  opcional.
- `Base`: `branch`, `sha`.
- `Epic`: `slug`, `issue`, `url`.
- `Read(path)`: ausencia del archivo se distingue de error de lectura/JSON; un
  JSON mal formado devuelve error.
- `Problems(links)`: lista de incumplimientos de forma; un `links.json`
  sintácticamente válido pero incompleto sigue siendo legible y recuperable.

Reglas de forma (todas offline, sin red ni resolución canónica):

| Campo | Regla |
|-------|-------|
| `provider` | No vacío y exactamente `github` (v1). |
| `repo` | `owner/name` con la misma gramática de identificadores que usa el remoto de GitHub; sin host, `.git` ni barra extra. |
| `issue` | Entero mayor que cero. |
| `url` | URL absoluta `http`/`https` con host y ruta no vacíos. No se exige que coincida con `repo`/`issue` (los renames y redirects son legítimos). |
| `base` | Si está presente, objeto con `branch` no vacío y `sha` hexadecimal de 7–40 caracteres. Ausente = índice v1. |
| `epic` | `null` o ausente; si es objeto, `slug` no vacío, `issue` > 0 y `url` absoluta. |
| Campos extra | Se ignoran para no romper la compatibilidad futura. |

Decisión: la validación de forma no emite `FAIL` ni bloquea. Un `links.json`
inválido es un índice recuperable: la issue se puede reconstruir por marcador
(playbook) y el CLI solo informa.

`show`, `cli` y `validate` construyen la ruta con `paths.SpecDir(id)` y llaman
a `links.Read`/`links.Problems`; no se duplica el parseo.

## 2. `rei item show`

Se añade una sección `Enlaces:` después de los metadatos y antes de
`Documentos:`.

- Válido: `issue: #N`, `url`, `base: branch @ sha` (o «sin registrar» si falta),
  `epic: slug (#N)` y su URL (o «sin épica»).
- Sin archivo: una sola línea «Enlaces: (sin links.json)».
- Inválido o ilegible: «Enlaces: nota: links.json inválido (causa); se ignora».
  La ficha continúa y el código de salida sigue dependiendo solo de la
  validación ya existente.

`show` ya llama a `validate.WorkItem`, así que un `links.json` inválido también
aparecerá como `WARN` en `Validación:`. Es intencional: la nota de `Enlaces`
explica el contenido y el `WARN` lo integra en la validación; ninguno es `FAIL`.

## 3. `rei items status`

`state.ItemStatus` incorpora `Type` (leído del `meta.json` que `ListWorkItems`
ya carga). `cmdItems` conserva el orden y el marcador `<< ACTIVO`; por cada
Work Item intenta leer `links.json`:

- válido con issue: añade `issue: #N`;
- con base: añade `base: branch@sha`;
- inválido/ilegible: añade `links inválido` y no aborta;
- sin archivo: solo tipo.

Formato previsto: `- <id> [<status>] (type: <tipo>[, issue: #N][, base: b@sha])
<< ACTIVO`. El aviso de más de un `in_progress` se mantiene exactamente igual:
se calcula con `state.InProgressItems`, se escribe en stderr y devuelve 1.

## 4. `rei github enable`

`github.Enable` pasa a recibir opciones explícitas (`Repo`, `Base`, `Strategy`;
vacío = no informado). `cmdGitHub` parsea `--repo <valor>`, `--base <valor>` y
`--strategy <valor>` solo para `enable`; los flags de `disable`/`status`, los
flags desconocidos, los valores ausentes y los flags repetidos son uso
incorrecto (2). Los valores se validan antes de resolver el proyecto: repo
`owner/name`, base no vacía y sin espacios, estrategia exacta entre `squash`,
`merge` y `rebase`. Los errores semánticos de valores también son uso incorrecto
(2); los fallos operativos (config inválida, `origin` no resoluble, escritura)
son 1 con causa y siguiente paso.

Resolución de cada valor (se aplica por separado):

| Valor | Orden |
|-------|-------|
| `repo` | flag → `integrations.github.repo` previo → `git remote get-url origin`; si nada resuelve, error 1 sin escribir. |
| `base` | flag → `base` previo → `origin/HEAD` detectado → `main`; el aviso de fallback solo aparece cuando se usó la detección y falló. |
| `mergeStrategy` | flag → `mergeStrategy` previo → `squash`. |

Consecuencia deliberada: si ya hay `repo`/`base`/`mergeStrategy` previos y no se
pasa flag, **no** se detecta `origin` para sobrescribirlos. `enable` sin flags
sigue funcionando y escribe la misma estructura de cuatro claves. Si no hay
`repo` previo ni flag, `origin` sigue siendo obligatorio.

La escritura reutiliza `config.SetGitHub`, que ya reescribe solo
`integrations.github` y conserva el resto del JSON de forma atómica. No se
modifica `internal/config`. `Disable` y `Status` no cambian salvo el texto de
cierre de `Status`: la rama sin pendientes pasa a
«Configuración completa; sin acciones de configuración pendientes.». No se
insinúa nada sobre issues/PRs.

## 5. `rei validate` y `links.json`

`validate.WorkItem` llama a `links.Read` y `links.Problems` sobre
`.rei/specs/<id>/links.json`:

- archivo ausente: sin hallazgo;
- error de lectura/JSON o problemas de forma: `WARN` con la causa y recordatorio
  de que es recuperable;
- nunca `FAIL` y nunca altera el resto de la validación.

La comprobación es offline y no consulta `gh` ni el remoto.

## 6. `done → blocked`

En `meta.TransitionAllowed`, la regla especial «todo → blocked» deja de aplicar
a `done`: `done → blocked` devuelve `false`. El resto de estados sigue pudiendo
bloquearse. `state.Transition` no cambia: si la transición no es estándar y
`force=false`, devuelve `ErrTransitionNotAllowed` antes de escribir; con
`--force`, `forced=true` y se aplica. `done` queda como terminal en la práctica,
pero `--force` sigue siendo la vía explícita para una transición no estándar.

Se actualizan los tests de `meta` y `state` (incluido el caso forzado ya
existente) y la documentación de `workflow.md`, `meta.md`, `help.go`
(`status set`) y `leader.md`.

## 7. `rei session start`

`template.StartSession` valida antes de tocar el disco:

1. el Work Item existe y `meta.json` se puede cargar;
2. el `type` solicitado coincide con `meta.json.type`;
3. el estado es `pending`.

Solo después hace `MkdirAll` y renderiza `current.md`; cualquier rechazo deja el
archivo actual intacto (sin escrituras parciales). Los mensajes incluyen causa y
siguiente paso:

| Estado | Siguiente paso sugerido |
|--------|-------------------------|
| `pending` | inicia la sesión (caso normal). |
| `ready` | no reiniciar: la planificación ya está lista; aprobar o revisar (Caso C). |
| `in_progress`, `review`, `changes_requested` | retomar la sesión/etapa existente (`rei item show`), no crear una nueva. |
| `done`, `blocked` | Work Item cerrado: crear uno nuevo con `rei new`. |

`cmdSession` ya traduce el error a stderr y código 1. El fixture de
`template/render_test.go` debe crear el `meta.json` en `pending` antes de
`StartSession`; se añaden casos de inexistente, tipo distinto y cada estado no
permitido.

## 8. Repositorio canónico

Se cambian `repoPath` y `releasesPage` en `internal/update/github.go` a
`k1wi777/REI-Harness` y
`https://github.com/k1wi777/REI-Harness/releases`. Los tests de update ya usan
`repoPath` para el mux, así que seguirán pasando; se añade una prueba de
regresión que fije ambos valores canónicos. `README.md` y `.rei/docs/usage.md`
actualizan sus dos enlaces de instalación. El module path y `.goreleaser.yaml`
quedan fuera de alcance (P3).

## 9. Documentación y adaptadores

- `help.go`: ayuda de `item` (enlaces), `items` (tipo/issue/base), `session`
  (solo `pending`), `status` (`done` terminal y `--force`) y `github` (flags y
  precedencia). La forma corta de `github` debe seguir conteniendo
  `github <enable|disable|status>` para no romper la ayuda general.
- `README.md`: tabla con `rei new`, `rei session`, `rei commit` y `rei github`;
  enlaces canónicos.
- `.rei/docs/usage.md`: flags de `enable`, estados resultantes, validación de
  `session start`, enlaces de `item show`/`items status` y enlaces canónicos.
- `github.md`: comandos con flags y precedencia; aclarar que `--base` fija el
  fallback y no la base por Work Item; eliminar el opt-in de Tasks.
- `core.md`: eliminar el opt-in y el «type == feature (o opt-in)»; documentar
  que un `links.json` inválido se trata como índice recuperable y que
  `rei validate` lo avisa.
- `errors.md`: añadir el caso de `links.json` inválido con causa y recuperación.
- `workflow.md`: `done` terminal; `done → blocked` solo con `--force`.
- `meta.md`: misma regla en `status`.
- `leader.md`: eliminar el opt-in de Tasks y reforzar que `done` no se fuerza
  sin orden explícita; al cambiar el contrato se regeneran
  `.opencode/agents/leader.md` y `.claude/agents/leader.md` con
  `rei init opencode` / `rei init claude` y se comprueba con `--check`.
- `progress.md`: `session start` solo para Work Items `pending` y qué hacer con
  el resto.
- `AGENTS.md`: no requiere cambios; no documenta flags ni comandos del CLI.

## 10. Limpieza

Se elimina `package-lock.json` (sin `package.json` ni referencias en Makefile,
CI, GoReleaser o configuración). Las menciones históricas en
`.rei/specs/**` y `.rei/progress/**` no se tocan.

## Archivos afectados

| Archivo | Cambio |
|---------|--------|
| `internal/links/links.go`, `links_test.go` | Nuevo paquete hoja: modelo, lectura y problemas de forma. |
| `internal/show/show.go`, `show_test.go` | Sección `Enlaces:` y casos válido/ausente/inválido. |
| `internal/state/state.go`, `state_test.go` | `ItemStatus.Type`; cobertura de `done → blocked` forzado. |
| `internal/cli/cli.go`, `cli_test.go` | Parseo de flags de `github enable`; `items status` con tipo/enlaces. |
| `internal/github/github.go`, `github_test.go` | Opciones de `Enable`, precedencia, validaciones y frase de `status`. |
| `internal/validate/validate.go`, `validate_test.go` | `WARN` de `links.json`; nunca `FAIL`. |
| `internal/meta/meta.go`, `meta_test.go` | `done` no transiciona a `blocked` sin forzar. |
| `internal/template/render.go`, `render_test.go` | Validación de `StartSession`; fixture en `pending`. |
| `internal/update/github.go`, `update_test.go` | Repositorio canónico y regresión. |
| `internal/cli/help.go`, `help_test.go` | Ayuda de los comandos afectados. |
| `README.md`, `.rei/docs/usage.md` | Tabla, flags, estados y enlaces canónicos. |
| `.rei/docs/harness/integrations/github.md`, `github/core.md`, `github/errors.md` | Flags, precedencia, `links.json` y Task sin issue. |
| `.rei/docs/harness/workflow.md`, `meta.md`, `progress.md`, `.rei/agents/leader.md` | `done`, `session start` y Task sin opt-in. |
| `.opencode/agents/leader.md`, `.claude/agents/leader.md` | Regenerados desde el contrato. |
| `package-lock.json` | Eliminado. |

## Alternativas descartadas

- **Tener la lógica de `links.json` en `show` y duplicarla en `validate`:** dos
  lectores divergirían y ya hay tres consumidores; se prefiere un paquete hoja.
- **Añadir `links` a `state.ListWorkItems`:** acoplaría el estado del workflow a
  una integración opcional; basta con `Type` en el modelo y leer los enlaces en
  la capa que presenta.
- **Reutilizar `github.ParseRemote` para validar `--repo` y `repo` de
  `links.json`:** acepta hosts y remotos, no el formato `owner/name` puro, y
  obligaría a `links` a depender de `github`.
- **Validar el `url` contra `repo`/`issue`:** los renames de GitHub mantienen
  redirects válidos; una discrepancia no debe convertirse en ruido de validación.
- **Resolver `repo` con detección aunque haya valor previo:** rompe la
  precedencia acordada y hace que `enable` sin flags cambie un repo ya fijado.
- **Rechazar `done → blocked` incluso con `--force`:** el acuerdo mantiene
  `--force` como vía explícita para transiciones no estándar; solo se retira del
  conjunto estándar.

## Verificación

Tests junto al código en cada paquete: `links` (forma, ausencia, JSON roto),
`show` (válido/ausente/inválido y código de salida), `state`/`cli` (tipo,
issue/base, dos `in_progress`), `github` (sin flags, cada flag, precedencia,
valores inválidos sin escritura, preservación), `validate` (WARN y no FAIL),
`meta`/`state` (`done → blocked` estándar rechazado y forzado permitido),
`template` (existe/tipo/estado, sin escritura parcial), `update` (canónico) y
`help` (textos clave). Cierre: `go test ./...`, `go vet ./...`, `make build`,
`rei check`, `rei init opencode --check` y `rei init claude --check`.
