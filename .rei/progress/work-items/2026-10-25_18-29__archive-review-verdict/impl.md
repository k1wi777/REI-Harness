# Implementación — Archive: incluir el veredicto de review en history

## Resumen

`rei session archive` ahora solo archiva Work Items en estado terminal (`done` o
`blocked`) cuyo campo `Estado` de `current.md` coincide con `meta.json`. Si no se
cumple, devuelve `ErrArchiveNotAllowed` (causa + siguiente paso) sin escribir
nada. El contrato del Reviewer se reordena para que el veredicto quede en
`current.md` antes de archivar. Sin `--force`.

## Archivos modificados

- `internal/template/render.go` — guard en `ArchiveSession` + centinela.
- `internal/template/render_test.go` — test existente adaptado y tests del guard.
- `internal/cli/help.go` — detalle del subcomando `session archive`.
- `.rei/agents/reviewer.md` — orden de cierre, reglas duras y racional.
- `.rei/docs/harness/progress.md` — precondición de archivado y orden de cierre.
- `.opencode/agents/reviewer.md`, `.claude/agents/reviewer.md` — regenerados.

## Cambios

1. **Guard en `ArchiveSession`** (`internal/template/render.go`): tras resolver
   el Work Item activo y antes de abrir `history.md`, carga `meta.json` con
   `meta.Load(p.MetaFile(id))` y comprueba:
   - lectura correcta de `meta.json`;
   - `Status` ∈ {`done`, `blocked`};
   - `Estado` de `current.md` idéntico a `Status`.

   Los tres fallos envuelven `ErrArchiveNotAllowed` con `%w`, indican la causa y
   el comando correctivo (`rei status set`, `rei session reset`). No se toca
   `history.md` ni `current.md` al rechazar. Se mantiene la firma y el `("", nil)`
   sin sesión activa.
2. **Tests**: `TestStartAndArchive` ahora cierra en `done` (meta + `current.md`
   coherentes) y comprueba que `history.md` refleja `Estado: done`. Nuevos
   `TestArchiveSessionGuard` (tabla: `done`/`blocked` archivan; `review`,
   `in_progress`, `ready`, `pending` y `changes_requested` rechazan; coherencia
   rota en ambos sentidos; `meta.json` ausente),
   `TestArchiveSessionEstadoIncoherente` (el mensaje muestra ambos valores) y
   `TestArchiveSessionSinSesionActiva` (regresión). Cada rechazo verifica que
   `history.md` y `current.md` quedan intactos.
3. **Reviewer** (`.rei/agents/reviewer.md`): cierre pasa a ser veredicto en
   `current.md` → `rei status set <id> done` → `rei validate` → `rei session
   archive`. Reglas duras: archivado solo tras `status done` y nunca con estado
   no terminal o `current.md` incoherente. Racional de `## Referencia` actualizado.
4. **Documentación/CLI**: `progress.md` documenta la precondición (código 1 si no
   se cumple), que `history.md` refleja el estado final y el orden de cierre en
   la tabla de responsabilidades; `help.go` describe el nuevo comportamiento de
   `archive`.
5. **Agentes nativos**: regenerados con `./bin/rei init opencode` y `./bin/rei
   init claude` (solo cambió `reviewer.md`; idempotente en el resto).

## Verificación

| Comando | Resultado |
|---------|-----------|
| `go test ./internal/template/ -v` | PASS (12 casos) |
| `make test` | `ok` en todos los paquetes |
| `make vet` | sin avisos |
| `make build` | `bin/rei` compilado |
| `gofmt -l .` | sin salida |
| `rei init opencode --check` | `[OK]` todos los archivos |
| `rei init claude --check` | `[OK]` todos los archivos |
| `rei check --quiet` | exit 0 |

## Observaciones

- No se probó `rei session archive` contra la sesión viva del repositorio (según
  el plan); el guard se ejercita solo en directorios temporales de test.
- No hay `--force`; `rei session reset` cubre el descarte y `rei status set
  <id> <estado> --force` las transiciones no estándar.
- No se modificó el formato de `history.md` ni entradas pasadas.
