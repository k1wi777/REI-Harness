# Revisión — P2: UX del CLI, validación offline y documentación

## Resultado

**done** — re-revisión tras el rework de R7: el hallazgo previo queda resuelto y
no hay regresiones. R1–R18 se cumplen.

## Corrección de R7 (hallazgo previo)

- `github.Options` incorpora `RepoSet`/`BaseSet`/`StrategySet`; `ValidateOptions`
  distingue «flag informado» de «flag ausente» y rechaza el valor vacío antes de
  resolver el proyecto.
- Reproducido con el binario nuevo en proyecto temporal (config con `checks`,
  `extra`, `viewer` y `integrations.github` previa):

| Invocación | Salida | Config |
|------------|--------|--------|
| `enable --repo ""` | 2, `--repo "" inválido` | intacta |
| `enable --base ""` | 2, `--base "" inválida` | intacta |
| `enable --strategy ""` | 2, `--strategy "" inválida` | intacta |
| `enable` sin flags | 0, conserva `previo/repo`, `develop`, `merge` | preserva claves ajenas |
| `enable --repo acme/otro --base main --strategy rebase` | 0, aplica flags | escribe lo indicado |

- Tests nuevos: `TestParseGitHubEnableValorVacio` y
  `TestCmdGitHubEnableValorVacioNoEscribe` (exit 2, mención del flag, config
  intacta); `TestParseGitHubEnable` actualizado a los campos `*Set`.

## Verificaciones

| Checkpoint | Comando | Resultado |
|------------|---------|-----------|
| V1 | `go vet ./...` | pasa |
| V2 | `go test -count=1 ./...` | pasa (todos los paquetes) |
| V3 | `make build` | pasa |
| V4 | `make cross` | no aplica: sin código por plataforma ni release |

- `gofmt -l .`: sin salida; `./bin/rei check --quiet`: 0.
- `init opencode --check`, `init claude --check`, `doctor`, `init status`: 0.
- `rei validate <id>`: OK, sin `WARN` de presupuesto (design 256 líneas).
- `rei review-diff ... --full`: exit 0; el diff corresponde a T1–T12 más el
  rework de R7.
- Sin regresiones (proyecto temporal): `item show`/`items status` con
  `links.json` válido, ausente y roto (nota, exit 0); `items` exit 1 con dos
  `in_progress`; `validate` con índice roto → `[WARN] ... recuperable` +
  `Resultado: OK`; `github status` → «Configuración completa; sin acciones de
  configuración pendientes.»; `done → blocked` rechazado (1) y permitido con
  `--force`; `session start` rechaza inexistente/tipo distinto sin tocar
  `current.md` e inicia en `pending`.
- `skeleton/`, `.rei/config.json`, `go.mod`/`go.sum` sin cambios; sin
  dependencias; repo canónico `k1wi777/REI-Harness`; Task sin opt-in ni
  `rei:task`; `package-lock.json` eliminado sin referencias vigentes;
  documentación y agentes nativos coherentes con `--check`.

## Observaciones

- Sin acciones requeridas. V4 no aplica.
