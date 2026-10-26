# Plan — module path canónico `github.com/k1wi777/REI-Harness`

## Objetivo

Sustituir la identidad mecánica del módulo Go
`github.com/k1wi777/my-harness-SDD` por la canónica
`github.com/k1wi777/REI-Harness`, que coincide con el remoto git
(`git remote -v` → `git@github.com:k1wi777/REI-Harness.git`).

Es un cambio de identidad sin cambio de comportamiento, API ni dependencias:
solo module path, imports, `Makefile`, `.goreleaser.yaml` y cualquier referencia
viva en documentación/build. Los registros históricos de `.rei/specs/**` y
`.rei/progress/**` conservan el path antiguo y **no se tocan**.

## Archivos

Estado inventariado con
`grep -rl "my-harness-SDD" --exclude-dir=.git --exclude-dir=bin --exclude-dir=dist --exclude-dir=.rei .`:
**51 archivos, 118 ocurrencias** (115 en 48 archivos `.go`, más 3 no-Go).

### Go (48 archivos, 115 ocurrencias)

- Raíz del módulo: `embed_test.go` (1). `embed.go` **no** contiene el path
  (solo declara `package rei` y los `embed.FS`); no requiere cambios.
- `cmd/rei/main.go` (1).
- `internal/adapter/`: `adapter_test.go` (1), `claude.go` (1),
  `claude_test.go` (2), `generate.go` (3), `generate_test.go` (1),
  `opencode.go` (1).
- `internal/check/`: `check.go` (5), `check_test.go` (2).
- `internal/cli/`: `cli.go` (17), `cli_test.go` (3).
- `internal/doctor/`: `doctor.go` (7), `doctor_test.go` (3).
- `internal/github/`: `github.go` (2), `github_test.go` (2).
- `internal/gitx/`: `gitx.go` (2), `gitx_test.go` (2).
- `internal/initwizard/`: `embedded.go` (1), `initwizard.go` (2),
  `initwizard_test.go` (1), `installer.go` (1), `installer_test.go` (3),
  `manifest.go` (1), `manifest_test.go` (1), `update.go` (1), `update_test.go` (2).
- `internal/show/`: `show.go` (5), `show_test.go` (2).
- `internal/state/`: `state.go` (2), `state_test.go` (2).
- `internal/template/`: `render.go` (2), `render_test.go` (2).
- `internal/update/update.go` (1).
- `internal/validate/`: `validate.go` (4), `validate_test.go` (2).
- `internal/view/`: `changes.go` (2), `changes_test.go` (2), `lifecycle.go` (2),
  `lifecycle_test.go` (2), `render.go` (4), `render_test.go` (2), `server.go` (2),
  `server_test.go` (1), `toc_test.go` (1), `view.go` (2).
- `internal/workitem/`: `workitem.go` (2), `workitem_test.go` (2).

### No-Go (3 archivos)

- `go.mod` línea 1: `module github.com/k1wi777/my-harness-SDD`.
- `Makefile`: `PKG := github.com/k1wi777/my-harness-SDD/internal/cli`.
- `.goreleaser.yaml`: `ldflags` con `-X .../internal/cli.version`.

### Sin referencias (verificado)

`README.md`, `AGENTS.md`, `CLAUDE.md`, `DESIGN.md`, `PRODUCT.md`,
`.rei/docs/**` (incluido `usage.md`), `.github/workflows/**`, `skeleton/**`,
`.rei/agents/**`, `.rei/adapters/**`, `.rei/templates/**` y `.rei/config.json`.
`README.md` ya enlaza a `https://github.com/k1wi777/REI-Harness`.

## Cambios

1. **`go.mod`**: `module github.com/k1wi777/REI-Harness`. No cambiar la línea
   `go 1.27` ni el `require` de `github.com/yuin/goldmark v1.8.6`.
2. **Imports Go**: reemplazo mecánico del prefijo en los 48 archivos listados.
   Mantener alias existentes (`reiskel "github.com/k1wi777/REI-Harness"` en
   `internal/initwizard/embedded.go` y `internal/initwizard/update_test.go`).
3. **Comentario de `internal/cli/cli.go`**: la línea de ejemplo de `-ldflags`
   (`-X .../internal/cli.version=vX.Y.Z`) forma parte de las 17 ocurrencias y
   debe quedar con el path nuevo; es documentación del mecanismo de versión.
4. **`Makefile`**: actualizar `PKG` a
   `github.com/k1wi777/REI-Harness/internal/cli`. `LDFLAGS` se deriva de `PKG`,
   así que no se edita aparte.
5. **`.goreleaser.yaml`**: actualizar el `ldflags` a
   `-s -w -X github.com/k1wi777/REI-Harness/internal/cli.version={{ .Version }}`.
   No tocar `project_name`, `builds`, `archives` ni `checksum`.
6. **`go.sum`**: no debe cambiar. `go mod tidy` solo puede reescribir `go.mod`
   por el nuevo module path; no debe añadir ni quitar `require`.
