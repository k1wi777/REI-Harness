# Plan — CI: formato, vet y tests en PRs

## Objetivo

Añadir `.github/workflows/ci.yml` con un único job que valide formato, `go vet`
y tests en cada push a `main`/`dev`, cada PR contra esas ramas y bajo demanda.
Actualizar la documentación para distinguir **CI (validación)** de **CD
(release)** y dejar `V3`/`V4` como verificaciones manuales. No se modifican
código, `Makefile` ni el flujo de release.

## Contexto y decisiones

- CI **no construye binarios**: no ejecuta `make build`, `make cross` ni
  `go build`, no genera artefactos y no publica nada. El build local sigue siendo
  manual (`make build`) y el CD es `release.yml` (GoReleaser en tags `v*`).
- Workflow nuevo e independiente de `.github/workflows/release.yml`; `release.yml`
  no se toca.
- Un **único job `test`** en `ubuntu-latest` con `timeout-minutes`: el alcance es
  pequeño y no hay jobs lentos que separar.
- Se usan `actions/checkout@v4` y `actions/setup-go@v5`, las mismas versiones que
  `release.yml`; `go-version-file: go.mod` evita duplicar la versión (1.27).
- `permissions: contents: read` a nivel de workflow: el CI solo lee el repo.
- `concurrency` agrupada por workflow y ref con `cancel-in-progress: true`:
  pushes sucesivos a la misma rama/PR cancelan el run anterior.
- `gofmt` se comprueba con `gofmt -l .`; el paso falla si imprime archivos y los
  lista en el log. No se usa `gofmt -w` en CI.
- `go test -count=1 ./...` evita resultados cacheados y garantiza ejecución real.
- Sin `matrix`, sin caché manual (la de `setup-go` es suficiente), sin subida de
  artefactos y sin pasos condicionados por rama.
- **Decisión de alcance**: `make build`/`make cross` no se modifican ni se
  integran en CI. `make cross` produce cinco binarios (linux/darwin × amd64/arm64
  y windows × amd64); GoReleaser cubre las seis combinaciones en release. La
  documentación refleja esa diferencia sin prometer que CI compila.
- `V1` sigue siendo `rei check` (`.rei/config.json`); `V2` se apoya en el job de
  CI y también puede ejecutarse a mano. No se cambian IDs ni la configuración.
- No se añaden dependencias ni herramientas de linting externas.

## Archivos

| Archivo | Acción |
|---------|--------|
| `.github/workflows/ci.yml` | **Nuevo**: workflow con un único job `test`. |
| `.rei/docs/project/verification.md` | Separar CI (V2) de CD/manual (V3, V4); actualizar estrategia, tablas y evidencia. |
| `README.md` | Actualizar la línea de `.github/workflows/` del árbol de estructura. |
| `AGENTS.md` | Mención de CI y builds manuales en el bullet **Despliegue** (§3). |

No se modifica: `.github/workflows/release.yml`, `Makefile`, `.goreleaser.yaml`,
`skeleton/**`, `.rei/config.json`, código Go ni tests.

## Cambios

### 1. `.github/workflows/ci.yml` (nuevo)

- `name: ci`.
- Disparadores: `push` a `branches: [main, dev]`; `pull_request` a
  `branches: [main, dev]`; `workflow_dispatch`. Sin filtros `paths`.
- `permissions: contents: read`.
- `concurrency`: grupo que combine workflow y ref (p. ej.
  `ci-${{ github.workflow }}-${{ github.ref }}`), `cancel-in-progress: true`.
- Job `test` (`runs-on: ubuntu-latest`, `timeout-minutes: 15`), pasos en orden:
  1. Checkout (`actions/checkout@v4`).
  2. Setup Go (`actions/setup-go@v5`) con `go-version-file: go.mod`.
  3. Comprobar formato: capturar `gofmt -l .`; si la salida no está vacía,
     imprimirla y salir con código distinto de cero.
  4. `go vet ./...`.
  5. `go test -count=1 ./...`.
- Sin `make build`, `make cross`, `go build`, job `cross`, `matrix`, artefactos
  ni caché manual.

### 2. `.rei/docs/project/verification.md`

- **Estrategia**: sustituir «La verificación es local y reproducible: no hay
  workflow de CI…» por una descripción que distinga:
  - **CI** (`.github/workflows/ci.yml`): formato, `go vet` y tests en push/PR a
    `main`/`dev` y bajo demanda.
  - **CD** (`.github/workflows/release.yml`): GoReleaser publica binarios en tags
    `v*`.
  - **Manual**: `make build` (`V3`) y `make cross` (`V4`) antes de cerrar un
    Work Item cuando aplique.
- Añadir a «Se apoya en» el formato (`gofmt -l .`) como puerta del job `test` y
  precisar que tests se ejecutan en CI; build y cross no.
- **Checkpoints rápidos**: `V1` sigue siendo `rei check` local. Aclarar que CI
  ejecuta `go vet` como refuerzo, pero `.rei/config.json` no cambia.
- **Checkpoints lentos**: reescribir el encabezado para indicar que son manuales
  o de CD, no de CI. En la tabla:
  - `V2` (`go test ./...`): CI job `test` y también ejecutable a mano.
  - `V3` (`make build`): manual; no lo ejecuta CI.
  - `V4` (`make cross`): manual o release; no lo ejecuta CI. Cobertura real de
    `make cross`: linux/darwin amd64+arm64 y windows amd64 (5 binarios);
    GoReleaser cubre las 6 combinaciones (incluida windows/arm64) en release.
