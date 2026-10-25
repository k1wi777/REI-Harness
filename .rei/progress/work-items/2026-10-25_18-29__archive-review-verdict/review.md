# Review — Archive: incluir el veredicto de review en history

## Resultado

**Aprobado (`done`).** El guard de `ArchiveSession`, los tests, el contrato del
Reviewer, la documentación y los agentes nativos coinciden con `plan.md`. Sin
desviaciones ni cambios fuera de alcance.

## Verificaciones

| Check | Resultado | Evidencia |
|-------|-----------|-----------|
| `go test ./...` | pasa | todos los paquetes `ok` (Go 1.27 en contenedor: el host no tiene `go` en `PATH`) |
| `make vet` / `go vet ./...` | pasa | sin avisos |
| `make build` / `go build ./cmd/rei` | pasa | binario compilado |
| `gofmt -l .` | pasa | sin salida |
| `rei check --quiet` | pasa | exit 0 |
| `rei init opencode --check` | pasa | `[OK]` en todos los archivos |
| `rei init claude --check` | pasa | `[OK]` incl. `CLAUDE.md` |

Nota: `make` falla en el entorno de revisión porque `go` no está en `PATH`; los
tres comandos se ejecutaron con `golang:1.27-alpine` (misma versión que exige
`go.mod`). La evidencia del Implementer es consistente con el resultado.

## Comprobaciones de fondo

- **Guard** (`internal/template/render.go`): carga `meta.json` antes de abrir
  `history.md`; exige `Status ∈ {done, blocked}` y `Estado` de `current.md`
  idéntico; los tres fallos envuelven `ErrArchiveNotAllowed` (`%w`) con causa y
  siguiente paso. No escribe `history.md` ni `current.md` al rechazar. Sin
  `--force` ni escape hatch (`cmdSession` devuelve 1 con `ERROR:`).
- **Tests** (`internal/template/render_test.go`): tabla con `done`/`blocked`
  (archivan), `review`/`in_progress`/`ready`/`pending`/`changes_requested`
  (rechazan), incoherencia en ambos sentidos, `meta.json` ausente; test dedicado
  que verifica que el mensaje muestra ambos valores; regresión sin sesión
  (`("", nil)` y sin crear `history.md`). Cada rechazo comprueba inmutabilidad de
  `history.md` y `current.md`. `TestStartAndArchive` cubre el camino feliz y que
  `history.md` refleja `Estado: done`.
- **Contrato del Reviewer**: `.rei/agents/reviewer.md` ordena veredicto en
  `current.md` → `rei status set <id> done` → `rei validate` → `rei session
  archive`; reglas duras actualizadas (`archive` solo tras `done`, nunca con
  estado no terminal o `current.md` incoherente) y `## Referencia` coherente.
  Flujos de rechazo y bloqueo intactos.
- **Docs/CLI**: `help.go` describe el nuevo comportamiento; `progress.md`
  documenta la precondición, el contenido final de `history.md` y el orden en la
  tabla de responsabilidades.
- **Restricciones**: `go.mod` sin cambios (sin dependencias nuevas), formato de
  `history.md` intacto (solo se añadió el guard previo), `history.md` y entradas
  pasadas no modificadas, `.rei/templates/` sin tocar, agentes nativos
  regenerados y coincidentes.

## Observaciones

- El caso de tabla `meta=review, Estado=done` rechaza por estado no terminal
  antes de evaluar la coherencia, por lo que no compara ambos valores; el caso
  simétrico (`meta=done, Estado=review`) sí está cubierto por
  `TestArchiveSessionEstadoIncoherente`. No afecta al comportamiento exigido.
- No se archivó ni reinició la sesión viva durante las pruebas; el guard se
  ejercita solo en directorios temporales, como pedía el plan.

## Acciones requeridas

Ninguna.
