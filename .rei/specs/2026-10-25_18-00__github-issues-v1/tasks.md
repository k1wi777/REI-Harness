# Tasks — Integración opcional con GitHub Issues (v1)

> Work Item: `2026-10-25_18-00__github-issues-v1` (feature).
>
> Orden de ejecución. Cada tarea indica los requisitos que implementa.
> El Implementer marca `[x]` **inmediatamente** al terminar cada tarea.

---

## Fase 1 — Configuración

- [x] **T1 — Extender `internal/config` con el modelo de integraciones.**
  Añadir `GitHub`, `Integrations` y el campo `Integrations` a `Config`, con
  `GitHub *GitHub` para distinguir «ausente» de «`enabled:false`». Añadir
  `EffectiveBase()` (`main`) y `EffectiveMergeStrategy()` (`squash`).
  Implementar `SetGitHub` y `SetGitHubEnabled` sobre `map[string]any` para
  preservar `checks`, `viewer`, otras integraciones y claves desconocidas;
  serializar indentado y escribir de forma atómica. JSON inválido ⇒ error sin
  escribir. Añadir `ErrGitHubNotConfigured`.
  Tests en `internal/config/config_test.go`: ausencia/presencia, preservación de
  claves, idempotencia, valores por defecto y JSON inválido intacto.
  (R4, R8, R10, R12, R13)

## Fase 2 — Paquete `internal/github`

- [x] **T2 — Crear `internal/github` con `Runner`, parseo y detección.**
  Implementar `Runner` (`LookPath`, `Run`), `ExecRunner`, `Remote`,
  `ParseRemote`, `DetectOrigin` y `DetectBase`.
  `ParseRemote` acepta SSH (`git@github.com:owner/name.git`), HTTPS y
  `ssh://`, descarta `.git` y barras finales, exige exactamente `owner/name`,
  valida caracteres y rechaza hosts distintos de `github.com` sin exponer
  credenciales. `DetectOrigin` ejecuta `git remote get-url origin`.
  `DetectBase` ejecuta `git symbolic-ref --quiet --short
  refs/remotes/origin/HEAD` y cae a `main` con `fallback=true`.
  Tests table-driven con un `fakeRunner` en `github_test.go`; sin red ni `gh`
  real.
  (R6, R7, R8, R14, R56)

- [x] **T3 — Implementar `Enable`.**
  `Enable(p, runner, out)` lee la config, detecta `origin`, calcula
  repo/base/mergeStrategy (conservando valores previos no vacíos), escribe con
  `SetGitHub` y sondea `gh` (`LookPath` y `gh auth status`). Los avisos de `gh`
  no cambian el código de salida. Cualquier fallo de detección, lectura o
  escritura ⇒ mensaje con causa y siguiente paso y `1`, sin escribir.
  Salida con repo, base, mergeStrategy y siguiente paso.
  Tests: éxito, aviso sin `gh`, aviso sin auth (`0`), sin `origin` (`1`, sin
  config), host inválido (`1`), base/mergeStrategy conservados, JSON inválido
  (`1`, intacto).
  (R5, R7, R9, R12, R13, R49)

- [x] **T4 — Implementar `Disable` y `Status`.**
  `Disable` solo apaga si estaba configurada, conserva repo/base/mergeStrategy y
  demás claves, e informa si ya estaba desactivada. `Status` es de solo lectura:
  estado (no configurada/desactivada/activa), valores efectivos, comparación con
  el remoto detectado, disponibilidad y auth de `gh`, y siguiente paso. `status`
  sale `0` en estados legibles y `1` solo si el proyecto o el JSON fallan.
  Tests de todos los estados y de que `Status` no escribe.
  (R10, R11, R12, R49)

## Fase 3 — CLI

- [x] **T5 — Integrar `rei github` en el CLI.**
  Añadir `case "github"` y `cmdGitHub`: valida subcomando/argumentos antes de
  resolver el proyecto (`2` por uso incorrecto), resuelve con `project()`
  (`1` sin `.rei`) y delega en `github.Enable`/`Disable`/`Status`.
  Añadir la entrada `github` en `internal/cli/help.go` con uso, descripción,
  detalle de los tres subcomandos, códigos `0/1/2` y referencia al playbook.
  Tests: `cmdGitHub` sin subcomando/desconocido/extra ⇒ `2`; ayuda general
  incluye `github enable|disable|status`; ayuda del comando menciona los
  subcomandos y códigos.
  (R1, R2, R3, R14, R49)

## Fase 4 — Playbook y contratos