7. **Artefactos binarios** `bin/rei` y `dist/**` contienen el path antiguo
   embebido, pero están en `.gitignore` y se regeneran con `make build`/`make cross`;
   no se editan. `make clean` antes de las comprobaciones evita falsos positivos.
8. **Docs/build vivos**: no hay ocurrencias que corregir; el paso es de
   verificación, no de edición. `skeleton/**` está limpio y no se modifica.

## Notas de ejecución

- El cambio es de identidad de compilación: una sustitución parcial falla de
  inmediato en `go build`/`go test`, así que el compilador actúa como red de
  seguridad de los imports, pero no del comentario de `cli.go` ni de los
  `ldflags`; esos se revisan a mano.
- `go list ./...` antes y después debe devolver el mismo conjunto de paquetes,
  solo con el prefijo nuevo.
- `embed.go` no se modifica; si aparece en el reemplazo masivo, es un error.
- `internal/initwizard/embedded.go` y `internal/initwizard/update_test.go`
  importan el paquete raíz con alias `reiskel`; conservar el alias y cambiar
  solo la ruta.
- No usar `sed` sobre `.rei/specs/**` ni `.rei/progress/**`: excluirlos
  explícitamente del reemplazo (el inventario ya los separa).

## Restricciones

- No reescribir `.rei/specs/**` ni `.rei/progress/**` (registros históricos).
- No cambiar comportamiento, API pública, mensajes, flags ni `go 1.27`.
- No añadir/quitar dependencias; `go.mod`/`go.sum` solo por el module path.
- No renombrar el repositorio, el binario (`rei`) ni `project_name`.
- No introducir `replace`, `go.work` ni cambios de versión de Go.
- Un `grep` sin resultados devuelve exit code 1: eso es el resultado esperado.
- `plan.md` ≤ 200 líneas (presupuesto de Task).

## Verificación

Evidencia a registrar en `.rei/progress/work-items/<id>/impl.md` por checkpoint.

- **V1** — `go vet ./...` (y `rei check --quiet`).
- **V2** — `go test ./...` (sin caché: `go test -count=1 ./...`).
- **V3** — `make build`; `./bin/rei version` debe imprimir `dev`.
  Además `make build VERSION=v0.0.0-ldflags` + `./bin/rei version` debe imprimir
  `v0.0.0-ldflags`: prueba que `-X` apunta al package path correcto.
- **V4** — `CGO_ENABLED=0 make cross` (cinco binarios en `dist/`).
- Formato — `gofmt -l .` sin salida.
- CLI — `rei doctor`, `rei init status`, `rei validate`,
  `rei init opencode --check`, `rei init claude --check`.
- Tidy — `go mod tidy` y `git diff --stat go.mod go.sum`: `go.sum` intacto y
  `go.mod` solo con la línea `module` cambiada.
- Path antiguo — sin resultados en código/build/docs vivos:

  ```sh
  grep -rn "my-harness-SDD" \
    --exclude-dir=.git --exclude-dir=bin --exclude-dir=dist \
    cmd internal embed.go embed_test.go go.mod Makefile .goreleaser.yaml \
    README.md AGENTS.md CLAUDE.md DESIGN.md PRODUCT.md \
    .rei/docs .rei/agents .rei/adapters .rei/templates skeleton .github
  ```

  y `grep -rl "my-harness-SDD" --include='*.go' cmd internal embed_test.go`
  sin resultados.
- Diff acotado — `git diff --stat` debe listar solo los 51 archivos previstos
  (más los artefactos de la sesión en `.rei/`).

## Pasos

- [x] 1. `make clean` para retirar binarios con el path antiguo.
- [x] 2. Editar `go.mod` (línea `module`) al path canónico.
- [x] 3. Reemplazar el prefijo en los 48 archivos `.go` (115 ocurrencias),
      incluido el comentario `-ldflags` de `internal/cli/cli.go`, preservando
      alias como `reiskel`.
- [x] 4. Actualizar `PKG` en `Makefile`.
- [x] 5. Actualizar `ldflags` en `.goreleaser.yaml`.
- [x] 6. `gofmt -w .` y comprobar que `gofmt -l .` queda vacío.
- [x] 7. `go mod tidy`; verificar que `go.sum` no cambia y que `go.mod` solo
      difiere en la línea `module`.
- [x] 8. Ejecutar `go vet ./...` (V1) y `go test ./...` (V2).
- [x] 9. Ejecutar `make build` y `./bin/rei version` → `dev` (V3); luego
      `make build VERSION=v0.0.0-ldflags` y confirmar la versión sellada.
- [x] 10. Ejecutar `CGO_ENABLED=0 make cross` (V4).
- [x] 11. Ejecutar `rei check --quiet`, `rei doctor`, `rei init status`,
      `rei validate`, `rei init opencode --check` y `rei init claude --check`.
- [x] 12. Ejecutar los `grep` de verificación del path antiguo y revisar
      `git diff --stat`; confirmar que `.rei/specs/**` y `.rei/progress/**`
      históricos no aparecen en el diff.