- **Evidencia**: aceptar el run de CI para `V2` (URL del run, job `test`, commit
  SHA) y exigir ejecución manual documentada para `V3`/`V4`; el run debe
  corresponder al commit revisado.
- **Qué NO hacer**: añadir que no se cite un run de CI de otro commit, no se
  declare verificado un job fallido y no se atribuya a CI la ejecución de
  `V3`/`V4`. Mantener la prohibición de meter `V2`/`V3`/`V4` en `rei check`.
- No cambiar los IDs ni el comando de `V1`; no tocar `.rei/config.json`.

### 3. `README.md`

- En el árbol de estructura, reemplazar el comentario de `.github/workflows/`
  (hoy solo releases) por uno que mencione CI (formato, vet, tests) y releases
  con goreleaser.
- Mantener la referencia a `.goreleaser.yaml`; no corregir enlaces antiguos ni
  otras secciones (fuera de alcance).

### 4. `AGENTS.md` (§3, Despliegue)

- Ampliar el bullet **Despliegue** para indicar: CI
  (`.github/workflows/ci.yml`) valida formato, `go vet` y tests en push/PR;
  GoReleaser publica binarios en tags `v*`; `make build` y `make cross` son
  manuales (no los ejecuta CI).
- Mantener el resto de la sección y los comandos sin cambios.

## Restricciones

- Sin dependencias nuevas ni cambios en `go.mod`/`go.sum`.
- Sin cambios de código Go, tests ni assets.
- No tocar `release.yml`, `Makefile`, `.goreleaser.yaml`, `skeleton/**`,
  `.rei/config.json` ni el flujo de release.
- CI no compila ni publica artefactos; no añadir `make build`/`make cross`,
  `go build`, jobs `cross`, caché manual ni subida de artefactos.
- No añadir `actionlint`/`yamllint` ni otros linters externos.
- No inventar runs ni resultados de CI: GitHub Actions no se ejecuta en local.
- Mantener el plan dentro del presupuesto (≤ 200 líneas).

## Validación

- Parsear `.github/workflows/ci.yml` con un parser YAML disponible (p. ej.
  `npx --yes yaml@2 .github/workflows/ci.yml`; si no hay red, módulo Go temporal
  en `/tmp` con `gopkg.in/yaml.v3`, sin tocar `go.mod`). Comprobar claves `on`,
  `permissions`, `concurrency`, `jobs.test` y que no hay job `cross`.
- Revisar que los triggers son `push`/`pull_request` a `main` y `dev`, más
  `workflow_dispatch`, y que no hay pasos de build ni artefactos.
- `test -z "$(gofmt -l .)"` → sin salida.
- `go vet ./...`.
- `go test -count=1 ./...`.
- `rei check --quiet` (`V1`).
- `go test ./internal/adapter/...` (adaptadores).
- `rei doctor`, `rei init status`, `rei validate 2026-10-26_00-38__ci-tests-build`.
- `git status --porcelain`/`git diff --name-only`: solo `ci.yml`,
  `verification.md`, `README.md`, `AGENTS.md` y los documentos de sesión.
- `git diff --exit-code .github/workflows/release.yml Makefile .goreleaser.yaml
  skeleton/ .rei/config.json` (sin cambios).
- Limitación documentada: el flujo real de Actions solo se comprueba al subir el
  workflow; la evidencia local es el parseo y la validación estructural.

## Pasos

- [x] 1. Crear `.github/workflows/ci.yml` con `name`, `on` (main/dev + PR +
  dispatch), `permissions: contents: read` y `concurrency` cancelable.
- [x] 2. Añadir el job `test` con checkout, setup-go (`go-version-file`), gate de
  `gofmt -l`, `go vet ./...` y `go test -count=1 ./...`; sin build ni artefactos.
- [x] 3. Parsear el YAML y revisar triggers, permisos, ausencia de job `cross` y
  de pasos de compilación.
- [x] 4. Reescribir la estrategia y la lista «Se apoya en» de `verification.md`
  (CI valida; CD publica; sin «no hay workflow de CI»).
- [x] 5. Actualizar las tablas: `V1` local, `V2` CI/manual, `V3`/`V4` manuales o
  release, con la cobertura real de `make cross` (5) y GoReleaser (6).
- [x] 6. Actualizar evidencia y «Qué NO hacer» en `verification.md` (run de CI
  para `V2`, mismo commit; `V3`/`V4` manuales, nunca atribuidos a CI).
- [x] 7. Actualizar la línea de `.github/workflows/` en `README.md`.
- [x] 8. Actualizar el bullet **Despliegue** de `AGENTS.md` (CI valida,
  GoReleaser publica, builds manuales).
- [x] 9. Ejecutar `gofmt` check, `go vet ./...` y `go test -count=1 ./...`.
- [x] 10. Ejecutar `rei check --quiet`, `go test ./internal/adapter/...` y
  `rei doctor`/`rei init status`/`rei validate`.
- [x] 11. Verificar el alcance con `git diff --name-only` y confirmar que
  `release.yml`, `Makefile`, `.goreleaser.yaml`, `skeleton/` y
  `.rei/config.json` no cambian.
- [x] 12. Registrar la evidencia (`V1`, `V2` y verificaciones manuales de
  documentación) en `.rei/progress/work-items/2026-10-26_00-38__ci-tests-build/impl.md`.
