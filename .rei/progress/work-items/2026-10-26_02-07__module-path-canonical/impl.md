# Implementación — module path canónico `github.com/k1wi777/REI-Harness`

## Resumen

Sustitución mecánica del module path `github.com/k1wi777/my-harness-SDD` por
`github.com/k1wi777/REI-Harness`. Sin cambios de comportamiento, API ni
dependencias. Ejecutados los 12 pasos del `plan.md`.

## Archivos modificados

- **51 archivos previstos** (48 `.go` + `go.mod` + `Makefile` + `.goreleaser.yaml`),
  sin incluir artefactos de sesión.
- `go.mod`: solo la línea `module`; `go 1.27` y `require goldmark v1.8.6`
  intactos. `go.sum` sin cambios.
- `Makefile`: `PKG` al path nuevo (`LDFLAGS` se deriva).
- `.goreleaser.yaml`: solo `ldflags`; `project_name`, `builds`, `archives` y
  `checksum` sin tocar.
- Go: 48 archivos, 115 ocurrencias (incluido el comentario `-ldflags` de
  `internal/cli/cli.go`); alias `reiskel` preservados en
  `internal/initwizard/embedded.go` y `update_test.go`. `embed.go` no contiene el
  path y no se modificó.
- Referencias a `github.com/k1wi777/REI-Harness` ya existentes en
  `internal/update/github.go`/`github_test.go`: sin cambios.

## Cambios

1. `make clean` (retira `bin/` y `dist/` con el path antiguo embebido).
2. `go.mod` → `module github.com/k1wi777/REI-Harness`.
3. Reemplazo del prefijo en los 48 `.go` con `grep -rl | xargs sed` (excluidos
   `.rei`, `bin`, `dist`).
4. `PKG` en `Makefile` y `ldflags` en `.goreleaser.yaml`.
5. `gofmt -w .` y `go mod tidy`.
6. Rebuild y verificación completa.

## Verificación

- `gofmt -l .`: sin salida.
- `go vet ./...`: OK.
- `go test -count=1 ./...`: OK en todos los paquetes.
- `go mod tidy`: `go.mod` solo difiere en `module`; `go.sum` intacto.
- `make build` → `./bin/rei version` = `rei dev` (V3).
- `make build VERSION=v0.0.0-ldflags` → `rei v0.0.0-ldflags`: confirma que el
  package path del `-X` es correcto.
- `CGO_ENABLED=0 make cross`: 5 binarios en `dist/` (V4).
- `rei check --quiet` OK; `rei doctor` OK; `rei init status` OK;
  `rei validate 2026-10-26_02-07__module-path-canonical` OK;
  `rei init opencode --check` y `rei init claude --check` OK.
- Gaps: `grep -rn "my-harness-SDD"` en código/build/docs vivos → sin resultados
  (exit 1); también en `bin`/`dist` regenerados: 0.
- `git status`: 52 archivos modificados = 51 previstos + `.rei/progress/current.md`;
  más el directorio nuevo del Work Item. `.rei/specs/**` y `.rei/progress/**`
  históricos sin cambios.

## Observaciones

- `internal/update/github.go` y `github_test.go` ya usaban el path canónico; no
  forman parte del inventario de 51 y no se editaron.
- No se tocaron `README.md`, `AGENTS.md`, docs, `skeleton/`, `.github/**` ni
  dependencias (no contenían el path antiguo).
- Sin commit, push, PR ni merge.
