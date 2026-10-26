# Review — CI: formato, vet y tests en PRs

## Resultado

**Aprobado.** La implementación cumple el `plan.md` y el alcance del Work Item.

## Verificaciones

- **Alcance**: `rei review-diff --full` (base `31539b6`) = `ci.yml` (nuevo),
  `verification.md`, `README.md`, `AGENTS.md` + documentos de sesión. Nada fuera
  de lo planificado.
- **`ci.yml`** (válido con `npx --yes yaml@2 valid` → 0):
  - un único job `test` (`ubuntu-latest`, `timeout-minutes: 15`);
  - `push` y `pull_request` a `main`/`dev` + `workflow_dispatch`;
  - `permissions: contents: read`; `concurrency` con `cancel-in-progress: true`;
  - pasos: checkout@v4, setup-go@v5 (`go-version-file: go.mod`), gate
    `gofmt -l .` (falla si lista archivos), `go vet ./...`,
    `go test -count=1 ./...`.
  - Sin `make build`, `make cross`, `go build`, job `cross`, `matrix`,
    artefactos, `actions/cache` ni `gofmt -w` (grep vacío). Coherente con
    `release.yml` (v4/v5), que no se toca.
- **`verification.md`**: estrategia CI/CD/manual sin «no hay workflow de CI»;
  `V2` = job `test` de CI o manual; `V3`/`V4` manuales o release (CI no compila);
  evidencia de `V2` por run (URL, job, commit, mismo commit); cobertura real de
  `make cross` = 5 binarios y GoReleaser = 6 (incl. windows/arm64); `V1` intacto
  (fila de la tabla sin cambios, `.rei/config.json` sin tocar); tres
  prohibiciones nuevas en «Qué NO hacer».
- **`README.md` / `AGENTS.md` §3**: CI valida formato/vet/tests; GoReleaser
  publica en tags `v*`; `make build`/`make cross` manuales. Sin cambios colaterales.
- **Rutas protegidas**: `git diff --exit-code .github/workflows/release.yml
  Makefile .goreleaser.yaml skeleton/ .rei/config.json` → 0; `git status` no
  muestra código Go, tests ni esos archivos.
- **Checkpoints rápidos**: `V1` `./bin/rei check --quiet` → 0. `gofmt -l .` sin
  salida; `go vet ./...` → OK.
- **Checkpoints lentos**: `V2` `go test -count=1 ./...` → todos los paquetes OK
  (incl. `go test ./internal/adapter/...`). `V3`/`V4` no aplican: el cambio es de
  CI y documentación, sin código ni proceso de release; `verification.md` los
  deja como manuales y no se les atribuye a CI.
- **Harness**: `rei init opencode --check` y `rei init claude --check` → OK;
  `rei doctor` → sin fallos; `rei init status` → personalizado;
  `rei validate 2026-10-26_00-38__ci-tests-build` → OK (sin WARN).
- **Presupuesto**: `plan.md` 172 líneas (≤150–200), sin `requirements.md` ni
  `design.md`; `validate` no emite WARN.

## Observaciones

- GitHub Actions no se ejecuta en local: la evidencia es el parseo YAML y la
  validación estructural; el run real se comprobará al publicar el workflow.
- Nit menor no bloqueante: en `verification.md` «el job de CI solo ejecuta `V2`»
  se lee en el contexto de checkpoints lentos (CI también corre `gofmt` y
  `go vet` como puerta/refuerzo), coherente con el párrafo de `V1`.

## Acciones requeridas

Ninguna.
