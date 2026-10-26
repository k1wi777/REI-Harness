# Implementación — CI: formato, vet y tests en PRs

## Resumen

Se añade `.github/workflows/ci.yml`: un único job `test` (`ubuntu-latest`) que
valida formato (`gofmt -l .`), `go vet ./...` y `go test -count=1 ./...` en push
a `main`/`dev`, PR contra esas ramas y `workflow_dispatch`. No compila ni publica
binarios. La documentación distingue CI (valida) de CD (`release.yml`, publica) y
deja `V3`/`V4` como verificaciones manuales. No se modifican código, `Makefile`,
`release.yml`, `.goreleaser.yaml`, `skeleton/` ni `.rei/config.json`.

## Archivos

- `.github/workflows/ci.yml` (nuevo) — workflow de validación.
- `.rei/docs/project/verification.md` — estrategia, tablas, evidencia y prohibiciones.
- `README.md` — comentario de `.github/workflows/` en el árbol.
- `AGENTS.md` §3 — bullet **Despliegue**.
- Sesión: `plan.md` (pasos marcados), `current.md`, este `impl.md`.

## Cambios

1. **`ci.yml`**: `name: ci`; `on` con `push`/`pull_request` a `main` y `dev` más
   `workflow_dispatch`; `permissions: contents: read`; `concurrency`
   (`ci-${{ github.workflow }}-${{ github.ref }}`, `cancel-in-progress: true`).
   Job `test` con `timeout-minutes: 15`: checkout@v4, setup-go@v5
   (`go-version-file: go.mod`), gate de formato (`gofmt -l .`; si imprime archivos
   los lista y sale con 1), `go vet ./...` y `go test -count=1 ./...`. Sin
   `make build`/`make cross`/`go build`, sin job `cross`, `matrix`, caché manual
   ni artefactos.
2. **`verification.md`**: nueva estrategia CI/CD/manual (sin «no hay workflow de
   CI»); «Se apoya en» incluye el formato como puerta del job `test`; `V1` sigue
   siendo el único checkpoint de `rei check` (CI ejecuta `go vet` como refuerzo,
   sin tocar `.rei/config.json`); `V2` = job `test` o manual; `V3`/`V4` manuales o
   release, con cobertura real de `make cross` (5 binarios) y GoReleaser (6,
   incluida windows/arm64); evidencia por run de CI (URL, job, commit); tres
   prohibiciones nuevas en «Qué NO hacer».
3. **`README.md`**: línea de `.github/workflows/` = CI (formato, vet, tests) y
   releases con goreleaser.
4. **`AGENTS.md` §3**: CI valida formato/vet/tests; GoReleaser publica en tags
   `v*`; `make build`/`make cross` manuales.

## Verificación

- **YAML**: `npx --yes yaml@2 valid` → exit 0; parseo a JSON y comprobación
  estructural (claves `name`/`on`/`permissions`/`concurrency`/`jobs`, job único
  `test`, triggers `main`/`dev` + PR + dispatch, `contents: read`, cancelable, sin
  `cross`, sin `make build`/`make cross`/`go build`, sin `matrix`/artefactos/caché).
- **Gate de formato**: el script del paso ejecutado con `bash -e` localmente → OK.
- **V1** `./bin/rei check --quiet` → exit 0.
- **V2** `go test -count=1 ./...` → ok (todos los paquetes); `go test
  ./internal/adapter/...` → ok.
- `test -z "$(gofmt -l .)"` → sin salida; `go vet ./...` → OK.
- `./bin/rei init opencode --check` y `./bin/rei init claude --check` → OK;
  `./bin/rei doctor` → «Sin fallos»; `./bin/rei init status` → personalizado;
  `./bin/rei validate 2026-10-26_00-38__ci-tests-build` → OK.
- **Alcance**: `git diff --name-only` = `verification.md`, `current.md`,
  `README.md`, `AGENTS.md`; nuevo `ci.yml`. `git diff --exit-code` sobre
  `release.yml`, `Makefile`, `.goreleaser.yaml`, `skeleton/`, `.rei/config.json`
  → sin cambios.
- **V3/V4**: no ejecutados (cambio solo de CI y documentación); quedan como
  manuales según `verification.md`.

## Observaciones

- GitHub Actions no se puede ejecutar en local: la evidencia es el parseo YAML y
  la validación estructural; el run real se comprobará al subir el workflow.
- El parser usado (`yaml@2`) trabaja en YAML 1.2, por lo que `on` se conserva como
  clave (no como booleano de YAML 1.1).
- Sin commit ni push, según instrucciones de la sesión.
