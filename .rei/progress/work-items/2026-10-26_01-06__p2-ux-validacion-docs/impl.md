# Implementación — P2: UX del CLI, validación offline y documentación

## Resumen

Se ejecutaron T1–T12 según `design.md`: lector compartido de `links.json`, su
integración en `item show`, `items status` y `validate`, flags de
`github enable`, rechazo estándar de `done → blocked`, validación de
`session start`, repositorio canónico, documentación y limpieza.

## Archivos modificados

- Nuevos: `internal/links/links.go`, `internal/links/links_test.go`.
- Código: `internal/{show,state,cli,github,validate,meta,template,update}` y sus
  tests (`show`, `state`, `cli`, `github`, `validate`, `meta`, `template`,
  `update`, `help`).
- Docs: `README.md`, `.rei/docs/usage.md`,
  `.rei/docs/harness/{workflow.md,meta.md,progress.md}`,
  `.rei/docs/harness/integrations/{github.md,github/core.md,github/errors.md}`,
  `.rei/agents/leader.md` y los nativos `.opencode/agents/leader.md` y
  `.claude/agents/leader.md`.
- Eliminado: `package-lock.json`.

## Cambios

- **T1** `internal/links` (hoja, solo stdlib): `Links`, `Read` (ausencia vía
  `fs.ErrNotExist`), `Problems` con las reglas de forma acordadas.
- **T2** `item show` añade `Enlaces:` (issue/url/base/épica); ausente o inválido
  se muestra como nota y no cambia el código de salida.
- **T3** `state.ItemStatus.Type`; `items status` muestra `type`, `issue` y
  `base: branch@sha`, conserva `<< ACTIVO` y el exit 1 con >1 `in_progress`.
- **T4** `github status` cierra con «Configuración completa; sin acciones de
  configuración pendientes.».
- **T5** `github.Options`/`ValidateOptions` y parseo CLI de flags; precedencia
  flag > previo > detección > default; uso incorrecto 2 sin escribir; `*Set`
  distingue «flag informado» de «ausente» (rework de R7).
- **T6** `validate.WorkItem` avisa `WARN` de `links.json` inválido; nunca `FAIL`.
- **T7** `done → blocked` deja de ser estándar (`TransitionAllowed`); sin
  `--force` se rechaza y con `--force` se aplica.
- **T8** `StartSession` exige Work Item existente, tipo coincidente y `pending`
  antes de escribir, con causa y siguiente paso.
- **T9** `repoPath`/`releasesPage` → `k1wi777/REI-Harness` y test de regresión.
- **T10** Ayuda de `item/items/session/status/github`, tabla de `README.md` y
  comandos/flags en `.rei/docs/usage.md`.
- **T11** Docs de GitHub (flags/precedencia, Task sin issue ni opt-in,
  `links.json` recuperable), `done` terminal y `session start` solo `pending`;
  agentes nativos regenerados.
- **T12** `package-lock.json` eliminado; sin referencias vigentes.

## Verificación

- V1 `go vet ./...`, V2 `go test -count=1 ./...`, V3 `make build`: pasan;
  `./bin/rei check --quiet`, `init opencode --check`, `init claude --check`,
  `doctor` e `init status`: exit 0; `validate <id>`: OK; `gofmt -l .`: sin salida.
- Manual (proyecto temporal): `item show`/`items status` con `links.json`
  válido, ausente y roto (nota + exit 0); `validate` con índice roto emite
  `WARN` y `Resultado: OK`; `session start` rechaza inexistente, tipo distinto
  y estado `ready` con causa y siguiente paso; `done → blocked` rechazado sin
  `--force` (1) y permitido con `--force` (0); `github enable` con flags aplica
  `flag/repo`, `develop`, `merge`, sin flags conserva lo previo y
  `--strategy octopus` sale 2 sin escribir.
- V4 no aplica: sin cambios de plataforma ni de release.

## Revisión (rework de R7)

- Causa: `""` era el centinela de «no informado» y los flags vacíos caían en la
  precedencia; corrección: `Options` añade `RepoSet/BaseSet/StrategySet`, el
  parser los marca y `ValidateOptions` rechaza el vacío (2) antes de resolver el
  proyecto. Sin flags (vacío = ausente) precedencia y preservación no cambian.
- Tests: valor vacío en `parseGitHubEnable` y `cmdGitHub` (exit 2, stderr con el
  flag, config intacta), tabla en `ValidateOptions`/`Enable` y caso de flag
  ausente conservado.
- Reejecutado: V1–V3, `gofmt`, `check`, adaptadores `--check`, `doctor`,
  `init status` y `validate` en verde.
- Manual (`sha256sum`): los tres flags vacíos → exit 2 y hash idéntico de
  `.rei/config.json`; `enable` sin flags conserva lo previo.

## Observaciones

- Sin cambios en `skeleton/`, `.rei/config.json`, `go.mod`/`go.sum`; `bin/rei`
  recompilado para las comprobaciones; sin commit, push, PR ni merge.