- [x] **T6 — Escribir la primera mitad de
  `.rei/docs/harness/integrations/github.md`.**
  Secciones: principio «GitHub es una proyección», alcance v1, configuración y
  comandos `rei github` (con lectura de `integrations.github`), vocabulario fijo
  (marcadores, labels v1 y reservadas v2), puerta única (pregunta exacta, orden
  «sí» con búsqueda por marcador, labels, creación de issue, `links.json` y rama
  `feat/<n>-<slug>`), plantilla autosuficiente del cuerpo de la issue con
  `design.md` plegado, camino «no» y adopción de issues existentes (lectura,
  aclaraciones, marcador, labels, `links.json`).
  Incluir comandos `gh`/`git` exactos, la regla de no duplicar y la de no crear
  rama desde una base distinta.
  (R20–R28, R40–R43, R50, R55)

- [x] **T7 — Escribir la segunda mitad del playbook.**
  Secciones: comentarios automáticos de implementación y revisión (marcadores,
  plantillas, búsqueda del comentario por REST y actualización en sitio para no
  duplicar), label `rei:blocked` sin comentarios, PR bajo orden explícita, merge
  squash con asunto `<título> (#<n>)` y borrado de rama, verificación y cierre
  de la issue, prohibición de `--admin`/auto-merge, tabla de errores con causa y
  siguiente paso (sin `.rei`, desactivada, incompleta, `gh` ausente/sin auth,
  repo, issue/rama/PR, permisos, branch protection, red, rate limit),
  verificación de resultados, aislamiento del estado de REI, reintentos sin
  duplicados y recordatorios en `current.md`.
  (R29–R39, R44–R48, R50, R54)

- [x] **T8 — Actualizar el contrato del Leader.**
  En `.rei/agents/leader.md`, añadir la sección «Integración con GitHub
  (opcional)» y los pasos correspondientes:
  - Caso C: entre la aprobación y `rei status set in_progress`, aplicar la
    puerta única si `enabled` es `true`; si está desactivada, no mencionar
    GitHub.
  - Tras el Implementer: comentario de implementación idempotente y
    actualización de «Próximo paso».
  - Tras el Reviewer: comentario de revisión idempotente.
  - Adopción de issues al arrancar cuando el usuario pide «trabajemos en la #N».
  - PR y merge solo bajo orden explícita; nunca `--admin`; verificar la issue
    tras el merge.
  - Errores: causa + siguiente paso, sin bloquear REI, reintento sin duplicar.
  - Recordatorios en `current.md` mientras haya sesión y reporte en la respuesta
    tras `done`.
  Mantener el Contrato autosuficiente y remitir al playbook para los comandos
  exactos. No añadir herramientas nuevas.
  (R20–R39, R43–R48, R52, R54, R55)

- [x] **T9 — Actualizar el contrato del `initializer`.**
  En `.rei/agents/initializer.md`, añadir «Configuración de GitHub Issues» junto
  al visor: preguntar si se activa; si sí, ejecutar `rei github enable` y
  mostrar repo/base/avisos; si no, no escribir `integrations.github`; si ya
  existe, mostrar y pedir confirmación; nunca editar el JSON a mano; un fallo de
  `gh` se informa sin bloquear.
  (R15–R19, R52)

## Fase 5 — Índice, adaptadores y verificación

- [x] **T10 — Declarar el playbook en el harness.**
  Añadir `.rei/docs/harness/integrations/github.md` a
  `check.RequiredFiles` y una fila en `.rei/docs/harness/README.md`. Añadir en
  `embed_test.go` la comprobación de que el playbook está embebido.
  (R51, R56)

- [x] **T11 — Regenerar agentes nativos y manifiesto.**
  Ejecutar `make build`, después `./bin/rei init opencode` y
  `./bin/rei init claude` para regenerar `.opencode/agents/*.md`,
  `.claude/agents/*.md` y `.rei/install-manifest.json`. Verificar con
  `./bin/rei init opencode --check` y `./bin/rei init claude --check`.
  (R53, R56)

- [x] **T12 — Verificación final.**
  Ejecutar y dejar en verde:
  - `gofmt -l internal` (sin salida);
  - `make vet`;
  - `make test`;
  - `make build`;
  - `rei check --quiet`;
  - `./bin/rei init opencode --check`;
  - `./bin/rei init claude --check`.
  Prueba manual en un proyecto temporal: sin `.rei` ⇒ `1`; sin config ⇒
  `status` `0` «no configurada»; `origin` SSH y HTTPS ⇒ `enable` escribe y
  preserva claves; `disable` conserva; JSON inválido ⇒ `1` sin sobrescribir;
  argumentos inválidos ⇒ `2`. Revisar que el playbook cubre toda la tabla de
  errores y que el flujo con la integración desactivada no cambia.
  (R1–R56)
